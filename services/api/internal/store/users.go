package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
)

type User struct {
	ID            string
	ZaloID        string
	Handle        *string
	Email         *string
	DisplayName   *string
	AvatarURL     *string
	DailyGoal     int
	Locale        string
	Acquisition   *string
	StravaUserID  *string
	Status        string
	FraudScore    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UsersStore struct{ pool *db.Pool }

var ErrNotFound = errors.New("not found")

const userColumns = `id, zalo_id, handle, email, display_name, avatar_url, daily_goal, locale, acquisition, strava_user_id, status, fraud_score, created_at, updated_at`

func scanUser(row pgx.Row) (*User, error) {
	u := &User{}
	if err := row.Scan(&u.ID, &u.ZaloID, &u.Handle, &u.Email, &u.DisplayName, &u.AvatarURL,
		&u.DailyGoal, &u.Locale, &u.Acquisition, &u.StravaUserID, &u.Status, &u.FraudScore,
		&u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *UsersStore) GetByID(ctx context.Context, id string) (*User, error) {
	return scanUser(s.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1`, id))
}

func (s *UsersStore) GetByZaloID(ctx context.Context, zaloID string) (*User, error) {
	return scanUser(s.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE zalo_id=$1`, zaloID))
}

func (s *UsersStore) HandleAvailable(ctx context.Context, handle string) (bool, error) {
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE handle=$1`, handle).Scan(&n); err != nil {
		return false, err
	}
	return n == 0, nil
}

// UpsertByZaloID tạo nếu chưa có, return user + isNew.
func (s *UsersStore) UpsertByZaloID(ctx context.Context, zaloID, displayName, avatarURL, locale string) (*User, bool, error) {
	existing, err := s.GetByZaloID(ctx, zaloID)
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, false, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (zalo_id, display_name, avatar_url, locale, daily_goal, status)
		VALUES ($1, NULLIF($2,''), NULLIF($3,''), COALESCE(NULLIF($4,''), 'vi-VN'), 10000, 'active')
		RETURNING `+userColumns,
		zaloID, displayName, avatarURL, locale)
	u, err := scanUser(row)
	if err != nil {
		return nil, false, err
	}
	return u, true, nil
}

type UserPatch struct {
	Handle      *string
	DisplayName *string
	AvatarURL   *string
	DailyGoal   *int
	Locale      *string
	Email       *string
	Acquisition *string
}

func (s *UsersStore) Patch(ctx context.Context, id string, p UserPatch) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET
			handle        = COALESCE($2, handle),
			display_name  = COALESCE($3, display_name),
			avatar_url    = COALESCE($4, avatar_url),
			daily_goal    = COALESCE($5, daily_goal),
			locale        = COALESCE($6, locale),
			email         = COALESCE($7, email),
			acquisition   = COALESCE($8, acquisition),
			updated_at    = now()
		WHERE id=$1
		RETURNING `+userColumns,
		id, p.Handle, p.DisplayName, p.AvatarURL, p.DailyGoal, p.Locale, p.Email, p.Acquisition)
	return scanUser(row)
}

func (s *UsersStore) IncrementFraudScore(ctx context.Context, id string, delta int) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET fraud_score = LEAST(100, fraud_score + $2), updated_at=now() WHERE id=$1`, id, delta)
	return err
}

func (s *UsersStore) Suspend(ctx context.Context, id, reason string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET status='suspended', updated_at=now() WHERE id=$1`, id)
	return err
}

func (s *UsersStore) SetStravaUserID(ctx context.Context, id, stravaUserID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET strava_user_id=$2, updated_at=now() WHERE id=$1`, id, stravaUserID)
	return err
}
