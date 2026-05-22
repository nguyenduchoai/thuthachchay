-- Chạy khi container postgres init lần đầu.
-- DB `buocvang` đã được tạo qua POSTGRES_DB env. Đây chỉ thêm extension.
\connect buocvang;
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";
