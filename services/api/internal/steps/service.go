package steps

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/buocvang/api/internal/antifraud"
	"github.com/buocvang/api/internal/config"
	"github.com/buocvang/api/internal/db"
	"github.com/buocvang/api/internal/leaderboard"
	"github.com/buocvang/api/internal/store"
)

type Service interface {
	Ingest(ctx context.Context, req IngestRequest) (*IngestResult, error)
	History(ctx context.Context, userID string, from, to time.Time) ([]store.DailySteps, error)
	TodayTotal(ctx context.Context, userID string) (int, error)
}

type service struct {
	pool *db.Pool
	st   *store.Store
	lb   *leaderboard.Client
	af   *antifraud.Checker
	cfg  *config.Config
}

func NewService(pool *db.Pool, st *store.Store, lb *leaderboard.Client, af *antifraud.Checker, cfg *config.Config) Service {
	return &service{pool: pool, st: st, lb: lb, af: af, cfg: cfg}
}

type IngestRequest struct {
	UserID   string
	Day      time.Time
	Source   string
	Chunks   []Chunk
	Device   Device
}

type Chunk struct {
	StartedAt   time.Time
	EndedAt     time.Time
	Steps       int
	ClientNonce string
	SensorHash  string
	Cadence     int // average period in ms
}

type Device struct {
	OS         string
	Model      string
	AppVersion string
}

type IngestResult struct {
	Accepted      int
	Rejected      int
	DayTotal      int
	Flagged       bool
	ScoreDelta    int
	FlagReasons   []string
}

func (s *service) Ingest(ctx context.Context, req IngestRequest) (*IngestResult, error) {
	if req.UserID == "" || req.Source == "" || len(req.Chunks) == 0 {
		return nil, errors.New("invalid ingest payload")
	}
	if req.Source != "zmp" && req.Source != "strava" {
		return nil, ErrInvalidSource
	}
	day := req.Day
	if day.IsZero() {
		day = time.Now().UTC().Truncate(24 * time.Hour)
	}
	res := &IngestResult{}
	err := db.InTx(ctx, s.pool, func(tx pgx.Tx) error {
		acceptedSteps := 0
		for _, ch := range req.Chunks {
			flagged, reasons := s.af.Check(ch.Steps, ch.Cadence, ch.StartedAt, ch.EndedAt)
			ev := store.StepEvent{
				UserID: req.UserID, Source: req.Source, Steps: ch.Steps,
				StartedAt: ch.StartedAt, EndedAt: ch.EndedAt,
				ClientNonce: ch.ClientNonce, Cadence: ch.Cadence,
				Flagged: flagged,
			}
			if len(reasons) > 0 {
				ev.FlagReason = reasons[0]
				res.FlagReasons = append(res.FlagReasons, reasons...)
			}
			if err := s.st.Steps.InsertEvent(ctx, tx, ev); err != nil {
				if errors.Is(err, store.ErrDuplicateNonce) {
					res.Rejected++
					continue
				}
				return err
			}
			res.Accepted++
			acceptedSteps += ch.Steps
			if flagged {
				res.Flagged = true
			}
		}
		if acceptedSteps > 0 {
			if err := s.st.Steps.UpsertDaily(ctx, tx, req.UserID, day, acceptedSteps, req.Source, res.Flagged); err != nil {
				return err
			}
			if err := s.st.Steps.IncrementChallengeProgress(ctx, tx, req.UserID, day, acceptedSteps); err != nil {
				return err
			}
		}
		if res.Flagged {
			res.ScoreDelta = 5
			_ = s.st.Users.IncrementFraudScore(ctx, req.UserID, 5)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Leaderboard update (best effort).
	if s.lb != nil && res.Accepted > 0 {
		stepsTotal, _ := s.st.Steps.GetDailyTotal(ctx, req.UserID, day)
		res.DayTotal = stepsTotal
		_ = s.lb.AddSteps(ctx, leaderboard.GlobalKey(), req.UserID, float64(sumChunkSteps(req.Chunks)))
	}
	return res, nil
}

func sumChunkSteps(cs []Chunk) int {
	n := 0
	for _, c := range cs {
		n += c.Steps
	}
	return n
}

func (s *service) History(ctx context.Context, userID string, from, to time.Time) ([]store.DailySteps, error) {
	return s.st.Steps.ListDailyForRange(ctx, userID, from, to)
}

func (s *service) TodayTotal(ctx context.Context, userID string) (int, error) {
	return s.st.Steps.GetDailyTotal(ctx, userID, time.Now().UTC().Truncate(24*time.Hour))
}
