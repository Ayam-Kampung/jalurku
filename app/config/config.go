package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config func to get env value
func Config(key string) string {
	// load .env file
    if os.Getenv("APP_ENV") != "production" {
        if err := godotenv.Load(".env"); err != nil {
            fmt.Println("Warning: .env file not found")
        }
    }
	return os.Getenv(key)
}