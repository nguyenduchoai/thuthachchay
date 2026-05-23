// Package server wire dependencies (config → db, redis, store, services, handlers) và mount routes.
package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"github.com/buocvang/api/internal/antifraud"
	"github.com/buocvang/api/internal/auth"
	"github.com/buocvang/api/internal/challenges"
	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/db"
	"github.com/buocvang/api/internal/httpx"
	"github.com/buocvang/api/internal/leaderboard"
	"github.com/buocvang/api/internal/middleware"
	"github.com/buocvang/api/internal/notifications"
	"github.com/buocvang/api/internal/referrals"
	"github.com/buocvang/api/internal/steps"
	"github.com/buocvang/api/internal/store"
	"github.com/buocvang/api/internal/strava"
	"github.com/buocvang/api/internal/users"
	"github.com/buocvang/api/internal/vouchers"
	"github.com/buocvang/api/internal/wallet"
)

type Deps struct {
	Cfg    *config.Config
	Pool   *db.Pool
	Redis  *redis.Client
	Store  *store.Store
	JWT    *auth.JWTManager
	Zalo   *auth.ZaloClient
	LB     *leaderboard.Client
	AF     *antifraud.Checker
	Notif  *notifications.Client
	Strava *strava.Client
}

// Build dựng tất cả services + handlers + register routes lên Fiber app.
func Build(ctx context.Context, cfg *config.Config) (*fiber.App, *Deps, error) {
	pool, err := db.New(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	rdb, err := leaderboard.Dial(cfg.RedisURL)
	if err != nil {
		return nil, nil, err
	}
	st := store.New(pool)
	jm, err := auth.LoadJWTManager(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath, cfg.JWTAccessTTL)
	if err != nil {
		return nil, nil, err
	}
	zaloCli := auth.NewZaloClient(cfg.ZaloAppID, cfg.ZaloAppSecret, cfg.AppEnv == "dev")
	lb := leaderboard.New(rdb)
	af := antifraud.NewChecker(0, cfg.AntiFraudMinCadenceMs, cfg.AntiFraudMaxCadenceMs)
	notif := notifications.New(cfg)
	stravaCli := strava.New(cfg, st.StravaTokens)

	deps := &Deps{
		Cfg: cfg, Pool: pool, Redis: rdb, Store: st,
		JWT: jm, Zalo: zaloCli, LB: lb, AF: af, Notif: notif, Strava: stravaCli,
	}

	app := fiber.New(fiber.Config{
		AppName:      "buocvang-api",
		ErrorHandler: httpx.ErrorHandler,
	})
	httpx.RegisterMiddleware(app, cfg)

	authSvc := auth.NewService(st.Users, st.Sessions, jm, zaloCli, cfg.JWTRefreshTTL)
	authH := auth.NewHandler(authSvc)
	usersH := users.NewHandler(st.Users)
	chSvc := challenges.NewService(challenges.Deps{Pool: pool, Store: st, LB: lb})
	chH := challenges.NewHandler(chSvc, st, lb)
	stepsSvc := steps.NewService(pool, st, lb, af, cfg)
	stepsH := steps.NewHandler(stepsSvc)
	walletSvc := wallet.NewService(pool, st)
	walletH := wallet.NewHandler(walletSvc, st)
	vouH := vouchers.NewHandler(pool, st)
	refH := referrals.NewHandler(pool, st)
	stravaH := strava.NewHandler(stravaCli, cfg, stepsSvc)

	registerRoutes(app, deps, authH, usersH, chH, stepsH, walletH, vouH, refH, stravaH, notif)
	return app, deps, nil
}

func registerRoutes(app *fiber.App, d *Deps,
	authH *auth.Handler, usersH *users.Handler, chH *challenges.Handler,
	stepsH *steps.Handler, walletH *wallet.Handler, vouH *vouchers.Handler,
	refH *referrals.Handler, stravaH *strava.Handler, notif *notifications.Client,
) {
	_ = notif
	// Health
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/readyz", func(c *fiber.Ctx) error {
		if err := d.Pool.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"db": err.Error()})
		}
		if _, err := d.Redis.Ping(c.Context()).Result(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"redis": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"service": "buocvang-api", "version": "0.1.0"})
	})

	v1 := app.Group("/v1")

	// Auth (public)
	v1.Post("/auth/zalo", authH.LoginZalo)
	v1.Post("/auth/refresh", authH.Refresh)
	v1.Post("/auth/sign-out", authH.SignOut)

	// Username check (public — onboarding)
	v1.Post("/username/check", usersH.CheckHandle)

	// Authenticated routes
	jwtMw := middleware.RequireActiveJWT(d.JWT, d.Store.Users)
	priv := v1.Group("", jwtMw)

	priv.Get("/me", usersH.Me)
	priv.Patch("/me", usersH.Patch)
	priv.Post("/me/attribution", usersH.Attribution)
	priv.Get("/me/referral", refH.Me)

	priv.Get("/steps/today", stepsH.Today)
	priv.Post("/steps/ingest", stepsH.Ingest)
	priv.Get("/steps/me", stepsH.History)

	priv.Get("/challenges", chH.List)
	priv.Get("/challenges/:id", chH.Get)
	priv.Post("/challenges", chH.Create)
	priv.Post("/challenges/:id/join", chH.Join)
	priv.Get("/challenges/:id/leaderboard", chH.Leaderboard)

	priv.Get("/leaderboards/global", chH.GlobalLeaderboard)

	priv.Get("/wallet", walletH.Balance)
	priv.Get("/wallet/ledger", walletH.Ledger)

	priv.Get("/vouchers", vouH.List)
	priv.Get("/vouchers/mine", vouH.Mine)
	priv.Post("/vouchers/:id/redeem", vouH.Redeem)

	priv.Post("/referrals/track", refH.Track)

	priv.Get("/strava/oauth/url", stravaH.AuthURL)
	priv.Post("/strava/oauth/callback", stravaH.Callback)
	v1.Get("/strava/webhook", stravaH.WebhookVerify)
	v1.Post("/strava/webhook", stravaH.WebhookEvent)
}
