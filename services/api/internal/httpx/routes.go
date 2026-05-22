package httpx

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/config"
)

// RegisterRoutes mount tất cả route theo openapi.yaml.
// MVP: handler trả 501 Not Implemented, ngoại trừ healthz/readyz/version.
// Sẽ wire dần khi từng package internal/<domain> sẵn sàng.
func RegisterRoutes(app *fiber.App, cfg *config.Config) {
	app.Get("/healthz", healthz)
	app.Get("/readyz", readyz)
	app.Get("/version", version)

	v1 := app.Group("/v1")

	// Auth
	auth := v1.Group("/auth")
	auth.Post("/zalo", notImplemented)
	auth.Post("/refresh", notImplemented)
	auth.Post("/sign-out", notImplemented)

	// Profile
	me := v1.Group("/me")
	me.Get("", notImplemented)
	me.Patch("", notImplemented)
	me.Get("/today", notImplemented)
	me.Get("/history", notImplemented)
	me.Get("/referral", notImplemented)
	me.Get("/host-stats", notImplemented)
	me.Patch("/notifications", notImplemented)
	me.Post("/attribution", notImplemented)

	// Username
	v1.Post("/username/check", notImplemented)

	// Steps
	steps := v1.Group("/steps")
	steps.Post("/ingest", notImplemented)

	// Challenges
	challenges := v1.Group("/challenges")
	challenges.Get("", notImplemented)
	challenges.Post("", notImplemented)
	challenges.Get("/:id", notImplemented)
	challenges.Post("/:id/join", notImplemented)

	// Leaderboards
	v1.Get("/leaderboards/global", notImplemented)
	v1.Get("/leaderboards/challenge/:id", notImplemented)

	// Wallet & vouchers
	v1.Get("/wallet", notImplemented)
	v1.Get("/vouchers", notImplemented)
	v1.Get("/vouchers/mine", notImplemented)
	v1.Post("/vouchers/:id/redeem", notImplemented)

	// Transactions
	v1.Get("/transactions/:id", notImplemented)

	// Referral
	v1.Post("/referrals/track", notImplemented)

	// Strava
	strava := v1.Group("/strava")
	strava.Get("/oauth/url", notImplemented)
	strava.Post("/oauth/callback", notImplemented)
	strava.Post("/webhook", notImplemented) // POST từ Strava → handler validate verify_token
	strava.Get("/webhook", notImplemented)  // GET subscription verify

	// Upload (avatar/cover)
	v1.Post("/upload", notImplemented)
}

func healthz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// readyz: kiểm tra DB + Redis. MVP trả ok; sẽ wire khi store/redis ready.
func readyz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"checked": time.Now().UTC().Format(time.RFC3339),
	})
}

func version(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service": "buocvang-api",
		"version": "0.1.0",
		"commit":  "dev",
	})
}

func notImplemented(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "not_implemented",
			"message": "Handler chưa được implement. Xem openapi.yaml và lộ trình MVP.",
		},
		"path": c.Path(),
	})
}
