// Package middleware tách JWT-auth middleware ra để các handler khác chỉ phụ thuộc Service.
package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/auth"
)

const UserIDLocal = "user_id"

type UserStatusLookup interface {
	UserStatus(ctx context.Context, id string) (string, error)
}

// RequireJWT trả về middleware xác thực Bearer token. 401 nếu thiếu/invalid.
func RequireJWT(jm *auth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get(fiber.HeaderAuthorization)
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{"code": "unauthorized", "message": "missing bearer token"},
			})
		}
		tok := strings.TrimPrefix(h, "Bearer ")
		claims, err := jm.Parse(tok)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{"code": "unauthorized", "message": err.Error()},
			})
		}
		c.Locals(UserIDLocal, claims.UserID)
		return c.Next()
	}
}

// RequireActiveJWT xác thực Bearer token và chặn user suspended/banned.
func RequireActiveJWT(jm *auth.JWTManager, users UserStatusLookup) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get(fiber.HeaderAuthorization)
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{"code": "unauthorized", "message": "missing bearer token"},
			})
		}
		tok := strings.TrimPrefix(h, "Bearer ")
		claims, err := jm.Parse(tok)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{"code": "unauthorized", "message": err.Error()},
			})
		}
		status, err := users.UserStatus(c.Context(), claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{"code": "unauthorized", "message": "user not found"},
			})
		}
		if status != "active" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fiber.Map{"code": "user_disabled", "message": "user is not active"},
			})
		}
		c.Locals(UserIDLocal, claims.UserID)
		return c.Next()
	}
}

// UserID lấy user id từ context. Trả "" nếu chưa qua JWT middleware.
func UserID(c *fiber.Ctx) string {
	if v, ok := c.Locals(UserIDLocal).(string); ok {
		return v
	}
	return ""
}
