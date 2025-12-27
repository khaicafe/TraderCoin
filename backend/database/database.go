package database

import (
	"context"
	"fmt"
	"log"
	"tradercoin/backend/config"
	"tradercoin/backend/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
	cfg := config.Load()

	var db *gorm.DB
	var err error

	// GORM config
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	switch cfg.DBType {
	case "postgresql":
		// PostgreSQL connection
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresDB,
			cfg.PostgresSSLMode,
		)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
		}
		log.Println("🐘 Using PostgreSQL database")

	case "sqlite":
		// SQLite connection
		dbPath := cfg.DBPath
		if dbPath == "" {
			dbPath = "./tradercoin.db"
		}
		db, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
		}
		log.Println("📦 Using SQLite database")

	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: sqlite, postgresql)", cfg.DBType)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("✅ Database connected successfully")
	return db, nil
}

func InitRedis() *redis.Client {
	cfg := config.Load()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		// Redis is optional - log warning and return nil
		fmt.Printf("⚠️  Warning: Redis not available: %v\n", err)
		fmt.Println("ℹ️  System will run without Redis caching")
		return nil
	}

	fmt.Println("✅ Redis connected successfully")
	return client
}

func RunMigrations(db *gorm.DB) error {
	// Auto migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.ExchangeKey{},
		&models.TradingConfig{},
		&models.Order{},
		&models.Transaction{},
		&models.Admin{},
		&models.TradingSignal{},
		&models.WebhookPrefix{},
		&models.SystemLog{},
		&models.ExchangeAPIConfig{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed")
	return nil
}
