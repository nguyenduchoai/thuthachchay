package auth

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

type loginReq struct {
	ZaloAccessToken string `json:"zalo_access_token"`
	Locale          string `json:"locale,omitempty"`
}

type loginResp struct {
	User         map[string]any `json:"user"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int            `json:"expires_in"`
}

func (h *Handler) LoginZalo(c *fiber.Ctx) error {
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "bad_request", "message": err.Error()}})
	}
	ctx := WithLoginContext(c.Context(), LoginContext{
		UserAgent: c.Get(fiber.HeaderUserAgent),
		IP:        c.IP(),
	})
	res, err := h.svc.LoginWithZalo(ctx, req.ZaloAccessToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "auth_failed", "message": err.Error()}})
	}
	return c.JSON(loginResp{
		User:         map[string]any{"id": res.UserID},
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    res.ExpiresIn,
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req refreshReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "bad_request"}})
	}
	ctx := WithLoginContext(c.Context(), LoginContext{
		UserAgent: c.Get(fiber.HeaderUserAgent),
		IP:        c.IP(),
	})
	res, err := h.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "refresh_failed", "message": err.Error()}})
	}
	return c.JSON(fiber.Map{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_in":    res.ExpiresIn,
	})
}

func (h *Handler) SignOut(c *fiber.Ctx) error {
	var req refreshReq
	if err := c.BodyParser(&req); err != nil {
		// best-effort sign-out, ignore parse error
	}
	_ = h.svc.SignOut(c.Context(), req.RefreshToken)
	return c.SendStatus(fiber.StatusNoContent)
}
