-- +goose Up
-- +goose StatementBegin

-- ─── Extensions ──────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "citext";    -- case-insensitive text cho handle

-- ─── Users ───────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zalo_id         TEXT NOT NULL UNIQUE,
    handle          CITEXT UNIQUE,
    display_name    TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT,
    email           CITEXT,
    phone           TEXT,
    daily_goal      INT NOT NULL DEFAULT 10000 CHECK (daily_goal BETWEEN 1000 AND 50000),
    acquisition     TEXT,
    locale          TEXT NOT NULL DEFAULT 'vi',
    flags           JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','deleted')),
    fraud_score     INT NOT NULL DEFAULT 0 CHECK (fraud_score BETWEEN 0 AND 100),
    referral_code   CITEXT UNIQUE,
    referred_by     UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_status         ON users(status);
CREATE INDEX idx_users_referred_by    ON users(referred_by) WHERE referred_by IS NOT NULL;
CREATE INDEX idx_users_fraud_score    ON users(fraud_score DESC) WHERE fraud_score >= 50;

-- ─── Sessions (refresh token) ────────────────────────────────────────────────
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_hash    TEXT NOT NULL,                   -- sha256 hash của refresh token
    device_info     JSONB NOT NULL DEFAULT '{}',
    ip              INET,
    user_agent      TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user        ON sessions(user_id);
CREATE INDEX idx_sessions_active      ON sessions(user_id, expires_at) WHERE revoked_at IS NULL;

-- ─── Strava tokens ───────────────────────────────────────────────────────────
CREATE TABLE strava_tokens (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    strava_id       TEXT NOT NULL UNIQUE,
    access_token    TEXT NOT NULL,
    refresh_token   TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    scope           TEXT NOT NULL DEFAULT 'read,activity:read',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─── Daily steps (aggregate hàng ngày, sources: zmp / strava) ────────────────
CREATE TABLE daily_steps (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day             DATE NOT NULL,
    zmp_steps       INT NOT NULL DEFAULT 0,
    strava_steps    INT NOT NULL DEFAULT 0,
    merged_steps    INT NOT NULL DEFAULT 0,
    cadence_variance NUMERIC(6,3),
    flagged         BOOLEAN NOT NULL DEFAULT false,
    flag_reasons    TEXT[],
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, day)
);
CREATE INDEX idx_daily_steps_user_day ON daily_steps(user_id, day DESC);
CREATE INDEX idx_daily_steps_flagged  ON daily_steps(day, flagged) WHERE flagged = true;

-- ─── Step ingest events (raw, anti-replay) ───────────────────────────────────
CREATE TABLE step_ingest_events (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day             DATE NOT NULL,
    source          TEXT NOT NULL CHECK (source IN ('zmp','strava','manual_admin')),
    steps           INT NOT NULL CHECK (steps >= 0),
    client_nonce    TEXT NOT NULL,
    sensor_hash     TEXT,
    raw             JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, client_nonce)
);
CREATE INDEX idx_step_events_user_day ON step_ingest_events(user_id, day);

-- ─── Challenges ──────────────────────────────────────────────────────────────
CREATE TABLE challenges (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id             UUID REFERENCES users(id),
    visibility          TEXT NOT NULL CHECK (visibility IN ('public','private')),
    name                TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    cover_url           TEXT,
    daily_steps_target  INT NOT NULL CHECK (daily_steps_target BETWEEN 1000 AND 50000),
    duration_days       INT NOT NULL CHECK (duration_days BETWEEN 1 AND 90),
    entry_points        INT NOT NULL DEFAULT 0 CHECK (entry_points >= 0),
    prize_pool          INT NOT NULL DEFAULT 0 CHECK (prize_pool >= 0),
    max_participants    INT,
    start_date          DATE NOT NULL,
    end_date            DATE NOT NULL,
    status              TEXT NOT NULL DEFAULT 'open'
                        CHECK (status IN ('draft','open','live','settled','cancelled')),
    settled_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date >= start_date)
);
CREATE INDEX idx_challenges_status   ON challenges(status, start_date);
CREATE INDEX idx_challenges_host     ON challenges(host_id) WHERE host_id IS NOT NULL;

