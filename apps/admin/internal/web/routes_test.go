package web

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestIndex(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app)
	req := httptest.NewRequest("GET", "/", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	bs := string(body)
	// Phải có Tailwind CDN + brand colour config + heading.
	for _, want := range []string{
		"cdn.tailwindcss.com",
		"brand: '#ff9500'",
		"Bước Vàng — Admin",
		"htmx.org",
	} {
		if !strings.Contains(bs, want) {
			t.Errorf("index body missing %q", want)
		}
	}
}

func TestHealthz(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app)
	req := httptest.NewRequest("GET", "/healthz", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

// /lang phải set cookie và redirect 303.
func TestSetLang(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app)
	req := httptest.NewRequest("GET", "/lang?to=en&back=/admin/users", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if res.StatusCode != 303 {
		t.Fatalf("expected 303, got %d", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); loc != "/admin/users" {
		t.Fatalf("expected redirect to /admin/users, got %q", loc)
	}
	hasLang := false
	for _, ck := range res.Cookies() {
		if ck.Name == "lang" && ck.Value == "en" {
			hasLang = true
		}
	}
	if !hasLang {
		t.Fatalf("expected lang=en cookie to be set")
	}
}

// Không có DATABASE_URL → /admin/* trả 503 để tránh chạy admin thiếu DB.
func TestAdminNoDatabaseFallback(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app)
	for _, path := range []string{
		"/admin/users",
		"/admin/challenges",
		"/admin/vouchers",
		"/admin/audit",
		"/admin/reports/dashboard",
		"/admin/reports/dso",
		"/admin/reports/csv/challenges",
	} {
		req := httptest.NewRequest("GET", path, nil)
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("app.Test %s: %v", path, err)
		}
		if res.StatusCode != 503 {
			t.Fatalf("%s: expected 503 without database, got %d", path, res.StatusCode)
		}
	}
}

// ─── pure helpers ────────────────────────────────────────────────────────────

func TestIsUUID(t *testing.T) {
	cases := map[string]bool{
		"00000000-0000-0000-0000-000000000000": true,
		"abcd1234-ab12-cd34-ef56-abcdef123456": true,
		"not-a-uuid":                           false,
		"":                                     false,
		"abcd1234ab12cd34ef56abcdef123456":     false,
		"abcd1234-ab12-cd34-ef56-abcdefXXXXXX": false,
	}
	for in, want := range cases {
		if got := isUUID(in); got != want {
			t.Errorf("isUUID(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestPrettyJSON(t *testing.T) {
	cases := map[string]string{
		`{"a":1,"b":2}`: `{"a":1,"b":2}`,
		`not json`:      `not json`,
		``:              ``,
		`{"x":  "y" }`:  `{"x":"y"}`,
	}
	for in, want := range cases {
		if got := prettyJSON(in); got != want {
			t.Errorf("prettyJSON(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatInt(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		100:     "100",
		1000:    "1.000",
		1234567: "1.234.567",
		-500000: "-500.000",
	}
	for in, want := range cases {
		if got := formatInt(in); got != want {
			t.Errorf("formatInt(%d) = %q, want %q", in, got, want)
		}
	}
}

// TestStockPillClass — sanity cho tier màu stock.
func TestStockPillClass(t *testing.T) {
	cases := map[int]string{
		0:  "bg-red-100 text-red-700",
		10: "bg-red-100 text-red-700",
		20: "bg-amber-100 text-amber-800",
		50: "bg-emerald-100 text-emerald-800",
	}
	for in, want := range cases {
		if got := stockPillClass(in); got != want {
			t.Errorf("stockPillClass(%d) = %q, want %q", in, got, want)
		}
	}
}

// TestShortID — bảo vệ vs UUID < 8 chars.
func TestShortID(t *testing.T) {
	if got := shortID("abcd-1234-..."); got != "abcd-123" {
		t.Errorf("shortID full = %q", got)
	}
	if got := shortID("abc"); got != "abc" {
		t.Errorf("shortID short = %q", got)
	}
}
