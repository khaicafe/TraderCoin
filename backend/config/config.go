package config

import (
	"os"
	"time"
)

type Config struct {
	Port string

	// Database Settings
	DBType string // "sqlite" or "postgresql"

	// SQLite Configuration
	DBPath string

	// PostgreSQL Configuration
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// Encryption (for API keys and secrets)
	EncryptionKey string

	// Exchange APIs
	BinanceAPIURL string
	BittrexAPIURL string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),

		// =====================================================
		// 🔧 DATABASE CONFIGURATION - CHỌN Ở ĐÂY!
		// =====================================================
		// Đổi giá trị này để chọn database:
		// - "sqlite"      : SQLite (Development - nhanh, đơn giản)
		// - "postgresql"  : PostgreSQL (Production - mạnh mẽ, scale tốt)
		DBType: "sqlite", // 👈 THAY ĐỔI GIÁ TRỊ NÀY!

		// SQLite Configuration (dùng khi DBType = "sqlite")
		DBPath: getEnv("DB_PATH", "./tradercoin.db"),

		// PostgreSQL Configuration (dùng khi DBType = "postgresql")
		PostgresHost:     getEnv("DB_HOST", "localhost"),
		PostgresPort:     getEnv("DB_PORT", "5432"),
		PostgresUser:     getEnv("DB_USER", "tradercoin"),
		PostgresPassword: getEnv("DB_PASSWORD", "tradercoin123"),
		PostgresDB:       getEnv("DB_NAME", "tradercoin_db"),
		PostgresSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Redis (optional)
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,

		JWTSecret:     getEnv("JWT_SECRET", "your_secret_key"),
		JWTExpiration: 24 * time.Hour,

		// Encryption key must be 32 bytes for AES-256
		EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-byte-encryption-key-1234"),

		BinanceAPIURL: getEnv("BINANCE_API_URL", "https://api.binance.com"),
		BittrexAPIURL: getEnv("BITTREX_API_URL", "https://api.bittrex.com"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
