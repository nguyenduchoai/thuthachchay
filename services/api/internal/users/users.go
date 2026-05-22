// Package users: handlers cho /me, /username/check, /me/attribution.
package users

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/middleware"
	"github.com/buocvang/api/internal/store"
)

type Handler struct {
	users *store.UsersStore
}

func NewHandler(users *store.UsersStore) *Handler { return &Handler{users: users} }

var handleRegex = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)

func (h *Handler) Me(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "unauthorized"}})
	}
	u, err := h.users.GetByID(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fiber.Map{"code": "user_not_found"}})
	}
	return c.JSON(toAPI(u))
}

type patchReq struct {
	Handle      *string `json:"handle,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	DailyGoal   *int    `json:"daily_goal,omitempty"`
	Locale      *string `json:"locale,omitempty"`
	Email       *string `json:"email,omitempty"`
}

func (h *Handler) Patch(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var req patchReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "bad_request"}})
	}
	if req.Handle != nil {
		hl := strings.ToLower(strings.TrimSpace(*req.Handle))
		if !handleRegex.MatchString(hl) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_handle"}})
		}
		req.Handle = &hl
	}
	if req.DailyGoal != nil && (*req.DailyGoal < 1000 || *req.DailyGoal > 50000) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_daily_goal"}})
	}
	u, err := h.users.Patch(c.Context(), uid, store.UserPatch{
		Handle: req.Handle, DisplayName: req.DisplayName, AvatarURL: req.AvatarURL,
		DailyGoal: req.DailyGoal, Locale: req.Locale, Email: req.Email,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"code": "update_failed", "message": err.Error()}})
	}
	return c.JSON(toAPI(u))
}

type attribReq struct {
	Source string `json:"source"`
}

func (h *Handler) Attribution(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var req attribReq
	_ = c.BodyParser(&req)
	if req.Source == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if _, err := h.users.Patch(c.Context(), uid, store.UserPatch{Acquisition: &req.Source}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(fiber.Map{"ok": true})
}

type checkReq struct {
	Handle string `json:"handle"`
}

func (h *Handler) CheckHandle(c *fiber.Ctx) error {
	var req checkReq
	_ = c.BodyParser(&req)
	hl := strings.ToLower(strings.TrimSpace(req.Handle))
	if !handleRegex.MatchString(hl) {
		return c.JSON(fiber.Map{"available": false, "reason": "invalid_format"})
	}
	avail, err := h.users.HandleAvailable(c.Context(), hl)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(fiber.Map{"available": avail, "handle": hl})
}

func toAPI(u *store.User) fiber.Map {
	return fiber.Map{
		"id":           u.ID,
		"zalo_id":      u.ZaloID,
		"handle":       u.Handle,
		"email":        u.Email,
		"display_name": u.DisplayName,
		"avatar_url":   u.AvatarURL,
		"daily_goal":   u.DailyGoal,
		"locale":       u.Locale,
		"status":       u.Status,
		"created_at":   u.CreatedAt,
	}
}
