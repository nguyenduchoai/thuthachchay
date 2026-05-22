// Bước Vàng public API.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/logger"
	"github.com/buocvang/api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	log.Info().Str("env", cfg.AppEnv).Str("addr", cfg.HTTPListenAddr).Msg("buocvang api starting")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, _, err := server.Build(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("build server")
	}

	go func() {
		if err := app.Listen(cfg.HTTPListenAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("fiber listen")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received")
	shCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown")
	}
	log.Info().Msg("bye")
}
