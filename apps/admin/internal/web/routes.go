// Package web mount routes admin web (Go HTML + HTMX).
//
// Truy cập DB trực tiếp qua pgxpool. Auth tạm thời Basic Auth env ADMIN_USER:ADMIN_PASS.
// Production sẽ chuyển sang OIDC Google Workspace (xem STATUS.md "còn lại").
//
// File chỉ chứa entry points + middleware. Handlers nằm trong:
//   - layout.go    (page, i18n, pagination, helpers)
//   - users.go     (list/detail/update/suspend/activate/adjust)
//   - challenges.go (list/detail+settle preview/cancel/trigger)
//   - vouchers.go  (list/detail/upload/update/codes/disable)
//   - fraud.go     (queue/decide)
//   - audit.go     (viewer)
//   - reports.go   (dashboard/dso/csv-users/csv-ledger/csv-challenges)
package web

import (
	"context"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool *pgxpool.Pool
}

// RegisterRoutes giữ chữ ký cũ để main.go không break.
func RegisterRoutes(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/", index)
	app.Get("/lang", setLang)
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		mountNoDatabaseFallback(app)
		return
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		mountNoDatabaseFallback(app)
		return
	}
	RegisterWithDeps(app, Deps{Pool: pool})
}

// RegisterWithDeps mount full admin với dep injection.
func RegisterWithDeps(app *fiber.App, d Deps) {
	app.Use(basicAuth())
	admin := app.Group("/admin")

	// Users
	admin.Get("/users", func(c *fiber.Ctx) error { return listUsers(c, d) })
	admin.Get("/users/:id", func(c *fiber.Ctx) error { return userDetailH(c, d) })
	admin.Post("/users/:id/update", func(c *fiber.Ctx) error { return updateUser(c, d) })
	admin.Post("/users/:id/suspend", func(c *fiber.Ctx) error { return suspendUser(c, d) })
	admin.Post("/users/:id/activate", func(c *fiber.Ctx) error { return activateUser(c, d) })
	admin.Post("/users/:id/adjust", func(c *fiber.Ctx) error { return adjustPoints(c, d) })

	// Challenges
	admin.Get("/challenges", func(c *fiber.Ctx) error { return listChallenges(c, d) })
	admin.Get("/challenges/:id", func(c *fiber.Ctx) error { return challengeDetail(c, d) })
	admin.Post("/challenges/:id/cancel", func(c *fiber.Ctx) error { return cancelChallenge(c, d) })
	admin.Post("/challenges/:id/settle", func(c *fiber.Ctx) error { return triggerSettle(c, d) })

	// Fraud
	admin.Get("/fraud-queue", func(c *fiber.Ctx) error { return listFraudQueue(c, d) })
	admin.Post("/fraud-queue/:event_id/decide", func(c *fiber.Ctx) error { return decideFraud(c, d) })

	// Vouchers
	admin.Get("/vouchers", func(c *fiber.Ctx) error { return listVouchers(c, d) })
	admin.Get("/vouchers/new", func(c *fiber.Ctx) error { return uploadForm(c) })
	admin.Post("/vouchers", func(c *fiber.Ctx) error { return uploadVoucher(c, d) })
	admin.Get("/vouchers/:id", func(c *fiber.Ctx) error { return voucherDetail(c, d) })
	admin.Post("/vouchers/:id/update", func(c *fiber.Ctx) error { return updateVoucher(c, d) })
	admin.Post("/vouchers/:id/codes", func(c *fiber.Ctx) error { return addVoucherCodes(c, d) })
	admin.Post("/vouchers/:id/disable", func(c *fiber.Ctx) error { return disableVoucher(c, d) })

	// Audit
	admin.Get("/audit", func(c *fiber.Ctx) error { return listAudit(c, d) })

	// Reports
	admin.Get("/reports/dashboard", func(c *fiber.Ctx) error { return dashboard(c, d) })
	admin.Get("/reports/dso", func(c *fiber.Ctx) error { return reportsDSO(c, d) })
	admin.Get("/reports/csv/users", func(c *fiber.Ctx) error { return csvUsers(c, d) })
	admin.Get("/reports/csv/ledger", func(c *fiber.Ctx) error { return csvLedger(c, d) })
	admin.Get("/reports/csv/challenges", func(c *fiber.Ctx) error { return csvChallenges(c, d) })
}

func mountNoDatabaseFallback(app *fiber.App) {
	app.Get("/admin/*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "set DATABASE_URL to enable admin"})
	})
}

// ─── auth ────────────────────────────────────────────────────────────────────

func basicAuth() fiber.Handler {
	user := os.Getenv("ADMIN_USER")
	pass := os.Getenv("ADMIN_PASS")
	if user == "" {
		user = "admin"
	}
	if pass == "" {
		pass = "buocvang-dev"
	}
	return func(c *fiber.Ctx) error {
		if !strings.HasPrefix(c.Path(), "/admin") {
			return c.Next()
		}
		u, p, ok := basicAuthDecode(c.Get("Authorization"))
		if !ok || u != user || p != pass {
			c.Set("WWW-Authenticate", `Basic realm="buocvang admin"`)
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.Next()
	}
}

func basicAuthDecode(h string) (user, pass string, ok bool) {
	if !strings.HasPrefix(h, "Basic ") {
		return
	}
	dec, err := decodeB64(strings.TrimPrefix(h, "Basic "))
	if err != nil {
		return
	}
	parts := strings.SplitN(dec, ":", 2)
	if len(parts) != 2 {
		return
	}
	return parts[0], parts[1], true
}
