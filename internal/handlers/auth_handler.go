package handlers

import (
	"github.com/gofiber/fiber/v2"
	"zeitpass/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(as *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

func (h *AuthHandler) RequestMagicLink(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil || input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
		})
	}

	// Always return 200 to prevent email enumeration
	_ = h.authService.SendMagicLink(input.Email)

	return c.JSON(fiber.Map{
		"message": "If this email is associated with an account, a sign-in link has been sent.",
	})
}

func (h *AuthHandler) VerifyMagicLink(c *fiber.Ctx) error {
	var input struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&input); err != nil || input.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "token is required",
		})
	}

	result, err := h.authService.VerifyMagicLink(input.Token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired token",
		})
	}

	return c.JSON(result)
}
