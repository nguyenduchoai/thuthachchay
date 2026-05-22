package httpx

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/config"
)

func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	cfg := &config.Config{AppEnv: "dev"}
	RegisterMiddleware(app, cfg)
	RegisterRoutes(app, cfg)
	return app
}

func TestHealthz(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest("GET", "/healthz", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", got["status"])
	}
}

func TestVersion(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest("GET", "/version", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestNotImplemented_ReturnsJSON501(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest("POST", "/v1/steps/ingest", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if res.StatusCode != 501 {
		t.Fatalf("expected 501, got %d", res.StatusCode)
	}
}
