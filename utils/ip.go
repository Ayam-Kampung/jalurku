package utils

import (
    "strings"
    "github.com/gofiber/fiber/v2"
)

// Pake proxy cloudflare segala ohmaygad
func GetRealIP(c *fiber.Ctx) string {
    if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
        return cfIP
    }
    if xff := c.Get("X-Forwarded-For"); xff != "" {
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }
    return c.IP()
}