// Package web mount routes admin web (Go HTML + HTMX).
//
// Truy cập DB trực tiếp qua pgxpool. Auth tạm thời Basic Auth env ADMIN_USER:ADMIN_PASS.
// Production sẽ chuyển sang OIDC Google Workspace.
package web

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
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
	admin.Get("/users/:id", func(c *fiber.Ctx) error { return userDetail(c, d) })
	admin.Post("/users/:id/suspend", func(c *fiber.Ctx) error { return suspendUser(c, d) })
	admin.Post("/users/:id/activate", func(c *fiber.Ctx) error { return activateUser(c, d) })
	admin.Post("/users/:id/adjust", func(c *fiber.Ctx) error { return adjustPoints(c, d) })

	admin.Get("/challenges", func(c *fiber.Ctx) error { return listChallenges(c, d) })
	admin.Post("/challenges/:id/cancel", func(c *fiber.Ctx) error { return cancelChallenge(c, d) })
	admin.Post("/challenges/:id/settle", func(c *fiber.Ctx) error { return triggerSettle(c, d) })

	admin.Get("/fraud-queue", func(c *fiber.Ctx) error { return listFraudQueue(c, d) })
	admin.Post("/fraud-queue/:event_id/decide", func(c *fiber.Ctx) error { return decideFraud(c, d) })

	admin.Get("/vouchers", func(c *fiber.Ctx) error { return listVouchers(c, d) })
	admin.Get("/vouchers/new", func(c *fiber.Ctx) error { return uploadForm(c) })
	admin.Post("/vouchers", func(c *fiber.Ctx) error { return uploadVoucher(c, d) })

	admin.Get("/reports/csv/users", func(c *fiber.Ctx) error { return csvUsers(c, d) })
	admin.Get("/reports/csv/ledger", func(c *fiber.Ctx) error { return csvLedger(c, d) })
	admin.Get("/reports/dashboard", func(c *fiber.Ctx) error { return dashboard(c, d) })
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

// ─── HTML helpers ─────────────────────────────────────────────────────────────

func page(title, body string) string {
	return `<!doctype html><html lang="vi"><head><meta charset="utf-8"><title>` + title + ` · Bước Vàng Admin</title>
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
<style>
body{font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,sans-serif;margin:0;background:#f5f5f7;color:#1c1c1e}
header.top{padding:14px 24px;background:#fff;border-bottom:1px solid #e5e5ea;display:flex;gap:14px;align-items:center}
header.top a{color:#ff9500;text-decoration:none;font-weight:600;font-size:14px}
header.top a:hover{text-decoration:underline}
main{padding:24px;max-width:1280px;margin:0 auto}
h1{font-size:22px;letter-spacing:-.4px;margin:0 0 16px}
h2{font-size:16px;margin-top:24px}
table{width:100%;border-collapse:collapse;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 6px rgba(0,0,0,.04)}
th,td{padding:10px 14px;text-align:left;border-bottom:1px solid #f0f0f3;font-size:13px}
th{background:#fafafc;font-weight:600;text-transform:uppercase;letter-spacing:.4px;font-size:11px;color:#6b6b70}
tr:last-child td{border-bottom:0}
.pill{display:inline-block;padding:2px 8px;border-radius:999px;font-size:11px;font-weight:600}
.pill.green{background:#dcfce7;color:#1e8a3e}
.pill.red{background:#fee2e2;color:#dc2626}
.pill.orange{background:#ffedd5;color:#c2410c}
.pill.gray{background:#e5e7eb;color:#374151}
.btn,button{padding:6px 12px;border-radius:8px;border:1px solid #d1d1d6;background:#fff;cursor:pointer;font-size:12px;font-family:inherit}
.btn.primary,button.primary{background:#ff9500;color:#fff;border:0}
.btn.danger,button.danger{background:#dc2626;color:#fff;border:0}
.btn.ghost{background:transparent}
.field{display:flex;flex-direction:column;gap:4px;margin-bottom:12px;max-width:360px}
.field label{font-size:11px;color:#6b6b70;text-transform:uppercase;letter-spacing:.4px}
.field input,.field textarea,.field select{padding:8px 10px;border:1px solid #d1d1d6;border-radius:8px;font-family:inherit;font-size:13px}
.row{display:flex;gap:8px;align-items:center}
.actions{display:flex;gap:6px}
.card{background:#fff;border-radius:12px;padding:16px;box-shadow:0 2px 6px rgba(0,0,0,.04);margin-bottom:16px}
.kpi{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:12px}
.kpi .tile{background:#fff;padding:14px;border-radius:12px;box-shadow:0 2px 6px rgba(0,0,0,.04)}
.kpi .tile .lab{font-size:11px;color:#6b6b70;text-transform:uppercase;letter-spacing:.4px}
.kpi .tile .v{font-size:22px;font-weight:700;margin-top:4px}
small.muted{color:#6b6b70}
form.inline{display:inline-flex;gap:6px;align-items:center;margin:0}
form.inline input[type=number]{width:80px;padding:4px 6px;border:1px solid #d1d1d6;border-radius:6px;font-size:12px}
</style>
</head><body>
<header class="top">
  <a href="/admin/reports/dashboard"><b>📊 Dashboard</b></a>
  <a href="/admin/users">Users</a>
  <a href="/admin/challenges">Challenges</a>
  <a href="/admin/fraud-queue">Fraud</a>
  <a href="/admin/vouchers">Vouchers</a>
  <a href="/admin/vouchers/new">+ Voucher</a>
  <a href="/admin/reports/csv/users" style="margin-left:auto">CSV Users</a>
  <a href="/admin/reports/csv/ledger">CSV Ledger</a>
</header>
<main>` + body + `</main></body></html>`
}

