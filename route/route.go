package route

import (
	"time"

	"jalurku/controller"
	"jalurku/middleware"
	"jalurku/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// SetupRoutes setup all application routes
func SetupRoutes(app *fiber.App) {

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))

	// API Group
	api := app.Group("/api")
	api.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))

	// Health check
	api.Get("/", controller.Hello)

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	auth.Post("/login", controller.Login)
	auth.Post("/register", controller.Register)

	// User routes
	user := api.Group("/user")
	user.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	user.Get("/:id", controller.GetUser)
	// Hanya pengguna terautentikasi
	user.Get("/me", middleware.Protected(), controller.GetCurrentUser)
	user.Put("/:id", middleware.Protected(), controller.UpdateUser)
	user.Delete("/:id", middleware.Protected(), controller.DeleteUser)

	angket := api.Group("/angket")
	angket.Use(middleware.Optional())
	angket.Post("/mulai", controller.StartAngket)
	angket.Post("/submit", controller.SubmitJawaban)
	angket.Post("/selesai", controller.FinishAngket)

	// Rute Pertanyaan
	pertanyaan := api.Group("/pertanyaan")
	pertanyaan.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	pertanyaan.Get("/", controller.GetPertanyaans)
	pertanyaan.Get("/:id", controller.GetPertanyaanByID)
	// Hanya Admin
	pertanyaan.Post("/", middleware.Protected(), middleware.AdminOnly(), controller.CreatePertanyaan)
	pertanyaan.Put("/:id", middleware.Protected(), middleware.AdminOnly(), controller.UpdatePertanyaan)
	pertanyaan.Delete("/:id", middleware.Protected(), middleware.AdminOnly(), controller.DeletePertanyaan)

	// Admin routes (protected + admin only)
	admin := api.Group("/admin", middleware.Protected(), middleware.AdminOnly())
	admin.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	// admin.Get("/dashboard", controller.GetAdminDashboard)                    // Admin dashboard
}