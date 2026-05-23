package steps

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/middleware"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

type ingestReqDTO struct {
	Day    string            `json:"day"` // YYYY-MM-DD
	Source string            `json:"source"`
	Chunks []chunkDTO        `json:"chunks"`
	Device map[string]string `json:"device"`
}

type chunkDTO struct {
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Steps       int       `json:"steps"`
	ClientNonce string    `json:"client_nonce"`
	SensorHash  string    `json:"sensor_hash"`
	Cadence     int       `json:"cadence_avg_ms"`
}

func (h *Handler) Ingest(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var req ingestReqDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	day, _ := time.Parse("2006-01-02", req.Day)
	chunks := make([]Chunk, 0, len(req.Chunks))
	for _, c := range req.Chunks {
		chunks = append(chunks, Chunk{
			StartedAt: c.Start, EndedAt: c.End, Steps: c.Steps,
			ClientNonce: c.ClientNonce, SensorHash: c.SensorHash, Cadence: c.Cadence,
		})
	}
	res, err := h.svc.Ingest(c.Context(), IngestRequest{
		UserID: uid, Day: day, Source: req.Source, Chunks: chunks,
		Device: Device{OS: req.Device["os"], Model: req.Device["model"], AppVersion: req.Device["app_version"]},
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(fiber.Map{
		"accepted": res.Accepted, "rejected": res.Rejected,
		"day_total": res.DayTotal, "flagged": res.Flagged,
		"score_delta": res.ScoreDelta, "flag_reasons": res.FlagReasons,
	})
}

func (h *Handler) Today(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	n, _ := h.svc.TodayTotal(c.Context(), uid)
	return c.JSON(fiber.Map{"day_total": n})
}

func (h *Handler) History(c *fiber.Ctx) error {
	uid := middleware.UserID(c)
	if uid == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	from, _ := time.Parse("2006-01-02", c.Query("from"))
	to, _ := time.Parse("2006-01-02", c.Query("to"))
	if from.IsZero() {
		from = time.Now().UTC().AddDate(0, 0, -30)
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}
	list, err := h.svc.History(c.Context(), uid, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	out := make([]fiber.Map, 0, len(list))
	for _, item := range list {
		out = append(out, fiber.Map{
			"user_id": item.UserID,
			"day":     item.Day.Format("2006-01-02"),
			"steps":   item.Steps,
			"source":  item.Source,
			"flagged": item.Flagged,
		})
	}
	return c.JSON(fiber.Map{"items": out})
}
