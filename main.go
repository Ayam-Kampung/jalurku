package main

import (
	"flag"
	"log"
	"os"

	"jalurku/config"
	"jalurku/database"
	"jalurku/model"
	"jalurku/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Command line flags
	// seed := flag.Bool("seed", false, "Run database seeder")
	flag.Parse()

	// Load config (will auto-detect production vs development)
	config.LoadConfig()

	// Connect to database
	database.ConnectDB()

	// Auto migrate model
	err := database.DB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Question{},
		&model.Option{},
		&model.Reflection{},
		&model.Answer{},
		&model.CategoryScore{},
		&model.Recommendation{},
		&model.ScoreThreshold{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	log.Println("✅ Database migration completed")

	// // Run seeder if flag is set
	// if *seed {
	// 	database.SeedDatabase()
	// 	return
	// }

	// Create Fiber app with production-ready config
	app := fiber.New(fiber.Config{
		ErrorHandler:          customErrorHandler,
		AppName:               "jalurku v1.0",
		DisableStartupMessage: config.IsProduction(), // Disable startup message in production
		ServerHeader:          "Fiber",
		StrictRouting:         false,
		CaseSensitive:         false,
		BodyLimit:             4 * 1024 * 1024, // 4MB
	})

	// Middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: config.IsDevelopment(),
	}))

	// Logger middleware (different format for production)
	if config.IsProduction() {
		app.Use(logger.New(logger.Config{
			Format: "${time} | ${status} | ${latency} | ${method} ${path}\n",
		}))
	} else {
		app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} ${latency}\n",
		}))
	}

	// CORS middleware
	corsConfig := cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}

	// Production CORS (more restrictive)
	if config.IsProduction() {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins != "" {
			corsConfig.AllowOrigins = allowedOrigins
		}
	}

	app.Use(cors.New(corsConfig))

	// Setup routes
	routes.SetupRoutes(app)

	// Health check endpoint (useful for cloud platforms)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":      "ok",
			"environment": config.ConfigWithDefault("APP_ENV", "development"),
			"version":     "1.0.0",
		})
	})

	// Start server
	port := config.ConfigWithDefault("PORT", "3000")

	if config.IsProduction() {
		log.Printf("🚀 Server running in PRODUCTION mode on port %s", port)
	} else {
		log.Printf("🚀 Server running in DEVELOPMENT mode on http://localhost:%s", port)
		log.Println("📚 API Documentation: http://localhost:" + port + "/api")
	}

	log.Fatal(app.Listen(":" + port))
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error in production
	if config.IsProduction() {
		log.Printf("ERROR: %v\n", err)
	}

	return c.Status(code).JSON(fiber.Map{
		"status":  "error",
		"message": message,
		"data":    nil,
	})
}