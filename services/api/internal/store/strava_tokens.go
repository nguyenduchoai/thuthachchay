package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
)

type StravaToken struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Scope        string
	AthleteID    string
}

type StravaTokensStore struct{ pool *db.Pool }

func (s *StravaTokensStore) Upsert(ctx context.Context, t StravaToken) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO strava_tokens (user_id, access_token, refresh_token, expires_at, scope, athlete_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
		  access_token=EXCLUDED.access_token,
		  refresh_token=EXCLUDED.refresh_token,
		  expires_at=EXCLUDED.expires_at,
		  scope=EXCLUDED.scope,
		  athlete_id=EXCLUDED.athlete_id,
		  updated_at=now()`,
		t.UserID, t.AccessToken, t.RefreshToken, t.ExpiresAt, t.Scope, t.AthleteID)
	return err
}

func (s *StravaTokensStore) Get(ctx context.Context, userID string) (*StravaToken, error) {
	row := s.pool.QueryRow(ctx, `SELECT user_id, access_token, refresh_token, expires_at, scope, athlete_id
		FROM strava_tokens WHERE user_id=$1`, userID)
	var t StravaToken
	if err := row.Scan(&t.UserID, &t.AccessToken, &t.RefreshToken, &t.ExpiresAt, &t.Scope, &t.AthleteID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (s *StravaTokensStore) FindByAthleteID(ctx context.Context, athleteID string) (*StravaToken, error) {
	row := s.pool.QueryRow(ctx, `SELECT user_id, access_token, refresh_token, expires_at, scope, athlete_id
		FROM strava_tokens WHERE athlete_id=$1`, athleteID)
	var t StravaToken
	if err := row.Scan(&t.UserID, &t.AccessToken, &t.RefreshToken, &t.ExpiresAt, &t.Scope, &t.AthleteID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}
