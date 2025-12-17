package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"tradercoin/backend/config"
)

// ExchangeAdapter interface for different exchanges
type ExchangeAdapter interface {
	CreateListenKey(apiKey, apiSecret string) (string, error)
	KeepAliveListenKey(apiKey, apiSecret, listenKey string) error
	CloseListenKey(apiKey, apiSecret, listenKey string) error
	GetWSURL(tradingMode, listenKey string) string
}

// BinanceAdapter implements ExchangeAdapter for Binance
type BinanceAdapter struct {
	Config        *config.BinanceConfig
	IsTestnet     bool
	SpotAPIURL    string
	FuturesAPIURL string
	SpotWSURL     string
	FuturesWSURL  string
}

// NewBinanceAdapter creates a new Binance adapter
func NewBinanceAdapter(isTestnet bool) *BinanceAdapter {
	cfg := config.Load()
	binanceCfg := cfg.Exchanges.Binance

	adapter := &BinanceAdapter{
		Config:    &binanceCfg,
		IsTestnet: isTestnet,
	}

	if isTestnet {
		adapter.SpotAPIURL = binanceCfg.TestnetSpotAPIURL
		adapter.FuturesAPIURL = binanceCfg.TestnetFuturesAPIURL
		adapter.SpotWSURL = binanceCfg.TestnetSpotWSURL
		adapter.FuturesWSURL = binanceCfg.TestnetFuturesWSURL
	} else {
		adapter.SpotAPIURL = binanceCfg.SpotAPIURL
		adapter.FuturesAPIURL = binanceCfg.FuturesAPIURL
		adapter.SpotWSURL = binanceCfg.SpotWSURL
		adapter.FuturesWSURL = binanceCfg.FuturesWSURL
	}

	return adapter
}

// CreateListenKey creates a new listen key for user data stream
func (b *BinanceAdapter) CreateListenKey(apiKey, apiSecret string) (string, error) {
	endpoint := "/api/v3/userDataStream"
	url := b.SpotAPIURL + endpoint

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ListenKey string `json:"listenKey"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.ListenKey, nil
}

// KeepAliveListenKey extends listen key validity
func (b *BinanceAdapter) KeepAliveListenKey(apiKey, apiSecret, listenKey string) error {
	endpoint := "/api/v3/userDataStream"

	params := url.Values{}
	params.Set("listenKey", listenKey)

	url := fmt.Sprintf("%s%s?%s", b.SpotAPIURL, endpoint, params.Encode())

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// CloseListenKey closes a listen key
func (b *BinanceAdapter) CloseListenKey(apiKey, apiSecret, listenKey string) error {
	endpoint := "/api/v3/userDataStream"

	params := url.Values{}
	params.Set("listenKey", listenKey)

	url := fmt.Sprintf("%s%s?%s", b.SpotAPIURL, endpoint, params.Encode())

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetWSURL returns WebSocket URL for Binance
func (b *BinanceAdapter) GetWSURL(tradingMode, listenKey string) string {
	if tradingMode == "futures" {
		return fmt.Sprintf("%s/%s", b.FuturesWSURL, listenKey)
	}
	return fmt.Sprintf("%s/%s", b.SpotWSURL, listenKey)
}

// OKXAdapter implements ExchangeAdapter for OKX
type OKXAdapter struct {
	Config *config.OKXConfig
	APIURL string
	WSURL  string
}

// NewOKXAdapter creates a new OKX adapter
func NewOKXAdapter() *OKXAdapter {
	cfg := config.Load()
	okxCfg := cfg.Exchanges.OKX

	return &OKXAdapter{
		Config: &okxCfg,
		APIURL: okxCfg.APIURL,
		WSURL:  okxCfg.WSURL,
	}
}

// CreateListenKey creates a new listen key for OKX
func (o *OKXAdapter) CreateListenKey(apiKey, apiSecret string) (string, error) {
	// OKX uses different authentication mechanism
	// Generate token or use API key directly
	return apiKey, nil // Simplified for now
}

// KeepAliveListenKey for OKX (not needed, uses different mechanism)
func (o *OKXAdapter) KeepAliveListenKey(apiKey, apiSecret, listenKey string) error {
	return nil
}

// CloseListenKey for OKX
func (o *OKXAdapter) CloseListenKey(apiKey, apiSecret, listenKey string) error {
	return nil
}

// GetWSURL returns WebSocket URL for OKX
func (o *OKXAdapter) GetWSURL(tradingMode, listenKey string) string {
	return o.WSURL
}

// BybitAdapter implements ExchangeAdapter for Bybit
type BybitAdapter struct {
	Config *config.BybitConfig
	APIURL string
	WSURL  string
}

// NewBybitAdapter creates a new Bybit adapter
func NewBybitAdapter() *BybitAdapter {
	cfg := config.Load()
	bybitCfg := cfg.Exchanges.Bybit

	return &BybitAdapter{
		Config: &bybitCfg,
		APIURL: bybitCfg.APIURL,
		WSURL:  bybitCfg.WSURL,
	}
}

// CreateListenKey creates a new listen key for Bybit
func (b *BybitAdapter) CreateListenKey(apiKey, apiSecret string) (string, error) {
	// Bybit uses different authentication mechanism
	return apiKey, nil // Simplified for now
}

// KeepAliveListenKey for Bybit
func (b *BybitAdapter) KeepAliveListenKey(apiKey, apiSecret, listenKey string) error {
	return nil
}

// CloseListenKey for Bybit
func (b *BybitAdapter) CloseListenKey(apiKey, apiSecret, listenKey string) error {
	return nil
}

// GetWSURL returns WebSocket URL for Bybit
func (b *BybitAdapter) GetWSURL(tradingMode, listenKey string) string {
	return b.WSURL
}

// GetExchangeAdapter returns appropriate adapter for exchange
func GetExchangeAdapter(exchange string, isTestnet bool) ExchangeAdapter {
	switch exchange {
	case "binance":
		return NewBinanceAdapter(isTestnet)
	case "okx":
		return NewOKXAdapter()
	case "bybit":
		return NewBybitAdapter()
	default:
		return nil
	}
}

// Helper function to create HMAC signature
func createHMAC(secret, message string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// Helper function to get timestamp
func getTimestamp() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}
