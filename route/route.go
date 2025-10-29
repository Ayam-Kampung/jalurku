package route

import (
	"jalurku/controller"
	"jalurku/middleware"

	"github.com/gofiber/fiber/v2"
)

// Menyetel semua rute REST API jalurku
func SetupRoutes(app *fiber.App) {

	// Kelompokkan menjadi /api
	api := app.Group("/api")

	// Ini buat apa si yah?
	api.Get("/", controller.Hello)

	// Rute: Autentikasi
	auth := api.Group("/auth")
	auth.Post("/login", controller.Login)
	auth.Post("/register", controller.Register)

	// Rute: Jurusan
	jurus := api.Group("/jurusan")
	jurus.Get("/", controller.GetJurusan)

	// Rute: Pengguna
	user := api.Group("/user")
	user.Get("/me", middleware.Protected(), controller.GetCurrentUser)
	user.Put("/:id", middleware.Protected(), controller.UpdateUser)
	user.Delete("/:id", middleware.Protected(), controller.DeleteUser)

	// Rute: Angket
	angket := api.Group("/angket")
	angket.Use(middleware.Optional())
	angket.Post("/mulai", controller.StartAngket)
	angket.Post("/submit", controller.SubmitJawaban)
	angket.Post("/selesai", controller.FinishAngket)

	// Rute: Pertanyaan
	pertanyaan := api.Group("/pertanyaan")
	pertanyaan.Get("/rand", controller.GetRandPertanyaans)
	pertanyaan.Get("/", controller.GetPertanyaans)
	pertanyaan.Get("/:id", controller.GetPertanyaanByID)

	// Rute: ytta
	pertanyaan.Post("/", middleware.Protected(), middleware.AdminOnly(), controller.CreatePertanyaan)
	pertanyaan.Put("/:id", middleware.Protected(), middleware.AdminOnly(), controller.UpdatePertanyaan)
	pertanyaan.Delete("/:id", middleware.Protected(), middleware.AdminOnly(), controller.DeletePertanyaan)

	// Rute: ngecek yang ytta
	admin := api.Group("/admin", middleware.Protected(), middleware.AdminOnly())
	admin.Get("/check", controller.IsAdmin)
}