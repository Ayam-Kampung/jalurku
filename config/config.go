package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Memuat nilai variabel lingkungan
func LoadConfig() {
	appEnv := os.Getenv("APP_ENV")
	
	if appEnv == "production" {
		// Production: tidak perlu load .env, langsung pakai variabel lingkungan
		log.Println("🌐 Berjalan di mode PRODUKSI - menggunakan variabel lingkungan cloud")
		return
	}

	// Development: load dari .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Tidak ada .env, sebagai gantinya menggunakan variabel lingkungan sistem")
	} else {
		log.Println("✅ Konfigurasi variabel lingkungan termuat")
	}
}

// Mengambil nilai variabel, nilai ke nilai
func Config(key string) string {
	return os.Getenv(key)
}

// Jika tidak ada nilai variabel lingkungan, maka gunakan defaultValue
func ConfigWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Apakah aplikasi berjalan di lingkungan produksi?
func IsProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}

// Apakah aplikasi berjalan di lingkungan pengembangan?
func IsDevelopment() bool {
	return os.Getenv("APP_ENV") != "production"
}
