// Package antifraud chấm fraud_score và đẩy event lên admin queue.
//
// Vector → biện pháp (xem PHAN_TICH_KIEN_TRUC.md §8):
//   - Tài khoản ảo: 1 zalo_id = 1 user, IP cluster score.
//   - Lắc tay giả: cadence_variance bất thường.
//   - Replay nonce: UNIQUE constraint + Redis bloom 24h.
//   - Strava giả: whitelist activity type, blacklist device.
//   - Brute redeem: rate limit 5/phút/user.
//
// Quy trình:
//   - Server gắn flagged=true mà vẫn cộng vào daily_steps (không UX xấu).
//   - Worker tính fraud_score mỗi đêm.
//   - score >= 70: vào admin queue.
//   - score >= 90 trong 3 ngày liên tiếp: auto suspend + email.
package antifraud

import (
	"context"
	"math"
)

type Service interface {
	ScoreUser(ctx context.Context, userID string) (int, []string, error)
}

// CadenceFlag trả về (flagged, reason) dựa trên period trung bình (ms).
// Bộ ngưỡng từ doc §8: chạy thật cadence 140–200 step/phút = period 300–428ms;
// đi bộ 80–130 = 461–750ms. Period < 250 hoặc > 1000 = bất thường.
func CadenceFlag(avgPeriodMs float64, variance float64) (bool, string) {
	if math.IsNaN(avgPeriodMs) || avgPeriodMs <= 0 {
		return true, "cadence_unknown"
	}
	if avgPeriodMs < 250 {
		return true, "cadence_too_fast"
	}
	if avgPeriodMs > 1000 {
		return true, "cadence_too_slow"
	}
	if variance < 5 {
		return true, "cadence_too_uniform" // robot-like
	}
	if variance > 250 {
		return true, "cadence_too_chaotic" // lắc tay
	}
	return false, ""
}
