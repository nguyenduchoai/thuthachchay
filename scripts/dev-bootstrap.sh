#!/usr/bin/env bash
# Bootstrap local dev environment trong 1 lệnh.
# Yêu cầu: docker, go, node ≥ 20.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "→ Check toolchain"
for cmd in go node npm; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "❌ $cmd missing"; exit 1; }
done

if ! command -v docker >/dev/null 2>&1; then
  echo "⚠️  Docker không có — bỏ qua infra-up. Cài Docker Desktop và chạy lại."
  HAS_DOCKER=0
else
  HAS_DOCKER=1
fi

if [ ! -f .env ]; then
  echo "→ Tạo .env từ .env.example"
  cp .env.example .env
fi

echo "→ npm install"
npm install --no-audit --no-fund

echo "→ go mod download (api)"
(cd services/api && go mod download)
echo "→ go mod download (admin)"
(cd apps/admin && go mod download)

if [ "$HAS_DOCKER" -eq 1 ]; then
  echo "→ Khởi động infra (postgres + redis + minio)"
  docker compose -f deploy/docker-compose.yml up -d

  echo "→ Đợi Postgres ready..."
  for _ in $(seq 1 30); do
    if docker compose -f deploy/docker-compose.yml exec -T postgres pg_isready -U buocvang >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done

  echo "→ Chạy migration"
  if command -v goose >/dev/null 2>&1; then
    (cd services/api && goose -dir migrations postgres "$(grep DATABASE_URL ../../.env | cut -d= -f2-)" up)
  else
    echo "⚠️  goose chưa cài. go install github.com/pressly/goose/v3/cmd/goose@latest"
  fi
fi

cat <<'MSG'

✓ Setup xong.

Chạy 3 terminal:
  make api       # http://localhost:8080
  make admin     # http://localhost:8081
  make miniapp   # http://localhost:5173

Smoke:
  curl http://localhost:8080/healthz
MSG
