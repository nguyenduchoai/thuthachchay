package web

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

// ─── i18n ────────────────────────────────────────────────────────────────────

// langOf trả về "vi" hoặc "en" dựa trên cookie. Mặc định "vi".
func langOf(c *fiber.Ctx) string {
	if v := c.Cookies("lang"); v == "en" {
		return "en"
	}
	return "vi"
}

// t trả chuỗi theo ngôn ngữ hiện tại.
func t(c *fiber.Ctx, vi, en string) string {
	if langOf(c) == "en" {
		return en
	}
	return vi
}

// setLang đổi cookie lang rồi redirect về back.
func setLang(c *fiber.Ctx) error {
	to := c.Query("to", "vi")
	if to != "en" {
		to = "vi"
	}
	c.Cookie(&fiber.Cookie{
		Name:     "lang",
		Value:    to,
		Path:     "/",
		HTTPOnly: false,
		MaxAge:   60 * 60 * 24 * 365,
	})
	back := c.Query("back", "/admin/reports/dashboard")
	if !strings.HasPrefix(back, "/") {
		back = "/admin/reports/dashboard"
	}
	return c.Redirect(back, fiber.StatusSeeOther)
}

// ─── pagination ──────────────────────────────────────────────────────────────

const perPageDefault = 50

// paginate đọc ?page=&pp= và trả về limit, offset, page hiện tại.
func paginate(c *fiber.Ctx) (limit, offset, pageN int) {
	p, _ := strconv.Atoi(c.Query("page", "1"))
	if p < 1 {
		p = 1
	}
	pp, _ := strconv.Atoi(c.Query("pp", strconv.Itoa(perPageDefault)))
	if pp < 10 {
		pp = perPageDefault
	}
	if pp > 200 {
		pp = 200
	}
	return pp, (p - 1) * pp, p
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// statusPillClass trả về Tailwind class theo status, dùng trong templ.
func statusPillClass(s string) string {
	switch s {
	case "active", "open", "live", "ok", "in":
		return "bg-emerald-100 text-emerald-800"
	case "suspended", "banned", "cancelled", "disabled":
		return "bg-red-100 text-red-700"
	case "settling", "draft":
		return "bg-amber-100 text-amber-800"
	default:
		return "bg-gray-200 text-gray-700"
	}
}

// stockPillClass — màu pill theo lượng stock voucher.
func stockPillClass(stock int) string {
	switch {
	case stock <= 10:
		return "bg-red-100 text-red-700"
	case stock <= 30:
		return "bg-amber-100 text-amber-800"
	default:
		return "bg-emerald-100 text-emerald-800"
	}
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func activeIf(c *fiber.Ctx, prefix string) string {
	if strings.HasPrefix(c.Path(), prefix) {
		return "active"
	}
	return ""
}

// isUUID đơn giản (8-4-4-4-12 hex).
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				return false
			}
		}
	}
	return true
}

// prettyJSON compact-format JSON cho display (giữ nguyên nếu lỗi).
func prettyJSON(s string) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	out, err := json.Marshal(v)
	if err != nil {
		return s
	}
	return string(out)
}

// shortID lấy 8 ký tự đầu (cho hiển thị UUID rút gọn).
func shortID(s string) string {
	return s[:minInt(len(s), 8)]
}

// ─── handler /  ──────────────────────────────────────────────────────────────

func index(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return indexPage(c).Render(c.Context(), c.Response().BodyWriter())
}
