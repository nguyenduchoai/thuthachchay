// Package strava: OAuth code exchange + webhook verify + activity fetch.
// Docs: https://developers.strava.com/docs/authentication/ + webhooks.
package strava

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/store"
)

type Client struct {
	HTTP   *http.Client
	Cfg    *config.Config
	Tokens *store.StravaTokensStore
}

func New(cfg *config.Config, tokens *store.StravaTokensStore) *Client {
	return &Client{HTTP: &http.Client{Timeout: 10 * time.Second}, Cfg: cfg, Tokens: tokens}
}

// AuthorizeURL trả về URL Strava OAuth redirect.
func (c *Client) AuthorizeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", c.Cfg.StravaClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", c.Cfg.StravaRedirectURI)
	q.Set("approval_prompt", "auto")
	q.Set("scope", "activity:read")
	q.Set("state", state)
	return "https://www.strava.com/oauth/authorize?" + q.Encode()
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Athlete      struct {
		ID int64 `json:"id"`
	} `json:"athlete"`
	Scope string `json:"scope"`
}

// ExchangeCode đổi authorization code lấy token và lưu vào DB cho userID.
func (c *Client) ExchangeCode(ctx context.Context, userID, code string) error {
	if code == "" {
		return errors.New("missing code")
	}
	body := url.Values{}
	body.Set("client_id", c.Cfg.StravaClientID)
	body.Set("client_secret", c.Cfg.StravaClientSecret)
	body.Set("code", code)
	body.Set("grant_type", "authorization_code")
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.strava.com/oauth/token", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("strava exchange: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("strava exchange status %d", resp.StatusCode)
	}
	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return fmt.Errorf("decode strava token: %w", err)
	}
	return c.Tokens.Upsert(ctx, store.StravaToken{
		UserID: userID, AccessToken: tr.AccessToken, RefreshToken: tr.RefreshToken,
		ExpiresAt: time.Unix(tr.ExpiresAt, 0), Scope: tr.Scope,
		AthleteID: fmt.Sprintf("%d", tr.Athlete.ID),
	})
}

// RefreshIfNeeded refresh token nếu sắp hết hạn.
func (c *Client) RefreshIfNeeded(ctx context.Context, t *store.StravaToken) (*store.StravaToken, error) {
	if time.Until(t.ExpiresAt) > 5*time.Minute {
		return t, nil
	}
	body := url.Values{}
	body.Set("client_id", c.Cfg.StravaClientID)
	body.Set("client_secret", c.Cfg.StravaClientSecret)
	body.Set("refresh_token", t.RefreshToken)
	body.Set("grant_type", "refresh_token")
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.strava.com/oauth/token", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava refresh status %d", resp.StatusCode)
	}
	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	t.AccessToken = tr.AccessToken
	t.RefreshToken = tr.RefreshToken
	t.ExpiresAt = time.Unix(tr.ExpiresAt, 0)
	if err := c.Tokens.Upsert(ctx, *t); err != nil {
		return nil, err
	}
	return t, nil
}

type Activity struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	StartDate  string `json:"start_date"`
	Distance   float64 `json:"distance"`
	MovingTime int    `json:"moving_time"`
}

// GetActivity lấy chi tiết activity. (steps không có trực tiếp; ước tính từ moving_time*cadence hoặc distance/stride.)
func (c *Client) GetActivity(ctx context.Context, accessToken string, activityID int64) (*Activity, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://www.strava.com/api/v3/activities/%d", activityID), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava activity status %d", resp.StatusCode)
	}
	var a Activity
	if err := json.NewDecoder(resp.Body).Decode(&a); err != nil {
		return nil, err
	}
	return &a, nil
}

// EstimateSteps ước tính bước từ distance và type. Đi bộ ~0.75m/bước, chạy ~1m/bước.
func EstimateSteps(a *Activity) int {
	if a.Distance <= 0 {
		return 0
	}
	stride := 0.75
	switch a.Type {
	case "Run", "VirtualRun", "TrailRun":
		stride = 1.0
	case "Walk", "Hike":
		stride = 0.75
	default:
		return 0 // bỏ qua type không liên quan
	}
	return int(a.Distance / stride)
}
