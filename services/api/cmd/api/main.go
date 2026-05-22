package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/httpx"
	"github.com/buocvang/api/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	logger.Init(cfg.LogLevel, cfg.AppEnv)
	log.Info().Str("env", cfg.AppEnv).Str("addr", cfg.HTTPListenAddr).Msg("buocvang api starting")

	app := fiber.New(fiber.Config{
		AppName:               "buocvang-api",
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		ErrorHandler:          httpx.ErrorHandler,
	})

	httpx.RegisterMiddleware(app, cfg)
	httpx.RegisterRoutes(app, cfg)

	go func() {
		if err := app.Listen(cfg.HTTPListenAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("fiber listen")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Info().Msg("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown")
	}
	log.Info().Msg("bye")

	// Tránh unused import nếu zerolog không được dùng trực tiếp.
	_ = zerolog.TraceLevel
}
