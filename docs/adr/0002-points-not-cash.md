# 0002 — Reward = điểm + voucher, không tiền mặt P2P

- **Status**: Accepted
- **Date**: 2026-05-22

## Context

Mẫu Steppa gốc trả "$5 referral, $X prize". Việt Nam có rủi ro pháp lý:
"pool tiền + chia thưởng dựa trên kết quả ngẫu nhiên/skill" có thể bị xếp
vào nhóm cờ bạc/khuyến mại có thưởng cần xin phép.

## Decision

- Đơn vị nội bộ là **điểm** (không quy đổi tiền mặt).
- Điểm chỉ được đổi sang **voucher** từ pool đã ký hợp đồng với đối tác
  (F&B, vé phim, ví điện tử top-up…).
- T&C ghi rõ "điểm không có giá trị quy đổi tiền mặt".
- Referral, challenge payout đều trả bằng điểm.

## Consequences

- Tránh được rủi ro pháp lý lớn nhất cho MVP.
- Mất khả năng cashout trực tiếp (giảm appeal so với Steppa gốc).
- Phải vận hành **voucher inventory** ở admin (upload CSV mã, theo dõi
  stock, expire) — đã thiết kế trong `voucher_products`/`voucher_codes`.
- Cần đối tác cung cấp mã trước khi launch — risk: chậm hợp tác → kho rỗng.
