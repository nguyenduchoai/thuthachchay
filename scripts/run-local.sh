#!/usr/bin/env bash
# Bước Vàng — script chạy đầy đủ local trên Mac.
# Yêu cầu trước: docker, go ≥ 1.22, node ≥ 20, goose (go install github.com/pressly/goose/v3/cmd/goose@latest).

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export APP_ENV=dev
export LOG_LEVEL=debug
export HTTP_LISTEN_ADDR=:8080
export ADMIN_LISTEN_ADDR=:8081
export DATABASE_URL="postgres://buocvang:buocvang@localhost:5432/buocvang?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"
export ADMIN_USER="admin"
export ADMIN_PASS="buocvang-dev"

color() { printf "\033[1;33m%s\033[0m\n" "$*"; }
ok()    { printf "\033[1;32m✓\033[0m %s\n" "$*"; }
fail()  { printf "\033[1;31m✗\033[0m %s\n" "$*"; exit 1; }

color "1/6 ▸ Kiểm tra dependencies"
command -v docker  >/dev/null || fail "Cần Docker Desktop"
command -v go      >/dev/null || fail "Cần Go ≥ 1.22"
command -v node    >/dev/null || fail "Cần Node ≥ 20"
command -v goose   >/dev/null || {
  color "  → goose chưa có, cài bằng: go install github.com/pressly/goose/v3/cmd/goose@latest"
  fail "Cài goose rồi chạy lại"
}
ok "Đủ tool"

color "2/6 ▸ Khởi động Postgres + Redis (docker compose)"
docker compose -f deploy/docker-compose.yml up -d postgres redis
echo "Chờ Postgres healthy..."
for i in {1..30}; do
  if docker exec buocvang-postgres pg_isready -U buocvang -d buocvang >/dev/null 2>&1; then
    ok "Postgres ready"
    break
  fi
  sleep 1
done

color "3/6 ▸ Chạy migration + seed (goose up)"
goose -dir services/api/migrations postgres "$DATABASE_URL" up
ok "DB schema + seed sẵn sàng"

color "4/6 ▸ Build Go binaries"
(cd services/api && go mod tidy && go build -o bin/api ./cmd/api && go build -o bin/worker ./cmd/worker)
(cd apps/admin && go mod tidy && go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate -path internal/web && go build -o bin/admin ./cmd/admin)
ok "Đã build api / worker / admin"

color "5/6 ▸ Cài deps Mini App (nếu chưa)"
if [ ! -d apps/miniapp/node_modules ]; then
  (cd apps/miniapp && npm install --no-fund --no-audit)
fi
ok "Mini App deps OK"

color "6/6 ▸ Khởi chạy 4 process (mỗi terminal 1 lệnh)"
cat <<EOF

Mở 4 terminal tab và chạy lần lượt:

  ① cd $ROOT && ./services/api/bin/api
       → http://localhost:8080  (public API)

  ② cd $ROOT && ./services/api/bin/worker
       → background cron (settle / leaderboard / referral)

  ③ cd $ROOT && DATABASE_URL="$DATABASE_URL" ADMIN_USER=admin ADMIN_PASS=buocvang-dev ./apps/admin/bin/admin
       → http://localhost:8081  (basic auth: admin / buocvang-dev)

  ④ cd $ROOT/apps/miniapp && VITE_API_BASE_URL=http://localhost:8080 npm run dev
       → http://localhost:5173  (Mini App dev — xem trong browser)

Smoke test (terminal khác):

  curl -X POST http://localhost:8080/v1/auth/zalo \\
       -H 'Content-Type: application/json' \\
       -d '{"zalo_access_token":"dev:u1"}'

Dừng stack: docker compose -f deploy/docker-compose.yml down
EOF