func index(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Bảng điều khiển", `<h1>Bước Vàng — Admin</h1>
<p>Chào mừng. Chọn mục ở header. <a href="/admin/reports/dashboard">→ Vào dashboard</a></p>`))
}

// ─── DASHBOARD ───────────────────────────────────────────────────────────────

func dashboard(c *fiber.Ctx, d Deps) error {
	var users, active, susp, challenges, openCh, vouchers, redem int
	var totalPoints int
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM users`).Scan(&users)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM users WHERE status='active'`).Scan(&active)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM users WHERE status='suspended'`).Scan(&susp)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM challenges`).Scan(&challenges)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM challenges WHERE status IN ('open','live')`).Scan(&openCh)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM vouchers`).Scan(&vouchers)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM voucher_redemptions`).Scan(&redem)
	_ = d.Pool.QueryRow(c.Context(), `SELECT COALESCE(SUM(delta_points),0)::int FROM ledger_entries WHERE delta_points > 0`).Scan(&totalPoints)

	body := fmt.Sprintf(`<h1>Dashboard</h1>
<div class="kpi">
  <div class="tile"><div class="lab">Users</div><div class="v">%d</div><small class="muted">active %d · suspended %d</small></div>
  <div class="tile"><div class="lab">Challenges</div><div class="v">%d</div><small class="muted">đang chạy %d</small></div>
  <div class="tile"><div class="lab">Vouchers</div><div class="v">%d</div><small class="muted">đã đổi %d</small></div>
  <div class="tile"><div class="lab">Points lưu hành</div><div class="v">%s đ</div></div>
</div>`, users, active, susp, challenges, openCh, vouchers, redem,
		formatInt(totalPoints))
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Dashboard", body))
}

// ─── USERS ───────────────────────────────────────────────────────────────────

