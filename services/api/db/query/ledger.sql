-- name: InsertLedgerEntry :one
INSERT INTO ledger_entries
    (user_id, delta_points, reason, reference_type, reference_id, idempotency_key, note)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, idempotency_key) DO NOTHING
RETURNING *;

-- name: ListLedger :many
SELECT * FROM ledger_entries
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
