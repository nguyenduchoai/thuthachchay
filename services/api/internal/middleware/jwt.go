// Package middleware tách JWT-auth middleware ra để các handler khác chỉ phụ thuộc Service.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/buocvang/api/internal/auth"
)

const UserIDLocal = "user_id"

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

// UserID lấy user id từ context. Trả "" nếu chưa qua JWT middleware.
func UserID(c *fiber.Ctx) string {
	if v, ok := c.Locals(UserIDLocal).(string); ok {
		return v
	}
	return ""
}