CREATE TABLE challenge_participants (
    challenge_id    UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    total_steps     INT NOT NULL DEFAULT 0,
    days_completed  INT NOT NULL DEFAULT 0,
    won             BOOLEAN NOT NULL DEFAULT false,
    payout_points   INT NOT NULL DEFAULT 0,
    PRIMARY KEY (challenge_id, user_id)
);
CREATE INDEX idx_cp_user ON challenge_participants(user_id);

-- ─── Wallet ledger (double-entry an toàn cho điểm) ───────────────────────────
CREATE TABLE ledger_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    delta_points    INT NOT NULL,                       -- âm = trừ, dương = cộng
    reason          TEXT NOT NULL,                      -- 'challenge_join','challenge_payout','voucher_redeem','referral','admin_adjust'
    reference_type  TEXT,
    reference_id    TEXT,
    idempotency_key TEXT,
    note            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);
CREATE INDEX idx_ledger_user_time ON ledger_entries(user_id, created_at DESC);

-- View: số dư hiện tại
CREATE VIEW user_balances AS
SELECT u.id AS user_id, COALESCE(SUM(l.delta_points), 0) AS balance_points
FROM users u
LEFT JOIN ledger_entries l ON l.user_id = u.id
GROUP BY u.id;

-- ─── Vouchers ────────────────────────────────────────────────────────────────
CREATE TABLE voucher_products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner         TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    cost_points     INT NOT NULL CHECK (cost_points > 0),
    cover_url       TEXT,
    terms_url       TEXT,
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE voucher_codes (
    id              BIGSERIAL PRIMARY KEY,
    product_id      UUID NOT NULL REFERENCES voucher_products(id) ON DELETE CASCADE,
    code            TEXT NOT NULL,
    redeemed_by     UUID REFERENCES users(id),
    redeemed_at     TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, code)
);
CREATE INDEX idx_voucher_codes_unused
    ON voucher_codes(product_id) WHERE redeemed_by IS NULL;

CREATE TABLE redemptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    product_id      UUID NOT NULL REFERENCES voucher_products(id),
    code_id         BIGINT REFERENCES voucher_codes(id),
    cost_points     INT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','complete','failed','refunded')),
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);

-- ─── Notifications subscription (Zalo OA) ────────────────────────────────────
CREATE TABLE notification_subs (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    challenges      BOOLEAN NOT NULL DEFAULT true,
    rewards         BOOLEAN NOT NULL DEFAULT true,
    streaks         BOOLEAN NOT NULL DEFAULT true,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─── Audit log ───────────────────────────────────────────────────────────────
CREATE TABLE audit_log (
    id              BIGSERIAL PRIMARY KEY,
    actor_id        UUID,                               -- admin user UUID hoặc null nếu system
    actor_type      TEXT NOT NULL CHECK (actor_type IN ('admin','system','user')),
    action          TEXT NOT NULL,
    target_type     TEXT,
    target_id       TEXT,
    payload         JSONB NOT NULL DEFAULT '{}',
    ip              INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_target ON audit_log(target_type, target_id);
CREATE INDEX idx_audit_actor  ON audit_log(actor_id, created_at DESC);

-- ─── Updated_at trigger ──────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION trg_set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at         BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
CREATE TRIGGER daily_steps_updated_at   BEFORE UPDATE ON daily_steps
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
CREATE TRIGGER challenges_updated_at    BEFORE UPDATE ON challenges
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
CREATE TRIGGER strava_tokens_updated_at BEFORE UPDATE ON strava_tokens
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS strava_tokens_updated_at ON strava_tokens;
DROP TRIGGER IF EXISTS challenges_updated_at    ON challenges;
DROP TRIGGER IF EXISTS daily_steps_updated_at   ON daily_steps;
DROP TRIGGER IF EXISTS users_updated_at         ON users;
DROP FUNCTION IF EXISTS trg_set_updated_at();

DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notification_subs;
DROP TABLE IF EXISTS redemptions;
DROP TABLE IF EXISTS voucher_codes;
DROP TABLE IF EXISTS voucher_products;
DROP VIEW  IF EXISTS user_balances;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS challenge_participants;
DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS step_ingest_events;
DROP TABLE IF EXISTS daily_steps;
DROP TABLE IF EXISTS strava_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
