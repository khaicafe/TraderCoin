package main

import (
	"log"
	"os"

	"tradercoin/backend/config"
	"tradercoin/backend/database"
	"tradercoin/backend/middleware"
	api "tradercoin/backend/routes"
	"tradercoin/backend/services"
	"tradercoin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize encryption key for API credentials
	utils.InitEncryptionKey(cfg.EncryptionKey)
	log.Println("Encryption initialized for API credentials")

	// Initialize database with GORM
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying SQL DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	defer sqlDB.Close()

	// Initialize Redis (optional)
	redisClient := database.InitRedis()
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Seed sample data (admin and user accounts)
	if err := database.SeedData(db); err != nil {
		log.Println("Warning: Failed to seed sample data:", err)
	}

	// Initialize services
	services := &services.Services{
		DB:    db,
		Redis: redisClient,
	}

	// Setup Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimiter())

	// API routes
	api.SetupRoutes(router, services)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Display configuration info
	log.Println("========================================")
	log.Printf("🚀 Server Configuration:")
	log.Printf("   - Port: %s", port)
	log.Printf("   - Database: %s", cfg.DBType)
	if cfg.DBType == "sqlite" {
		log.Printf("   - SQLite Path: %s", cfg.DBPath)
	} else if cfg.DBType == "postgresql" {
		log.Printf("   - PostgreSQL Host: %s:%s", cfg.PostgresHost, cfg.PostgresPort)
		log.Printf("   - PostgreSQL DB: %s", cfg.PostgresDB)
	}
	log.Println("========================================")

	log.Printf("🌐 Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
