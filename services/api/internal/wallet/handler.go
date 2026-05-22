package wallet

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/middleware"
	"github.com/buocvang/api/internal/store"
)

type Handler struct {
	svc Service
	st  *store.Store
}

func NewHandler(svc Service, st *store.Store) *Handler { return &Handler{svc: svc, st: st} }

func (h *Handler) Balance(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	bal, err := h.svc.Balance(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(fiber.Map{"balance": bal, "currency": "POINT"})
}

func (h *Handler) Ledger(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	limit, _ := strconv.Atoi(c.Query("limit", "30"))
	items, err := h.st.Wallet.RecentEntries(c.Context(), uid, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(fiber.Map{"items": items})
}
