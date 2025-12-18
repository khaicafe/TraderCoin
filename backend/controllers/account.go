package controllers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"tradercoin/backend/config"
	"tradercoin/backend/models"
	"tradercoin/backend/services"
	"tradercoin/backend/utils"

	"github.com/gin-gonic/gin"
)

// AccountInfoResponse represents the account information from exchange
type AccountInfoResponse struct {
	Exchange string              `json:"exchange"`
	Spot     *TradingAccountInfo `json:"spot,omitempty"`
	Futures  *TradingAccountInfo `json:"futures,omitempty"`
	// Legacy fields (deprecated, for backward compatibility)
	TotalBalance     float64       `json:"total_balance,omitempty"`
	AvailableBalance float64       `json:"available_balance,omitempty"`
	InOrder          float64       `json:"in_order,omitempty"`
	Balances         []BalanceInfo `json:"balances,omitempty"`
}

// TradingAccountInfo represents account info for a specific trading mode
type TradingAccountInfo struct {
	TotalBalance     float64       `json:"total_balance"`
	AvailableBalance float64       `json:"available_balance"`
	InOrder          float64       `json:"in_order"`
	Balances         []BalanceInfo `json:"balances"`
}

// BalanceInfo represents individual asset balance
type BalanceInfo struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
	Total  float64 `json:"total"`
}

