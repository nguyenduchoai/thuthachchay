// Package referrals: track + stats. Bonus 500đ trả 2 phía sau khi invitee hoàn thành challenge đầu.
package referrals

import (
	"context"

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

func (h *Handler) Me(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	u, err := h.st.Users.GetByID(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fiber.Map{"code": "user_not_found"}})
	}
	code := uid
	if u.Handle != nil && *u.Handle != "" {
		code = *u.Handle
	}
	invited, joined, earned, _ := h.st.Referrals.Stats(c.Context(), uid)
	return c.JSON(fiber.Map{
		"code":  code,
		"stats": fiber.Map{"invited": invited, "joined": joined, "earned": earned},
	})
}

type trackReq struct {
	Code string `json:"code"`
}

func (h *Handler) Track(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var req trackReq
	_ = c.BodyParser(&req)
	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "missing_code"}})
	}
	var inviterID string
	if err := h.pool.QueryRow(c.Context(), `SELECT id FROM users WHERE handle=$1`, req.Code).Scan(&inviterID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fiber.Map{"code": "inviter_not_found"}})
	}
	if inviterID == uid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "self_referral"}})
	}
	if err := h.st.Referrals.Track(c.Context(), inviterID, uid, req.Code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(fiber.Map{"ok": true, "inviter": inviterID})
}

// PayBonusOnFirstChallengeComplete được gọi từ worker khi invitee finish 1 thử thách.
// Idempotent: AppendEntry dùng idempotency_key có invitee_id; gọi nhiều lần OK.
func PayBonusOnFirstChallengeComplete(ctx context.Context, pool *db.Pool, st *store.Store, inviteeID string, bonusPoints int) error {
	r, err := st.Referrals.FindUnpaidByInvitee(ctx, inviteeID)
	if err != nil {
		return nil // không có referral hoặc đã trả → bỏ qua
	}
	return db.InTx(ctx, pool, func(tx pgx.Tx) error {
		if err := st.Wallet.AppendEntry(ctx, tx, store.LedgerEntry{
			UserID: r.InviterID, DeltaPoints: bonusPoints, Reason: "referral",
			ReferenceType: "referral", ReferenceID: r.InviteeID,
			IdempotencyKey: "ref-inviter:" + r.InviterID + ":" + r.InviteeID,
		}); err != nil {
			return err
		}
		if err := st.Wallet.AppendEntry(ctx, tx, store.LedgerEntry{
			UserID: r.InviteeID, DeltaPoints: bonusPoints, Reason: "referral",
			ReferenceType: "referral", ReferenceID: r.InviterID,
			IdempotencyKey: "ref-invitee:" + r.InviterID + ":" + r.InviteeID,
		}); err != nil {
			return err
		}
		return st.Referrals.MarkPaid(ctx, tx, r.InviterID, r.InviteeID)
	})
}
