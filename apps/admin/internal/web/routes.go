// Package web mount routes admin web (Go HTML + HTMX).
//
// Truy cập DB trực tiếp qua pgxpool. Auth tạm thời Basic Auth env ADMIN_USER:ADMIN_PASS.
// Production sẽ chuyển sang OIDC Google Workspace.
package web

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool *pgxpool.Pool
}

// RegisterRoutes giữ chữ ký cũ để main.go không break, dùng env DATABASE_URL khi gọi RegisterWithDeps.
func RegisterRoutes(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/", index)
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		mountStub(app)
		return
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		mountStub(app)
		return
	}
	RegisterWithDeps(app, Deps{Pool: pool})
}

// RegisterWithDeps mount full admin với dep injection.
func RegisterWithDeps(app *fiber.App, d Deps) {
	app.Use(basicAuth())
	admin := app.Group("/admin")
	admin.Get("/users", func(c *fiber.Ctx) error { return listUsers(c, d) })
	admin.Patch("/users/:id", func(c *fiber.Ctx) error { return patchUser(c, d) })
	admin.Post("/users/:id/adjust", func(c *fiber.Ctx) error { return adjustPoints(c, d) })
	admin.Get("/challenges", func(c *fiber.Ctx) error { return listChallenges(c, d) })
	admin.Post("/challenges/:id/cancel", func(c *fiber.Ctx) error { return cancelChallenge(c, d) })
	admin.Get("/fraud-queue", func(c *fiber.Ctx) error { return listFraudQueue(c, d) })
	admin.Get("/vouchers", func(c *fiber.Ctx) error { return listVouchers(c, d) })
	admin.Get("/reports/csv/users", func(c *fiber.Ctx) error { return csvUsers(c, d) })
	admin.Get("/reports/csv/ledger", func(c *fiber.Ctx) error { return csvLedger(c, d) })
}

func mountStub(app *fiber.App) {
	app.Get("/admin/*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "set DATABASE_URL to enable admin"})
	})
}

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

// --- HTML helpers ---

func page(title, body string) string {
	return `<!doctype html><html lang="vi"><head><meta charset="utf-8"><title>` + title + ` · Bước Vàng Admin</title>
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
<style>
body{font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,sans-serif;margin:0;background:#f5f5f7;color:#1c1c1e}
header{padding:14px 24px;background:#fff;border-bottom:1px solid #e5e5ea;display:flex;gap:12px;align-items:center}
header a{color:#ff9500;text-decoration:none;font-weight:600}
main{padding:24px;max-width:1200px;margin:0 auto}
h1{font-size:22px;letter-spacing:-.4px;margin:0 0 16px}
table{width:100%;border-collapse:collapse;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 6px rgba(0,0,0,.04)}
th,td{padding:10px 14px;text-align:left;border-bottom:1px solid #f0f0f3;font-size:13px}
th{background:#fafafc;font-weight:600;text-transform:uppercase;letter-spacing:.4px;font-size:11px;color:#6b6b70}
.pill{display:inline-block;padding:2px 8px;border-radius:999px;font-size:11px;font-weight:600}
.pill.green{background:#dcfce7;color:#1e8a3e}
.pill.red{background:#fee2e2;color:#ef4444}
.pill.orange{background:#fff3e0;color:#ff9500}
button,.btn{padding:6px 12px;border-radius:8px;border:1px solid #d1d1d6;background:#fff;cursor:pointer;font-size:12px}
button.primary{background:#ff9500;color:#fff;border:0}
</style>
</head><body>
<header>
  <a href="/admin/users">Users</a>
  <a href="/admin/challenges">Challenges</a>
  <a href="/admin/fraud-queue">Fraud</a>
  <a href="/admin/vouchers">Vouchers</a>
  <a href="/admin/reports/csv/users">Export CSV</a>
</header>
<main>` + body + `</main></body></html>`
}

