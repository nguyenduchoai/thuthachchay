package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/logger"
)

// Worker tách 3 loại job:
//  1. settle: hàng đêm tính winner challenge đã hết hạn.
//  2. strava: poll/handle webhook backlog.
//  3. antifraud: tính fraud_score đêm + auto-suspend.
//
// MVP: chạy bằng ticker. Lên prod chuyển sang asynq (Redis-backed queue).
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	log.Info().Str("env", cfg.AppEnv).Msg("buocvang worker starting")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(3)

	go runLoop(ctx, &wg, "settle", 5*time.Minute, runSettle)
	go runLoop(ctx, &wg, "strava", 1*time.Minute, runStravaSync)
	go runLoop(ctx, &wg, "antifraud", 10*time.Minute, runAntiFraud)

	<-ctx.Done()
	log.Info().Msg("shutdown signal received, waiting for in-flight jobs")
	wg.Wait()
	log.Info().Msg("bye")
}

type jobFn func(ctx context.Context) error

func runLoop(ctx context.Context, wg *sync.WaitGroup, name string, interval time.Duration, fn jobFn) {
	defer wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			if err := fn(ctx); err != nil {
				log.Error().Err(err).Str("job", name).Msg("job failed")
				continue
			}
			log.Debug().Str("job", name).Dur("elapsed", time.Since(start)).Msg("job ok")
		}
	}
}

func runSettle(_ context.Context) error {
	// TODO: query challenges where end_date <= now() AND status='live' → settle
	return nil
}

func runStravaSync(_ context.Context) error {
	// TODO: drain Redis queue strava:pending → fetch activities → merge daily_steps
	return nil
}

func runAntiFraud(_ context.Context) error {
	// TODO: compute fraud_score per user, push to admin queue, auto-suspend if criteria met
	return nil
}
