// Package steps xử lý ingest, merge ZMP+Strava, anti-fraud chấm điểm cadence.
//
// Merge logic (xem PHAN_TICH_KIEN_TRUC.md §7.2):
//   - Nếu ngày X có cả ZMP và Strava: merged = max(zmp_steps, strava_steps).
//   - Nếu chênh > 30%: flagged=true, reason="source_diverge".
//   - Cadence variance: chuẩn của step period theo ms. Quá thấp (robot) hoặc
//     quá cao (lắc tay nhiễu) → flagged.
package steps

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNonceReplay   = errors.New("client_nonce đã được dùng")
	ErrInvalidSource = errors.New("source phải là zmp hoặc strava")
)

type IngestRequest struct {
	UserID       string
	Day          time.Time
	Source       string // 'zmp' | 'strava'
	Steps        int
	ClientNonce  string
	SensorHash   string
}

type DailySteps struct {
	Day          time.Time
	ZMPSteps     int
	StravaSteps  int
	MergedSteps  int
	Flagged      bool
	FlagReasons  []string
}

type Service interface {
	Ingest(ctx context.Context, req IngestRequest) (*DailySteps, error)
}

// Merge: hợp nhất zmp + strava theo quy tắc max, đánh flag khi diverge.
func Merge(zmp, strava int) (merged int, flagged bool, reasons []string) {
	merged = zmp
	if strava > merged {
		merged = strava
	}
	if zmp > 0 && strava > 0 {
		hi, lo := zmp, strava
		if lo > hi {
			hi, lo = lo, hi
		}
		if hi > 0 && float64(hi-lo)/float64(hi) > 0.3 {
			flagged = true
			reasons = append(reasons, "source_diverge")
		}
	}
	return merged, flagged, reasons
}
