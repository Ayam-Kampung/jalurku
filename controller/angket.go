package controller

import (
	"fmt"
	"jalurku/database"
	"jalurku/model"
	"math/rand"
	"time"

	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Membuat sesi angket baru, dan disimpan di Redis.
// Sesi akan hilang, jika tidak digunakan dalam jangka waktu 1 jam
func StartAngket(c *fiber.Ctx) error {
	sessionID := uuid.New().String()

	ctx := context.Background()
	key := fmt.Sprintf("session:%s:started", sessionID)

	// simpan di Redis (berlaku 1 jam)
	if err := database.RedisClient.Set(ctx, key, true, time.Hour).Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal membuat session"})
	}

	return c.JSON(fiber.Map{
		"message":    "Session angket dimulai",
		"session_id": sessionID,
	})
}

// Submit jawaban ke Redis terlebih dahulu.
func SubmitJawaban(c *fiber.Ctx) error {
	var req model.SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	ctx := context.Background()

	// Apakah sesi valid?
	sessionKey := fmt.Sprintf("session:%s:started", req.SessionID)
	exists, err := database.RedisClient.Exists(ctx, sessionKey).Result()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal memeriksa session"})
	}
	if exists == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "session tidak valid atau sudah expired"})
	}

	// Apakah pertanyaannya valid?
	var q model.Pertanyaan
	if err := database.DB.Where("id = ?", req.QuestionID).First(&q).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "pertanyaan tidak ditemukan"})
	}

	// 💾 Simpan jawaban 
	dataKey := fmt.Sprintf("session:%s:started", req.SessionID)
	existing, _ := database.RedisClient.Get(ctx, dataKey).Result()
	var sessionData []model.SubmitRequest
	if existing != "" {
		json.Unmarshal([]byte(existing), &sessionData)
	}
	sessionData = append(sessionData, req)
	jsonData, _ := json.Marshal(sessionData)

	if err := database.RedisClient.Set(ctx, dataKey, jsonData, time.Hour).Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal menyimpan jawaban"})
	}

	// ⏱️ Perpanjang juga TTL session utama
	database.RedisClient.Expire(ctx, sessionKey, time.Hour)

	return c.JSON(fiber.Map{
		"message": "Jawaban tersimpan dan session diperpanjang",
		"data":    req,
	})
}

// Jika semua angket sudah terjawab,
// maka hapus sesi angketnya,
// dan hitung skor angketnya
func FinishAngket(c *fiber.Ctx) error {
	type FinishRequest struct {
		SessionID string `json:"session_id"`
	}

	var req FinishRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	ctx := context.Background()
	key := fmt.Sprintf("session:%s:started", req.SessionID)

	// Ambil semua jawaban dari Redis
	data, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "data sesi tidak ditemukan"})
	}

	var answers []model.SubmitRequest
	json.Unmarshal([]byte(data), &answers)

	// Map jurusan_id -> total skor
	skorJurusan := make(map[int]int)

	for _, ans := range answers {
		var p model.Pertanyaan
		if err := database.DB.First(&p, "id = ?", ans.QuestionID).Error; err != nil {
			continue
		}
		skorJurusan[p.JurusanID] += ans.SelectedOption
	}

	// Jika tidak ada data valid
	if len(skorJurusan) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "tidak ada jawaban valid"})
	}

	// Cari skor tertinggi
	maxScore := 0
	for _, total := range skorJurusan {
		if total > maxScore {
			maxScore = total
		}
	}

	// Ambil semua jurusan yang punya skor tertinggi
	var kandidat []int
	for jurusanID, total := range skorJurusan {
		if total == maxScore {
			kandidat = append(kandidat, jurusanID)
		}
	}

	// Pilih satu jurusan secara random dari yang seri
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	chosenJurusanID := kandidat[r.Intn(len(kandidat))]

	// Ambil nama jurusan terbaik
	var jurusanTerbaik model.Jurusan
	database.DB.First(&jurusanTerbaik, "id = ?", chosenJurusanID)

	// Hapus Redis
	if err := database.RedisClient.Del(ctx, key).Err(); err != nil {
    	fmt.Println("⚠️ gagal menghapus redis key:", err)
	}

	// 🔐 Cek apakah user login
	var userID uuid.UUID
	userToken := c.Locals("user")
	if userToken != nil {
		token := userToken.(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		uidStr, ok := claims["user_id"].(string)
		if ok {
			parsedID, err := uuid.Parse(uidStr)
			if err == nil {
				userID = parsedID
			}
		}
	}

	// 💾 Jika user login, simpan hasil ke tabel HasilAngket
	if userID != uuid.Nil {
		has := model.HasilAngket{
			ID:        uuid.New(),
			UserID:    userID,
			JurusanID: chosenJurusanID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := database.DB.Create(&has).Error; err != nil {
			fmt.Println("⚠️ Gagal menyimpan hasil angket:", err)
		}
	}

	fmt.Println("🧠 Akan simpan hasil untuk user:", userID)
	fmt.Println("🎯 Jurusan terbaik:", chosenJurusanID)

	// Kirim hasil akhir
	return c.JSON(fiber.Map{
		"message": "Angket selesai 🎯",
		"hasil": fiber.Map{
			"session_id":          req.SessionID,
			"jurusan_terbaik":  jurusanTerbaik.Name,
			"total_skor":       maxScore,
			"detail_skor":      skorJurusan,
		},
	})
}

// GET: Dapatkan banyak pertanyaan
func GetPertanyaans(c *fiber.Ctx) error {
	var ids []uuid.UUID // atau []int, tergantung tipe ID kamu

	db := database.DB

	// Ambil hanya kolom id dan acak urutannya
	if err := db.Model(&model.Pertanyaan{}).
		Select("id").
		Order("RANDOM()").
		Pluck("id", &ids).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data pertanyaan",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Berhasil mengambil semua ID pertanyaan (acak)",
		"data":    ids,
	})
}


// GET: Dapatkan satu pertanyaan
func GetPertanyaanByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var pertanyaan model.Pertanyaan
	db := database.DB

	if err := db.Preload("Jurusan").
		Where("id = ?", id).
		First(&pertanyaan).Error; err != nil {

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

	if input.ID == uuid.Nil {
		input.ID = uuid.New()
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

func UpdatePertanyaan(c *fiber.Ctx) error {
	idParam := c.Params("id")

	// ✅ Validasi format UUID agar tidak error di query
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Format ID tidak valid (bukan UUID)",
		})
	}

	var pertanyaan model.Pertanyaan
	db := database.DB

	// ✅ Gunakan Where agar UUID diperlakukan sebagai string, bukan angka
	if err := db.Where("id = ?", id).First(&pertanyaan).Error; err != nil {
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

	// Update field yang boleh diubah
	if updateData.Text != "" {
		pertanyaan.Text = updateData.Text
	}
	if updateData.JurusanID != 0 {
		pertanyaan.JurusanID = updateData.JurusanID
	}

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
	idParam := c.Params("id")

	// ✅ Validasi UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Format ID tidak valid (bukan UUID)",
		})
	}

	var pertanyaan model.Pertanyaan
	db := database.DB

	// ✅ Gunakan Where agar cocok untuk UUID
	if err := db.Where("id = ?", id).First(&pertanyaan).Error; err != nil {
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