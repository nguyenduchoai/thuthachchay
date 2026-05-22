// Worker chạy cron settle thử thách + refresh leaderboard + trả referral bonus.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/challenges"
	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/db"
	"github.com/buocvang/api/internal/leaderboard"
	"github.com/buocvang/api/internal/logger"
	"github.com/buocvang/api/internal/referrals"
	"github.com/buocvang/api/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	log.Info().Msg("buocvang worker starting")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.New(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("db")
	}
	rdb, err := leaderboard.Dial(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("redis")
	}
	st := store.New(pool)
	lb := leaderboard.New(rdb)
	chSvc := challenges.NewService(challenges.Deps{Pool: pool, Store: st, LB: lb})

	settleTick := time.NewTicker(5 * time.Minute)
	lbTick := time.NewTicker(1 * time.Hour)
	refTick := time.NewTicker(10 * time.Minute)
	defer settleTick.Stop()
	defer lbTick.Stop()
	defer refTick.Stop()

	runSettle(ctx, chSvc, st)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("worker bye")
			return
		case <-settleTick.C:
			runSettle(ctx, chSvc, st)
		case <-lbTick.C:
			runRefreshLeaderboard(ctx, st, lb)
		case <-refTick.C:
			runReferralBonuses(ctx, pool, st)
		}
	}
}

func runSettle(ctx context.Context, svc challenges.Service, st *store.Store) {
	list, err := st.Challenges.EligibleForSettle(ctx)
	if err != nil {
		log.Error().Err(err).Msg("eligible settle")
		return
	}
	for _, ch := range list {
		res, err := svc.Settle(ctx, ch.ID)
		if err != nil {
			log.Error().Err(err).Str("challenge", ch.ID).Msg("settle failed")
			continue
		}
		log.Info().Str("challenge", ch.ID).Int("winners", len(res.Winners)).Int("share", res.SharePoints).Msg("settled")
	}
}

func runRefreshLeaderboard(ctx context.Context, st *store.Store, lb *leaderboard.Client) {
	from := time.Now().AddDate(0, 0, -30)
	rows, err := st.Pool.Query(ctx, `SELECT user_id::text, COALESCE(SUM(steps),0)::int FROM daily_steps WHERE day >= $1 GROUP BY user_id`, from)
	if err != nil {
		log.Error().Err(err).Msg("rebuild lb")
		return
	}
	defer rows.Close()
	key := leaderboard.GlobalKey()
	_ = lb.Reset(ctx, key)
	count := 0
	for rows.Next() {
		var uid string
		var n int
		if err := rows.Scan(&uid, &n); err != nil {
			continue
		}
		_ = lb.AddSteps(ctx, key, uid, float64(n))
		count++
	}
	log.Info().Int("entries", count).Msg("leaderboard refreshed")
}

func runReferralBonuses(ctx context.Context, pool *db.Pool, st *store.Store) {
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT r.invitee_id::text
		FROM referrals r
		JOIN challenge_participants cp ON cp.user_id = r.invitee_id
		WHERE r.bonus_paid = false AND cp.state IN ('won','lost')
		LIMIT 100`)
	if err != nil {
		log.Error().Err(err).Msg("scan referrals")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var inviteeID string
		if err := rows.Scan(&inviteeID); err != nil {
			continue
		}
		if err := referrals.PayBonusOnFirstChallengeComplete(ctx, pool, st, inviteeID, 500); err != nil {
			log.Error().Err(err).Str("invitee", inviteeID).Msg("pay referral bonus")
		}
	}
}
