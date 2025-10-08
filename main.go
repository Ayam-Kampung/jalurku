package main

import (
	"jalurku/app/db"
	"jalurku/app/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// Berkas Template seperti .html
	// Template digunakan untuk patokan isi konten yang akan ditampilkan
	// Templatenya nanti akan bisa diubah isinya
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/", "./static")

	app.Use(cors.New())

	database.ConnectDB()

	router.SetupRoutes(app)

	app.Listen(":3000")
}