// BinanceAccountResponse represents Binance API response
type BinanceAccountResponse struct {
	Balances []struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"balances"`
}

// BittrexBalance represents Bittrex API response
type BittrexBalance struct {
	CurrencySymbol string  `json:"currencySymbol"`
	Total          float64 `json:"total"`
	Available      float64 `json:"available"`
}

// GetAccountInfo - Get account information from exchange using bot config
func GetAccountInfo(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configIDStr := c.Param("id")
		configID, err := strconv.Atoi(configIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		// Get bot config
		var config models.TradingConfig
		if err := services.DB.Where("id = ? AND user_id = ?", configID, userID).First(&config).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot config not found"})
			return
		}

		// Decrypt API credentials
		apiKey, err := utils.DecryptString(config.APIKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API key"})
			return
		}

		apiSecret, err := utils.DecryptString(config.APISecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API secret"})
			return
		}

		var accountInfo AccountInfoResponse

		// Fetch account info based on exchange
		switch config.Exchange {
		case "binance":
			accountInfo, err = getBinanceAccountInfo(apiKey, apiSecret)
		case "bittrex":
			accountInfo, err = getBittrexAccountInfo(apiKey, apiSecret)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported exchange"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch account info from exchange",
				"details": err.Error(),
			})
			return
		}

		accountInfo.Exchange = config.Exchange
		c.JSON(http.StatusOK, accountInfo)
	}
}

// getBinanceAccountInfo fetches account information from Binance
func getBinanceAccountInfoBK(apiKey, apiSecret string) (AccountInfoResponse, error) {
	cfg := config.Load()
	baseURL := cfg.Exchanges.Binance.SpotAPIURL // Use production spot API
	endpoint := "/api/v3/account"

	// Create timestamp and signature
	timestamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("timestamp=%d", timestamp)

	// Create HMAC SHA256 signature
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	// Build full URL
	fullURL := fmt.Sprintf("%s%s?%s&signature=%s", baseURL, endpoint, queryString, signature)

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return AccountInfoResponse{}, err
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return AccountInfoResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountInfoResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return AccountInfoResponse{}, fmt.Errorf("binance API error: %s", string(body))
	}

	// Parse response
	var binanceResp BinanceAccountResponse
	if err := json.Unmarshal(body, &binanceResp); err != nil {
		return AccountInfoResponse{}, err
	}

	// Convert to our format
	var balances []BalanceInfo
	var totalBalance, availableBalance, inOrder float64

	for _, b := range binanceResp.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		total := free + locked

		// Only include assets with balance > 0
		if total > 0 {
			balances = append(balances, BalanceInfo{
				Asset:  b.Asset,
				Free:   free,
				Locked: locked,
				Total:  total,
			})

			availableBalance += free
			inOrder += locked
		}
	}

	totalBalance = availableBalance + inOrder

	return AccountInfoResponse{
		TotalBalance:     totalBalance,
		AvailableBalance: availableBalance,
		InOrder:          inOrder,
		Balances:         balances,
	}, nil
}

// getBinanceAccountInfo fetches account information from Binance (both Spot and Futures)
func getBinanceAccountInfo(apiKey, apiSecret string) (AccountInfoResponse, error) {
	cfg := config.Load()

	utils.LogInfo("🔄 Fetching Binance account info for SPOT and FUTURES...")

	// Fetch Spot account
	spotInfo, spotErr := fetchBinanceSpotAccount(apiKey, apiSecret, cfg.Exchanges.Binance.SpotAPIURL)
	if spotErr != nil {
		utils.LogError(fmt.Sprintf("❌ Failed to fetch Spot account: %v", spotErr))
	}

	// Fetch Futures account
	futuresInfo, futuresErr := fetchBinanceFuturesAccount(apiKey, apiSecret, cfg.Exchanges.Binance.FuturesAPIURL)
	if futuresErr != nil {
		utils.LogError(fmt.Sprintf("❌ Failed to fetch Futures account: %v", futuresErr))
	}

	// Return combined result
	response := AccountInfoResponse{
		Exchange: "binance",
	}

	if spotErr == nil {
		response.Spot = spotInfo
		utils.LogInfo(fmt.Sprintf("✅ Spot Account: Total=%.2f, Available=%.2f, Assets=%d",
			spotInfo.TotalBalance, spotInfo.AvailableBalance, len(spotInfo.Balances)))
	}

	if futuresErr == nil {
		response.Futures = futuresInfo
		utils.LogInfo(fmt.Sprintf("✅ Futures Account: Total=%.2f, Available=%.2f, Assets=%d",
			futuresInfo.TotalBalance, futuresInfo.AvailableBalance, len(futuresInfo.Balances)))
	}

	// If both failed, return error
	if spotErr != nil && futuresErr != nil {
		return response, fmt.Errorf("failed to fetch both Spot and Futures accounts")
	}

	return response, nil
}

// fetchBinanceSpotAccount fetches Spot account info
func fetchBinanceSpotAccount(apiKey, apiSecret, baseURL string) (*TradingAccountInfo, error) {
	endpoint := "/api/v3/account"

	// Create timestamp and signature
	timestamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("timestamp=%d", timestamp)

	// Create HMAC SHA256 signature
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	// Build full URL
	fullURL := fmt.Sprintf("%s%s?%s&signature=%s", baseURL, endpoint, queryString, signature)

	utils.LogInfo(fmt.Sprintf("📡 Binance SPOT API Request: %s%s", baseURL, endpoint))

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	utils.LogInfo(fmt.Sprintf("📥 Binance SPOT Response: Status=%d, Length=%d bytes", resp.StatusCode, len(body)))

	if resp.StatusCode != http.StatusOK {
		utils.LogError(fmt.Sprintf("❌ Binance SPOT API Error: %s", string(body)))
		return nil, fmt.Errorf("binance Spot API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var binanceResp BinanceAccountResponse
	if err := json.Unmarshal(body, &binanceResp); err != nil {
		return nil, err
	}

	// Convert to our format
	var balances []BalanceInfo
	var totalBalance, availableBalance, inOrder float64

	for _, b := range binanceResp.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		total := free + locked

		if total > 0 {
			balances = append(balances, BalanceInfo{
				Asset:  b.Asset,
				Free:   free,
				Locked: locked,
				Total:  total,
			})

			availableBalance += free
			inOrder += locked
		}
	}

	totalBalance = availableBalance + inOrder

	return &TradingAccountInfo{
		TotalBalance:     totalBalance,
		AvailableBalance: availableBalance,
		InOrder:          inOrder,
		Balances:         balances,
	}, nil
}

// fetchBinanceFuturesAccount fetches Futures account info
func fetchBinanceFuturesAccount(apiKey, apiSecret, baseURL string) (*TradingAccountInfo, error) {
	endpoint := "/fapi/v2/account"

	// Create timestamp and signature
	timestamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("timestamp=%d", timestamp)

	// Create HMAC SHA256 signature
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	// Build full URL
	fullURL := fmt.Sprintf("%s%s?%s&signature=%s", baseURL, endpoint, queryString, signature)

	utils.LogInfo(fmt.Sprintf("� Binance FUTURES API Request: %s%s", baseURL, endpoint))

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	utils.LogInfo(fmt.Sprintf("� Binance FUTURES Response: Status=%d, Length=%d bytes", resp.StatusCode, len(body)))

	if resp.StatusCode != http.StatusOK {
		utils.LogError(fmt.Sprintf("❌ Binance FUTURES API Error: %s", string(body)))
		return nil, fmt.Errorf("binance Futures API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 🔍 Log raw JSON for debugging
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		utils.LogInfo(fmt.Sprintf("📦 Futures Raw JSON:\n%s", prettyJSON.String()))
	}

	// Parse response (Futures has different structure)
	var futuresResp struct {
		Assets []struct {
			Asset            string `json:"asset"`
			WalletBalance    string `json:"walletBalance"`
			UnrealizedProfit string `json:"unrealizedProfit"`
			MarginBalance    string `json:"marginBalance"`
			AvailableBalance string `json:"availableBalance"`
		} `json:"assets"`
		Positions []struct {
			Symbol           string `json:"symbol"`
			PositionAmt      string `json:"positionAmt"`
			UnrealizedProfit string `json:"unrealizedProfit"`
		} `json:"positions"`
	}

	if err := json.Unmarshal(body, &futuresResp); err != nil {
		return nil, err
	}

	// Convert to our format
	var balances []BalanceInfo
	var totalBalance, availableBalance, inOrder float64

	for _, a := range futuresResp.Assets {
		walletBal, _ := strconv.ParseFloat(a.WalletBalance, 64)
		availableBal, _ := strconv.ParseFloat(a.AvailableBalance, 64)

		if walletBal > 0 {
			locked := walletBal - availableBal
			balances = append(balances, BalanceInfo{
				Asset:  a.Asset,
				Free:   availableBal,
				Locked: locked,
				Total:  walletBal,
			})

			availableBalance += availableBal
			inOrder += locked
		}
	}

	totalBalance = availableBalance + inOrder

	return &TradingAccountInfo{
		TotalBalance:     totalBalance,
		AvailableBalance: availableBalance,
		InOrder:          inOrder,
		Balances:         balances,
	}, nil
}

// getBittrexAccountInfo fetches account information from Bittrex
func getBittrexAccountInfo(apiKey, apiSecret string) (AccountInfoResponse, error) {
	cfg := config.Load()
	baseURL := cfg.Exchanges.Bittrex.APIURL
	endpoint := "/balances"

	// Create request
	fullURL := baseURL + endpoint
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return AccountInfoResponse{}, err
	}

	// Create timestamp and content hash
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	contentHash := sha256Hash("")

	// Create signature string
	preSign := timestamp + fullURL + "GET" + contentHash
	signature := hmacSha512(preSign, apiSecret)

	// Set headers
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Api-Timestamp", timestamp)
	req.Header.Set("Api-Content-Hash", contentHash)
	req.Header.Set("Api-Signature", signature)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return AccountInfoResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountInfoResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return AccountInfoResponse{}, fmt.Errorf("bittrex API error: %s", string(body))
	}

	// Parse response
	var bittrexBalances []BittrexBalance
	if err := json.Unmarshal(body, &bittrexBalances); err != nil {
		return AccountInfoResponse{}, err
	}

	// Convert to our format
	var balances []BalanceInfo
	var totalBalance, availableBalance, inOrder float64

	for _, b := range bittrexBalances {
		if b.Total > 0 {
			locked := b.Total - b.Available
			balances = append(balances, BalanceInfo{
				Asset:  b.CurrencySymbol,
				Free:   b.Available,
				Locked: locked,
				Total:  b.Total,
			})

			availableBalance += b.Available
			inOrder += locked
		}
	}

	totalBalance = availableBalance + inOrder

	return AccountInfoResponse{
		TotalBalance:     totalBalance,
		AvailableBalance: availableBalance,
		InOrder:          inOrder,
		Balances:         balances,
	}, nil
}

// Helper function to create SHA256 hash
func sha256Hash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// Helper function to create HMAC SHA512 signature
func hmacSha512(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
