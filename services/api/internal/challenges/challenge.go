// Package challenges xử lý CRUD thử thách + join + settle.
//
// Settle (xem PHAN_TICH_KIEN_TRUC.md §7.3):
//  winners = participants where days_completed = duration_days
//  share = prize_pool / len(winners)
//  for each: ledger.credit(user, share, ref=challenge_id)
//  0 winners → pool về system_revenue.
package challenges

import (
	"context"
	"time"
)

type Service interface {
	List(ctx context.Context, status string, limit, offset int) ([]Challenge, error)
	Get(ctx context.Context, id string) (*Challenge, error)
	Create(ctx context.Context, req CreateRequest) (*Challenge, error)
	Join(ctx context.Context, challengeID, userID, idempotencyKey string) error
	Settle(ctx context.Context, id string) (*SettleResult, error)
}

type Challenge struct {
	ID                 string
	HostID             *string
	Visibility         string
	Name               string
	Description        string
	CoverURL           *string
	DailyStepsTarget   int
	DurationDays       int
	EntryPoints        int
	PrizePool          int
	MaxParticipants    *int
	StartDate          time.Time
	EndDate            time.Time
	Status             string
}

type CreateRequest struct {
	HostID             string
	Visibility         string
	Name               string
	Description        string
	CoverURL           string
	DailyStepsTarget   int
	DurationDays       int
	EntryPoints        int
	PrizePool          int
	MaxParticipants    int
	StartDate          time.Time
}

type SettleResult struct {
	ChallengeID string
	Winners     []string
	SharePoints int
	Refunded    bool
}
