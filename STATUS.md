# Bước Vàng — Trạng thái scaffold hiện tại

> Cập nhật: 2026-05-22 · sau lượt scaffold mở rộng của Claude (Cowork)

## Kết quả số liệu

| Hạng mục | Số file | LOC |
|---|---|---|
| BE Go (services/api) | 48 | ~3,800 |
| Admin Go (apps/admin) | 3 | ~600 |
| Mini App TS (apps/miniapp) | 34 | ~1,330 |
| Migrations SQL (goose) | 2 | — |
| Tests (Go + TS) | 5 | — |
| **Tổng** | **92** | **~5,700** |

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
- `internal/web/routes.go` — auth basic, list users/challenges/fraud-queue/vouchers, cancel challenge, adjust points, CSV export users + ledger. HTML render server-side với HTMX-ready markup.
- Direct pgx (không phụ thuộc package `services/api`).

### Mini App `apps/miniapp`
- `services/endpoints.ts` — **24 endpoint** typed wrappers (auth, user, steps, challenges, wallet, voucher, referral, strava).
- `state/user.ts` — Zustand store user info, `refresh()` gọi `/me`.
- `pages/auth/SignIn.tsx` — gọi `zmp-sdk/apis getAccessToken`, fallback prompt dev mode.
- `pages/Home.tsx` — fetch song song today total + open challenges + leaderboard, render progress bar + cards.
- `pages/Wallet.tsx` — balance hero + voucher list + redeem flow.
- `pages/ChallengeDetail.tsx` — info + stats grid + sticky CTA Join.
- `pages/Invite.tsx` — referral code + share qua `zmp-sdk openShareSheet` hoặc `navigator.share`.
- 6 pages khác (Welcome, Splash, onboarding How/Source/Goal/Username/Strava/Notify) đã có UI thật từ scaffold trước.

### Tests
- `services/api/internal/httpx/routes_test.go` — healthz/version/501-not-implemented
- `services/api/internal/steps/counter_test.go` — Merge function
- `services/api/internal/antifraud/antifraud_test.go` — CadenceFlag
- `apps/admin/internal/web/routes_test.go` — index + healthz
- `apps/miniapp/src/sensors/stepCounter.test.ts` — peak detection

## Còn lại (chủ động không làm trong lượt này)

| Pages chưa có UI thật (vẫn `PagePlaceholder`) | Lý do |
|---|---|
| `Checkout.tsx` | Polling `/transactions/:id` chưa wire — endpoint backend chưa code |
| `Create.tsx`, `CreateNew.tsx` | Form tạo challenge — ưu tiên thấp cho MVP người chơi |
| `Discover.tsx` | Tương tự Home, có thể tái dùng list logic |
| `Profile.tsx`, `ProfileSettings.tsx`, `ProfileEdit.tsx` | CRUD đơn giản, chỉ cần wire `PATCH /me` |
| `onboarding/LeaderboardPreview.tsx` | Dùng `GET /v1/leaderboards/global` tương tự Home |

| Backend còn thiếu | Ghi chú |
|---|---|
| Upload avatar/cover S3/MinIO | `POST /v1/upload` chưa wire — cần SDK aws-sdk-go-v2 |
| Worker Strava webhook async (asynq) | Hiện webhook chỉ trả 200; logic ingest activity → daily_steps cần async queue |
| Notification batch (`requestSubscribeMessage`) | Chỉ có wrapper send; cần job nightly nhắc streak |
| Admin OIDC (Google Workspace) | Hiện Basic Auth env — production cần thay |
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
