package httpx

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/config"
)

func RegisterMiddleware(app *fiber.App, cfg *config.Config) {
	app.Use(recover.New(recover.Config{EnableStackTrace: cfg.AppEnv == "dev"}))
	app.Use(requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// MVP: allow tất cả ở dev. Prod: allowlist từ config.
			return cfg.AppEnv == "dev"
		},
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Idempotency-Key,X-Client",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
		MaxAge:       int((10 * time.Minute).Seconds()),
	}))
	app.Use(accessLog())
}

func accessLog() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info().
			Str("rid", c.Get("X-Request-ID")).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("elapsed", time.Since(start)).
			Str("ip", c.IP()).
			Msg("http")
		return err
	}
}
