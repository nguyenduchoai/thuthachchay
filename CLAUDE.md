# CLAUDE.md — Bước Vàng (Steppa Vietnam)

Tài liệu định hướng cho Claude (hoặc AI assistant) làm việc trong repo này.

## Repo là gì

Monorepo cho ứng dụng **Bước Vàng** — Zalo Mini App thử thách bước chân:
- `apps/miniapp/` — ZMP React+TS (path alias `@/` → `src/`).
- `apps/admin/` — Admin Go (Fiber, sẽ chuyển templ+HTMX).
- `services/api/` — Public API Go (Fiber, pgx, sqlc, goose).
- `packages/api-client/` — TS SDK gen từ OpenAPI.
- Thiết kế gốc: [`../PHAN_TICH_KIEN_TRUC.md`](../PHAN_TICH_KIEN_TRUC.md).
- ADR: `docs/adr/`.

## Quy tắc làm việc

1. **Đọc PHAN_TICH_KIEN_TRUC.md và ADR liên quan trước khi đổi kiến trúc.**
   Không thay đổi quyết định trong ADR mà chưa supersede bằng ADR mới.
2. **Migration**: tạo qua `goose -dir migrations create ... sql`; cả Up & Down;
   không xoá cột có data — luôn 2 bước.
3. **SQL**: viết ở `services/api/db/query/*.sql`, không string concat trong Go.
   Chạy `make sqlc` sau khi sửa.
4. **OpenAPI**: nếu sửa endpoint → cập nhật `services/api/openapi.yaml` + wire
   handler + chạy `make openapi` để regen TS SDK.
5. **Test**: mọi domain logic mới phải có unit test. Integration test dùng DB
   ephemeral (testcontainers), không mock.
6. **Bí mật**: không commit `.env`/credential. Thêm key mới vào `.env.example`
   kèm comment.
7. **Commit**: Conventional Commits (`feat(api): ...`, `fix(miniapp): ...`).
8. **Bundle ZMP**: phải ≤ 10MB sau build (CI có gate).
9. **Không dùng `eval`/`new Function()`** trong miniapp (Zalo cấm).
10. **i18n**: thêm string mới phải có cả `vi` lẫn `en` trong `apps/miniapp/src/i18n/`.

## Một số nguyên tắc kiến trúc (tóm tắt)

- **Reward bằng điểm + voucher**, không tiền mặt (ADR 0002).
- **Strava là nguồn bước chính** khi user connect (ADR 0003).
- **Anti-fraud**: flag mà vẫn cộng vào daily_steps (UX không bị block);
  worker chấm fraud_score đêm; ≥ 90 trong 3 ngày → auto suspend.
- **Auth**: Zalo Login (`getAccessToken` ZMP) → JWT RS256, access 15p, refresh 30d.
- **Admin**: OIDC Google Workspace + IP allowlist + audit_log mọi action mutation.

## Quy ước commit channel

| Type | Khi nào |
|---|---|
| `feat` | Tính năng mới user thấy |
| `fix` | Sửa lỗi |
| `refactor` | Refactor không đổi behavior |
| `chore` | Build, deps, infra, không ảnh hưởng product |
| `docs` | Chỉ tài liệu |
| `test` | Chỉ test |

Scope chuẩn: `api`, `worker`, `admin`, `miniapp`, `db`, `infra`, `ci`, `deps`.
