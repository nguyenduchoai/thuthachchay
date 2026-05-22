# Changelog

Format theo [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) và [SemVer](https://semver.org/).

## [Unreleased]

### Added
- Khởi tạo monorepo skeleton theo `PHAN_TICH_KIEN_TRUC.md` §10.
- Governance: README, CONTRIBUTING, SECURITY, LICENSE, .editorconfig, .gitattributes, .gitignore.
- Workspace npm cho `apps/miniapp` và `packages/api-client`.
- Go module skeleton cho `services/api` (cmd/api, cmd/worker) và `apps/admin`.
- Schema PostgreSQL khởi đầu (users, sessions, daily_steps, challenges, ledger, vouchers).
- OpenAPI 3.1 skeleton với các endpoint MVP.
- Docker Compose dev stack (Postgres 16 + Redis 7 + MinIO).
- CI GitHub Actions: lint + test + build cho cả Go và TS.
- Makefile orchestration.
