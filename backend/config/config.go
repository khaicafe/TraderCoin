package config

import (
	"os"
	"time"
)

type Config struct {
	Port string

	// Database - SQLite
	DBPath string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// Exchange APIs
	BinanceAPIURL string
	BittrexAPIURL string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),

		DBPath: getEnv("DB_PATH", "./tradercoin.db"),

		// RedisHost:     getEnv("REDIS_HOST", "localhost"),
		// RedisPort:     getEnv("REDIS_PORT", "6379"),
		// RedisPassword: getEnv("REDIS_PASSWORD", ""),
		// RedisDB:       0,

		JWTSecret:     getEnv("JWT_SECRET", "your_secret_key"),
		JWTExpiration: 24 * time.Hour,

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
