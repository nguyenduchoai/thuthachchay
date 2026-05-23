package web

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type auditRow struct {
	ID                          int64
	Admin, Action, Target, Diff string
	CreatedAt                   time.Time
}

type auditFilter struct {
	Admin, Action, Target, From, To string
}

func listAudit(c *fiber.Ctx, d Deps) error {
	f := auditFilter{
		Admin:  strings.TrimSpace(c.Query("admin", "")),
		Action: strings.TrimSpace(c.Query("action", "")),
		Target: strings.TrimSpace(c.Query("target", "")),
		From:   strings.TrimSpace(c.Query("from", "")),
		To:     strings.TrimSpace(c.Query("to", "")),
	}
	limit, offset, pageN := paginate(c)

	args := []any{}
	where := "WHERE 1=1"
	if f.Admin != "" {
		args = append(args, f.Admin)
		where += fmt.Sprintf(" AND admin_id=$%d", len(args))
	}
	if f.Action != "" {
		args = append(args, f.Action+"%")
		where += fmt.Sprintf(" AND action ILIKE $%d", len(args))
	}
	if f.Target != "" {
		args = append(args, "%"+f.Target+"%")
		where += fmt.Sprintf(" AND target ILIKE $%d", len(args))
	}
	if f.From != "" {
		if ts, err := time.Parse("2006-01-02", f.From); err == nil {
			args = append(args, ts)
			where += fmt.Sprintf(" AND created_at >= $%d", len(args))
		}
	}
	if f.To != "" {
		if ts, err := time.Parse("2006-01-02", f.To); err == nil {
			args = append(args, ts.Add(24*time.Hour))
			where += fmt.Sprintf(" AND created_at < $%d", len(args))
		}
	}

	var total int
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM audit_log `+where, args...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := d.Pool.Query(c.Context(),
		`SELECT id, COALESCE(admin_id,''), action, COALESCE(target,''), COALESCE(diff::text,''), created_at
		 FROM audit_log `+where+` ORDER BY id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var entries []auditRow
	for rows.Next() {
		var r auditRow
		_ = rows.Scan(&r.ID, &r.Admin, &r.Action, &r.Target, &r.Diff, &r.CreatedAt)
		entries = append(entries, r)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return auditListPage(c, entries, total, pageN, limit, f).Render(c.Context(), c.Response().BodyWriter())
}

// auditTargetURL: nếu target là UUID, link sang detail tương ứng.
func auditTargetURL(action, target string) string {
	if !isUUID(target) {
		return ""
	}
	switch {
	case strings.HasPrefix(action, "user.") || strings.HasPrefix(action, "points."):
		return "/admin/users/" + target
	case strings.HasPrefix(action, "challenge."):
		return "/admin/challenges/" + target
	case strings.HasPrefix(action, "voucher."):
		return "/admin/vouchers/" + target
	}
	return ""
}
