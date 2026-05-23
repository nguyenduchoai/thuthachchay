package web

import (
	"encoding/csv"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type dashboardData struct {
	Users, Active, Suspended            int
	Challenges, OpenChallenges          int
	Vouchers, Redemptions               int
	TotalPoints, FraudOpen              int
}

type dsoData struct {
	From, To                                          time.Time
	Issued, Burned, NetDelta                          int
	Redemptions, ActiveUsers, NewUsers, ARPU          int
	BrandRows                                         []dsoBrandRow
	DailyRows                                         []dsoDailyRow
}

type dsoBrandRow struct {
	Brand string
	Count int
	Burn  int
}

type dsoDailyRow struct {
	Day            time.Time
	Issued, Burned int
}

// ─── DASHBOARD ───────────────────────────────────────────────────────────────

func dashboard(c *fiber.Ctx, d Deps) error {
	var data dashboardData
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM users`).Scan(&data.Users)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM users WHERE status='active'`).Scan(&data.Active)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM users WHERE status='suspended'`).Scan(&data.Suspended)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM challenges`).Scan(&data.Challenges)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM challenges WHERE status IN ('open','live')`).Scan(&data.OpenChallenges)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM vouchers`).Scan(&data.Vouchers)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM voucher_redemptions`).Scan(&data.Redemptions)
	_ = d.Pool.QueryRow(c.Context(), `SELECT COALESCE(SUM(delta_points),0)::int FROM ledger_entries WHERE delta_points > 0`).Scan(&data.TotalPoints)
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM step_events WHERE flagged=true`).Scan(&data.FraudOpen)

	c.Set("Content-Type", "text/html; charset=utf-8")
	return dashboardPage(c, data).Render(c.Context(), c.Response().BodyWriter())
}

// ─── DSO ─────────────────────────────────────────────────────────────────────

