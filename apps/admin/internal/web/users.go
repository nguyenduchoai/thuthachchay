package web

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ─── data shapes (cho templ) ─────────────────────────────────────────────────

type userRow struct {
	ID, ZaloID, Handle, Name, Status string
	DailyGoal, FraudScore, Balance   int
	CreatedAt                        time.Time
}

type userDetail struct {
	ID, ZaloID, Handle, Name, Status string
	DailyGoal, FraudScore            int
	CreatedAt                        time.Time
}

type ledgerRow struct {
	Delta             int
	Reason, Ref, Note string
	CreatedAt         time.Time
}

// ─── handlers ────────────────────────────────────────────────────────────────

func listUsers(c *fiber.Ctx, d Deps) error {
	q := strings.TrimSpace(c.Query("q", ""))
	statusF := c.Query("status", "")
	limit, offset, pageN := paginate(c)

	args := []any{}
	where := "WHERE 1=1"
	if q != "" {
		args = append(args, "%"+q+"%")
		where += fmt.Sprintf(" AND (u.zalo_id ILIKE $%d OR u.handle::text ILIKE $%d OR u.display_name ILIKE $%d)",
			len(args), len(args), len(args))
	}
	if statusF != "" {
		args = append(args, statusF)
		where += fmt.Sprintf(" AND u.status=$%d", len(args))
	}

	var total int
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM users u `+where, args...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := d.Pool.Query(c.Context(), `
		SELECT u.id::text, u.zalo_id, COALESCE(u.handle::text,''), COALESCE(u.display_name,''),
		       u.daily_goal, u.status, u.fraud_score, u.created_at,
		       COALESCE((SELECT SUM(delta_points)::int FROM ledger_entries le WHERE le.user_id=u.id), 0)
		FROM users u `+where+` ORDER BY u.created_at DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var users []userRow
	for rows.Next() {
		var u userRow
		_ = rows.Scan(&u.ID, &u.ZaloID, &u.Handle, &u.Name, &u.DailyGoal, &u.Status, &u.FraudScore, &u.CreatedAt, &u.Balance)
		users = append(users, u)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return usersListPage(c, users, total, pageN, limit, q, statusF).Render(c.Context(), c.Response().BodyWriter())
}

func userDetailH(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	var u userDetail
	u.ID = id
	err := d.Pool.QueryRow(c.Context(),
		`SELECT zalo_id, COALESCE(handle::text,''), COALESCE(display_name,''), status, daily_goal, fraud_score, created_at FROM users WHERE id=$1`, id).
		Scan(&u.ZaloID, &u.Handle, &u.Name, &u.Status, &u.DailyGoal, &u.FraudScore, &u.CreatedAt)
	if err != nil {
		return c.Status(404).SendString("not found")
	}

	lrows, _ := d.Pool.Query(c.Context(),
		`SELECT delta_points, reason, COALESCE(reference_id,''), COALESCE(note,''), created_at FROM ledger_entries WHERE user_id=$1 ORDER BY created_at DESC LIMIT 30`, id)
	defer lrows.Close()
	var ledger []ledgerRow
	for lrows.Next() {
		var l ledgerRow
		_ = lrows.Scan(&l.Delta, &l.Reason, &l.Ref, &l.Note, &l.CreatedAt)
		ledger = append(ledger, l)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return userDetailPage(c, u, ledger).Render(c.Context(), c.Response().BodyWriter())
}

// updateUser implement spec §6 PATCH /admin/users/:id { status, fraud_score, note }.
func updateUser(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	status := c.FormValue("status")
	fraudScore, _ := strconv.Atoi(c.FormValue("fraud_score", "-1"))
	note := strings.TrimSpace(c.FormValue("note", ""))

	if status != "active" && status != "suspended" && status != "banned" {
		return c.Status(400).SendString("invalid status")
	}
	if fraudScore < 0 || fraudScore > 100 {
		return c.Status(400).SendString("fraud_score out of range")
	}

	var oldStatus string
	var oldFraud int
	_ = d.Pool.QueryRow(c.Context(),
		`SELECT status, fraud_score FROM users WHERE id=$1`, id).Scan(&oldStatus, &oldFraud)

	if _, err := d.Pool.Exec(c.Context(),
		`UPDATE users SET status=$1, fraud_score=$2, updated_at=now() WHERE id=$3`,
		status, fraudScore, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','user.update',$1,
		 jsonb_build_object('status_old',$2::text,'status_new',$3::text,'fraud_old',$4::int,'fraud_new',$5::int,'note',$6::text))`,
		id, oldStatus, status, oldFraud, fraudScore, note)
	return c.Redirect("/admin/users/"+id, fiber.StatusSeeOther)
}

func suspendUser(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(),
		`UPDATE users SET status='suspended', updated_at=now() WHERE id=$1`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target) VALUES('admin','user.suspend',$1)`, id)
	return c.Redirect("/admin/users", fiber.StatusSeeOther)
}

func activateUser(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(),
		`UPDATE users SET status='active', updated_at=now() WHERE id=$1`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target) VALUES('admin','user.activate',$1)`, id)
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
	if _, err := d.Pool.Exec(c.Context(),
		`INSERT INTO ledger_entries(user_id, delta_points, reason, idempotency_key, note) VALUES($1,$2,'admin_adjust',$3,$4)`,
		id, delta, idem, note); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','points.adjust',$1, jsonb_build_object('delta',$2::int,'note',$3::text))`,
		id, delta, note)
	return c.Redirect("/admin/users", fiber.StatusSeeOther)
}
