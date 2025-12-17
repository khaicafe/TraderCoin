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

	// Exchange Configurations
	Exchanges ExchangeConfig
}

// ExchangeConfig holds all exchange API and WebSocket configurations
type ExchangeConfig struct {
	Binance BinanceConfig
	OKX     OKXConfig
	Bybit   BybitConfig
	Kraken  KrakenConfig
	Bittrex BittrexConfig
}

// BinanceConfig for Binance exchange
type BinanceConfig struct {
	// Production URLs
	SpotAPIURL    string
	FuturesAPIURL string
	SpotWSURL     string
	FuturesWSURL  string

	// Testnet URLs
	TestnetSpotAPIURL    string
	TestnetFuturesAPIURL string
	TestnetSpotWSURL     string
	TestnetFuturesWSURL  string

	// Ticker API endpoints (for real-time price)
	SpotTickerAPI    string // e.g., /api/v3/ticker/price
	FuturesTickerAPI string // e.g., /fapi/v1/ticker/price
}

// OKXConfig for OKX exchange
type OKXConfig struct {
	APIURL string
	WSURL  string
}

// BybitConfig for Bybit exchange
type BybitConfig struct {
	APIURL string
	WSURL  string
}

// KrakenConfig for Kraken exchange
type KrakenConfig struct {
	APIURL    string
	WSURL     string
	WSAuthURL string
}

// BittrexConfig for Bittrex exchange
type BittrexConfig struct {
	APIURL string
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

		// Exchange Configurations
		Exchanges: ExchangeConfig{
			Binance: BinanceConfig{
				// Production
				// SpotAPIURL:    "https://api.binance.com",
				// FuturesAPIURL: "https://fapi.binance.com",
				// SpotWSURL:     "wss://stream.binance.com:9443/ws",
				// FuturesWSURL:  "wss://fstream.binance.com/ws",

				// Testnet - Using correct URLs from Binance docs
				// Base APIs for REST
				SpotAPIURL:           "https://testnet.binance.vision",
				FuturesAPIURL:        "https://testnet.binancefuture.com",
				// Market data WS (combined streams). For user data WS use TestnetSpotWSURL/TestnetFuturesWSURL
				SpotWSURL:            "wss://stream.testnet.binance.vision/ws",
				FuturesWSURL:         "wss://stream.binancefuture.com/ws",
				// Explicit testnet fields (used by adapters when isTestnet=true)
				TestnetSpotAPIURL:    "https://testnet.binance.vision",
				TestnetFuturesAPIURL: "https://testnet.binancefuture.com",
				// User data WS endpoints (listenKey based)
				TestnetSpotWSURL:     "wss://testnet.binance.vision/ws",
				TestnetFuturesWSURL:  "wss://stream.binancefuture.com/ws",
			},
			OKX: OKXConfig{
				APIURL: "https://www.okx.com",
				WSURL:  "wss://ws.okx.com:8443/ws/v5/private",
			},
			Bybit: BybitConfig{
				APIURL: "https://api.bybit.com",
				WSURL:  "wss://stream.bybit.com/v5/private",
			},
			Kraken: KrakenConfig{
				APIURL:    "https://api.kraken.com",
				WSURL:     "wss://ws.kraken.com",
				WSAuthURL: "wss://ws-auth.kraken.com",
			},
			Bittrex: BittrexConfig{
				APIURL: "https://api.bittrex.com/v3",
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
