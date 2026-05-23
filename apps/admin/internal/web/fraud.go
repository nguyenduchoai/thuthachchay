package web

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type fraudRow struct {
	ID         int64
	UserID     string
	Source     string
	Steps      int
	StartedAt  time.Time
	FlagReason string
}

func listFraudQueue(c *fiber.Ctx, d Deps) error {
	limit, offset, pageN := paginate(c)

	var total int
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM step_events WHERE flagged=true`).Scan(&total)

	rows, err := d.Pool.Query(c.Context(), `
		SELECT id, user_id::text, source, steps, started_at, COALESCE(flag_reason,'')
		FROM step_events WHERE flagged=true ORDER BY id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var events []fraudRow
	for rows.Next() {
		var e fraudRow
		_ = rows.Scan(&e.ID, &e.UserID, &e.Source, &e.Steps, &e.StartedAt, &e.FlagReason)
		events = append(events, e)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return fraudListPage(c, events, total, pageN, limit).Render(c.Context(), c.Response().BodyWriter())
}

func decideFraud(c *fiber.Ctx, d Deps) error {
	id := c.Params("event_id")
	decision := c.FormValue("decision")
	if decision != "approve" && decision != "reject" {
		return c.Status(400).SendString("bad decision")
	}
	var uid string
	if err := d.Pool.QueryRow(c.Context(),
		`SELECT user_id::text FROM step_events WHERE id=$1`, id).Scan(&uid); err != nil {
		return c.Status(404).SendString("event not found")
	}
	if decision == "approve" {
		_, _ = d.Pool.Exec(c.Context(), `UPDATE step_events SET flagged=false, flag_reason=NULL WHERE id=$1`, id)
		_, _ = d.Pool.Exec(c.Context(), `UPDATE users SET fraud_score=GREATEST(0, fraud_score-10) WHERE id=$1`, uid)
	} else {
		_, _ = d.Pool.Exec(c.Context(), `UPDATE users SET status='suspended', fraud_score=100 WHERE id=$1`, uid)
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','fraud.decide',$1, jsonb_build_object('decision',$2::text,'user_id',$3::text))`,
		id, decision, uid)
	return c.Redirect("/admin/fraud-queue", fiber.StatusSeeOther)
}
