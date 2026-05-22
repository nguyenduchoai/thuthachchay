package strava

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/middleware"
)

type Handler struct {
	c   *Client
	cfg *config.Config
}

func NewHandler(c *Client, cfg *config.Config) *Handler { return &Handler{c: c, cfg: cfg} }

// GET /v1/strava/oauth/url → { url, state }
func (h *Handler) AuthURL(c *fiber.Ctx) error {
	state := randomState()
	return c.JSON(fiber.Map{"url": h.c.AuthorizeURL(state), "state": state})
}

type cbReq struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// POST /v1/strava/oauth/callback { code, state }
func (h *Handler) Callback(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var req cbReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	if err := h.c.ExchangeCode(c.Context(), uid, req.Code); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// GET /v1/strava/webhook — Strava verify subscription.
func (h *Handler) WebhookVerify(c *fiber.Ctx) error {
	hubMode := c.Query("hub.mode")
	hubChallenge := c.Query("hub.challenge")
	hubToken := c.Query("hub.verify_token")
	if hubMode == "subscribe" && hubToken == h.cfg.StravaVerifyToken {
		return c.JSON(fiber.Map{"hub.challenge": hubChallenge})
	}
	return c.SendStatus(fiber.StatusForbidden)
}

type webhookEvent struct {
	ObjectType string         `json:"object_type"`
	ObjectID   int64          `json:"object_id"`
	AspectType string         `json:"aspect_type"`
	OwnerID    int64          `json:"owner_id"`
	Updates    map[string]any `json:"updates"`
}

// POST /v1/strava/webhook — Strava push event. Trả 200 ngay, xử lý async (TODO worker queue).
func (h *Handler) WebhookEvent(c *fiber.Ctx) error {
	var ev webhookEvent
	if err := c.BodyParser(&ev); err != nil {
		return c.SendStatus(fiber.StatusOK)
	}
	// TODO: gửi vào asynq queue để worker xử lý activity → steps ingest.
	_ = ev
	return c.SendStatus(fiber.StatusOK)
}

func randomState() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
