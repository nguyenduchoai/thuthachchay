// Package vouchers: list, my, redeem (idempotent).
package vouchers

import (
	"errors"

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
	return c.JSON(fiber.Map{"items": items})
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
	return c.JSON(fiber.Map{"items": items})
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
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
		}
	}
	return c.JSON(fiber.Map{"code": code, "voucher_id": v.ID, "brand": v.Brand, "title": v.Title})
}
