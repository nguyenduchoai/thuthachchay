package web

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type voucherRow struct {
	ID, Brand, Title, Expires string
	Cost, Stock               int
}

type voucherDetailData struct {
	ID, Brand, Title, Expires string
	Cost, Stock               int
	CreatedAt                 time.Time
	CodesTotal, CodesUsed     int
}

type voucherCodeRow struct {
	Code   string
	UsedBy string
	UsedAt *time.Time
}

type voucherRedemptionRow struct {
	UserID, Handle, Code string
	RedeemedAt           time.Time
}

// ─── LIST ────────────────────────────────────────────────────────────────────

func listVouchers(c *fiber.Ctx, d Deps) error {
	limit, offset, pageN := paginate(c)

	var total int
	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*) FROM vouchers`).Scan(&total)

	rows, err := d.Pool.Query(c.Context(),
		`SELECT id::text, brand, title, cost_points, stock, COALESCE(expires_at::text,'') FROM vouchers ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var list []voucherRow
	for rows.Next() {
		var v voucherRow
		_ = rows.Scan(&v.ID, &v.Brand, &v.Title, &v.Cost, &v.Stock, &v.Expires)
		list = append(list, v)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return vouchersListPage(c, list, total, pageN, limit).Render(c.Context(), c.Response().BodyWriter())
}

// ─── UPLOAD ──────────────────────────────────────────────────────────────────

func uploadForm(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return voucherUploadPage(c).Render(c.Context(), c.Response().BodyWriter())
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

	var voucherID string
	args := []any{brand, title, cost, len(codes)}
	sqlIns := `INSERT INTO vouchers(brand, title, cost_points, stock) VALUES($1,$2,$3,$4) RETURNING id::text`
	if expStr != "" {
		if ts, err := time.Parse("2006-01-02", expStr); err == nil {
			args = []any{brand, title, cost, len(codes), ts}
			sqlIns = `INSERT INTO vouchers(brand, title, cost_points, stock, expires_at) VALUES($1,$2,$3,$4,$5) RETURNING id::text`
		}
	}
	if err := d.Pool.QueryRow(c.Context(), sqlIns, args...).Scan(&voucherID); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	inserted, err := insertCodes(c, d, voucherID, codes)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if inserted != len(codes) {
		_, _ = d.Pool.Exec(c.Context(), `UPDATE vouchers SET stock=$1 WHERE id=$2`, inserted, voucherID)
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','voucher.upload',$1, jsonb_build_object('codes',$2::int))`,
		voucherID, inserted)
	return c.Redirect("/admin/vouchers/"+voucherID, fiber.StatusSeeOther)
}

// ─── DETAIL ──────────────────────────────────────────────────────────────────

