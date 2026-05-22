package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/buocvang/api/internal/db"
)

type DailySteps struct {
	UserID  string
	Day     time.Time
	Steps   int
	Source  string
	Flagged bool
}

type StepEvent struct {
	UserID      string
	Source      string
	Steps       int
	StartedAt   time.Time
	EndedAt     time.Time
	ClientNonce string
	Cadence     int // ms avg
	Flagged     bool
	FlagReason  string
}

type StepsStore struct{ pool *db.Pool }

func (s *StepsStore) UpsertDaily(ctx context.Context, tx pgx.Tx, userID string, day time.Time, steps int, source string, flagged bool) error {
	q := `INSERT INTO daily_steps (user_id, day, steps, source, flagged)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, day) DO UPDATE SET
		  steps = GREATEST(daily_steps.steps, EXCLUDED.steps),
		  source = CASE
		    WHEN daily_steps.source = EXCLUDED.source THEN EXCLUDED.source
		    ELSE 'merged' END,
		  flagged = daily_steps.flagged OR EXCLUDED.flagged`
	if tx != nil {
		_, err := tx.Exec(ctx, q, userID, day, steps, source, flagged)
		return err
	}
	_, err := s.pool.Exec(ctx, q, userID, day, steps, source, flagged)
	return err
}

func (s *StepsStore) GetDailyTotal(ctx context.Context, userID string, day time.Time) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COALESCE(steps,0) FROM daily_steps WHERE user_id=$1 AND day=$2`, userID, day).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return n, err
}

// InsertEvent ghi raw event. Trả ErrDuplicateNonce nếu client_nonce trùng.
var ErrDuplicateNonce = errors.New("client_nonce trùng")

func (s *StepsStore) InsertEvent(ctx context.Context, tx pgx.Tx, e StepEvent) error {
	q := `INSERT INTO step_events (user_id, source, steps, started_at, ended_at, client_nonce, cadence_avg_ms, flagged, flag_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9,''))`
	args := []any{e.UserID, e.Source, e.Steps, e.StartedAt, e.EndedAt, e.ClientNonce, e.Cadence, e.Flagged, e.FlagReason}
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, q, args...)
	} else {
		_, err = s.pool.Exec(ctx, q, args...)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateNonce
		}
		return err
	}
	return nil
}

// IncrementChallengeProgress cập nhật total_steps cho mọi participant active mà user đang join.
func (s *StepsStore) IncrementChallengeProgress(ctx context.Context, tx pgx.Tx, userID string, day time.Time, steps int) error {
	q := `UPDATE challenge_participants cp
		SET total_steps = total_steps + $3
		FROM challenges c
		WHERE cp.challenge_id = c.id
		  AND cp.user_id = $1
		  AND cp.state = 'in'
		  AND c.start_date <= $2 AND c.end_date >= $2
		  AND c.status IN ('open','live')`
	if tx != nil {
		_, err := tx.Exec(ctx, q, userID, day, steps)
		return err
	}
	_, err := s.pool.Exec(ctx, q, userID, day, steps)
	return err
}

func (s *StepsStore) ListDailyForRange(ctx context.Context, userID string, from, to time.Time) ([]DailySteps, error) {
	rows, err := s.pool.Query(ctx, `SELECT user_id, day, steps, source, flagged FROM daily_steps
		WHERE user_id=$1 AND day BETWEEN $2 AND $3 ORDER BY day DESC`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DailySteps{}
	for rows.Next() {
		var d DailySteps
		if err := rows.Scan(&d.UserID, &d.Day, &d.Steps, &d.Source, &d.Flagged); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
