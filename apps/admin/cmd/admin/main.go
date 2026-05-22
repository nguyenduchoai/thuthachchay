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
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"

	"github.com/buocvang/admin/internal/web"
)

func main() {
	addr := os.Getenv("ADMIN_LISTEN_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	app := fiber.New(fiber.Config{
		AppName:               "buocvang-admin",
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          15 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Static("/static", "./static")

	web.RegisterRoutes(app)

	go func() {
		log.Info().Str("addr", addr).Msg("buocvang admin starting")
		if err := app.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("fiber listen")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown")
	}
}