func voucherDetail(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	v := voucherDetailData{ID: id}
	err := d.Pool.QueryRow(c.Context(),
		`SELECT brand, title, cost_points, stock, COALESCE(expires_at::text,''), created_at FROM vouchers WHERE id=$1`, id).
		Scan(&v.Brand, &v.Title, &v.Cost, &v.Stock, &v.Expires, &v.CreatedAt)
	if err != nil {
		return c.Status(404).SendString("not found")
	}

	_ = d.Pool.QueryRow(c.Context(), `SELECT count(*), count(used_at) FROM voucher_codes WHERE voucher_id=$1`, id).
		Scan(&v.CodesTotal, &v.CodesUsed)

	cRows, _ := d.Pool.Query(c.Context(),
		`SELECT code, COALESCE(used_by_user_id::text,''), used_at FROM voucher_codes WHERE voucher_id=$1 ORDER BY id DESC LIMIT 100`, id)
	defer cRows.Close()
	var codes []voucherCodeRow
	for cRows.Next() {
		var rc voucherCodeRow
		_ = cRows.Scan(&rc.Code, &rc.UsedBy, &rc.UsedAt)
		codes = append(codes, rc)
	}

	rRows, _ := d.Pool.Query(c.Context(),
		`SELECT vr.user_id::text, COALESCE(u.handle::text, u.zalo_id), vr.code, vr.redeemed_at
		 FROM voucher_redemptions vr JOIN users u ON u.id=vr.user_id
		 WHERE vr.voucher_id=$1 ORDER BY vr.redeemed_at DESC LIMIT 100`, id)
	defer rRows.Close()
	var redems []voucherRedemptionRow
	for rRows.Next() {
		var r voucherRedemptionRow
		_ = rRows.Scan(&r.UserID, &r.Handle, &r.Code, &r.RedeemedAt)
		redems = append(redems, r)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return voucherDetailPage(c, v, codes, redems).Render(c.Context(), c.Response().BodyWriter())
}

// ─── UPDATE / CODES / DISABLE ────────────────────────────────────────────────

func updateVoucher(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	brand := strings.TrimSpace(c.FormValue("brand"))
	title := strings.TrimSpace(c.FormValue("title"))
	cost, _ := strconv.Atoi(c.FormValue("cost_points"))
	stock, _ := strconv.Atoi(c.FormValue("stock"))
	expStr := strings.TrimSpace(c.FormValue("expires"))

	if brand == "" || title == "" || cost <= 0 || stock < 0 {
		return c.Status(400).SendString("invalid fields")
	}

	if expStr != "" {
		ts, err := time.Parse("2006-01-02", expStr)
		if err != nil {
			return c.Status(400).SendString("bad expires date")
		}
		_, err = d.Pool.Exec(c.Context(),
			`UPDATE vouchers SET brand=$1, title=$2, cost_points=$3, stock=$4, expires_at=$5 WHERE id=$6`,
			brand, title, cost, stock, ts, id)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
	} else {
		_, err := d.Pool.Exec(c.Context(),
			`UPDATE vouchers SET brand=$1, title=$2, cost_points=$3, stock=$4, expires_at=NULL WHERE id=$5`,
			brand, title, cost, stock, id)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','voucher.update',$1, jsonb_build_object('cost',$2::int,'stock',$3::int,'expires',$4::text))`,
		id, cost, stock, expStr)
	return c.Redirect("/admin/vouchers/"+id, fiber.StatusSeeOther)
}

func addVoucherCodes(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	codes := parseCodes(c)
	if len(codes) == 0 {
		return c.Status(400).SendString("no codes provided")
	}
	inserted, err := insertCodes(c, d, id, codes)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if _, err := d.Pool.Exec(c.Context(), `UPDATE vouchers SET stock = stock + $1 WHERE id=$2`, inserted, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target, diff) VALUES('admin','voucher.codes_add',$1, jsonb_build_object('codes',$2::int))`,
		id, inserted)
	return c.Redirect("/admin/vouchers/"+id, fiber.StatusSeeOther)
}

func disableVoucher(c *fiber.Ctx, d Deps) error {
	id := c.Params("id")
	tx, err := d.Pool.Begin(c.Context())
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer func() { _ = tx.Rollback(c.Context()) }()
	if _, err := tx.Exec(c.Context(), `DELETE FROM voucher_codes WHERE voucher_id=$1 AND used_at IS NULL`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if _, err := tx.Exec(c.Context(), `UPDATE vouchers SET stock=0 WHERE id=$1`, id); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if err := tx.Commit(c.Context()); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	_, _ = d.Pool.Exec(c.Context(),
		`INSERT INTO audit_log(admin_id, action, target) VALUES('admin','voucher.disable',$1)`, id)
	return c.Redirect("/admin/vouchers/"+id, fiber.StatusSeeOther)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

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

func insertCodes(c *fiber.Ctx, d Deps, voucherID string, codes []string) (int, error) {
	tx, err := d.Pool.Begin(c.Context())
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(c.Context()) }()
	inserted := 0
	for _, code := range codes {
		tag, err := tx.Exec(c.Context(),
			`INSERT INTO voucher_codes(voucher_id, code) VALUES($1, $2) ON CONFLICT DO NOTHING`,
			voucherID, code)
		if err != nil {
			return 0, fmt.Errorf("insert code: %w", err)
		}
		inserted += int(tag.RowsAffected())
	}
	return inserted, tx.Commit(c.Context())
}
