package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/buocvang/api/internal/db"
)

type LedgerEntry struct {
	ID             string
	UserID         string
	DeltaPoints    int
	Reason         string
	ReferenceType  string
	ReferenceID    string
	IdempotencyKey string
	Note           *string
	CreatedAt      time.Time
}

type WalletStore struct{ pool *db.Pool }

var ErrDuplicateIdempotencyKey = errors.New("idempotency_key trùng")

// Balance dùng SUM ledger.
func (s *WalletStore) Balance(ctx context.Context, userID string) (int, error) {
	var n *int
	if err := s.pool.QueryRow(ctx, `SELECT SUM(delta_points)::int FROM ledger_entries WHERE user_id=$1`, userID).Scan(&n); err != nil {
		return 0, err
	}
	if n == nil {
		return 0, nil
	}
	return *n, nil
}

// AppendEntry idempotent qua unique (user_id, idempotency_key).
func (s *WalletStore) AppendEntry(ctx context.Context, tx pgx.Tx, e LedgerEntry) error {
	q := `INSERT INTO ledger_entries (user_id, delta_points, reason, reference_type, reference_id, idempotency_key, note)
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), $6, NULLIF($7,''))`
	args := []any{e.UserID, e.DeltaPoints, e.Reason, e.ReferenceType, e.ReferenceID, e.IdempotencyKey, ""}
	if e.Note != nil {
		args[len(args)-1] = *e.Note
	}
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, q, args...)
	} else {
		_, err = s.pool.Exec(ctx, q, args...)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateIdempotencyKey
		}
		return err
	}
	return nil
}

func (s *WalletStore) RecentEntries(ctx context.Context, userID string, limit int) ([]LedgerEntry, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `SELECT id, user_id, delta_points, reason, COALESCE(reference_type,''), COALESCE(reference_id,''), idempotency_key, note, created_at
		FROM ledger_entries WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LedgerEntry{}
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.DeltaPoints, &e.Reason, &e.ReferenceType, &e.ReferenceID, &e.IdempotencyKey, &e.Note, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
