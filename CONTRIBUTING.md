# Đóng góp cho Bước Vàng

## Workflow

1. Tạo branch từ `main`: `feat/<scope>-<short-desc>`.
2. Commit theo [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat(api): add step ingest endpoint`
   - `fix(miniapp): handle accelerometer permission denied`
   - `chore(deploy): bump postgres to 16.4`
3. Đảm bảo `make lint test` pass trước khi mở PR.
4. PR cần ≥ 1 review + tất cả status check xanh.

## Code style

- **Go**: `gofmt` + `golangci-lint` (config ở `.golangci.yml`). Không bỏ `error`. Truyền `context.Context` xuyên call.
- **TypeScript**: ESLint + Prettier (config ở root). `strict: true`. Tránh `any`.
- **SQL**: dùng `sqlc` — viết SQL ở `services/api/db/query/*.sql`, không viết SQL string trong Go code.
- **HTML templates**: dùng `templ` cho admin, không string concatenation.

## Test

- **Unit test**: cùng package với source, tên `*_test.go` (Go) hoặc `*.test.ts` (TS).
- **Integration test**: ở `services/api/internal/<pkg>/integration_test.go`, dùng `dockertest`/`testcontainers-go` cho Postgres ephemeral.
- **E2E test miniapp**: `apps/miniapp/e2e/` dùng Playwright (chạy với ZMP simulator).
- Ngưỡng coverage tối thiểu: 60% (lý tưởng 80% cho domain logic).

## Branch protection

Branch `main`:
- Cấm force-push.
- Yêu cầu PR + 1 review.
- Yêu cầu status check: `lint`, `test`, `build`.

## Bí mật & cấu hình

- Không commit `.env` hay credential. Mỗi env var phải có entry trong `.env.example` với comment giải thích.
- Production secret quản lý bằng Doppler/Vault — xem `docs/runbook-secrets.md`.

## Tạo migration mới

```bash
cd services/api
goose -dir migrations create <ten_migration> sql
```

Migration phải reversible (cả `Up` và `Down`).

## Sinh code

- Sau khi sửa `services/api/db/query/*.sql`: chạy `make sqlc`.
- Sau khi sửa `apps/admin/templates/*.templ`: chạy `make templ`.
- Sau khi sửa `services/api/openapi.yaml`: chạy `make openapi`.

Code sinh ra **được commit vào repo** để CI không cần chạy generator.
