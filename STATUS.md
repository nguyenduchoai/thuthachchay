# Bước Vàng — Trạng thái MVP hiện tại

> Cập nhật: 2026-05-23 · lượt mở rộng Admin + migration templ + Tailwind (Claude Cowork)

## Kết quả số liệu

| Hạng mục | Số file | LOC |
|---|---|---|
| BE Go (services/api) | 48 | ~3,800 |
| Admin Go (apps/admin, viết tay) | 15 | ~1,950 |
| Admin Go (apps/admin, `*_templ.go` sinh) | 7 | ~4,790 |
| Mini App TS (apps/miniapp) | 34 | ~1,330 |
| Migrations SQL (goose) | 2 | — |
| Tests (Go + TS) | 5 | — |
| **Tổng (viết tay)** | **104** | **~7,080** |

## Đã làm xong

### Backend `services/api`
- `internal/db` — pgxpool + `InTx` helper transaction
- `internal/store` — repo cho `users, challenges, steps, wallet, vouchers, referrals, sessions, strava_tokens` (truy vấn SQL viết tay theo phong cách sqlc, không cần generator).
- `internal/auth` — JWT RS256 (issue + parse, ephemeral key cho dev), ZaloClient (verify access_token, dev mode chấp nhận `dev:<id>`), service Login/Refresh/SignOut + handler.
- `internal/middleware/jwt.go` — Fiber middleware `RequireJWT`.
- `internal/users` — handler `/me`, `PATCH /me`, `/me/attribution`, `/username/check` với regex validate handle.
- `internal/challenges` — service + handler list/get/create/join/leaderboard, settle logic chia pool ÷ winners (host nhận 10% nếu public).
- `internal/steps` — service ingest với anti-fraud cadence + replay-nonce, merge logic + history.
- `internal/wallet` — service Balance/Credit/Debit qua ledger idempotent + handler `/wallet`, `/wallet/ledger`.
- `internal/vouchers` — redeem atomic (trừ điểm + allocate code + insert redemption trong 1 tx).
- `internal/referrals` — track + stats + `PayBonusOnFirstChallengeComplete` cho worker.
- `internal/leaderboard` — Redis ZSET wrapper (Top, Rank, AddSteps).
- `internal/notifications` — Zalo OA send template (no-op trong dev mode).
- `internal/strava` — OAuth URL + ExchangeCode + Refresh + GetActivity + EstimateSteps + webhook verify.
- `internal/server/server.go` — dependency wiring + route registration (31 routes thật + healthz/readyz/version).
- `cmd/api/main.go` — wire qua `server.Build`, graceful shutdown.
- `cmd/worker/main.go` — cron 5'/1h/10' cho settle + rebuild leaderboard + pay referral bonus.
- `migrations/20260522120000_init_schema.sql` — 13 bảng + indexes + view `user_balances`.
- `migrations/20260522120100_seed_dev.sql` — 5 user, 3 challenge, 4 voucher (300 codes), 5,000đ welcome.

### Admin `apps/admin`

**Stack đúng spec §6**: `templ` + HTMX + **Tailwind CSS** (qua CDN, custom theme color `brand: #ff9500`).

**Tách module theo domain** — mỗi cặp `xxx.go` (handler slim: fetch data + render) + `xxx.templ` (UI components, được templ generate thành `xxx_templ.go`):

- `routes.go` (133 LOC) — wiring 27 endpoint + Basic Auth + mountStub.
- `layout.go` + `layout.templ` — `@page(c, title)` wrapper với HTMX + Tailwind CDN + Tailwind theme config + `data-confirm` JS listener global; `@nav(c)` highlight active section + lang toggle vi/en; `@pager`, `@kpiTile`, `@kpiTileSub`, `@pillC`, `@statusPillC`; helpers `langOf/t/setLang/paginate/statusPillClass/stockPillClass/formatInt/decodeB64/isUUID/prettyJSON/shortID`.
- `users.go` + `users.templ` — `@usersListPage` (filter q+status+pagination) · `@userDetailPage` (info + form update + ledger 30 mới nhất) · `@userRowTR` · handlers cho list/detail/update/suspend/activate/adjust. **Endpoint mới spec §6**: `POST /users/:id/update` gộp status+fraud_score+note vào 1 form.
- `challenges.go` + `challenges.templ` — `@challengesListPage` (filter status+pagination) · **`@challengeDetailPage` mới** với **settle preview** (winners = đủ daily_target mọi ngày · host 10% nếu public · per-winner) · `@participantTR` · handlers list/detail/cancel/trigger.
- `vouchers.go` + `vouchers.templ` — `@vouchersListPage`+pagination · **`@voucherDetailPage` mới** (KPI grid + edit form + add-codes form + codes inventory 100 gần nhất + redemptions 100 gần nhất) · `@voucherUploadPage` · `@voucherEditForm` · `@voucherAddCodesForm` · handlers list/detail/upload/update/codes/disable. **Endpoints mới spec §6**: `POST /vouchers/:id/update`, `POST /vouchers/:id/codes`, `POST /vouchers/:id/disable`.
- `fraud.go` + `fraud.templ` — `@fraudListPage` + `@fraudRowTR` + decide approve/reject + pagination.
- `audit.go` + `audit.templ` — **mới**: `@auditListPage` với filter admin/action/target/from/to + pagination + `@auditRowTR` auto-link target UUID sang user/challenge/voucher.
- `reports.go` + `reports.templ` — `@dashboardPage` (thêm KPI fraud queue + 5 nav button) · **`@dsoPage` mới** (KPI: issued/burned/net/redemptions/active/new/ARPU + top brands by burn + daily breakdown 60 ngày) · CSV exports users/ledger/challenges (CSV không qua templ, viết trực tiếp).

