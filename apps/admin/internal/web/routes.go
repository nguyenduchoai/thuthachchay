// Package web mount routes admin web (Go templ + HTMX).
//
// MVP: render placeholder HTML; sẽ chuyển sang templ khi domain layer kết nối
// vào public API qua RPC nội bộ hoặc cùng package store.
package web

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/", index)
	app.Get("/healthz", healthz)

	admin := app.Group("/admin")
	admin.Get("/users", listUsers)
	admin.Patch("/users/:id", patchUser)
	admin.Post("/users/:id/adjust", adjustPoints)
	admin.Get("/challenges", listChallenges)
	admin.Post("/challenges/:id/cancel", cancelChallenge)
	admin.Post("/challenges/:id/settle", settleChallenge)
	admin.Get("/fraud-queue", listFraudQueue)
	admin.Post("/fraud-queue/:event_id/decide", decideFraud)
	admin.Get("/vouchers", listVouchers)
	admin.Post("/vouchers", uploadVouchers)
	admin.Patch("/vouchers/:id", patchVoucher)
	admin.Get("/reports/dso", reportDSO)
	admin.Get("/reports/csv/users", csvUsers)
	admin.Get("/reports/csv/challenges", csvChallenges)
}

func healthz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func index(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(`<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8" />
  <title>Bước Vàng — Admin</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 40px; color: #1c1c1e; }
    h1 { font-size: 28px; }
    code { background: #f4f4f5; padding: 2px 6px; border-radius: 4px; }
    ul li { margin: 6px 0; }
  </style>
</head>
<body>
  <h1>Bước Vàng — Admin Console</h1>
  <p>Skeleton. Sẽ chuyển sang <code>templ + HTMX + Tailwind</code> ở milestone tiếp theo.</p>
  <h2>Endpoints khả dụng</h2>
  <ul>
    <li><code>GET /admin/users</code></li>
    <li><code>GET /admin/challenges</code></li>
    <li><code>GET /admin/fraud-queue</code></li>
    <li><code>GET /admin/vouchers</code></li>
    <li><code>GET /admin/reports/dso</code></li>
  </ul>
</body>
</html>`)
}

func notImplemented(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   "not_implemented",
		"path":    c.Path(),
		"method":  c.Method(),
	})
}

func listUsers(c *fiber.Ctx) error        { return notImplemented(c) }
func patchUser(c *fiber.Ctx) error        { return notImplemented(c) }
func adjustPoints(c *fiber.Ctx) error     { return notImplemented(c) }
func listChallenges(c *fiber.Ctx) error   { return notImplemented(c) }
func cancelChallenge(c *fiber.Ctx) error  { return notImplemented(c) }
func settleChallenge(c *fiber.Ctx) error  { return notImplemented(c) }
func listFraudQueue(c *fiber.Ctx) error   { return notImplemented(c) }
func decideFraud(c *fiber.Ctx) error      { return notImplemented(c) }
func listVouchers(c *fiber.Ctx) error     { return notImplemented(c) }
func uploadVouchers(c *fiber.Ctx) error   { return notImplemented(c) }
func patchVoucher(c *fiber.Ctx) error     { return notImplemented(c) }
func reportDSO(c *fiber.Ctx) error        { return notImplemented(c) }
func csvUsers(c *fiber.Ctx) error         { return notImplemented(c) }
func csvChallenges(c *fiber.Ctx) error    { return notImplemented(c) }
