// Package notifications: gửi Zalo OA template message theo loại event.
package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/config"
)

type Template string

const (
	TplChallengeReminder Template = "challenge_reminder"
	TplRewardEarned      Template = "reward_earned"
	TplStreakAtRisk      Template = "streak_at_risk"
	TplWelcome           Template = "welcome"
)

type Client struct {
	HTTP   *http.Client
	Cfg    *config.Config
	NoOp   bool // dev mode → chỉ log
}

func New(cfg *config.Config) *Client {
	return &Client{
		HTTP: &http.Client{Timeout: 6 * time.Second},
		Cfg:  cfg,
		NoOp: cfg.AppEnv == "dev" || cfg.ZaloAppSecret == "",
	}
}

// Send gửi 1 template tới zaloUserID.
func (c *Client) Send(ctx context.Context, zaloUserID string, tpl Template, data map[string]string) error {
	if c.NoOp {
		log.Info().Str("zalo_user", zaloUserID).Str("tpl", string(tpl)).Interface("data", data).Msg("notif (noop)")
		return nil
	}
	body := map[string]any{
		"recipient":  map[string]string{"user_id": zaloUserID},
		"message":    map[string]any{"text": data["text"]},
		"template":   tpl,
	}
	buf, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://openapi.zalo.me/v3.0/oa/message", strings.NewReader(string(buf)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", c.Cfg.ZaloAppSecret)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("zalo notif: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("zalo notif status %d", resp.StatusCode)
	}
	return nil
}