func listUsers(c *fiber.Ctx, d Deps) error {
	q := c.Query("q", "")
	statusF := c.Query("status", "")
	args := []any{}
	where := "WHERE 1=1"
	if q != "" {
		args = append(args, "%"+q+"%")
		where += fmt.Sprintf(" AND (u.zalo_id ILIKE $%d OR u.handle::text ILIKE $%d OR u.display_name ILIKE $%d)", len(args), len(args), len(args))
	}
	if statusF != "" {
		args = append(args, statusF)
		where += fmt.Sprintf(" AND u.status=$%d", len(args))
	}
	rows, err := d.Pool.Query(c.Context(), `
		SELECT u.id::text, u.zalo_id, COALESCE(u.handle::text,''), COALESCE(u.display_name,''),
		       u.daily_goal, u.status, u.fraud_score, u.created_at,
		       COALESCE((SELECT SUM(delta_points)::int FROM ledger_entries le WHERE le.user_id=u.id), 0)
		FROM users u `+where+` ORDER BY u.created_at DESC LIMIT 200`, args...)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	var b strings.Builder
	fmt.Fprintf(&b, `<h1>Users</h1>
<form method="get" class="card row">
  <div class="field"><label>Tìm</label><input type="search" name="q" value="%s" placeholder="zalo_id, handle, tên"></div>
  <div class="field"><label>Trạng thái</label><select name="status">
    <option value="">Tất cả</option>
    <option value="active" %s>active</option>
    <option value="suspended" %s>suspended</option>
    <option value="banned" %s>banned</option>
  </select></div>
  <button class="primary" type="submit">Lọc</button>
</form>
<table><thead><tr><th>Handle</th><th>Zalo ID</th><th>Tên</th><th>Goal</th><th>Status</th><th>Fraud</th><th>Balance</th><th>Tham gia</th><th>Hành động</th></tr></thead><tbody>`,
		escape(q), sel(statusF, "active"), sel(statusF, "suspended"), sel(statusF, "banned"))
	for rows.Next() {
		var id, zalo, handle, name, status string
		var goal, fraud, bal int
		var created time.Time
		_ = rows.Scan(&id, &zalo, &handle, &name, &goal, &status, &fraud, &created, &bal)
		fmt.Fprintf(&b, `<tr><td><a href="/admin/users/%s">@%s</a></td><td><code>%s</code></td><td>%s</td><td>%d</td><td>%s</td><td>%d</td><td>%s đ</td><td>%s</td><td class="actions">`,
			id, escape(handle), escape(zalo), escape(name), goal, statusPill(status), fraud, formatInt(bal), created.Format("02/01/06"))
		if status == "active" {
			fmt.Fprintf(&b, `<form method="post" action="/admin/users/%s/suspend" class="inline"><button class="danger" onclick="return confirm('Suspend user này?')">Suspend</button></form>`, id)
		} else {
			fmt.Fprintf(&b, `<form method="post" action="/admin/users/%s/activate" class="inline"><button class="primary">Activate</button></form>`, id)
		}
		fmt.Fprintf(&b, `<form method="post" action="/admin/users/%s/adjust" class="inline"><input type="number" name="delta" placeholder="±delta" required><input type="hidden" name="note" value="manual adjust"><button>±Pt</button></form>`, id)
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Users", b.String()))
}

func userDetail(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	var zalo, handle, name, status string
	var goal, fraud int
	var created time.Time
	err := d.Pool.QueryRow(c.Context(), `SELECT zalo_id, COALESCE(handle::text,''), COALESCE(display_name,''), status, daily_goal, fraud_score, created_at FROM users WHERE id=$1`, id).
		Scan(&zalo, &handle, &name, &status, &goal, &fraud, &created)
	if err != nil {
		return c.Status(404).SendString("not found")
	}

	// Recent ledger
	lrows, _ := d.Pool.Query(c.Context(), `SELECT delta_points, reason, COALESCE(reference_id,''), created_at FROM ledger_entries WHERE user_id=$1 ORDER BY created_at DESC LIMIT 30`, id)
	defer lrows.Close()
	var ledger strings.Builder
	ledger.WriteString(`<table><thead><tr><th>Delta</th><th>Reason</th><th>Ref</th><th>Time</th></tr></thead><tbody>`)
	for lrows.Next() {
		var delta int
		var reason, ref string
		var t time.Time
		_ = lrows.Scan(&delta, &reason, &ref, &t)
		fmt.Fprintf(&ledger, `<tr><td>%+d</td><td>%s</td><td><code>%s</code></td><td>%s</td></tr>`, delta, reason, ref, t.Format("02/01 15:04"))
	}
	ledger.WriteString(`</tbody></table>`)

	body := fmt.Sprintf(`<h1>@%s <small class="muted">%s</small></h1>
<div class="card">
  <p>Name: <b>%s</b> · Status: %s · Fraud: %d · Goal: %d</p>
  <p>Tham gia: %s</p>
</div>
<h2>Ledger gần đây</h2>
%s`, escape(handle), escape(zalo), escape(name), statusPill(status), fraud, goal, created.Format("02/01/2006 15:04"), ledger.String())
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("User", body))
}

