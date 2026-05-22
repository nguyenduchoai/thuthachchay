# Changelog

Format theo [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) và [SemVer](https://semver.org/).

## [Unreleased]

### Added — Milestone 1 (scaffold)
- Monorepo skeleton theo `PHAN_TICH_KIEN_TRUC.md` §10.
- Governance: README, CONTRIBUTING, SECURITY, LICENSE, .editorconfig, .gitattributes, .gitignore, .prettierrc, .env.example, Makefile, .golangci.yml.
- Workspace npm cho `apps/miniapp` và `packages/api-client`.
- Go modules: `services/api` (cmd/api + cmd/worker) và `apps/admin`.
- Schema PostgreSQL khởi đầu — users, sessions, strava_tokens, daily_steps, step_ingest_events, challenges, challenge_participants, ledger_entries, voucher_products, voucher_codes, redemptions, notification_subs, audit_log + indexes + updated_at trigger.
- sqlc.yaml + queries cho users, steps, challenges, ledger, vouchers, leaderboards.
- OpenAPI 3.1 với ≥30 endpoint MVP.
- HTTPx middleware: recover, requestid, cors, access log + AppError JSON envelope.
- Unit test: `steps.Merge`, `antifraud.CadenceFlag`, healthz/version/501 routing, admin index/healthz.

### Added — Milestone 2 (miniapp)
- Vite + React 18 + TS strict + react-router + i18next vi/en + zustand.
- 22 route khớp §1 phân tích kiến trúc.
- Auth store persist + axios interceptor JWT refresh inflight-dedup.
- StepCounter peak-detection (4 test pass).
- BottomNav 5 tab, style Liquid Glass theo `mockup.html`.
- `@buocvang/api-client` placeholder types.
- Vite test exclude `._*` (macOS metadata exFAT).

### Added — Milestone 3 (infra + ops)
- `deploy/docker-compose.yml`: Postgres 16 + Redis 7 + MinIO + bucket auto-init.
- 4 Dockerfile multi-stage (api, worker, admin, miniapp) — distroless nonroot.
- `.github/workflows/`: ci.yml (4 job parallel) + security.yml (gitleaks, govulncheck, npm audit, cron weekly).
- dependabot.yml (gomod x2 + npm groups + actions + docker), CODEOWNERS, PR & issue templates.
- ADR 0001 process, 0002 điểm-không-tiền, 0003 Strava-primary.
- `docs/`: onboarding, runbook oncall, deploy.
- CLAUDE.md + `scripts/dev-bootstrap.sh` idempotent.
