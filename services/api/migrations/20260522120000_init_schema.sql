-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  zalo_id         TEXT NOT NULL UNIQUE,
  handle          CITEXT UNIQUE,
  email           CITEXT UNIQUE,
  display_name    TEXT,
  avatar_url      TEXT,
  daily_goal      INT  NOT NULL DEFAULT 10000,
  locale          TEXT NOT NULL DEFAULT 'vi-VN',
  acquisition     TEXT,
  strava_user_id  TEXT UNIQUE,
  status          TEXT NOT NULL DEFAULT 'active',
  fraud_score     INT  NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
  id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token_hash  TEXT NOT NULL UNIQUE,
  user_agent          TEXT,
  ip                  INET,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at          TIMESTAMPTZ NOT NULL,
  revoked_at          TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS sessions_user_idx ON sessions(user_id);

CREATE TABLE IF NOT EXISTS strava_tokens (
  user_id        UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  access_token   TEXT NOT NULL,
  refresh_token  TEXT NOT NULL,
  expires_at     TIMESTAMPTZ NOT NULL,
  scope          TEXT NOT NULL DEFAULT '',
  athlete_id     TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS strava_tokens_athlete_idx ON strava_tokens(athlete_id) WHERE athlete_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS challenges (
  id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  host_id              UUID REFERENCES users(id) ON DELETE SET NULL,
  visibility           TEXT NOT NULL,
  name                 TEXT NOT NULL,
  description          TEXT,
  cover_url            TEXT,
  daily_steps_target   INT  NOT NULL,
  duration_days        INT  NOT NULL,
  start_date           DATE NOT NULL,
  end_date             DATE NOT NULL,
  entry_points         INT  NOT NULL DEFAULT 0,
  prize_pool           INT  NOT NULL DEFAULT 0,
  sponsor_id           UUID,
  max_participants     INT,
  status               TEXT NOT NULL DEFAULT 'open',
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS challenges_status_idx ON challenges(status, start_date);

CREATE TABLE IF NOT EXISTS challenge_participants (
  challenge_id  UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
  user_id       UUID NOT NULL REFERENCES users(id)      ON DELETE CASCADE,
  joined_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  entry_paid    INT NOT NULL DEFAULT 0,
  state         TEXT NOT NULL DEFAULT 'in',
  total_steps   BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (challenge_id, user_id)
);
CREATE INDEX IF NOT EXISTS cp_user_idx ON challenge_participants(user_id);
CREATE INDEX IF NOT EXISTS cp_lb_idx   ON challenge_participants(challenge_id, total_steps DESC);

CREATE TABLE IF NOT EXISTS daily_steps (
  user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  day      DATE NOT NULL,
  steps    INT  NOT NULL DEFAULT 0,
  source   TEXT NOT NULL DEFAULT 'zmp',
  flagged  BOOLEAN NOT NULL DEFAULT false,
  PRIMARY KEY (user_id, day)
);

CREATE TABLE IF NOT EXISTS step_events (
  id              BIGSERIAL PRIMARY KEY,
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  source          TEXT NOT NULL,
  steps           INT  NOT NULL,
  started_at      TIMESTAMPTZ NOT NULL,
  ended_at        TIMESTAMPTZ NOT NULL,
  client_nonce    TEXT NOT NULL,
  cadence_avg_ms  INT  NOT NULL DEFAULT 0,
  flagged         BOOLEAN NOT NULL DEFAULT false,
  flag_reason     TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, client_nonce)
);
CREATE INDEX IF NOT EXISTS step_events_user_time_idx ON step_events(user_id, started_at DESC);

CREATE TABLE IF NOT EXISTS ledger_entries (
  id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  delta_points     INT  NOT NULL,
  reason           TEXT NOT NULL,
  reference_type   TEXT,
  reference_id     TEXT,
  idempotency_key  TEXT NOT NULL,
  note             TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS ledger_user_time_idx ON ledger_entries(user_id, created_at DESC);

CREATE OR REPLACE VIEW user_balances AS
  SELECT user_id, COALESCE(SUM(delta_points), 0)::int AS balance
  FROM ledger_entries
  GROUP BY user_id;

CREATE TABLE IF NOT EXISTS vouchers (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  brand         TEXT NOT NULL,
  title         TEXT NOT NULL,
  cost_points   INT  NOT NULL,
  stock         INT  NOT NULL DEFAULT 0,
  cover_url     TEXT,
  expires_at    DATE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS voucher_codes (
  id              BIGSERIAL PRIMARY KEY,
  voucher_id      UUID NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
  code            TEXT NOT NULL,
  used_by_user_id UUID REFERENCES users(id),
  used_at         TIMESTAMPTZ,
  UNIQUE (voucher_id, code)
);

CREATE TABLE IF NOT EXISTS voucher_redemptions (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  voucher_id   UUID NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
  code         TEXT NOT NULL,
  redeemed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS voucher_redem_user_idx ON voucher_redemptions(user_id);

CREATE TABLE IF NOT EXISTS referrals (
  inviter_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  invitee_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  code        TEXT NOT NULL,
  bonus_paid  BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (inviter_id, invitee_id)
);

CREATE TABLE IF NOT EXISTS notif_subscriptions (
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  template   TEXT NOT NULL,
  granted    BOOLEAN NOT NULL DEFAULT true,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, template)
);

CREATE TABLE IF NOT EXISTS audit_log (
  id          BIGSERIAL PRIMARY KEY,
  admin_id    TEXT,
  action      TEXT NOT NULL,
  target      TEXT,
  diff        JSONB,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notif_subscriptions;
DROP TABLE IF EXISTS referrals;
DROP TABLE IF EXISTS voucher_redemptions;
DROP TABLE IF EXISTS voucher_codes;
DROP TABLE IF EXISTS vouchers;
DROP VIEW  IF EXISTS user_balances;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS step_events;
DROP TABLE IF EXISTS daily_steps;
DROP TABLE IF EXISTS challenge_participants;
DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS strava_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
