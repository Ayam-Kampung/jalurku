package main

import (
	"github.com/gofiber/fiber/v2"
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

	// Menyediakan file static seperti .css, .js, dll
	app.Static("/", "./static")

	// Rute / (root)
	// Kita mengubah {{.Title}} menjadi "Website Sekolah"
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Ayam Kampung—Beranda",
		}, "layouts/main")
	})

	app.Get("/rencanaku", func(c *fiber.Ctx) error {
		return c.Render("rencanaku", fiber.Map{
			"Title": "Ayam Kampung—RencanaKu",
		}, "layouts/main")
	})

	app.Get("/rencanaku/rpl", func(c *fiber.Ctx) error {
		return c.Render("rpl", fiber.Map{
			"Title": "Ayam Kampung—RencanaKu",
		}, "layouts/main")
	})

	app.Use(func(c *fiber.Ctx) error {
        return c.Render("404", fiber.Map{
			"Title": "Ayam Kampung—404",
		}, "layouts/main")
    })

	app.Listen(":3000")
}
