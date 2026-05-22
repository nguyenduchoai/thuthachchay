// Package strava xử lý OAuth + webhook + sync activity → daily_steps.
//
// Whitelist activity types: Run, Walk, VirtualRun.
// Validate: moving_time, distance, device_name not in blacklist.
// Rate-limit Strava: 100 req / 15 min, 1000 / day → ưu tiên webhook, không poll.
package strava

import "context"

type Service interface {
	OAuthURL(ctx context.Context, userID string) (url, state string, err error)
	OAuthCallback(ctx context.Context, code, state string) error
	HandleWebhook(ctx context.Context, payload []byte) error
	SyncActivity(ctx context.Context, userID, activityID string) error
}
