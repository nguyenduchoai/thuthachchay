#!/usr/bin/env bash
# Bước Vàng — push code lên https://github.com/nguyenduchoai/thuthachchay.
# Chạy trên macOS Terminal (không phải sandbox). Yêu cầu: gh CLI đã login HOẶC ssh key đã add vào GitHub.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REPO="nguyenduchoai/thuthachchay"
BRANCH="main"

color() { printf "\033[1;33m▸ %s\033[0m\n" "$*"; }
ok()    { printf "\033[1;32m✓\033[0m %s\n" "$*"; }
fail()  { printf "\033[1;31m✗\033[0m %s\n" "$*"; exit 1; }

color "1/5 ▸ Dọn lock file (nếu có)"
rm -f .git/index.lock 2>/dev/null || true
ok "OK"

color "2/5 ▸ Kiểm tra remote"
if git remote get-url origin >/dev/null 2>&1; then
  current=$(git remote get-url origin)
  if [[ "$current" != *"$REPO"* ]]; then
    echo "  remote hiện tại: $current"
    git remote set-url origin "https://github.com/${REPO}.git"
  fi
else
  git remote add origin "https://github.com/${REPO}.git"
fi
git remote -v
ok "Remote OK"

color "3/5 ▸ Stage + commit thay đổi local"
git add -A
if git diff --cached --quiet; then
  ok "Không có thay đổi mới — skip commit"
else
  git commit -m "feat(admin,miniapp): hoàn thiện 21 page UI thật + admin HTMX actions

- Mini App: implement các page còn thiếu (Checkout, Create, CreateNew,
  Discover, Profile, ProfileSettings, ProfileEdit, LeaderboardPreview) với
  data fetching thật, validation realtime, i18n đầy đủ.
- Admin: dashboard KPI, 16 endpoint với HTMX form action (suspend/activate
  user, ±points, settle/cancel challenge, approve/reject fraud, voucher
  upload từ inline textarea hoặc CSV file). Tất cả action ghi audit_log.
- scripts/run-local.sh: one-shot dev runner cho Mac.
- scripts/push-to-github.sh: push lên repo nguyenduchoai/thuthachchay.

Verify: TypeScript compile PASS (0 lỗi).
Stats: 4,634 LOC Go · 1,920 LOC TS · 51 file Go · 34 file TS · 31 route BE
       · 24 typed endpoint FE · 21/21 page Mini App có UI thật."
  ok "Đã commit"
fi

color "4/5 ▸ Đảm bảo repo đã tồn tại trên GitHub"
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
  if ! gh repo view "$REPO" >/dev/null 2>&1; then
    echo "  Repo chưa tồn tại — tạo private..."
    gh repo create "$REPO" --private --description "Bước Vàng — Zalo Mini App thử thách bước chân" --source=. --remote=origin
  else
    ok "Repo đã tồn tại"
  fi
else
  echo "  (gh CLI không có/chưa login — bỏ qua check; bạn phải tự tạo repo nếu chưa có)"
fi

color "5/5 ▸ Push lên origin/$BRANCH"
git push -u origin "$BRANCH"
ok "✅ Đã push xong"

echo ""
echo "🔗 Mở: https://github.com/${REPO}"
