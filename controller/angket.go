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
		"message":    "angket_started",
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
		"message": "jawaban_saved_temp",
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
		"message": "angket_finished",
		"hasil": fiber.Map{
			"session_id":       	req.SessionID,
			"jurusan_terbaik":  	jurusanTerbaik.Name,
			"jurusan_terbaik_id":  	jurusanTerbaik.ID,
			"total_skor":       	maxScore,
			"detail_skor":      	skorJurusan,
		},
	})
}

// GET: Dapatkan banyak pertanyaan secara acak
func GetRandPertanyaans(c *fiber.Ctx) error {
	db := database.DB
	var pertanyaan []model.Pertanyaan

	if err := db.Preload("Jurusan").
		Order("RANDOM()").
		Find(&pertanyaan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_fetched_allrand_failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "pertanyaan_fetched_allrand",
		"data":    pertanyaan,
	})
}

// GET: Dapatkan banyak pertanyaan secara urut
func GetPertanyaans(c *fiber.Ctx) error {
	db := database.DB
	var pertanyaan []model.Pertanyaan

	if err := db.Preload("Jurusan").
		Find(&pertanyaan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_fetched_all_failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "pertanyaan_fetched_all",
		"data":    pertanyaan,
	})
}

func GetJurusan(c *fiber.Ctx) error {
	db := database.DB
	var jurusan []model.Jurusan

	if err := db.Find(&jurusan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "jurusan_fetched_all_failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "jurusan_fetched_all",
		"data":    jurusan,
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
			"message": "pertanyaan_fetched_failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "pertanyaan_fetched",
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
			"message": "pertanyaan_create_input_failed",
			"error":   err.Error(),
		})
	}

	if input.Text == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_input_insufficient",
		})
	}

	if input.ID == uuid.Nil {
		input.ID = uuid.New()
	}

	if err := db.Preload("Jurusan").Create(&input).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_create_failed",
			"error":   err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "pertanyaan_create",
		"data":    input,
	})
}

type BulkPertanyaanInput struct {
	Pertanyaan []model.Pertanyaan `json:"pertanyaan"`
}

func CreatePertanyaanBulk(c *fiber.Ctx) error {
	var input BulkPertanyaanInput
	db := database.DB

	// Parse input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_bulk_create_input_failed",
			"error":   err.Error(),
		})
	}

	// Validasi: harus ada tepat 4 pertanyaan
	if len(input.Pertanyaan) != 4 {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_must_be_4",
			"error":   "Harus ada tepat 4 pertanyaan (satu untuk setiap jurusan)",
		})
	}

	// Ambil semua jurusan yang tersedia
	var jurusanList []model.Jurusan
	if err := db.Find(&jurusanList).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "jurusan_fetch_failed",
			"error":   err.Error(),
		})
	}

	// Validasi: harus ada tepat 4 jurusan
	if len(jurusanList) != 4 {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "jurusan_count_invalid",
			"error":   "Database harus memiliki tepat 4 jurusan",
		})
	}

	// Buat map untuk track jurusan yang sudah diisi
	jurusanMap := make(map[int]bool)
	for _, j := range jurusanList {
		jurusanMap[j.ID] = false
	}

	// Validasi setiap pertanyaan
	var validatedPertanyaan []model.Pertanyaan
	for i, p := range input.Pertanyaan {
		// Validasi text tidak kosong
		if p.Text == "" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "pertanyaan_text_empty",
				"error":   "Pertanyaan ke-" + string(rune(i+1)) + " tidak boleh kosong",
			})
		}

		// Validasi jurusan_id valid
		if _, exists := jurusanMap[p.JurusanID]; !exists {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "jurusan_id_invalid",
				"error":   "Jurusan ID tidak valid: " + string(rune(p.JurusanID)),
			})
		}

		// Validasi tidak ada duplikat jurusan_id
		if jurusanMap[p.JurusanID] {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "jurusan_duplicate",
				"error":   "Jurusan ID duplikat ditemukan: " + string(rune(p.JurusanID)),
			})
		}

		// Mark jurusan sebagai sudah diisi
		jurusanMap[p.JurusanID] = true

		// Generate UUID jika belum ada
		if p.ID == uuid.Nil {
			p.ID = uuid.New()
		}

		validatedPertanyaan = append(validatedPertanyaan, p)
	}

	// Validasi semua jurusan sudah terisi
	for jurusanID, filled := range jurusanMap {
		if !filled {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "jurusan_missing",
				"error":   "Pertanyaan untuk Jurusan ID " + string(rune(jurusanID)) + " belum diisi",
			})
		}
	}

	// Simpan semua pertanyaan dalam satu transaksi
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i := range validatedPertanyaan {
		if err := tx.Create(&validatedPertanyaan[i]).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "pertanyaan_create_failed",
				"error":   err.Error(),
			})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_commit_failed",
			"error":   err.Error(),
		})
	}

	// Reload dengan relasi Jurusan
	var createdPertanyaan []model.Pertanyaan
	ids := make([]uuid.UUID, len(validatedPertanyaan))
	for i, p := range validatedPertanyaan {
		ids[i] = p.ID
	}
	db.Preload("Jurusan").Where("id IN ?", ids).Find(&createdPertanyaan)

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "pertanyaan_bulk_create_success",
		"data":    createdPertanyaan,
		"count":   len(createdPertanyaan),
	})
}

func UpdatePertanyaan(c *fiber.Ctx) error {
	idParam := c.Params("id")

	// ✅ Validasi format UUID agar tidak error di query
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_update_uuid_failed",
		})
	}

	var pertanyaan model.Pertanyaan
	db := database.DB

	// ✅ Gunakan Where agar UUID diperlakukan sebagai string, bukan angka
	if err := db.Preload("Jurusan").Where("id = ?", id).First(&pertanyaan).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_update_notfound",
			"error":   err.Error(),
		})
	}

	var updateData model.Pertanyaan
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_update_input_failed",
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
	// 👇 TAMBAHKAN INI
	pertanyaan.Meta = updateData.Meta     // Meta boleh kosong
	pertanyaan.Image = updateData.Image   // Image boleh kosong

	if err := db.Save(&pertanyaan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_update_failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "pertanyaan_update",
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
			"message": "pertanyaan_delete_uuid_failed",
		})
	}

	var pertanyaan model.Pertanyaan
	db := database.DB

	// ✅ Gunakan Where agar cocok untuk UUID
	if err := db.Preload("Jurusan").Where("id = ?", id).First(&pertanyaan).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_delete_notfound",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&pertanyaan).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "pertanyaan_delete_failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Pertanyaan berhasil dihapus",
	})
}