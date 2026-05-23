# Onboarding cho dev mới

## 0. Đọc trước

1. [`PHAN_TICH_KIEN_TRUC.md`](../../PHAN_TICH_KIEN_TRUC.md) — bản kiến trúc, lộ trình MVP 8 tuần.
2. [`SDLC_CHECKLIST_BIZINO.md`](../../SDLC_CHECKLIST_BIZINO.md) — chuẩn SDLC team.
3. [`mockup.html`](../../mockup.html) — wireframe Liquid Glass.
4. ADR ở [`docs/adr/`](./adr/README.md).

## 1. Setup máy (15 phút)

```bash
# Cài toolchain
brew install go node@20 git
brew install --cask docker
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/a-h/templ/cmd/templ@latest
npm i -g zmp-cli

# Clone & cài deps
git clone <repo-url> buoc-vang
cd buoc-vang
cp .env.example .env
npm install
(cd services/api && go mod download)
(cd apps/admin && go mod download)
```

## 2. Boot dev stack (5 phút)

```bash
make infra-up                    # postgres + redis + minio
make db-up                       # chạy migration
# 3 terminal:
make api                         # http://localhost:8080
make admin                       # http://localhost:8081
make miniapp                     # http://localhost:5173
```

Smoke:
```bash
curl http://localhost:8080/healthz   # {"status":"ok"}
curl http://localhost:8080/version
open http://localhost:5173
open http://localhost:8081
```

## 3. Tạo migration mới

```bash
cd services/api
goose -dir migrations create add_something sql
# Sửa file, viết cả Up & Down
make db-up
make sqlc                        # nếu thêm bảng/cột → gen lại store
```

## 4. Sửa SQL queries

1. Thêm/sửa `services/api/db/query/<domain>.sql`.
2. Chạy `make sqlc`.
3. Code Go ở `internal/store/` được generate lại → commit kèm.

## 5. Sửa OpenAPI

1. Sửa `services/api/openapi.yaml`.
2. Wire handler thật trong `services/api/internal/server` và package domain tương ứng.
3. Cập nhật `packages/api-client/src/index.ts` để SDK khớp spec.
4. Miniapp import API qua `@buocvang/api-client` từ `apps/miniapp/src/services/endpoints.ts`.
5. Chạy `npx --yes @redocly/cli@latest lint services/api/openapi.yaml`.

## 6. Test convention

- **Unit**: cùng package, `*_test.go` (Go) / `*.test.ts` (TS).
- **Integration**: `services/api/internal/<pkg>/integration_test.go` với
  `dockertest`/`testcontainers-go` cho Postgres ephemeral.
- **E2E miniapp**: `apps/miniapp/e2e/` Playwright + ZMP simulator.

## 7. Git workflow

- Branch: `feat/<scope>-<short>`, `fix/...`, `chore/...`.
- Commit: Conventional Commits (xem `CONTRIBUTING.md`).
- PR: template tự fill, cần ≥ 1 review, CI xanh.

## 8. Câu hỏi nóng

> **Đếm bước thật chính xác đến đâu?**
> Tham khảo ADR 0003. Strava là nguồn chính khi user connect; ZMP phụ.

> **Tại sao không P2P tiền?**
> ADR 0002 — pháp lý.

> **Admin login bằng gì?**
> Google Workspace OIDC + IP allowlist + audit log.
