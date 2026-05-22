# 0003 — Strava là nguồn bước chân chính, ZMP accelerometer là phụ

- **Status**: Accepted
- **Date**: 2026-05-22

## Context

Có 2 nguồn đếm bước:
1. **ZMP accelerometer** (`zmp-sdk.onAccelerometerChange`): chỉ chạy khi
   miniapp foreground/màn hình bật.
2. **Strava API**: hậu kiểm activity, có GPS, có moving_time.

ZMP có giới hạn nghiêm trọng: khi user khoá màn hình hoặc thoát app, sensor
ngừng cập nhật → mất dữ liệu lớn. Đây là rủi ro #1 trong §12 của phân tích.

## Decision

- **Strava là nguồn ưu tiên** cho user đã connect.
- ZMP chỉ tính **bổ sung khi Strava không có activity ngày đó**.
- Merge: nếu cả 2 đều có, lấy `max(zmp, strava)`. Chênh > 30% → `flagged`.
- UI nhắc "Mở app khi đang đi bộ" cho user chưa connect Strava.

## Consequences

- Phụ thuộc Strava → cần handle rate limit (100 req/15p, 1000/day) bằng
  webhook + cache, không poll.
- Strava OAuth thêm friction onboarding → cho phép skip, nhắc lại sau 3 ngày.
- Đếm bước realtime trên ZMP vẫn cần (để UI hiển thị tiến độ trong ngày);
  reconciliation theo Strava xảy ra cuối ngày qua worker.
