package controller

import (
	"github.com/gofiber/fiber/v2"
)

// ============================================
// API STATUS
// ============================================

// Contoh API
func Hello(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "success", "message": "Halo, dari tim Ayam Kampung! ❤️", "data": "Untuk JHIC"})
}