func suspendUser(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(), `UPDATE users SET status='suspended', updated_at=now() WHERE id=$1`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target) VALUES('admin','user.suspend',$1)`, id)
	return c.Redirect("/admin/users", fiber.StatusSeeOther)
}

func activateUser(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(), `UPDATE users SET status='active', updated_at=now() WHERE id=$1`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target) VALUES('admin','user.activate',$1)`, id)
	return c.Redirect("/admin/users", fiber.StatusSeeOther)
}

func adjustPoints(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	delta, _ := strconv.Atoi(c.FormValue("delta"))
	note := c.FormValue("note", "manual adjust")
	if delta == 0 {
		return c.Redirect("/admin/users", fiber.StatusSeeOther)
	}
	idem := fmt.Sprintf("admin-adjust:%s:%d:%d", id, delta, time.Now().UnixNano())
	if _, err := d.Pool.Exec(c.Context(), `INSERT INTO ledger_entries(user_id, delta_points, reason, idempotency_key, note) VALUES($1,$2,'admin_adjust',$3,$4)`, id, delta, idem, note); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','points.adjust',$1, jsonb_build_object('delta',$2))`, id, delta)
	return c.Redirect("/admin/users", fiber.StatusSeeOther)
}

// ─── CHALLENGES ──────────────────────────────────────────────────────────────

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
	b.WriteString(`<h1>Challenges</h1><table><thead><tr><th>Name</th><th>Vis</th><th>Status</th><th>Daily</th><th>Days</th><th>Entry</th><th>Pool</th><th>Players</th><th>Range</th><th>Actions</th></tr></thead><tbody>`)
	for rows.Next() {
		var id, name, vis, status string
		var daily, days, entry, pool, participants int
		var start, end time.Time
		_ = rows.Scan(&id, &name, &vis, &status, &daily, &days, &entry, &pool, &start, &end, &participants)
		actions := ""
		if status == "open" || status == "live" {
			actions += fmt.Sprintf(`<form method="post" action="/admin/challenges/%s/settle" class="inline"><button class="primary" onclick="return confirm('Settle ngay?')">Settle</button></form>`, id)
			actions += fmt.Sprintf(`<form method="post" action="/admin/challenges/%s/cancel" class="inline"><button class="danger" onclick="return confirm('Cancel?')">Cancel</button></form>`, id)
		}
		fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%s đ</td><td>%d</td><td>%s → %s</td><td class="actions">%s</td></tr>`,
			escape(name), vis, statusPill(status), daily, days, entry, formatInt(pool), participants, start.Format("02/01"), end.Format("02/01"), actions)
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Challenges", b.String()))
}

func cancelChallenge(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(), `UPDATE challenges SET status='cancelled' WHERE id=$1`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target) VALUES('admin','challenge.cancel',$1)`, id)
	return c.Redirect("/admin/challenges", fiber.StatusSeeOther)
}

