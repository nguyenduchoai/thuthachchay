# Bước Vàng (Steppa Vietnam)

> Zalo Mini App thử thách bước chân — đổi bước chân thành điểm và voucher.
> Kiến trúc đầy đủ trong [`../PHAN_TICH_KIEN_TRUC.md`](../PHAN_TICH_KIEN_TRUC.md).

## Monorepo layout

```
buoc-vang/
├─ apps/
│  ├─ miniapp/        # Zalo Mini App (React 18 + TS + zmp-ui)
│  └─ admin/          # Admin Go (Fiber + templ + HTMX + Tailwind)
├─ services/
│  └─ api/            # Public API Go (Fiber + sqlc + Postgres + Redis)
│     ├─ cmd/api/     # HTTP server binary
│     └─ cmd/worker/  # Background worker binary (cron, Strava webhooks)
├─ packages/
│  └─ api-client/     # TS SDK (generated from openapi.yaml)
├─ deploy/            # Docker Compose, Dockerfiles
├─ docs/              # ADR, runbooks, onboarding
└─ .github/workflows/ # CI pipelines
```

## Yêu cầu môi trường

| Tool | Version | Mục đích |
|---|---|---|
| Node.js | ≥ 20 (LTS) | miniapp, npm workspace |
| Go | ≥ 1.22 | services/api, apps/admin |
| Docker | ≥ 24 | Postgres, Redis, MinIO local |
| `goose` | latest | DB migration: `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| `sqlc` | ≥ 1.27 | Codegen từ SQL: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| `templ` | ≥ 0.2.7 | Templ HTML cho admin: `go install github.com/a-h/templ/cmd/templ@latest` |
| `zmp-cli` | latest | Build & deploy ZMP: `npm i -g zmp-cli` |

## Quickstart

```bash
# 1. Cài dependencies (mỗi workspace)
npm install                       # miniapp + api-client
go mod download                   # services/api + apps/admin

# 2. Khởi động infra local
docker compose -f deploy/docker-compose.yml up -d

# 3. Chạy migration
make db-up

# 4. Chạy 3 process song song (mở 3 terminal)
make api          # http://localhost:8080
make admin        # http://localhost:8081
make miniapp      # http://localhost:5173 (ZMP simulator)
```

## Make targets

```bash
make help          # Liệt kê tất cả targets
make lint          # Lint tất cả (eslint + golangci-lint)
make test          # Unit test tất cả
make build         # Build binaries + miniapp dist/
make sqlc          # Sinh code Go từ SQL queries
make templ         # Sinh code Go từ templates
make openapi       # Sinh TS SDK từ openapi.yaml
```

## Git workflow

- `main`: protected, mỗi PR cần ≥ 1 review + status check pass.
- Branch convention: `feat/<scope>-<desc>`, `fix/<scope>-<desc>`, `chore/<desc>`.
- Commit theo Conventional Commits.

## Đóng góp

Xem [`CONTRIBUTING.md`](./CONTRIBUTING.md). Báo cáo lỗ hổng bảo mật theo [`SECURITY.md`](./SECURITY.md).
