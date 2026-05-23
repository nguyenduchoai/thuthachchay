// Package auth xử lý Zalo Login, JWT issue/refresh, session lifecycle.
//
// Flow Zalo Login (xem PHAN_TICH_KIEN_TRUC.md §7.1):
//  1. Client gọi zmp-sdk getAccessToken() → access_token.
//  2. Client POST /v1/auth/zalo { access_token }.
//  3. Server gọi Zalo OpenAPI để đổi thành zalo_id + profile.
//  4. Upsert users; tạo session; trả về { access_jwt, refresh_token }.
//
// JWT format: RS256, claims { sub: user_id, ver: int, exp }.
// `ver` tăng lên khi user sign-out all để invalidate hết JWT cũ.
package auth

import "context"

type Service interface {
	LoginWithZalo(ctx context.Context, zaloAccessToken string) (*LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (*LoginResult, error)
	SignOut(ctx context.Context, refreshToken string) error
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int    // seconds
	UserID       string // uuid
}
