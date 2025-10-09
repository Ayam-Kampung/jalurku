package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"context"

	"github.com/redis/go-redis/v9"
	redisStorage "github.com/gofiber/storage/redis/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the database connection
var DB *gorm.DB

// ConnectDB establishes database connection
func ConnectDB() {
	var err error
	dbDriver := os.Getenv("DB_DRIVER")
	appEnv := os.Getenv("APP_ENV")
	
	if dbDriver == "" {
		dbDriver = "sqlite" // Default to SQLite for easy setup
	}

	// Log environment
	if appEnv == "production" {
		log.Println("🌐 Environment: PRODUCTION (using cloud env variables)")
	} else {
		log.Println("💻 Environment: DEVELOPMENT (using .env file)")
	}

	DB, err = connectPostgres()

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("✅ Database connected successfully")
}

func connectPostgres() (*gorm.DB, error) {
	var dsn string
	appEnv := os.Getenv("APP_ENV")

	if appEnv == "production" {
		// Production: gunakan DATABASE_URL dari cloud (Railway, Heroku, dll)
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL != "" {
			dsn = databaseURL
			log.Println("📦 Using DATABASE_URL from cloud environment")
		} else {
			// Fallback: compose from individual env vars
			host := os.Getenv("DB_HOST")
			port := os.Getenv("DB_PORT")
			user := os.Getenv("DB_USER")
			password := os.Getenv("DB_PASSWORD")
			dbname := os.Getenv("DB_NAME")
			sslmode := os.Getenv("DB_SSLMODE")
			
			if sslmode == "" {
				sslmode = "require" // Default untuk production
			}

			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
				host, port, user, password, dbname, sslmode)
			log.Println("📦 Using composed connection string from cloud env")
		}
	} else {
		// Development: gunakan .env file
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		if sslmode == "" {
			sslmode = "disable" // Default untuk development
		}

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
		log.Println("📄 Using .env file for database connection")
	}

	// Set logger mode based on environment
	logMode := logger.Info
	if appEnv == "production" {
		logMode = logger.Error // Only log errors in production
	}

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
}

// Caching menggunakan Redis biar keren
var RedisClient *redis.Client

func RedisStore() *redisStorage.Storage {
    host := os.Getenv("REDIS_HOST")
    port := os.Getenv("REDIS_PORT")
    password := os.Getenv("REDIS_PASSWORD")
    dbStr := os.Getenv("REDIS_DATABASE")

    db, err := strconv.Atoi(dbStr)
    if err != nil {
        log.Fatalf("❌ Gagal mengonversi REDIS_DATABASE: %v", err)
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
        log.Printf("❌ Redis gagal terkoneksi ke %s: %v", addr, err)
    } else {
        log.Printf("✅ Redis terkoneksi di %s (DB %d)", addr, db)
    }

    // Return Fiber Storage
    portInt, _ := strconv.Atoi(port)
    return redisStorage.New(redisStorage.Config{
        Host:     host,
        Port:     portInt,
        Password: password,
        Database: db,
    })
}