package wallet

import (
	"context"
	"errors"

	"github.com/buocvang/api/internal/db"
	"github.com/buocvang/api/internal/store"
)

type service struct {
	pool *db.Pool
	st   *store.Store
}

func NewService(pool *db.Pool, st *store.Store) Service {
	return &service{pool: pool, st: st}
}

func (s *service) Balance(ctx context.Context, userID string) (int, error) {
	return s.st.Wallet.Balance(ctx, userID)
}

func (s *service) Credit(ctx context.Context, e Entry) error {
	err := s.st.Wallet.AppendEntry(ctx, nil, store.LedgerEntry{
		UserID: e.UserID, DeltaPoints: e.Points, Reason: e.Reason,
		ReferenceType: e.ReferenceType, ReferenceID: e.ReferenceID,
		IdempotencyKey: e.IdempotencyKey, Note: ptr(e.Note),
	})
	return mapStoreErr(err)
}

func (s *service) Debit(ctx context.Context, e Entry) error {
	err := s.st.Wallet.AppendEntry(ctx, nil, store.LedgerEntry{
		UserID: e.UserID, DeltaPoints: -e.Points, Reason: e.Reason,
		ReferenceType: e.ReferenceType, ReferenceID: e.ReferenceID,
		IdempotencyKey: e.IdempotencyKey, Note: ptr(e.Note),
	})
	return mapStoreErr(err)
}

func mapStoreErr(err error) error {
	switch {
	case errors.Is(err, store.ErrInsufficientBalance):
		return ErrInsufficientBalance
	case errors.Is(err, store.ErrDuplicateIdempotencyKey):
		return ErrDuplicateIdempotency
	default:
		return err
	}
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