**Hạ tầng templ + tooling**:
- `//go:generate go run github.com/a-h/templ/cmd/templ@latest generate` trong `layout.go`.
- `make templ` target dùng `go run github.com/a-h/templ/cmd/templ@latest` (không yêu cầu cài CLI riêng).
- `make build-go` và `make admin` đều depend `templ` → tự động regen trước khi build/run.
- File `*_templ.go` được check-in (per templ convention) để CI build mà không cần chạy generate.

- Direct pgx (không phụ thuộc package `services/api`).
- i18n vi/en qua cookie `lang` (rule CLAUDE.md #10), toggle ở góc phải header — mọi text user-facing đều dùng `t(c, vi, en)`.
- Mọi mutation ghi `audit_log` với `diff` jsonb (đã đủ 10 action: user.suspend/activate/update, points.adjust, challenge.cancel/settle_trigger, fraud.decide, voucher.upload/update/codes_add/disable).

### Mini App `apps/miniapp`
- `services/endpoints.ts` — dùng `@buocvang/api-client` fetch SDK thật cho auth, user, steps, challenges, wallet, voucher, referral, strava.
- `state/user.ts` — Zustand store user info, `refresh()` gọi `/me`.
- `pages/auth/SignIn.tsx` — gọi `zmp-sdk/apis getAccessToken`, fallback prompt dev mode.
- `pages/Home.tsx` — fetch song song today total + open challenges + leaderboard, render progress bar + cards.
- `pages/Wallet.tsx` — balance hero + voucher list + redeem flow.
- `pages/ChallengeDetail.tsx` — info + stats grid + sticky CTA Join.
- `pages/Invite.tsx` — referral code + share qua `zmp-sdk openShareSheet` hoặc `navigator.share`.
- 6 pages khác (Welcome, Splash, onboarding How/Source/Goal/Username/Strava/Notify) đã có UI thật từ scaffold trước.

### Tests
- `services/api/internal/httpx/middleware_test.go` — CORS allowlist middleware
- `services/api/internal/steps/counter_test.go` — Merge function
- `services/api/internal/antifraud/antifraud_test.go` — CadenceFlag
- `apps/admin/internal/web/routes_test.go` — index render (verify Tailwind CDN + brand + htmx loaded) + healthz + /lang redirect+cookie + /admin/* no-database fallback + isUUID + prettyJSON + formatInt + stockPillClass + shortID (9 tests)
- `apps/miniapp/src/sensors/stepCounter.test.ts` — peak detection

## Còn lại trước VPS

Không còn màn runtime rỗng trong Mini App. Các màn player chính đã đi qua API thật bằng `@buocvang/api-client`.

| Backend còn thiếu | Ghi chú |
|---|---|
| Upload avatar/cover S3/MinIO | `POST /v1/upload` chưa wire — cần SDK aws-sdk-go-v2 |
| Queue bền vững cho Strava webhook | Hiện webhook đã sync activity → daily_steps bằng goroutine trong process; nếu chạy nhiều replica hoặc traffic cao nên chuyển Redis/asynq |
| Notification batch (`requestSubscribeMessage`) | Chỉ có wrapper send; cần job nightly nhắc streak |
| Admin OIDC (Google Workspace) + IP allowlist | Hiện Basic Auth env — production cần thay (spec §6). Cần Google Workspace client_id/secret + middleware IP whitelist. |
| Tailwind production build | Hiện qua **CDN** (script `cdn.tailwindcss.com` + custom theme inline). Hoạt động tốt nội bộ. Production khuyến nghị chuyển standalone CLI (binary download + compile sang static CSS) để tránh JS runtime cost. |
| KYC review UI | Cột `kyc_status` chưa có trong schema — cần migration trước khi build UI. |
| `go.work` workspace | Admin & API là 2 module riêng; nếu muốn share types có thể thêm |

## Chạy local

```bash
cd /Volumes/NEOM/ThuThachChay/buoc-vang

# 1) Lên Postgres + Redis + MinIO
cd deploy && docker compose up -d && cd ..

# 2) Migrations
export DATABASE_URL=postgres://buocvang:buocvang@localhost:5432/buocvang?sslmode=disable
goose -dir services/api/migrations postgres "$DATABASE_URL" up

# 3) API + worker + admin
make api      # http://localhost:8080
make worker   # background
make admin    # http://localhost:8081 (basic auth: admin / buocvang-dev)

# 4) Mini App
cd apps/miniapp && npm i && npm run dev
# → http://localhost:5173 (cần VITE_API_BASE_URL=http://localhost:8080)
```

## Smoke test E2E (dev mode)

```bash
# 1. Login dev — dùng zalo_id giả "u1"
curl -X POST http://localhost:8080/v1/auth/zalo \
  -H 'Content-Type: application/json' \
  -d '{"zalo_access_token":"dev:u1"}'
# → { access_token, refresh_token, expires_in }

# 2. Lấy profile
curl http://localhost:8080/v1/me -H "Authorization: Bearer $TOKEN"

# 3. List challenges
curl http://localhost:8080/v1/challenges -H "Authorization: Bearer $TOKEN"

# 4. Wallet
curl http://localhost:8080/v1/wallet -H "Authorization: Bearer $TOKEN"
# → { "balance": 5000, "currency": "POINT" } (sau seed)
```
