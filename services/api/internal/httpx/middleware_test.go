package httpx

import (
	"testing"

	"github.com/buocvang/api/internal/config"
)

func TestAllowOrigin(t *testing.T) {
	if !allowOrigin(&config.Config{AppEnv: "dev"}, "https://evil.example") {
		t.Fatalf("dev should allow any origin")
	}
	cfg := &config.Config{AppEnv: "prod", CORSAllowedOrigins: "https://app.example, https://zalo.example"}
	if !allowOrigin(cfg, "https://zalo.example") {
		t.Fatalf("prod should allow configured origin")
	}
	if allowOrigin(cfg, "https://evil.example") {
		t.Fatalf("prod should reject unconfigured origin")
	}
}
