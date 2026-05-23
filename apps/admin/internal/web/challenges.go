package web

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type challengeRow struct {
	ID, Name, Visibility, Status            string
	DailyTarget, Duration, Entry, Pool      int
	Participants                            int
	StartDate, EndDate                      time.Time
}

type challengeDetailData struct {
	ID, Name, Visibility, Status, Description string
	DailyTarget, Duration, Entry, Pool, MaxP  int
	StartDate, EndDate                        time.Time
	HostID                                    *string
}

type participantRow struct {
	UserID, Handle string
	JoinedAt       time.Time
	State          string
	TotalSteps     int64
	DaysOnGoal     int
	Winner         bool
}

type settlePreview struct {
	Players      int
	Winners      int
	HostShare    int
	PerWinner    int
	RequiredDays int
}

// ─── LIST ────────────────────────────────────────────────────────────────────

func listChallenges(c *fiber.Ctx, d Deps) error {
	statusF := c.Query("status", "")
	limit, offset, pageN := paginate(c)
	args := []any{}
	where := "WHERE 1=1"
	if statusF != "" {
		args = append(args, statusF)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}

	var total int
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM challenges `+where, args...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := d.Pool.Query(c.Context(), `
		SELECT id::text, name, visibility, status, daily_steps_target, duration_days, entry_points, prize_pool, start_date, end_date,
		       (SELECT count(*) FROM challenge_participants WHERE challenge_id=challenges.id)
		FROM challenges `+where+` ORDER BY start_date DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var list []challengeRow
	for rows.Next() {
		var r challengeRow
		_ = rows.Scan(&r.ID, &r.Name, &r.Visibility, &r.Status, &r.DailyTarget, &r.Duration,
			&r.Entry, &r.Pool, &r.StartDate, &r.EndDate, &r.Participants)
		list = append(list, r)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return challengesListPage(c, list, total, pageN, limit, statusF).Render(c.Context(), c.Response().BodyWriter())
}

// ─── DETAIL ──────────────────────────────────────────────────────────────────

func challengeDetail(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	ch := challengeDetailData{ID: id}
	err := d.Pool.QueryRow(c.Context(), `
		SELECT name, visibility, status, COALESCE(description,''),
		       daily_steps_target, duration_days, entry_points, prize_pool, COALESCE(max_participants,0),
		       start_date, end_date, host_id::text
		FROM challenges WHERE id=$1`, id).
		Scan(&ch.Name, &ch.Visibility, &ch.Status, &ch.Description,
			&ch.DailyTarget, &ch.Duration, &ch.Entry, &ch.Pool, &ch.MaxP,
			&ch.StartDate, &ch.EndDate, &ch.HostID)
	if err != nil {
		return c.Status(404).SendString("not found")
	}

	rows, err := d.Pool.Query(c.Context(), `
		SELECT cp.user_id::text, COALESCE(u.handle::text, u.zalo_id), cp.joined_at, cp.state, cp.total_steps,
		       (SELECT count(*) FROM daily_steps ds
		         WHERE ds.user_id=cp.user_id AND ds.day BETWEEN $2 AND $3 AND ds.steps >= $4)
		FROM challenge_participants cp
		JOIN users u ON u.id=cp.user_id
		WHERE cp.challenge_id=$1
		ORDER BY cp.total_steps DESC
		LIMIT 500`, id, ch.StartDate, ch.EndDate, ch.DailyTarget)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var parts []participantRow
	for rows.Next() {
		var p participantRow
		_ = rows.Scan(&p.UserID, &p.Handle, &p.JoinedAt, &p.State, &p.TotalSteps, &p.DaysOnGoal)
		parts = append(parts, p)
	}

	requiredDays := ch.Duration
	if requiredDays <= 0 {
		requiredDays = 1
	}
	winners := 0
	for i := range parts {
		if parts[i].DaysOnGoal >= requiredDays && parts[i].State == "in" {
			parts[i].Winner = true
			winners++
		}
	}
	prev := settlePreview{
		Players:      len(parts),
		Winners:      winners,
		RequiredDays: requiredDays,
	}
	playerPool := ch.Pool
	if ch.Visibility == "public" && ch.HostID != nil {
		prev.HostShare = ch.Pool / 10
		playerPool = ch.Pool - prev.HostShare
	}
	if winners > 0 {
		prev.PerWinner = playerPool / winners
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return challengeDetailPage(c, ch, parts, prev).Render(c.Context(), c.Response().BodyWriter())
}

// ─── ACTIONS ─────────────────────────────────────────────────────────────────

func cancelChallenge(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(), `UPDATE challenges SET status='cancelled' WHERE id=$1`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target) VALUES('admin','challenge.cancel',$1)`, id)
	return c.Redirect("/admin/challenges", fiber.StatusSeeOther)
}

func triggerSettle(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	if _, err := d.Pool.Exec(c.Context(),
		`UPDATE challenges SET status='settling' WHERE id=$1 AND status IN ('open','live')`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target) VALUES('admin','challenge.settle_trigger',$1)`, id)
	return c.Redirect("/admin/challenges/"+id, fiber.StatusSeeOther)
}
