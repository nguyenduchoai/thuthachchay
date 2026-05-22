// Package wallet quản lý điểm (point ledger).
// Mọi mutation đi qua InsertLedgerEntry với idempotency_key.
// Balance = SUM(delta_points) qua view user_balances.
package wallet

import (
	"context"
	"errors"
)

var (
	ErrInsufficientBalance = errors.New("không đủ điểm")
	ErrDuplicateIdempotency = errors.New("idempotency key đã dùng")
)

type Service interface {
	Balance(ctx context.Context, userID string) (int, error)
	Credit(ctx context.Context, e Entry) error
	Debit(ctx context.Context, e Entry) error
}

type Entry struct {
	UserID         string
	Points         int    // luôn dương; Credit cộng vào, Debit trừ ra
	Reason         string // 'challenge_join'|'challenge_payout'|'voucher_redeem'|'referral'|'admin_adjust'
	ReferenceType  string
	ReferenceID    string
	IdempotencyKey string
	Note           string
}
