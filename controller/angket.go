package controller

import (
	"jalurku/database"
	"jalurku/model"

	"github.com/gofiber/fiber/v2"
)

// GET: Dapatkan banyak pertanyaan
func GetPertanyaans(c *fiber.Ctx) error {
	var pertanyaans []model.Pertanyaan
	db := database.DB

	if err := db.Preload("Jurusan").Find(&pertanyaans).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data pertanyaan",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Berhasil mengambil semua pertanyaan",
		"data":    pertanyaans,
	})
}

// GET: Dapatkan satu pertanyaan
func GetPertanyaanByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var pertanyaan model.Pertanyaan
	db := database.DB

	if err := db.Preload("Jurusan").First(&pertanyaan, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Pertanyaan tidak ditemukan",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Berhasil mengambil data pertanyaan",
		"data":    pertanyaan,
	})
}

// POST: Membuat pertanyaan
func CreatePertanyaan(c *fiber.Ctx) error {
	var input model.Pertanyaan
	db := database.DB

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membaca input",
			"error":   err.Error(),
		})
	}

	if input.Text == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Teks dan JurusanID wajib diisi",
		})
	}

	if err := db.Create(&input).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membuat pertanyaan",
			"error":   err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Pertanyaan berhasil dibuat",
		"data":    input,
	})
}

// PUT: Memperbarui pertanyaan
func UpdatePertanyaan(c *fiber.Ctx) error {
	id := c.Params("id")
	var pertanyaan model.Pertanyaan
	db := database.DB

	if err := db.First(&pertanyaan, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Pertanyaan tidak ditemukan",
			"error":   err.Error(),
		})
	}

	var updateData model.Pertanyaan
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membaca input",
			"error":   err.Error(),
		})
	}

	pertanyaan.Text = updateData.Text
	pertanyaan.JurusanID = updateData.JurusanID

	if err := db.Save(&pertanyaan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal memperbarui pertanyaan",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Pertanyaan berhasil diperbarui",
		"data":    pertanyaan,
	})
}

// DELETE: Menghapus pertanyaan
func DeletePertanyaan(c *fiber.Ctx) error {
	id := c.Params("id")
	var pertanyaan model.Pertanyaan
	db := database.DB

	if err := db.First(&pertanyaan, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Pertanyaan tidak ditemukan",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&pertanyaan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menghapus pertanyaan",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Pertanyaan berhasil dihapus",
	})
}
