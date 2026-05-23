package auth

import (
	"context"
	"errors"
	"time"

	"github.com/buocvang/api/internal/store"
)

type service struct {
	users      *store.UsersStore
	sessions   *store.SessionsStore
	jwt        *JWTManager
	zalo       *ZaloClient
	refreshTTL time.Duration
}

var ErrUserDisabled = errors.New("user is not active")

// NewService dựng implementation chuẩn cho Service interface.
func NewService(users *store.UsersStore, sessions *store.SessionsStore, jwt *JWTManager, zalo *ZaloClient, refreshTTL time.Duration) Service {
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	return &service{
		users:      users,
		sessions:   sessions,
		jwt:        jwt,
		zalo:       zalo,
		refreshTTL: refreshTTL,
	}
}

// LoginInfo bổ sung cho UserAgent/IP để lưu session.
type LoginContext struct {
	UserAgent string
	IP        string
}

func ctxValue(ctx context.Context) LoginContext {
	if v, ok := ctx.Value(loginCtxKey{}).(LoginContext); ok {
		return v
	}
	return LoginContext{}
}

func WithLoginContext(ctx context.Context, lc LoginContext) context.Context {
	return context.WithValue(ctx, loginCtxKey{}, lc)
}

type loginCtxKey struct{}

func (s *service) LoginWithZalo(ctx context.Context, zaloAccessToken string) (*LoginResult, error) {
	if zaloAccessToken == "" {
		return nil, errors.New("missing zalo access_token")
	}
	zu, err := s.zalo.GetUser(ctx, zaloAccessToken)
	if err != nil {
		return nil, err
	}
	u, _, err := s.users.UpsertByZaloID(ctx, zu.ID, zu.Name, zu.Picture, "vi-VN")
	if err != nil {
		return nil, err
	}
	if u.Status != "active" {
		return nil, ErrUserDisabled
	}
	jwtStr, err := s.jwt.Issue(u.ID, 0)
	if err != nil {
		return nil, err
	}
	raw, hash, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}
	lc := ctxValue(ctx)
	expiresAt := time.Now().Add(s.refreshTTL)
	if _, err := s.sessions.Create(ctx, u.ID, hash, lc.UserAgent, lc.IP, expiresAt); err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken:  jwtStr,
		RefreshToken: raw,
		ExpiresIn:    s.jwt.AccessTTLSeconds(),
		UserID:       u.ID,
	}, nil
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	if refreshToken == "" {
		return nil, errors.New("missing refresh_token")
	}
	hash := HashRefreshToken(refreshToken)
	sess, err := s.sessions.FindByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if sess.RevokedAt != nil {
		return nil, errors.New("session revoked")
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil, errors.New("refresh expired")
	}
	u, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}
	if u.Status != "active" {
		return nil, ErrUserDisabled
	}
	jwtStr, err := s.jwt.Issue(sess.UserID, 0)
	if err != nil {
		return nil, err
	}
	// Rotate refresh token
	if err := s.sessions.Revoke(ctx, sess.ID); err != nil {
		return nil, err
	}
	raw, newHash, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}
	lc := ctxValue(ctx)
	if _, err := s.sessions.Create(ctx, sess.UserID, newHash, lc.UserAgent, lc.IP, time.Now().Add(s.refreshTTL)); err != nil {
		return nil, err
	}
	return &LoginResult{AccessToken: jwtStr, RefreshToken: raw, ExpiresIn: s.jwt.AccessTTLSeconds(), UserID: sess.UserID}, nil
}

func (s *service) SignOut(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	hash := HashRefreshToken(refreshToken)
	sess, err := s.sessions.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.sessions.Revoke(ctx, sess.ID)
}
