package middleware

import (
	"jalurku/config"
	"os"
	"strings"

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

func Optional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenStr string

		// 🔍 1. Coba ambil dari Authorization header
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 🍪 2. Kalau gak ada, coba ambil dari cookie "token"
		if tokenStr == "" {
			tokenStr = c.Cookies("token")
		}

		// ❌ Kalau dua-duanya kosong → lanjut (guest)
		if tokenStr == "" {
			return c.Next()
		}

		// ✅ 3. Parse token kalau ada
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("SECRET")), nil
		})

		// Jika token valid → simpan di context
		if err == nil && token.Valid {
			c.Locals("user", token)
		}

		// Apapun hasilnya (valid / invalid / kosong) tetap lanjut
		return c.Next()
	}
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