package middleware

import (
	"time"

    "jalurku/utils"
	"jalurku/database"

	"github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
)


// Konfigurasi limiter default
func DefaultLimiter() fiber.Handler {
    return limiter.New(limiter.Config{
		Max:        400,
		Expiration: 1 * time.Minute,
		KeyGenerator: utils.GetRealIP,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	})
}

// Strict limiter untuk endpoint sensitif
func EnforcedLimiter(max int, duration time.Duration, prefix string) fiber.Handler {
    return limiter.New(limiter.Config{
        Max:          max,
        Expiration:   duration,
        KeyGenerator: func(c *fiber.Ctx) string {
            if prefix != "" {
                return prefix + ":" + utils.GetRealIP(c)
            }
            return utils.GetRealIP(c)
        },
        Storage:    database.RedisStore(),
        LimitReached: func(c *fiber.Ctx) error {
            return c.SendStatus(fiber.StatusTooManyRequests)
        },
    })
}