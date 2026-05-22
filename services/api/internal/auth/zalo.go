package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ZaloClient gọi Zalo OpenAPI để verify access_token và lấy profile.
// Docs: https://developers.zalo.me/docs/api/social-api/tham-khao/lay-thong-tin-user
type ZaloClient struct {
	HTTP   *http.Client
	AppID  string
	Secret string
	// MockMode: dev — bỏ qua call Zalo API, trả về user id giả từ token.
	MockMode bool
}

type ZaloUser struct {
	ID       string
	Name     string
	Picture  string
}

func NewZaloClient(appID, secret string, mock bool) *ZaloClient {
	return &ZaloClient{
		HTTP:     &http.Client{Timeout: 8 * time.Second},
		AppID:    appID,
		Secret:   secret,
		MockMode: mock,
	}
}

func (c *ZaloClient) GetUser(ctx context.Context, accessToken string) (*ZaloUser, error) {
	if c.MockMode || accessToken == "" {
		// Dev mode: chấp nhận token "dev:<zalo_id>" để test nhanh.
		const devPrefix = "dev:"
		if len(accessToken) > len(devPrefix) && accessToken[:len(devPrefix)] == devPrefix {
			id := accessToken[len(devPrefix):]
			return &ZaloUser{ID: id, Name: "Dev " + id, Picture: ""}, nil
		}
		// Fallback completely fake user id để dev có thể chạy.
		return &ZaloUser{ID: "dev-anon-" + accessToken[:min(len(accessToken), 8)], Name: "Dev User"}, nil
	}
	u := "https://graph.zalo.me/v2.0/me?fields=id,name,picture&access_token=" + url.QueryEscape(accessToken)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zalo me: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zalo me status %d", resp.StatusCode)
	}
	var raw struct {
		Error   int    `json:"error"`
		Message string `json:"message"`
		ID      string `json:"id"`
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode zalo: %w", err)
	}
	if raw.Error != 0 {
		return nil, fmt.Errorf("zalo error %d: %s", raw.Error, raw.Message)
	}
	if raw.ID == "" {
		return nil, fmt.Errorf("zalo returned empty id")
	}
	return &ZaloUser{ID: raw.ID, Name: raw.Name, Picture: raw.Picture.Data.URL}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
