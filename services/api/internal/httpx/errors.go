package httpx

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// AppError là kiểu lỗi domain trả về client.
// Code dùng cho i18n (FE map sang text), Message là fallback en.
type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Code + ": " + e.Message }

func NewAppError(status int, code, msg string) *AppError {
	return &AppError{Status: status, Code: code, Message: msg}
}

var (
	ErrUnauthorized = NewAppError(fiber.StatusUnauthorized, "unauthorized", "Authentication required")
	ErrForbidden    = NewAppError(fiber.StatusForbidden, "forbidden", "Permission denied")
	ErrNotFound     = NewAppError(fiber.StatusNotFound, "not_found", "Resource not found")
	ErrConflict     = NewAppError(fiber.StatusConflict, "conflict", "Resource conflict")
	ErrRateLimit    = NewAppError(fiber.StatusTooManyRequests, "rate_limited", "Too many requests")
	ErrInternal     = NewAppError(fiber.StatusInternalServerError, "internal", "Internal server error")
)

// ErrorHandler convert mọi error trả về body JSON chuẩn.
func ErrorHandler(c *fiber.Ctx, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error":      appErr,
			"request_id": c.Get("X-Request-ID"),
		})
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "http_error",
				"message": fe.Message,
			},
			"request_id": c.Get("X-Request-ID"),
		})
	}

	log.Error().Err(err).Str("path", c.Path()).Msg("unhandled error")
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":      ErrInternal,
		"request_id": c.Get("X-Request-ID"),
	})
}
