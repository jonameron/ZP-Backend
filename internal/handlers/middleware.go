package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"zeitpass/internal/services"
)

// Admin email whitelist
var adminEmails = map[string]bool{
	"jstaude@11data.de": true,
}

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		email, _ := c.Locals("email").(string)
		if !adminEmails[email] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
		}
		return c.Next()
	}
}

func RequireAuth(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization header required",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization format",
			})
		}

		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("userDbID", claims.UserDbID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}