func reportsDSO(c *fiber.Ctx, d Deps) error {
	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")
	now := time.Now()
	from := now.AddDate(0, 0, -30)
	to := now
	if ts, err := time.Parse("2006-01-02", fromStr); err == nil {
		from = ts
	}
	if ts, err := time.Parse("2006-01-02", toStr); err == nil {
		to = ts.Add(24 * time.Hour)
	}

	dso := dsoData{From: from, To: to}
	_ = d.Pool.QueryRow(c.Context(),
		`SELECT COALESCE(SUM(delta_points),0)::int FROM ledger_entries WHERE created_at BETWEEN $1 AND $2 AND delta_points > 0`,
		from, to).Scan(&dso.Issued)
	_ = d.Pool.QueryRow(c.Context(),
		`SELECT COALESCE(-SUM(delta_points),0)::int FROM ledger_entries WHERE created_at BETWEEN $1 AND $2 AND delta_points < 0`,
		from, to).Scan(&dso.Burned)
	_ = d.Pool.QueryRow(c.Context(),
		`SELECT COALESCE(SUM(delta_points),0)::int FROM ledger_entries WHERE created_at BETWEEN $1 AND $2`,
		from, to).Scan(&dso.NetDelta)
	_ = d.Pool.QueryRow(c.Context(),
		`SELECT count(*) FROM voucher_redemptions WHERE redeemed_at BETWEEN $1 AND $2`,
		from, to).Scan(&dso.Redemptions)
	_ = d.Pool.QueryRow(c.Context(),
		`SELECT count(DISTINCT user_id) FROM ledger_entries WHERE created_at BETWEEN $1 AND $2`,
		from, to).Scan(&dso.ActiveUsers)
	_ = d.Pool.QueryRow(c.Context(),
		`SELECT count(*) FROM users WHERE created_at BETWEEN $1 AND $2`,
		from, to).Scan(&dso.NewUsers)

	if dso.ActiveUsers > 0 {
		dso.ARPU = dso.Issued / dso.ActiveUsers
	}

	bRows, _ := d.Pool.Query(c.Context(),
		`SELECT v.brand, count(*)::int, COALESCE(SUM(v.cost_points),0)::int
		 FROM voucher_redemptions vr JOIN vouchers v ON v.id=vr.voucher_id
		 WHERE vr.redeemed_at BETWEEN $1 AND $2
		 GROUP BY v.brand ORDER BY count(*) DESC LIMIT 20`, from, to)
	defer bRows.Close()
	for bRows.Next() {
		var r dsoBrandRow
		_ = bRows.Scan(&r.Brand, &r.Count, &r.Burn)
		dso.BrandRows = append(dso.BrandRows, r)
	}

	dRows, _ := d.Pool.Query(c.Context(), `
		SELECT date_trunc('day', created_at)::date AS day,
		       COALESCE(SUM(CASE WHEN delta_points>0 THEN delta_points END),0)::int AS issued,
		       COALESCE(-SUM(CASE WHEN delta_points<0 THEN delta_points END),0)::int AS burned
		FROM ledger_entries WHERE created_at BETWEEN $1 AND $2
		GROUP BY day ORDER BY day DESC LIMIT 60`, from, to)
	defer dRows.Close()
	for dRows.Next() {
		var r dsoDailyRow
		_ = dRows.Scan(&r.Day, &r.Issued, &r.Burned)
		dso.DailyRows = append(dso.DailyRows, r)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return dsoPage(c, dso).Render(c.Context(), c.Response().BodyWriter())
}

// ─── CSV exports ─────────────────────────────────────────────────────────────

func csvUsers(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(),
		`SELECT id::text, zalo_id, COALESCE(handle::text,''), COALESCE(display_name,''), status, fraud_score, created_at
		 FROM users ORDER BY created_at DESC`)
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
	rows, err := d.Pool.Query(c.Context(),
		`SELECT id::text, user_id::text, delta_points, reason, COALESCE(reference_id,''), idempotency_key, COALESCE(note,''), created_at
		 FROM ledger_entries ORDER BY created_at DESC LIMIT 50000`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="ledger.csv"`)
	w := csv.NewWriter(c.Response().BodyWriter())
	_ = w.Write([]string{"id", "user_id", "delta", "reason", "reference_id", "idempotency_key", "note", "created_at"})
	for rows.Next() {
		var id, uid, reason, ref, idem, note string
		var delta int
		var ts time.Time
		_ = rows.Scan(&id, &uid, &delta, &reason, &ref, &idem, &note, &ts)
		_ = w.Write([]string{id, uid, strconv.Itoa(delta), reason, ref, idem, note, ts.Format(time.RFC3339)})
	}
	w.Flush()
	return nil
}

func csvChallenges(c *fiber.Ctx, d Deps) error {
	rows, err := d.Pool.Query(c.Context(),
		`SELECT c.id::text, c.name, c.visibility, c.status, c.daily_steps_target, c.duration_days,
		        c.entry_points, c.prize_pool, c.start_date, c.end_date, c.created_at,
		        (SELECT count(*) FROM challenge_participants WHERE challenge_id=c.id)
		 FROM challenges c ORDER BY c.start_date DESC`)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="challenges.csv"`)
	w := csv.NewWriter(c.Response().BodyWriter())
	_ = w.Write([]string{"id", "name", "visibility", "status", "daily_target", "duration_days",
		"entry_points", "prize_pool", "start_date", "end_date", "created_at", "participants"})
	for rows.Next() {
		var id, name, vis, status string
		var daily, days, entry, pool, participants int
		var start, end, created time.Time
		_ = rows.Scan(&id, &name, &vis, &status, &daily, &days, &entry, &pool, &start, &end, &created, &participants)
		_ = w.Write([]string{
			id, name, vis, status,
			strconv.Itoa(daily), strconv.Itoa(days),
			strconv.Itoa(entry), strconv.Itoa(pool),
			start.Format("2006-01-02"), end.Format("2006-01-02"),
			created.Format(time.RFC3339), strconv.Itoa(participants),
		})
	}
	w.Flush()
	return nil
}
