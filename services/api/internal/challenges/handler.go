package challenges

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/leaderboard"
	"github.com/buocvang/api/internal/middleware"
	"github.com/buocvang/api/internal/store"
)

type Handler struct {
	svc   Service
	store *store.Store
	lb    *leaderboard.Client
}

func NewHandler(svc Service, st *store.Store, lb *leaderboard.Client) *Handler {
	return &Handler{svc: svc, store: st, lb: lb}
}

func (h *Handler) List(c *fiber.Ctx) error {
	phase := c.Query("phase", "")
	status := mapPhaseToStatus(phase)
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	list, err := h.svc.List(c.Context(), status, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	out := make([]fiber.Map, 0, len(list))
	for i := range list {
		out = append(out, challengeAPI(&list[i]))
	}
	return c.JSON(fiber.Map{"items": out})
}

func mapPhaseToStatus(phase string) string {
	switch phase {
	case "live":
		return "live"
	case "upcoming":
		return "open"
	case "wrapping":
		return "settling"
	default:
		return ""
	}
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	ch, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fiber.Map{"code": "not_found"}})
	}
	count, _ := h.store.Challenges.ParticipantCount(c.Context(), id)
	resp := challengeAPI(ch)
	resp["participants"] = count
	return c.JSON(resp)
}

type createReq struct {
	Visibility       string    `json:"visibility"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CoverURL         string    `json:"cover_url"`
	DailyStepsTarget int       `json:"daily_steps_target"`
	DurationDays     int       `json:"duration_days"`
	StartDate        time.Time `json:"start_date"`
	EntryPoints      int       `json:"entry_points"`
	MaxParticipants  int       `json:"max_participants"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var req createReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	ch, err := h.svc.Create(c.Context(), CreateRequest{
		HostID: uid, Visibility: req.Visibility, Name: req.Name,
		Description: req.Description, CoverURL: req.CoverURL,
		DailyStepsTarget: req.DailyStepsTarget, DurationDays: req.DurationDays,
		StartDate: req.StartDate, EntryPoints: req.EntryPoints,
		MaxParticipants: req.MaxParticipants,
	})
	if err != nil {
		if errors.Is(err, store.ErrInsufficientBalance) {
			return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": fiber.Map{"code": "insufficient_points", "message": err.Error()}})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.Status(fiber.StatusCreated).JSON(challengeAPI(ch))
}

func (h *Handler) Join(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	id := c.Params("id")
	idem := c.Get("X-Idempotency-Key")
	if err := h.svc.Join(c.Context(), id, uid, idem); err != nil {
		switch {
		case errors.Is(err, store.ErrInsufficientBalance):
			return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": fiber.Map{"code": "insufficient_points", "message": err.Error()}})
		case errors.Is(err, ErrChallengeFull):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": fiber.Map{"code": "challenge_full", "message": err.Error()}})
		default:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) Leaderboard(c *fiber.Ctx) error {
	id := c.Params("id")
	parts, err := h.store.Challenges.ListParticipants(c.Context(), id, 100)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	out := make([]fiber.Map, 0, len(parts))
	for i, p := range parts {
		out = append(out, fiber.Map{
			"rank": i + 1, "user_id": p.UserID, "total_steps": p.TotalSteps, "state": p.State,
		})
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) GlobalLeaderboard(c *fiber.Ctx) error {
	if h.lb == nil {
		return c.JSON(fiber.Map{"items": []any{}})
	}
	top, err := h.lb.Top(c.Context(), leaderboard.GlobalKey(), 50)
	if err != nil {
		return c.JSON(fiber.Map{"items": []any{}})
	}
	out := make([]fiber.Map, 0, len(top))
	for i, e := range top {
		out = append(out, fiber.Map{"rank": i + 1, "user_id": e.UserID, "steps": e.Steps})
	}
	return c.JSON(fiber.Map{"items": out})
}

func challengeAPI(c *Challenge) fiber.Map {
	return fiber.Map{
		"id":                 c.ID,
		"host_id":            c.HostID,
		"visibility":         c.Visibility,
		"name":               c.Name,
		"description":        c.Description,
		"cover_url":          c.CoverURL,
		"daily_steps_target": c.DailyStepsTarget,
		"duration_days":      c.DurationDays,
		"entry_points":       c.EntryPoints,
		"prize_pool":         c.PrizePool,
		"max_participants":   c.MaxParticipants,
		"start_date":         c.StartDate,
		"end_date":           c.EndDate,
		"status":             c.Status,
	}
}
