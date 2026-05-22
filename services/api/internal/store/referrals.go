package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/db"
)

type Referral struct {
	InviterID  string
	InviteeID  string
	Code       string
	BonusPaid  bool
	CreatedAt  time.Time
}

type ReferralsStore struct{ pool *db.Pool }

// Track ghi nhận khi invitee đăng ký lần đầu. inviterID lookup từ code (handle).
func (s *ReferralsStore) Track(ctx context.Context, inviterID, inviteeID, code string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO referrals (inviter_id, invitee_id, code, bonus_paid)
		VALUES ($1, $2, $3, false) ON CONFLICT (inviter_id, invitee_id) DO NOTHING`, inviterID, inviteeID, code)
	return err
}

// FindUnpaidByInvitee — invitee vừa join challenge đầu, trả về referral chưa thanh toán bonus.
func (s *ReferralsStore) FindUnpaidByInvitee(ctx context.Context, inviteeID string) (*Referral, error) {
	row := s.pool.QueryRow(ctx, `SELECT inviter_id, invitee_id, code, bonus_paid, created_at
		FROM referrals WHERE invitee_id=$1 AND bonus_paid=false LIMIT 1`, inviteeID)
	var r Referral
	if err := row.Scan(&r.InviterID, &r.InviteeID, &r.Code, &r.BonusPaid, &r.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *ReferralsStore) MarkPaid(ctx context.Context, tx pgx.Tx, inviterID, inviteeID string) error {
	q := `UPDATE referrals SET bonus_paid=true WHERE inviter_id=$1 AND invitee_id=$2`
	if tx != nil {
		_, err := tx.Exec(ctx, q, inviterID, inviteeID)
		return err
	}
	_, err := s.pool.Exec(ctx, q, inviterID, inviteeID)
	return err
}

func (s *ReferralsStore) Stats(ctx context.Context, inviterID string) (invited, joined, earned int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM referrals WHERE inviter_id=$1),
			(SELECT count(*) FROM referrals WHERE inviter_id=$1 AND bonus_paid=true),
			(SELECT COALESCE(SUM(delta_points),0)::int FROM ledger_entries WHERE user_id=$1 AND reason='referral')
		`, inviterID).Scan(&invited, &joined, &earned)
	return
}
