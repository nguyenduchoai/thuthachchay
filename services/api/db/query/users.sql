-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND status <> 'deleted';

-- name: GetUserByZaloID :one
SELECT * FROM users WHERE zalo_id = $1 AND status <> 'deleted';

-- name: GetUserByHandle :one
SELECT * FROM users WHERE handle = $1 AND status <> 'deleted';

-- name: CreateUser :one
INSERT INTO users (zalo_id, display_name, locale)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET handle       = COALESCE(sqlc.narg('handle'), handle),
    display_name = COALESCE(sqlc.narg('display_name'), display_name),
    avatar_url   = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    daily_goal   = COALESCE(sqlc.narg('daily_goal'), daily_goal),
    locale       = COALESCE(sqlc.narg('locale'), locale)
WHERE id = sqlc.arg('id') AND status <> 'deleted'
RETURNING *;

-- name: SetUserAttribution :exec
UPDATE users SET acquisition = $2 WHERE id = $1;

-- name: SuspendUser :exec
UPDATE users SET status = 'suspended' WHERE id = $1;

-- name: CheckHandleAvailable :one
SELECT NOT EXISTS (
    SELECT 1 FROM users WHERE handle = $1 AND status <> 'deleted'
) AS available;

-- name: GetUserBalance :one
SELECT balance_points FROM user_balances WHERE user_id = $1;