// triggerSettle chỉ mark trạng thái về 'settling'; worker (cmd/worker) sẽ pick up và chia pool.
func triggerSettle(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(), `UPDATE challenges SET status='settling' WHERE id=$1 AND status IN ('open','live')`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target) VALUES('admin','challenge.settle_trigger',$1)`, id)
	return c.Redirect("/admin/challenges", fiber.StatusSeeOther)
}

// ─── FRAUD ───────────────────────────────────────────────────────────────────

func listFraudQueue(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `
		SELECT id, user_id::text, source, steps, started_at, COALESCE(flag_reason,''), client_nonce
		FROM step_events WHERE flagged=true ORDER BY id DESC LIMIT 200`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	var b strings.Builder
	b.WriteString(`<h1>Fraud Queue</h1><table><thead><tr><th>ID</th><th>User</th><th>Source</th><th>Steps</th><th>Time</th><th>Reason</th><th>Action</th></tr></thead><tbody>`)
	for rows.Next() {
		var id int64
		var uid, src, reason, nonce string
		var steps int
		var t time.Time
		_ = rows.Scan(&id, &uid, &src, &steps, &t, &reason, &nonce)
		actions := fmt.Sprintf(
			`<form method="post" action="/admin/fraud-queue/%d/decide" class="inline">`+
				`<input type="hidden" name="decision" value="approve">`+
				`<button class="primary">Approve</button></form>`+
				`<form method="post" action="/admin/fraud-queue/%d/decide" class="inline">`+
				`<input type="hidden" name="decision" value="reject">`+
				`<button class="danger">Reject + Suspend</button></form>`, id, id)
		fmt.Fprintf(&b, `<tr><td>%d</td><td><a href="/admin/users/%s">%s…</a></td><td>%s</td><td>%d</td><td>%s</td><td><span class="pill red">%s</span></td><td class="actions">%s</td></tr>`,
			id, uid, uid[:min(len(uid), 8)], src, steps, t.Format("02/01 15:04"), reason, actions)
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Fraud", b.String()))
}

func decideFraud(c *fiber.Ctx, d Deps) error {
	id := c.Params("event_id")
	decision := c.FormValue("decision")
	if decision != "approve" && decision != "reject" {
		return c.Status(400).SendString("bad decision")
	}
	// Tìm user_id liên quan
	var uid string
	if err := d.Pool.QueryRow(c.Context(), `SELECT user_id::text FROM step_events WHERE id=$1`, id).Scan(&uid); err != nil {
		return c.Status(404).SendString("event not found")
	}
	if decision == "approve" {
		_, _ = d.Pool.Exec(c.Context(), `UPDATE step_events SET flagged=false, flag_reason=NULL WHERE id=$1`, id)
		_, _ = d.Pool.Exec(c.Context(), `UPDATE users SET fraud_score=GREATEST(0, fraud_score-10) WHERE id=$1`, uid)
	} else {
		_, _ = d.Pool.Exec(c.Context(), `UPDATE users SET status='suspended', fraud_score=100 WHERE id=$1`, uid)
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','fraud.decide',$1, jsonb_build_object('decision',$2))`, id, decision)
	return c.Redirect("/admin/fraud-queue", fiber.StatusSeeOther)
}

// ─── VOUCHERS ────────────────────────────────────────────────────────────────

