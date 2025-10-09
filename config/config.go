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
		// Production: tidak perlu load .env, langsung pakai environment variables
		log.Println("🌐 Running in PRODUCTION mode - using cloud environment variables")
		return
	}

	// Development: load dari .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	} else {
		log.Println("✅ Configuration loaded from .env file")
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
