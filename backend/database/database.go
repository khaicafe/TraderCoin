package database

import (
	"context"
	"database/sql"
	"fmt"
	"tradercoin/backend/config"

	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

func Connect() (*sql.DB, error) {
	cfg := config.Load()

	// SQLite database file
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = "./tradercoin.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

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

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(255),
			phone VARCHAR(50),
			status VARCHAR(50) DEFAULT 'active',
			subscription_end TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS exchange_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			exchange VARCHAR(50) NOT NULL,
			api_key VARCHAR(255) NOT NULL,
			api_secret VARCHAR(255) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, exchange)
		)`,

		`CREATE TABLE IF NOT EXISTS trading_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			exchange VARCHAR(50) NOT NULL,
			symbol VARCHAR(50) NOT NULL,
			stop_loss_percent DECIMAL(10, 2),
			take_profit_percent DECIMAL(10, 2),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			exchange VARCHAR(50) NOT NULL,
			symbol VARCHAR(50) NOT NULL,
			order_id VARCHAR(255),
			side VARCHAR(10) NOT NULL,
			type VARCHAR(20) NOT NULL,
			quantity DECIMAL(20, 8),
			price DECIMAL(20, 8),
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			amount DECIMAL(10, 2) NOT NULL,
			type VARCHAR(50) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS admins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(255),
			role VARCHAR(50) DEFAULT 'admin',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}
