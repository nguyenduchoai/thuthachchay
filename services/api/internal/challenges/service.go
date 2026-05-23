package challenges

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
	"github.com/buocvang/api/internal/leaderboard"
	"github.com/buocvang/api/internal/store"
)

type Deps struct {
	Pool               *db.Pool
	Store              *store.Store
	LB                 *leaderboard.Client
	PublicCostPoints   int // phí tạo public, default 100
	SponsorRevSharePct int // 10
}

type service struct{ d Deps }

var ErrChallengeFull = errors.New("challenge is full")

func NewService(d Deps) Service {
	if d.PublicCostPoints == 0 {
		d.PublicCostPoints = 100
	}
	if d.SponsorRevSharePct == 0 {
		d.SponsorRevSharePct = 10
	}
	return &service{d: d}
}

func (s *service) List(ctx context.Context, status string, limit, offset int) ([]Challenge, error) {
	chs, err := s.d.Store.Challenges.List(ctx, status, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]Challenge, 0, len(chs))
	for i := range chs {
		out = append(out, toDomain(&chs[i]))
	}
	return out, nil
}

func (s *service) Get(ctx context.Context, id string) (*Challenge, error) {
	c, err := s.d.Store.Challenges.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toDomain(c)
	return &out, nil
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Challenge, error) {
	if req.Name == "" || req.DurationDays <= 0 || req.DailyStepsTarget <= 0 || req.StartDate.IsZero() {
		return nil, errors.New("invalid challenge payload")
	}
	if req.Visibility != "private" && req.Visibility != "public" {
		req.Visibility = "private"
	}
	var created *store.Challenge
	err := db.InTx(ctx, s.d.Pool, func(tx pgx.Tx) error {
		// Trừ phí host nếu public.
		if req.Visibility == "public" {
			// Trừ điểm host.
			if err := s.d.Store.Wallet.AppendEntry(ctx, tx, store.LedgerEntry{
				UserID: req.HostID, DeltaPoints: -s.d.PublicCostPoints, Reason: "create_public_challenge",
				IdempotencyKey: "create:" + req.HostID + ":" + req.Name + ":" + req.StartDate.Format("2006-01-02"),
			}); err != nil {
				return fmt.Errorf("debit host: %w", err)
			}
		}
		c, err := s.d.Store.Challenges.Create(ctx, tx, store.CreateChallengeInput{
			HostID: req.HostID, Visibility: req.Visibility, Name: req.Name,
			Description: req.Description, CoverURL: req.CoverURL,
			DailyStepsTarget: req.DailyStepsTarget, DurationDays: req.DurationDays,
			StartDate: req.StartDate, EntryPoints: req.EntryPoints,
			MaxParticipants: req.MaxParticipants,
		})
		if err != nil {
			return err
		}
		created = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := toDomain(created)
	return &out, nil
}

func (s *service) Join(ctx context.Context, challengeID, userID, idem string) error {
	if idem == "" {
		idem = "join:" + challengeID + ":" + userID
	}
	return db.InTx(ctx, s.d.Pool, func(tx pgx.Tx) error {
		ch, err := s.d.Store.Challenges.GetForUpdate(ctx, tx, challengeID)
		if err != nil {
			return err
		}
		if ch.Status != "open" && ch.Status != "live" {
			return errors.New("challenge not joinable")
		}
		exists, err := s.d.Store.Challenges.ParticipantExists(ctx, tx, challengeID, userID)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		if ch.MaxParticipants != nil && *ch.MaxParticipants > 0 {
			count, err := s.d.Store.Challenges.ParticipantCountTx(ctx, tx, challengeID)
			if err != nil {
				return err
			}
			if count >= *ch.MaxParticipants {
				return ErrChallengeFull
			}
		}
		if time.Now().After(ch.StartDate.AddDate(0, 0, 0)) && ch.Status == "open" {
			// vẫn cho join sau start_date cho đơn giản MVP
		}
		// 1. Tạo participant trước, rollback nếu debit/pool fail.
		inserted, err := s.d.Store.Challenges.AddParticipant(ctx, tx, challengeID, userID, ch.EntryPoints)
		if err != nil {
			return err
		}
		if !inserted {
			return nil
		}
		// 2. Trừ điểm entry.
		if ch.EntryPoints > 0 {
			if err := s.d.Store.Wallet.AppendEntry(ctx, tx, store.LedgerEntry{
				UserID: userID, DeltaPoints: -ch.EntryPoints, Reason: "challenge_join",
				ReferenceType: "challenge", ReferenceID: challengeID,
				IdempotencyKey: idem,
			}); err != nil {
				return err
			}
		}
		// 3. Tăng pool.
		if err := s.d.Store.Challenges.AddPrizePool(ctx, tx, challengeID, ch.EntryPoints); err != nil {
			return err
		}
		return nil
	})
}

func (s *service) Settle(ctx context.Context, id string) (*SettleResult, error) {
	ch, err := s.d.Store.Challenges.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if ch.Status == "settled" || ch.Status == "cancelled" {
		return &SettleResult{ChallengeID: id}, nil
	}
	parts, err := s.d.Store.Challenges.ListParticipants(ctx, id, 1000)
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		_ = s.d.Store.Challenges.SetStatus(ctx, id, "settled")
		return &SettleResult{ChallengeID: id, Winners: nil, Refunded: false}, nil
	}
	// Winners: đạt daily target trong đủ số ngày của challenge.
	winners, err := s.winnersByDailyGoal(ctx, ch)
	if err != nil {
		return nil, err
	}
	winnerSet := make(map[string]struct{}, len(winners))
	for _, w := range winners {
		winnerSet[w] = struct{}{}
	}
	if len(winners) == 0 {
		err = db.InTx(ctx, s.d.Pool, func(tx pgx.Tx) error {
			for _, p := range parts {
				if err := s.d.Store.Challenges.SetParticipantState(ctx, tx, id, p.UserID, "lost"); err != nil {
					return err
				}
			}
			return s.d.Store.Challenges.SetStatusTx(ctx, tx, id, "settled")
		})
		if err != nil {
			return nil, err
		}
		return &SettleResult{ChallengeID: id, PrizeShared: 0, Refunded: false}, nil
	}
	hostCut := 0
	if ch.HostID != nil && ch.Visibility == "public" {
		hostCut = ch.PrizePool * s.d.SponsorRevSharePct / 100
	}
	prize := ch.PrizePool - hostCut
	share := prize / len(winners)
	err = db.InTx(ctx, s.d.Pool, func(tx pgx.Tx) error {
		for _, w := range winners {
			if share > 0 {
				if err := s.d.Store.Wallet.AppendEntry(ctx, tx, store.LedgerEntry{
					UserID: w, DeltaPoints: share, Reason: "challenge_payout",
					ReferenceType: "challenge", ReferenceID: id,
					IdempotencyKey: "payout:" + id + ":" + w,
				}); err != nil && !errors.Is(err, store.ErrDuplicateIdempotencyKey) {
					return err
				}
			}
		}
		if hostCut > 0 && ch.HostID != nil {
			err := s.d.Store.Wallet.AppendEntry(ctx, tx, store.LedgerEntry{
				UserID: *ch.HostID, DeltaPoints: hostCut, Reason: "challenge_host_revshare",
				ReferenceType: "challenge", ReferenceID: id,
				IdempotencyKey: "hostshare:" + id,
			})
			if err != nil && !errors.Is(err, store.ErrDuplicateIdempotencyKey) {
				return err
			}
		}
		for _, p := range parts {
			state := "lost"
			if _, ok := winnerSet[p.UserID]; ok {
				state = "won"
			}
			if err := s.d.Store.Challenges.SetParticipantState(ctx, tx, id, p.UserID, state); err != nil {
				return err
			}
		}
		return s.d.Store.Challenges.SetStatusTx(ctx, tx, id, "settled")
	})
	if err != nil {
		return nil, err
	}
	return &SettleResult{ChallengeID: id, Winners: winners, SharePoints: share, PrizeShared: prize}, nil
}

func (s *service) winnersByDailyGoal(ctx context.Context, ch *store.Challenge) ([]string, error) {
	requiredDays := ch.DurationDays
	if requiredDays <= 0 {
		requiredDays = 1
	}
	rows, err := s.d.Pool.Query(ctx, `
		SELECT cp.user_id::text
		FROM challenge_participants cp
		WHERE cp.challenge_id=$1
		  AND cp.state='in'
		  AND (
		    SELECT count(*)
		    FROM daily_steps ds
		    WHERE ds.user_id=cp.user_id
		      AND ds.day BETWEEN $2 AND $3
		      AND ds.steps >= $4
		  ) >= $5
		ORDER BY cp.total_steps DESC`,
		ch.ID, ch.StartDate, ch.EndDate, ch.DailyStepsTarget, requiredDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	winners := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		winners = append(winners, userID)
	}
	return winners, rows.Err()
}

// helpers
func toDomain(c *store.Challenge) Challenge {
	return Challenge{
		ID: c.ID, HostID: c.HostID, Visibility: c.Visibility, Name: c.Name,
		Description: derefStr(c.Description), CoverURL: c.CoverURL,
		DailyStepsTarget: c.DailyStepsTarget, DurationDays: c.DurationDays,
		EntryPoints: c.EntryPoints, PrizePool: c.PrizePool,
		MaxParticipants: c.MaxParticipants, StartDate: c.StartDate, EndDate: c.EndDate,
		Status: c.Status,
	}
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
