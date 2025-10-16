package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig memuat variabel lingkungan
func LoadConfig() {
	appEnv := os.Getenv("APP_ENV")
	
	if appEnv == "production" {
		log.Println("🌐 Running in PRODUCTION mode - using cloud environment variables")
		return
	}

	// Development: load dari .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system environment variables")
	} else {
		log.Println("✅ Environment variables loaded from .env")
	}
}

// Config mengambil nilai variabel lingkungan
func Config(key string) string {
	return os.Getenv(key)
}

// ConfigWithDefault mengambil nilai variabel atau default jika kosong
func ConfigWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsProduction mengecek apakah aplikasi berjalan di produksi
func IsProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}

// IsDevelopment mengecek apakah aplikasi berjalan di development
func IsDevelopment() bool {
	return os.Getenv("APP_ENV") != "production"
}