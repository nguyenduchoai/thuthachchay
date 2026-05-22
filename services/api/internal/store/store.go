// Package store gom truy vấn SQL theo từng entity.
// Tự viết tay (chưa generate sqlc) để dễ đọc + tránh dep generator.
package store

import (
	"github.com/buocvang/api/internal/db"
)

type Store struct {
	Pool *db.Pool

	Users         *UsersStore
	Challenges    *ChallengesStore
	Steps         *StepsStore
	Wallet        *WalletStore
	Vouchers      *VouchersStore
	Referrals     *ReferralsStore
	StravaTokens  *StravaTokensStore
	Sessions      *SessionsStore
}

func New(pool *db.Pool) *Store {
	return &Store{
		Pool:         pool,
		Users:        &UsersStore{pool: pool},
		Challenges:   &ChallengesStore{pool: pool},
		Steps:        &StepsStore{pool: pool},
		Wallet:       &WalletStore{pool: pool},
		Vouchers:     &VouchersStore{pool: pool},
		Referrals:    &ReferralsStore{pool: pool},
		StravaTokens: &StravaTokensStore{pool: pool},
		Sessions:     &SessionsStore{pool: pool},
	}
}
