package strava

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/middleware"
	"github.com/buocvang/api/internal/steps"
	"github.com/buocvang/api/internal/store"
)

type Handler struct {
	c        *Client
	cfg      *config.Config
	stepsSvc steps.Service
	states   sync.Map
}

func NewHandler(c *Client, cfg *config.Config, stepsSvc steps.Service) *Handler {
	return &Handler{c: c, cfg: cfg, stepsSvc: stepsSvc}
}

type oauthState struct {
	UserID    string
	ExpiresAt time.Time
}

// GET /v1/strava/oauth/url → { url, state }
func (h *Handler) AuthURL(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	state := randomState()
	h.states.Store(state, oauthState{UserID: uid, ExpiresAt: time.Now().Add(10 * time.Minute)})
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
	raw, ok := h.states.LoadAndDelete(req.State)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_state"}})
	}
	state, ok := raw.(oauthState)
	if !ok || state.UserID != uid || time.Now().After(state.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_state"}})
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

// POST /v1/strava/webhook — Strava push event. Trả 200 ngay, xử lý activity sync async.
func (h *Handler) WebhookEvent(c *fiber.Ctx) error {
	var ev webhookEvent
	if err := c.BodyParser(&ev); err != nil {
		return c.SendStatus(fiber.StatusOK)
	}
	if ev.ObjectType == "activity" && (ev.AspectType == "create" || ev.AspectType == "update") {
		go h.syncActivity(context.Background(), ev)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) syncActivity(ctx context.Context, ev webhookEvent) {
	if h.stepsSvc == nil || ev.OwnerID == 0 || ev.ObjectID == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tok, err := h.c.Tokens.FindByAthleteID(ctx, fmt.Sprintf("%d", ev.OwnerID))
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			log.Warn().Err(err).Int64("owner_id", ev.OwnerID).Msg("strava token lookup failed")
		}
		return
	}
	tok, err = h.c.RefreshIfNeeded(ctx, tok)
	if err != nil {
		log.Warn().Err(err).Str("user_id", tok.UserID).Msg("strava token refresh failed")
		return
	}
	activity, err := h.c.GetActivity(ctx, tok.AccessToken, ev.ObjectID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", tok.UserID).Int64("activity_id", ev.ObjectID).Msg("strava activity fetch failed")
		return
	}
	stepsN := EstimateSteps(activity)
	if stepsN <= 0 {
		return
	}
	startedAt, err := time.Parse(time.RFC3339, activity.StartDate)
	if err != nil {
		log.Warn().Err(err).Int64("activity_id", ev.ObjectID).Msg("strava activity bad start_date")
		return
	}
	endedAt := startedAt.Add(time.Duration(activity.MovingTime) * time.Second)
	_, err = h.stepsSvc.Ingest(ctx, steps.IngestRequest{
		UserID: tok.UserID,
		Day:    startedAt.UTC().Truncate(24 * time.Hour),
		Source: "strava",
		Chunks: []steps.Chunk{{
			StartedAt:   startedAt,
			EndedAt:     endedAt,
			Steps:       stepsN,
			ClientNonce: fmt.Sprintf("strava:%d", activity.ID),
		}},
	})
	if err != nil {
		log.Warn().Err(err).Str("user_id", tok.UserID).Int64("activity_id", ev.ObjectID).Msg("strava steps ingest failed")
	}
}

func randomState() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
