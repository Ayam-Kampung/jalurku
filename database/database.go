package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
)

// ConnectDB establishes database connection
func ConnectDB() {
	db, err := connectPostgres()
	if err != nil {
		log.Fatal("❌ Failed to connect to database: ", err)
	}

	DB = db
	log.Println("✅ Database connected successfully")
}

func connectPostgres() (*gorm.DB, error) {
	var dsn string

	// Prioritas DATABASE_URL (untuk cloud providers)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		dsn = databaseURL
		log.Println("📦 Using DATABASE_URL")
	} else {
		dsn = buildDSN()
		log.Println("📦 Using individual database env vars")
	}

	// Default ke logger.Info, override via env jika perlu
	logMode := logger.Info
	if os.Getenv("DB_LOG_MODE") == "error" {
		logMode = logger.Error
	}

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
}

// buildDSN membuat connection string dari env vars individual
func buildDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

// ConnectRedis menginisialisasi Redis Client
func ConnectRedis() {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DATABASE")

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Fatalf("❌ Failed to convert REDIS_DATABASE to int: %v", err)
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	// Inisialisasi Redis Client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test koneksi
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Redis connection failed to %s: %v", addr, err)
	}
	
	log.Printf("✅ Redis connected at %s (DB %d)", addr, db)
}
