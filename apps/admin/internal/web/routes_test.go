package web

import (
	"net/http/httptest"
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
