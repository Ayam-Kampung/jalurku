package router

import (
	"jalurku/app/handlers"
	"jalurku/app/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())

	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)

	// User
	user := api.Group("/user")
	user.Get("/:id", handlers.GetUser)
	user.Post("/", handlers.CreateUser)
	user.Patch("/:id", middleware.Protected(), handlers.UpdateUser)
	user.Delete("/:id", middleware.Protected(), handlers.DeleteUser)

	// Rute / (root)
	// Kita mengubah {{.Title}} menjadi "Website Sekolah"
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Ayam Kampung—Beranda",
		}, "layouts/main")
	})

	// RencanaKu
	rencanaku := app.Group("/rencanaku")
	rencanaku.Get("/", func(c *fiber.Ctx) error {
		return c.Render("rencanaku", fiber.Map{
			"Title": "Ayam Kampung—RencanaKu",
		}, "layouts/main")
	})

	rencanaku.Get("/tja", func(c *fiber.Ctx) error {
		return c.Render("tja", fiber.Map{
			"Title": "Ayam Kampung—RencanaKu",
		}, "layouts/main")
	})

	rencanaku.Get("/tkj", func(c *fiber.Ctx) error {
		return c.Render("tkj", fiber.Map{
			"Title": "Ayam Kampung—RencanaKu",
		}, "layouts/main")
	})

	rencanaku.Get("/pg", func(c *fiber.Ctx) error {
		return c.Render("pg", fiber.Map{
			"Title": "Ayam Kampung—RencanaKu",
		}, "layouts/main")
	})

	rencanaku.Get("/rpl", func(c *fiber.Ctx) error {
		return c.Render("rpl", fiber.Map{
			"Title": "Ayam Kampung—RencanaKu",
		}, "layouts/main")
	})

	// 404 HANDLER
	app.Use(func(c *fiber.Ctx) error {
		return c.Render("404", fiber.Map{
			"Title": "Ayam Kampung—404",
		}, "layouts/main")
	})
}