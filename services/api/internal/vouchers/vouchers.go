// Package vouchers: list, my, redeem (idempotent).
package vouchers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
	"github.com/buocvang/api/internal/middleware"
	"github.com/buocvang/api/internal/store"
)

type Handler struct {
	pool *db.Pool
	st   *store.Store
}

func NewHandler(pool *db.Pool, st *store.Store) *Handler { return &Handler{pool: pool, st: st} }

func (h *Handler) List(c *fiber.Ctx) error {
	items, err := h.st.Vouchers.List(c.Context(), 50)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	out := make([]fiber.Map, 0, len(items))
	for _, item := range items {
		out = append(out, voucherAPI(item))
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) Mine(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	items, err := h.st.Vouchers.ListMine(c.Context(), uid, 50)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	out := make([]fiber.Map, 0, len(items))
	for _, item := range items {
		out = append(out, fiber.Map{
			"id":          item.ID,
			"user_id":     item.UserID,
			"voucher_id":  item.VoucherID,
			"code":        item.Code,
			"redeemed_at": item.RedeemedAt,
		})
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) Redeem(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	id := c.Params("id")
	idem := c.Get("X-Idempotency-Key")
	if idem == "" {
		idem = "redeem:" + id + ":" + uid
	}
	v, err := h.st.Vouchers.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fiber.Map{"code": "voucher_not_found"}})
	}
	if v.ExpiresAt != nil && v.ExpiresAt.Before(time.Now().Truncate(24*time.Hour)) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": fiber.Map{"code": "voucher_expired"}})
	}
	var code string
	err = db.InTx(c.Context(), h.pool, func(tx pgx.Tx) error {
		// 1) trừ điểm
		if err := h.st.Wallet.AppendEntry(c.Context(), tx, store.LedgerEntry{
			UserID: uid, DeltaPoints: -v.CostPoints, Reason: "voucher_redeem",
			ReferenceType: "voucher", ReferenceID: v.ID,
			IdempotencyKey: idem,
		}); err != nil {
			return err
		}
		// 2) cấp code
		code2, err := h.st.Vouchers.AllocateCode(c.Context(), tx, v.ID, uid)
		if err != nil {
			return err
		}
		code = code2
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrOutOfStock):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": fiber.Map{"code": "out_of_stock"}})
		case errors.Is(err, store.ErrDuplicateIdempotencyKey):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": fiber.Map{"code": "duplicate"}})
		case errors.Is(err, store.ErrInsufficientBalance):
			return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": fiber.Map{"code": "insufficient_points"}})
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
		}
	}
	return c.JSON(fiber.Map{"code": code, "voucher_id": v.ID, "brand": v.Brand, "title": v.Title})
}

func voucherAPI(v store.Voucher) fiber.Map {
	return fiber.Map{
		"id":          v.ID,
		"brand":       v.Brand,
		"title":       v.Title,
		"cost_points": v.CostPoints,
		"stock":       v.Stock,
		"cover_url":   v.CoverURL,
		"expires_at":  v.ExpiresAt,
	}
}
