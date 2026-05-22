package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
)

type Session struct {
	ID                 string
	UserID             string
	RefreshTokenHash   string
	UserAgent          string
	IP                 string
	CreatedAt          time.Time
	ExpiresAt          time.Time
	RevokedAt          *time.Time
}

type SessionsStore struct{ pool *db.Pool }

func (s *SessionsStore) Create(ctx context.Context, userID, hash, ua, ip string, expiresAt time.Time) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip, expires_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`, userID, hash, ua, ip, expiresAt).Scan(&id)
	return id, err
}

func (s *SessionsStore) FindByHash(ctx context.Context, hash string) (*Session, error) {
	row := s.pool.QueryRow(ctx, `SELECT id, user_id, refresh_token_hash, user_agent, ip, created_at, expires_at, revoked_at
		FROM sessions WHERE refresh_token_hash=$1`, hash)
	var ss Session
	if err := row.Scan(&ss.ID, &ss.UserID, &ss.RefreshTokenHash, &ss.UserAgent, &ss.IP, &ss.CreatedAt, &ss.ExpiresAt, &ss.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ss, nil
}

func (s *SessionsStore) Revoke(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET revoked_at=now() WHERE id=$1`, id)
	return err
}

func (s *SessionsStore) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	return err
}
