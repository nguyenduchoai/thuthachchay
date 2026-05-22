package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
)

type Voucher struct {
	ID         string
	Brand      string
	Title      string
	CostPoints int
	Stock      int
	CoverURL   *string
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

type VoucherRedemption struct {
	ID         string
	UserID     string
	VoucherID  string
	Code       string
	RedeemedAt time.Time
}

type VouchersStore struct{ pool *db.Pool }

var ErrOutOfStock = errors.New("voucher hết kho")

func (s *VouchersStore) List(ctx context.Context, limit int) ([]Voucher, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `SELECT id, brand, title, cost_points, stock, cover_url, expires_at, created_at
		FROM vouchers WHERE stock > 0 AND (expires_at IS NULL OR expires_at >= CURRENT_DATE)
		ORDER BY cost_points ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Voucher{}
	for rows.Next() {
		var v Voucher
		if err := rows.Scan(&v.ID, &v.Brand, &v.Title, &v.CostPoints, &v.Stock, &v.CoverURL, &v.ExpiresAt, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *VouchersStore) Get(ctx context.Context, id string) (*Voucher, error) {
	row := s.pool.QueryRow(ctx, `SELECT id, brand, title, cost_points, stock, cover_url, expires_at, created_at FROM vouchers WHERE id=$1`, id)
	var v Voucher
	if err := row.Scan(&v.ID, &v.Brand, &v.Title, &v.CostPoints, &v.Stock, &v.CoverURL, &v.ExpiresAt, &v.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

// AllocateCode pop 1 mã chưa dùng, trừ stock, ghi redemption. Atomic trong tx.
func (s *VouchersStore) AllocateCode(ctx context.Context, tx pgx.Tx, voucherID, userID string) (string, error) {
	// Lấy 1 mã FOR UPDATE SKIP LOCKED, mark used.
	var code string
	err := tx.QueryRow(ctx, `
		UPDATE voucher_codes SET used_by_user_id=$2, used_at=now()
		WHERE id IN (
			SELECT id FROM voucher_codes
			WHERE voucher_id=$1 AND used_at IS NULL
			ORDER BY id LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING code`, voucherID, userID).Scan(&code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrOutOfStock
		}
		return "", err
	}
	if _, err := tx.Exec(ctx, `UPDATE vouchers SET stock = stock - 1 WHERE id=$1 AND stock > 0`, voucherID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO voucher_redemptions (user_id, voucher_id, code) VALUES ($1, $2, $3)`,
		userID, voucherID, code); err != nil {
		return "", err
	}
	return code, nil
}

func (s *VouchersStore) ListMine(ctx context.Context, userID string, limit int) ([]VoucherRedemption, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `SELECT id, user_id, voucher_id, code, redeemed_at FROM voucher_redemptions
		WHERE user_id=$1 ORDER BY redeemed_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []VoucherRedemption{}
	for rows.Next() {
		var r VoucherRedemption
		if err := rows.Scan(&r.ID, &r.UserID, &r.VoucherID, &r.Code, &r.RedeemedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
