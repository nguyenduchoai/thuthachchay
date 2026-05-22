-- name: ListActiveVoucherProducts :many
SELECT * FROM voucher_products
WHERE active = true
ORDER BY cost_points ASC;

-- name: GetVoucherProduct :one
SELECT * FROM voucher_products WHERE id = $1;

-- name: ClaimUnusedVoucherCode :one
-- SKIP LOCKED tránh race khi nhiều user redeem cùng product cùng lúc.
UPDATE voucher_codes
SET redeemed_by = $2, redeemed_at = now()
WHERE id = (
    SELECT id FROM voucher_codes
    WHERE product_id = $1 AND redeemed_by IS NULL
      AND (expires_at IS NULL OR expires_at > now())
    ORDER BY id
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: InsertRedemption :one
INSERT INTO redemptions
    (user_id, product_id, code_id, cost_points, status, idempotency_key)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, idempotency_key) DO NOTHING
RETURNING *;

-- name: ListMyRedemptions :many
SELECT r.*, p.name AS product_name, p.partner, c.code
FROM redemptions r
JOIN voucher_products p ON p.id = r.product_id
LEFT JOIN voucher_codes c ON c.id = r.code_id
WHERE r.user_id = $1
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;
