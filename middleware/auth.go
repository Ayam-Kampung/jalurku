package middleware

import (
	"jalurku/config"
	"os"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/golang-jwt/jwt/v5"
)

// Pengguna harus terdaftar dan terautentikasi melalui JWT
func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:  
		jwtware.SigningKey{
			Key: []byte(config.Config("SECRET")),
		},
		ErrorHandler: jwtError,
		ContextKey:   "user",
		// Sementara menggunaakan Token Bearer
		TokenLookup:  "header:Authorization",
		AuthScheme:   "Bearer",
	})
}

// Error handler JWT buatan
func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing or malformed JWT",
			"data":    nil,
		})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"status":  "error",
		"message": "Invalid or expired JWT",
		"data":    nil,
	})
}

// Apakah user memiliki role admin?
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		role := claims["role"].(string)

		// Mengecek apakah user memiliki role admin atau tidak
		if role != "admin" {
			return c.SendStatus(fiber.StatusForbidden)
		}

		return c.Next()
	}
}

// NOTE UNTUK FRONTEND: Harus menyediakan X-API-KEY di header
func ApiKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientKey := c.Get("X-API-Key")
		serverKey := os.Getenv("API_KEY")

		if clientKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "error",
				"error": "Missing API key",
				"data": nil,
			})
		}

		if clientKey != serverKey {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status": "error",
				"error": "Invalid API key",
				"data": nil,
			})
		}

		return c.Next()
	}
}