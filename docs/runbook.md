# Runbook cho oncall

## Cấu trúc service trên prod

| Service | Binary | Port | Healthcheck |
|---|---|---|---|
| Public API | `services/api/cmd/api` | 8080 | `GET /healthz` |
| Worker | `services/api/cmd/worker` | – | log heartbeat |
| Admin | `apps/admin/cmd/admin` | 8081 | `GET /healthz` |
| Miniapp | ZMP CDN (Zalo deploy) | – | – |

## Incident playbook

### API trả 5xx hàng loạt

1. Kiểm `GET /readyz` → fail ở component nào (db/redis).
2. Xem dashboard Grafana `api-latency` + `api-error-rate`.
3. Nếu DB down → kiểm pgbouncer + RDS metric.
4. Nếu Redis down → API vẫn chạy được (degrade), cache miss tăng. Khôi phục
   nhanh, tránh restart API hàng loạt.
5. Tag escalation #engineering, log incident vào `docs/incidents/YYYY-MM-DD.md`.

### Lượng "flagged" tăng đột biến

1. Check Grafana panel `daily_steps_flagged_rate`.
2. Truy vấn `flag_reasons` top:
   ```sql
   SELECT unnest(flag_reasons) AS r, COUNT(*) FROM daily_steps
   WHERE day = CURRENT_DATE GROUP BY r ORDER BY 2 DESC;
   ```
3. Nếu là `cadence_too_fast`: có thể bot đang scan endpoint `/v1/steps/ingest`
   → tăng rate limit hoặc tạm khoá user IP cluster.
4. Nếu `source_diverge`: Strava webhook trễ → check Strava status.

### Voucher cháy

1. Admin → Reports → DSO → voucher stock by product.
2. Khi stock < 10%: gửi alert #ops + tạm ẩn product khỏi miniapp.
3. Liên hệ đối tác refill mã.
4. Email user trong queue redeem chờ.

### Strava ngừng push webhook

1. Test endpoint webhook:
   ```bash
   curl "https://api.buocvang.vn/v1/strava/webhook?hub.mode=subscribe&hub.verify_token=$TOKEN&hub.challenge=test"
   ```
2. Nếu 200 OK nhưng vẫn không event → re-subscribe:
   ```bash
   curl -X POST https://api.strava.com/api/v3/push_subscriptions \
     -F client_id=$STRAVA_CLIENT_ID \
     -F client_secret=$STRAVA_CLIENT_SECRET \
     -F callback_url=https://api.buocvang.vn/v1/strava/webhook \
     -F verify_token=$TOKEN
   ```

### Đợt suspend nhầm

1. SQL khôi phục:
   ```sql
   UPDATE users SET status='active', fraud_score=0
   WHERE id = '<uuid>' AND status='suspended';
   ```
2. Gửi email xin lỗi, ghi audit_log `actor_type='admin', action='manual_unsuspend'`.

## Quay trở lại sau rollback

Sau khi rollback prod về version cũ:
1. Verify `goose -dir migrations postgres "$DATABASE_URL" status` — không có
   migration "Up" mà code cũ không hiểu.
2. Nếu schema đã forward — chạy `goose down` thủ công.

## Liên hệ

| Loại | Kênh | SLA phản hồi |
|---|---|---|
| P0 (down) | PagerDuty | 15 phút |
| P1 (degrade) | Slack #incidents | 1 giờ |
| P2 | Linear | 1 ngày |
| Bảo mật | `security@buocvang.vn` | 48 giờ |