func index(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Bảng điều khiển", `<h1>Bước Vàng — Admin</h1>
<p>Chào mừng. Chọn mục ở header.</p>`))
}

// --- Users ---

func listUsers(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `
		SELECT u.id::text, u.zalo_id, COALESCE(u.handle::text,''), COALESCE(u.display_name,''),
		       u.daily_goal, u.status, u.fraud_score, u.created_at,
		       COALESCE((SELECT SUM(delta_points)::int FROM ledger_entries le WHERE le.user_id=u.id), 0)
		FROM users u ORDER BY u.created_at DESC LIMIT 200`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	var b strings.Builder
	b.WriteString(`<h1>Users</h1><table><thead><tr><th>Handle</th><th>Zalo ID</th><th>Tên</th><th>Goal</th><th>Status</th><th>Fraud</th><th>Balance</th><th>Tham gia</th></tr></thead><tbody>`)
	for rows.Next() {
		var id, zalo, handle, name, status string
		var goal, fraud, bal int
		var created time.Time
		_ = rows.Scan(&id, &zalo, &handle, &name, &goal, &status, &fraud, &created, &bal)
		statusPill := `<span class="pill green">` + status + `</span>`
		if status == "suspended" {
			statusPill = `<span class="pill red">` + status + `</span>`
		}
		fmt.Fprintf(&b, `<tr><td>@%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td>%d</td><td>%d đ</td><td>%s</td></tr>`,
			handle, zalo, name, goal, statusPill, fraud, bal, created.Format("02/01/06"))
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Users", b.String()))
}

func patchUser(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	status := c.FormValue("status")
	_, err := d.Pool.Exec(c.Context(), `UPDATE users SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','user.patch',$1,$2)`, id, fmt.Sprintf(`{"status":"%s"}`, status))
	return c.JSON(fiber.Map{"ok": true})
}

func adjustPoints(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	delta, _ := strconv.Atoi(c.FormValue("delta"))
	note := c.FormValue("note")
	idem := fmt.Sprintf("admin-adjust:%s:%d:%d", id, delta, time.Now().Unix())
	_, err := d.Pool.Exec(c.Context(), `INSERT INTO ledger_entries(user_id, delta_points, reason, idempotency_key, note) VALUES($1,$2,'admin_adjust',$3,$4)`,
		id, delta, idem, note)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "delta": delta})
}

// --- Challenges ---

func listChallenges(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `
		SELECT id::text, name, visibility, status, daily_steps_target, duration_days, entry_points, prize_pool, start_date, end_date,
		       (SELECT count(*) FROM challenge_participants WHERE challenge_id=challenges.id)
		FROM challenges ORDER BY start_date DESC LIMIT 200`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	var b strings.Builder
	b.WriteString(`<h1>Challenges</h1><table><thead><tr><th>Name</th><th>Visibility</th><th>Status</th><th>Daily</th><th>Days</th><th>Entry</th><th>Pool</th><th>Người chơi</th><th>Range</th></tr></thead><tbody>`)
	for rows.Next() {
		var id, name, vis, status string
		var daily, days, entry, pool, participants int
		var start, end time.Time
		_ = rows.Scan(&id, &name, &vis, &status, &daily, &days, &entry, &pool, &start, &end, &participants)
		pill := `<span class="pill green">` + status + `</span>`
		if status == "settling" {
			pill = `<span class="pill orange">` + status + `</span>`
		} else if status == "cancelled" {
			pill = `<span class="pill red">` + status + `</span>`
		}
		fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%d đ</td><td>%d</td><td>%s → %s</td></tr>`,
			name, vis, pill, daily, days, entry, pool, participants, start.Format("02/01"), end.Format("02/01"))
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Challenges", b.String()))
}

func cancelChallenge(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	_, err := d.Pool.Exec(c.Context(), `UPDATE challenges SET status='cancelled' WHERE id=$1`, id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

// --- Fraud queue ---

func listFraudQueue(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `
		SELECT id, user_id::text, source, steps, started_at, COALESCE(flag_reason,''), client_nonce
		FROM step_events WHERE flagged=true ORDER BY id DESC LIMIT 200`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	var b strings.Builder
	b.WriteString(`<h1>Fraud Queue</h1><table><thead><tr><th>ID</th><th>User</th><th>Source</th><th>Steps</th><th>Time</th><th>Reason</th><th>Nonce</th></tr></thead><tbody>`)
	for rows.Next() {
		var id int64
		var uid, src, reason, nonce string
		var steps int
		var t time.Time
		_ = rows.Scan(&id, &uid, &src, &steps, &t, &reason, &nonce)
		fmt.Fprintf(&b, `<tr><td>%d</td><td><code>%s</code></td><td>%s</td><td>%d</td><td>%s</td><td><span class="pill red">%s</span></td><td>%s</td></tr>`,
			id, uid[:8], src, steps, t.Format("02/01 15:04"), reason, nonce[:min(len(nonce), 12)])
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Fraud Queue", b.String()))
}

// --- Vouchers ---

func listVouchers(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `SELECT id::text, brand, title, cost_points, stock, COALESCE(expires_at::text,'') FROM vouchers ORDER BY created_at DESC`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	var b strings.Builder
	b.WriteString(`<h1>Vouchers</h1><table><thead><tr><th>Brand</th><th>Title</th><th>Cost</th><th>Stock</th><th>Expires</th></tr></thead><tbody>`)
	for rows.Next() {
		var id, brand, title, exp string
		var cost, stock int
		_ = rows.Scan(&id, &brand, &title, &cost, &stock, &exp)
		stockPill := `<span class="pill green">` + strconv.Itoa(stock) + `</span>`
		if stock <= 10 {
			stockPill = `<span class="pill red">` + strconv.Itoa(stock) + `</span>`
		}
		fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td><td>%d đ</td><td>%s</td><td>%s</td></tr>`, brand, title, cost, stockPill, exp)
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Vouchers", b.String()))
}

