# Chính sách bảo mật

## Báo cáo lỗ hổng

Nếu phát hiện lỗ hổng bảo mật, **không mở issue công khai**. Gửi email tới `security@buocvang.vn` với:

- Mô tả lỗ hổng và tác động.
- Bước reproduce (PoC nếu có).
- Phiên bản/commit hash bị ảnh hưởng.
- Đề xuất fix (tùy chọn).

Chúng tôi cam kết:
- Phản hồi trong 48h làm việc.
- Patch P0 (Critical) trong 7 ngày.
- Patch P1 (High) trong 30 ngày.
- Ghi nhận credit trong release notes (nếu reporter đồng ý).

## Phạm vi

In-scope:
- `apps/miniapp` (client-side)
- `apps/admin` (admin web)
- `services/api` (public API, worker)
- Infrastructure (Docker images, CI pipeline)

Out-of-scope:
- Dependencies bên ngoài (báo upstream).
- Lỗi UI thuần (không có security impact).
- DoS bằng cách spam endpoint công khai mà không bypass rate-limit.

## Quy định nội bộ

- Secret rotation: tối đa 90 ngày.
- JWT key (RS256) rotation: 30 ngày, có overlap 7 ngày.
- Audit log: giữ ≥ 1 năm.
- Dependabot/Renovate alert P0 phải xử lý trong 48h.
- Pre-commit `gitleaks` + push protection bật trên GitHub.

## OWASP ASVS

Mục tiêu Level 1 cho MVP, Level 2 trước khi launch public.
