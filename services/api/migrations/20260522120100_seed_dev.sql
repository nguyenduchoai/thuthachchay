-- +goose Up
-- +goose StatementBegin
-- Seed dev: 5 user, 3 challenge, 4 voucher + codes.
INSERT INTO users (zalo_id, handle, display_name, daily_goal) VALUES
  ('dev:1', 'nguyenduchoai', 'Nguyễn Đức Hoài', 10000),
  ('dev:2', 'camhanes', 'Cam Hanes', 8400000),
  ('dev:3', 'marcus_w', 'Marcus W', 8000000),
  ('dev:4', 'joerogan', 'Joe Rogan', 6500000),
  ('dev:5', 'alex_t', 'Alex T', 5900000)
ON CONFLICT (zalo_id) DO NOTHING;

INSERT INTO challenges (host_id, visibility, name, description, daily_steps_target, duration_days, start_date, end_date, entry_points, prize_pool, status)
SELECT u.id, 'public', '75 Hard', 'Đi 7,500 bước mỗi ngày trong 75 ngày', 7500, 75, CURRENT_DATE, CURRENT_DATE + 74, 500, 6500, 'open'
FROM users u WHERE u.handle='nguyenduchoai' LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO challenges (host_id, visibility, name, description, daily_steps_target, duration_days, start_date, end_date, entry_points, prize_pool, status)
SELECT u.id, 'public', 'Memorial Miles', 'Đi 12,000 bước/ngày trong 4 ngày', 12000, 4, CURRENT_DATE + 1, CURRENT_DATE + 4, 100, 0, 'open'
FROM users u WHERE u.handle='marcus_w' LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO challenges (host_id, visibility, name, description, daily_steps_target, duration_days, start_date, end_date, entry_points, prize_pool, status)
SELECT u.id, 'public', 'Electric Steps', 'Sạc đầy năng lượng — 10K/ngày × 30 ngày', 10000, 30, CURRENT_DATE + 7, CURRENT_DATE + 36, 200, 0, 'open'
FROM users u WHERE u.handle='camhanes' LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO vouchers (brand, title, cost_points, stock) VALUES
  ('Highlands Coffee', 'Voucher 50.000đ', 5000, 100),
  ('Circle K', 'Voucher 20.000đ', 2000, 200),
  ('CGV', 'Vé xem phim 2D', 8000, 50),
  ('Grab', 'Mã giảm 30.000đ chuyến xe', 3000, 150)
ON CONFLICT DO NOTHING;

INSERT INTO voucher_codes (voucher_id, code)
SELECT v.id, 'HL-' || gs::text FROM vouchers v, generate_series(1, 100) AS gs WHERE v.brand='Highlands Coffee'
ON CONFLICT DO NOTHING;

INSERT INTO voucher_codes (voucher_id, code)
SELECT v.id, 'CK-' || gs::text FROM vouchers v, generate_series(1, 200) AS gs WHERE v.brand='Circle K'
ON CONFLICT DO NOTHING;

-- Seed mỗi user dev 5,000 điểm để demo
INSERT INTO ledger_entries (user_id, delta_points, reason, idempotency_key, note)
SELECT u.id, 5000, 'seed_bonus', 'seed:'||u.id::text, 'Welcome bonus' FROM users u WHERE u.zalo_id LIKE 'dev:%'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM voucher_codes;
DELETE FROM vouchers;
DELETE FROM ledger_entries WHERE reason='seed_bonus';
DELETE FROM challenges WHERE name IN ('75 Hard', 'Memorial Miles', 'Electric Steps');
DELETE FROM users WHERE zalo_id LIKE 'dev:%';
-- +goose StatementEnd