// --- CSV exports ---

func csvUsers(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `SELECT id::text, zalo_id, COALESCE(handle::text,''), COALESCE(display_name,''), status, fraud_score, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="users.csv"`)
	w := csv.NewWriter(c.Response().BodyWriter())
	_ = w.Write([]string{"id", "zalo_id", "handle", "name", "status", "fraud_score", "created_at"})
	for rows.Next() {
		var id, zid, handle, name, status string
		var fraud int
		var created time.Time
		_ = rows.Scan(&id, &zid, &handle, &name, &status, &fraud, &created)
		_ = w.Write([]string{id, zid, handle, name, status, strconv.Itoa(fraud), created.Format(time.RFC3339)})
	}
	w.Flush()
	return nil
}

func csvLedger(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `SELECT id::text, user_id::text, delta_points, reason, COALESCE(reference_id,''), idempotency_key, created_at FROM ledger_entries ORDER BY created_at DESC LIMIT 10000`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="ledger.csv"`)
	w := csv.NewWriter(c.Response().BodyWriter())
	_ = w.Write([]string{"id", "user_id", "delta", "reason", "reference_id", "idempotency_key", "created_at"})
	for rows.Next() {
		var id, uid, reason, ref, idem string
		var delta int
		var t time.Time
		_ = rows.Scan(&id, &uid, &delta, &reason, &ref, &idem, &t)
		_ = w.Write([]string{id, uid, strconv.Itoa(delta), reason, ref, idem, t.Format(time.RFC3339)})
	}
	w.Flush()
	return nil
}

func decodeB64(s string) (string, error) {
	// inline base64 decode to avoid import
	const tbl = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var idx [256]int
	for i := range idx {
		idx[i] = -1
	}
	for i, c := range tbl {
		idx[c] = i
	}
	clean := strings.TrimRight(s, "=")
	out := make([]byte, 0, len(clean)*3/4)
	val, bits := 0, 0
	for _, c := range clean {
		v := idx[byte(c)]
		if v < 0 {
			return "", fmt.Errorf("bad b64")
		}
		val = val<<6 | v
		bits += 6
		if bits >= 8 {
			bits -= 8
			out = append(out, byte(val>>bits))
			val &= (1 << bits) - 1
		}
	}
	return string(out), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