func listVouchers(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(), `SELECT id::text, brand, title, cost_points, stock, COALESCE(expires_at::text,'') FROM vouchers ORDER BY created_at DESC`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	var b strings.Builder
	b.WriteString(`<h1>Vouchers</h1><p><a href="/admin/vouchers/new" class="btn primary">+ Tải lên voucher mới</a></p>
<table><thead><tr><th>Brand</th><th>Title</th><th>Cost</th><th>Stock</th><th>Expires</th></tr></thead><tbody>`)
	for rows.Next() {
		var id, brand, title, exp string
		var cost, stock int
		_ = rows.Scan(&id, &brand, &title, &cost, &stock, &exp)
		stockPill := fmt.Sprintf(`<span class="pill green">%d</span>`, stock)
		if stock <= 10 {
			stockPill = fmt.Sprintf(`<span class="pill red">%d</span>`, stock)
		} else if stock <= 30 {
			stockPill = fmt.Sprintf(`<span class="pill orange">%d</span>`, stock)
		}
		fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td><td>%s đ</td><td>%s</td><td>%s</td></tr>`,
			escape(brand), escape(title), formatInt(cost), stockPill, exp)
	}
	b.WriteString(`</tbody></table>`)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Vouchers", b.String()))
}

func uploadForm(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page("Upload voucher", `<h1>Tải voucher mới</h1>
<form method="post" action="/admin/vouchers" enctype="multipart/form-data" class="card">
  <div class="field"><label>Brand</label><input name="brand" required maxlength="60" placeholder="Highlands Coffee"></div>
  <div class="field"><label>Title</label><input name="title" required maxlength="120" placeholder="Voucher 50.000đ"></div>
  <div class="field"><label>Cost (đ)</label><input name="cost" type="number" min="100" max="100000" required value="5000"></div>
  <div class="field"><label>Expires (YYYY-MM-DD, optional)</label><input name="expires" type="date"></div>
  <div class="field"><label>CSV codes (1 mã / dòng)</label><textarea name="codes_inline" rows="8" placeholder="HL-001&#10;HL-002&#10;..."></textarea></div>
  <div class="field"><label>Hoặc file CSV (cột "code")</label><input type="file" name="codes_file" accept=".csv,.txt"></div>
  <button type="submit" class="primary">Tạo voucher + Import codes</button>
</form>`))
}

func uploadVoucher(c *fiber.Ctx, d Deps) error {
	brand := strings.TrimSpace(c.FormValue("brand"))
	title := strings.TrimSpace(c.FormValue("title"))
	cost, _ := strconv.Atoi(c.FormValue("cost"))
	expStr := c.FormValue("expires")
	if brand == "" || title == "" || cost <= 0 {
		return c.Status(400).SendString("missing fields")
	}
	codes := parseCodes(c)
	if len(codes) == 0 {
		return c.Status(400).SendString("no codes provided")
	}

	// Insert voucher
	var voucherID string
	args := []any{brand, title, cost, len(codes)}
	sqlIns := `INSERT INTO vouchers(brand, title, cost_points, stock) VALUES($1,$2,$3,$4) RETURNING id::text`
	if expStr != "" {
		if t, err := time.Parse("2006-01-02", expStr); err == nil {
			args = []any{brand, title, cost, len(codes), t}
			sqlIns = `INSERT INTO vouchers(brand, title, cost_points, stock, expires_at) VALUES($1,$2,$3,$4,$5) RETURNING id::text`
		}
	}
	if err := d.Pool.QueryRow(c.Context(), sqlIns, args...).Scan(&voucherID); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	// Bulk insert codes
	tx, err := d.Pool.Begin(c.Context())
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	for _, code := range codes {
		if _, err := tx.Exec(c.Context(), `INSERT INTO voucher_codes(voucher_id, code) VALUES($1, $2) ON CONFLICT DO NOTHING`, voucherID, code); err != nil {
			_ = tx.Rollback(c.Context())
			return c.Status(500).SendString("insert code: " + err.Error())
		}
	}
	if err := tx.Commit(c.Context()); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(), `INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','voucher.upload',$1, jsonb_build_object('codes',$2))`, voucherID, len(codes))
	return c.Redirect("/admin/vouchers", fiber.StatusSeeOther)
}

func parseCodes(c *fiber.Ctx) []string {
	codes := map[string]struct{}{}
	if inline := c.FormValue("codes_inline"); inline != "" {
		for _, line := range strings.Split(inline, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				codes[line] = struct{}{}
			}
		}
	}
	if fh, err := c.FormFile("codes_file"); err == nil && fh != nil {
		if f, err := fh.Open(); err == nil {
			defer f.Close()
			data, _ := io.ReadAll(f)
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(strings.Trim(line, "\r"))
				if line == "code" || line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				// allow CSV "code,extra"
				if idx := strings.Index(line, ","); idx > 0 {
					line = strings.TrimSpace(line[:idx])
				}
				if line != "" {
					codes[line] = struct{}{}
				}
			}
		}
	}
	out := make([]string, 0, len(codes))
	for k := range codes {
		out = append(out, k)
	}
	return out
}

// ─── CSV exports ─────────────────────────────────────────────────────────────

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
	rows, err := d.Pool.Query(c.Context(), `SELECT id::text, user_id::text, delta_points, reason, COALESCE(reference_id,''), idempotency_key, created_at FROM ledger_entries ORDER BY created_at DESC LIMIT 50000`)
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

// ─── tiny helpers ────────────────────────────────────────────────────────────

func decodeB64(s string) (string, error) {
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

func statusPill(status string) string {
	cls := "gray"
	switch status {
	case "active", "open", "live":
		cls = "green"
	case "suspended", "banned", "cancelled":
		cls = "red"
	case "settling", "draft":
		cls = "orange"
	}
	return `<span class="pill ` + cls + `">` + status + `</span>`
}

func escape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func sel(current, target string) string {
	if current == target {
		return "selected"
	}
	return ""
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	s := strconv.Itoa(n)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}
