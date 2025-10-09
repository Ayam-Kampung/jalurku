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
	api := app.Group("/api", middleware.ApiKey())
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
		Max:        5,
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
	user.Get("/:id", controller.GetUser)                                    	// Public: get user by ID
	user.Get("/me", middleware.Protected(), controller.GetCurrentUser)       	// Protected: get current user
	user.Put("/:id", middleware.Protected(), controller.UpdateUser)          	// Protected: update user
	user.Delete("/:id", middleware.Protected(), controller.DeleteUser)       	// Protected: delete user

	// Category routes
	category := api.Group("/categories")
	category.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	category.Get("/", controller.GetCategories)                             	// Public: get all categories
	category.Get("/:id", controller.GetCategory)                             	// Public: get category by ID
	category.Post("/", middleware.Protected(), middleware.AdminOnly(), controller.CreateCategory)     	// Admin only
	category.Put("/:id", middleware.Protected(), middleware.AdminOnly(), controller.UpdateCategory)   	// Admin only
	category.Delete("/:id", middleware.Protected(), middleware.AdminOnly(), controller.DeleteCategory) 	// Admin only

	// Question routes
	question := api.Group("/questions")
	category.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	question.Get("/", controller.GetQuestions)                               // Public: get all questions
	question.Get("/:id", controller.GetQuestion)                             // Public: get question by ID
	question.Post("/", middleware.Protected(), middleware.AdminOnly(), controller.CreateQuestion)     	// Admin only
	question.Put("/:id", middleware.Protected(), middleware.AdminOnly(), controller.UpdateQuestion)   	// Admin only
	question.Delete("/:id", middleware.Protected(), middleware.AdminOnly(), controller.DeleteQuestion) 	// Admin only

	// Reflection routes (all protected)
	reflection := api.Group("/reflections", middleware.Protected())
	reflection.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	reflection.Get("/", controller.GetReflections)                           // Get user's reflections
	reflection.Post("/", controller.CreateReflection)                        // Create new reflection
	reflection.Get("/:id", controller.GetReflection)                         // Get reflection detail
	reflection.Post("/answer", controller.SubmitAnswer)                      // Submit answer
	reflection.Post("/:id/complete", controller.CompleteReflection)          // Complete reflection
	reflection.Get("/:id/report", controller.GetReflectionReport)            // Get reflection report
	reflection.Delete("/:id", controller.DeleteReflection)                   // Delete reflection

	// Statistics routes (protected)
	stats := api.Group("/statistics", middleware.Protected())
	category.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	stats.Get("/me", controller.GetUserStatistics)                           // Get user statistics
	stats.Get("/leaderboard", controller.GetLeaderboard)                     // Get leaderboard

	// Admin routes (protected + admin only)
	admin := api.Group("/admin", middleware.Protected(), middleware.AdminOnly())
	category.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		Storage:    database.RedisStore(),
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	}))
	admin.Get("/dashboard", controller.GetAdminDashboard)                    // Admin dashboard
}