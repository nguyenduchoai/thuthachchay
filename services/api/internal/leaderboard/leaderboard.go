// Package leaderboard tính top step qua Redis ZSET.
// Key:
//   lb:global:30d           — top 30 ngày toàn cầu (auto refresh mỗi 5')
//   lb:challenge:{id}       — leaderboard 1 thử thách
package leaderboard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Entry struct {
	UserID string  `json:"user_id"`
	Steps  float64 `json:"steps"`
}

type Client struct {
	R *redis.Client
}

func New(r *redis.Client) *Client { return &Client{R: r} }

func GlobalKey() string                  { return "lb:global:30d" }
func ChallengeKey(id string) string      { return "lb:challenge:" + id }

// AddSteps tăng score cho user (ZADD GT NX kết hợp INCR).
func (c *Client) AddSteps(ctx context.Context, key, userID string, delta float64) error {
	return c.R.ZIncrBy(ctx, key, delta, userID).Err()
}

// Top trả top N theo score giảm dần.
func (c *Client) Top(ctx context.Context, key string, n int) ([]Entry, error) {
	if n <= 0 {
		n = 20
	}
	res, err := c.R.ZRevRangeWithScores(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(res))
	for _, z := range res {
		uid, _ := z.Member.(string)
		out = append(out, Entry{UserID: uid, Steps: z.Score})
	}
	return out, nil
}

// Rank trả về vị trí (1-indexed) và score của user. (0, 0) nếu chưa có.
func (c *Client) Rank(ctx context.Context, key, userID string) (int64, float64, error) {
	rank, err := c.R.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	score, _ := c.R.ZScore(ctx, key, userID).Result()
	return rank + 1, score, nil
}

// Reset xoá key (dùng cho cron rebuild hàng đêm).
func (c *Client) Reset(ctx context.Context, key string) error {
	return c.R.Del(ctx, key).Err()
}

// SetExpire chỉnh TTL.
func (c *Client) SetExpire(ctx context.Context, key string, ttlSec int) error {
	return c.R.Expire(ctx, key, time.Duration(ttlSec)*time.Second).Err()
}

// Dial helper.
func Dial(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return redis.NewClient(opts), nil
}
