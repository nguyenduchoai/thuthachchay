package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
)

type Challenge struct {
	ID               string
	HostID           *string
	Visibility       string
	Name             string
	Description      *string
	CoverURL         *string
	DailyStepsTarget int
	DurationDays     int
	StartDate        time.Time
	EndDate          time.Time
	EntryPoints      int
	PrizePool        int
	SponsorID        *string
	MaxParticipants  *int
	Status           string
	CreatedAt        time.Time
}

type Participant struct {
	ChallengeID string
	UserID      string
	JoinedAt    time.Time
	EntryPaid   int
	State       string
	TotalSteps  int64
}

type ChallengesStore struct{ pool *db.Pool }

const challengeColumns = `id, host_id, visibility, name, description, cover_url, daily_steps_target, duration_days, start_date, end_date, entry_points, prize_pool, sponsor_id, max_participants, status, created_at`

func scanChallenge(row pgx.Row) (*Challenge, error) {
	c := &Challenge{}
	if err := row.Scan(&c.ID, &c.HostID, &c.Visibility, &c.Name, &c.Description, &c.CoverURL,
		&c.DailyStepsTarget, &c.DurationDays, &c.StartDate, &c.EndDate, &c.EntryPoints, &c.PrizePool,
		&c.SponsorID, &c.MaxParticipants, &c.Status, &c.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *ChallengesStore) Get(ctx context.Context, id string) (*Challenge, error) {
	return scanChallenge(s.pool.QueryRow(ctx, `SELECT `+challengeColumns+` FROM challenges WHERE id=$1`, id))
}

func (s *ChallengesStore) List(ctx context.Context, status string, limit, offset int) ([]Challenge, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+challengeColumns+` FROM challenges
		WHERE ($1='' OR status=$1)
		ORDER BY start_date DESC, created_at DESC
		LIMIT $2 OFFSET $3`, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Challenge{}
	for rows.Next() {
		c, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

type CreateChallengeInput struct {
	HostID           string
	Visibility       string
	Name             string
	Description      string
	CoverURL         string
	DailyStepsTarget int
	DurationDays     int
	StartDate        time.Time
	EntryPoints      int
	MaxParticipants  int
}

func (s *ChallengesStore) Create(ctx context.Context, tx pgx.Tx, in CreateChallengeInput) (*Challenge, error) {
	endDate := in.StartDate.AddDate(0, 0, in.DurationDays-1)
	q := `INSERT INTO challenges (host_id, visibility, name, description, cover_url, daily_steps_target, duration_days, start_date, end_date, entry_points, prize_pool, max_participants, status)
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), $6, $7, $8, $9, $10, $11, NULLIF($12,0), 'open')
		RETURNING ` + challengeColumns
	args := []any{in.HostID, in.Visibility, in.Name, in.Description, in.CoverURL, in.DailyStepsTarget, in.DurationDays, in.StartDate, endDate, in.EntryPoints, 0, in.MaxParticipants}
	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, q, args...)
	} else {
		row = s.pool.QueryRow(ctx, q, args...)
	}
	return scanChallenge(row)
}

func (s *ChallengesStore) AddPrizePool(ctx context.Context, tx pgx.Tx, id string, delta int) error {
	q := `UPDATE challenges SET prize_pool = prize_pool + $2 WHERE id=$1`
	if tx != nil {
		_, err := tx.Exec(ctx, q, id, delta)
		return err
	}
	_, err := s.pool.Exec(ctx, q, id, delta)
	return err
}

func (s *ChallengesStore) AddParticipant(ctx context.Context, tx pgx.Tx, challengeID, userID string, entryPaid int) error {
	_, err := tx.Exec(ctx, `INSERT INTO challenge_participants (challenge_id, user_id, entry_paid, state)
		VALUES ($1, $2, $3, 'in')
		ON CONFLICT (challenge_id, user_id) DO NOTHING`, challengeID, userID, entryPaid)
	return err
}

func (s *ChallengesStore) ParticipantCount(ctx context.Context, challengeID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM challenge_participants WHERE challenge_id=$1`, challengeID).Scan(&n)
	return n, err
}

func (s *ChallengesStore) ListParticipants(ctx context.Context, challengeID string, limit int) ([]Participant, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `SELECT challenge_id, user_id, joined_at, entry_paid, state, total_steps
		FROM challenge_participants WHERE challenge_id=$1 ORDER BY total_steps DESC LIMIT $2`, challengeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Participant{}
	for rows.Next() {
		var p Participant
		if err := rows.Scan(&p.ChallengeID, &p.UserID, &p.JoinedAt, &p.EntryPaid, &p.State, &p.TotalSteps); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *ChallengesStore) SetParticipantState(ctx context.Context, tx pgx.Tx, challengeID, userID, state string) error {
	q := `UPDATE challenge_participants SET state=$3 WHERE challenge_id=$1 AND user_id=$2`
	if tx != nil {
		_, err := tx.Exec(ctx, q, challengeID, userID, state)
		return err
	}
	_, err := s.pool.Exec(ctx, q, challengeID, userID, state)
	return err
}

func (s *ChallengesStore) SetStatus(ctx context.Context, id, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE challenges SET status=$2 WHERE id=$1`, id, status)
	return err
}

// EligibleForSettle: thử thách end_date <= today và status='live' hoặc 'open'.
func (s *ChallengesStore) EligibleForSettle(ctx context.Context) ([]Challenge, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+challengeColumns+` FROM challenges
		WHERE end_date < CURRENT_DATE AND status IN ('open','live') ORDER BY end_date ASC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Challenge{}
	for rows.Next() {
		c, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}
