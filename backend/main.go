package main

import (
	"log"
	"os"

	"tradercoin/backend/database"
	"tradercoin/backend/middleware"
	api "tradercoin/backend/routes"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

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

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
