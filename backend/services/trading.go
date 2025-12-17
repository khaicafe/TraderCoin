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
	"strings"
	"time"
	"tradercoin/backend/models"
)

// TradingService handles order placement on exchanges
type TradingService struct {
	APIKey    string
	APISecret string
	Exchange  string
}

// OrderResult represents the result of placing an order
type OrderResult struct {
	Success      bool        `json:"success"`
	OrderID      string      `json:"order_id"`
	Symbol       string      `json:"symbol"`
	Side         string      `json:"side"`
	Type         string      `json:"type"`
	Quantity     float64     `json:"quantity"`
	Price        float64     `json:"price"`
	FilledPrice  float64     `json:"filled_price"`
	Status       string      `json:"status"`
	Error        string      `json:"error,omitempty"`
	ErrorDetails interface{} `json:"error_details,omitempty"`
}

// NewTradingService creates a new trading service instance
func NewTradingService(apiKey, apiSecret, exchange string) *TradingService {
	return &TradingService{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Exchange:  exchange,
	}
}

// PlaceOrder places an order on the exchange
func (ts *TradingService) PlaceOrder(config *models.TradingConfig, side, orderType, symbol string, amount, price float64) OrderResult {
	switch ts.Exchange {
	case "binance":
		return ts.placeBinanceOrder(config, side, orderType, symbol, amount, price)
	case "bittrex":
		return ts.placeBittrexOrder(config, side, orderType, symbol, amount, price)
	default:
		return OrderResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported exchange: %s", ts.Exchange),
		}
	}
}

// placeBinanceOrder places an order on Binance
func (ts *TradingService) placeBinanceOrder(config *models.TradingConfig, side, orderType, symbol string, amount, price float64) OrderResult {
	// baseURL := "https://api.binance.com"
	endpoint := "/api/v3/order"
	baseURL := "https://testnet.binance.vision"

	// Convert to Binance format
	binanceSide := strings.ToUpper(side)      // buy -> BUY, sell -> SELL
	binanceType := strings.ToUpper(orderType) // market -> MARKET, limit -> LIMIT

	// Prepare parameters
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", binanceSide)
	params.Set("type", binanceType)

	if binanceType == "MARKET" {
		params.Set("type", "MARKET")
	} else {
		params.Set("type", "LIMIT")
		params.Set("timeInForce", "GTC")
		params.Set("price", fmt.Sprintf("%.8f", price))
	}

	params.Set("quantity", fmt.Sprintf("%.8f", amount))
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// Create signature
	queryString := params.Encode()
	h := hmac.New(sha256.New, []byte(ts.APISecret))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))
	params.Set("signature", signature)

	// Build full URL
	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	// Create request
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to create request",
			ErrorDetails: err.Error(),
		}
	}

	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to send request to exchange",
			ErrorDetails: err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to read response",
			ErrorDetails: err.Error(),
		}
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)

		// Try to get error message from Binance response
		errorMsg := fmt.Sprintf("Binance API error (status %d)", resp.StatusCode)
		if msg, ok := errorResp["msg"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, msg)
		}
		if code, ok := errorResp["code"].(float64); ok {
			errorMsg = fmt.Sprintf("%s [Code: %.0f]", errorMsg, code)
		}

		return OrderResult{
			Success: false,
			Error:   errorMsg,
			ErrorDetails: map[string]interface{}{
				"status_code": resp.StatusCode,
				"response":    errorResp,
				"raw_body":    string(body),
			},
		}
	}

	// Parse response
	var binanceResp struct {
		OrderID             int64  `json:"orderId"`
		Symbol              string `json:"symbol"`
		Side                string `json:"side"`
		Type                string `json:"type"`
		OrigQty             string `json:"origQty"`
		Price               string `json:"price"`
		ExecutedQty         string `json:"executedQty"`
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
		Status              string `json:"status"`
		Fills               []struct {
			Price string `json:"price"`
			Qty   string `json:"qty"`
		} `json:"fills"`
	}

	if err := json.Unmarshal(body, &binanceResp); err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to parse response",
			ErrorDetails: string(body),
		}
	}

	// Calculate filled price from fills
	filledPrice := 0.0
	if len(binanceResp.Fills) > 0 {
		filledPrice, _ = strconv.ParseFloat(binanceResp.Fills[0].Price, 64)
	} else if binanceResp.Price != "" {
		filledPrice, _ = strconv.ParseFloat(binanceResp.Price, 64)
	}

	quantity, _ := strconv.ParseFloat(binanceResp.OrigQty, 64)
	orderPrice, _ := strconv.ParseFloat(binanceResp.Price, 64)

	return OrderResult{
		Success:     true,
		OrderID:     strconv.FormatInt(binanceResp.OrderID, 10),
		Symbol:      binanceResp.Symbol,
		Side:        binanceResp.Side,
		Type:        binanceResp.Type,
		Quantity:    quantity,
		Price:       orderPrice,
		FilledPrice: filledPrice,
		Status:      binanceResp.Status,
	}
}

// placeBittrexOrder places an order on Bittrex
func (ts *TradingService) placeBittrexOrder(config *models.TradingConfig, side, orderType, symbol string, amount, price float64) OrderResult {
	baseURL := "https://api.bittrex.com/v3"
	endpoint := "/orders"

	// Prepare request body
	requestBody := map[string]interface{}{
		"marketSymbol": symbol,
		"direction":    side,
		"type":         orderType,
		"quantity":     amount,
		"timeInForce":  "GOOD_TIL_CANCELLED",
	}

	if orderType == "LIMIT" {
		requestBody["limit"] = price
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to create request body",
			ErrorDetails: err.Error(),
		}
	}

	// Create request
	fullURL := baseURL + endpoint
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to create request",
			ErrorDetails: err.Error(),
		}
	}

	// Create timestamp and content hash
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	contentHash := sha256Hash(string(bodyJSON))

	// Create signature
	preSign := timestamp + fullURL + "POST" + contentHash
	signature := hmacSha512(preSign, ts.APISecret)

	// Set headers
	req.Header.Set("Api-Key", ts.APIKey)
	req.Header.Set("Api-Timestamp", timestamp)
	req.Header.Set("Api-Content-Hash", contentHash)
	req.Header.Set("Api-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to send request to exchange",
			ErrorDetails: err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to read response",
			ErrorDetails: err.Error(),
		}
	}

	// Check status code
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)

		// Try to get error message from Bittrex response
		errorMsg := fmt.Sprintf("Bittrex API error (status %d)", resp.StatusCode)
		if msg, ok := errorResp["message"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, msg)
		}
		if code, ok := errorResp["code"].(string); ok {
			errorMsg = fmt.Sprintf("%s [Code: %s]", errorMsg, code)
		}

		return OrderResult{
			Success: false,
			Error:   errorMsg,
			ErrorDetails: map[string]interface{}{
				"status_code": resp.StatusCode,
				"response":    errorResp,
				"raw_body":    string(body),
			},
		}
	}

	// Parse response
	var bittrexResp struct {
		ID           string  `json:"id"`
		MarketSymbol string  `json:"marketSymbol"`
		Direction    string  `json:"direction"`
		Type         string  `json:"type"`
		Quantity     float64 `json:"quantity"`
		Limit        float64 `json:"limit"`
		FillQuantity float64 `json:"fillQuantity"`
		Status       string  `json:"status"`
	}

	if err := json.Unmarshal(body, &bittrexResp); err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to parse response",
			ErrorDetails: string(body),
		}
	}

	return OrderResult{
		Success:     true,
		OrderID:     bittrexResp.ID,
		Symbol:      bittrexResp.MarketSymbol,
		Side:        bittrexResp.Direction,
		Type:        bittrexResp.Type,
		Quantity:    bittrexResp.Quantity,
		Price:       bittrexResp.Limit,
		FilledPrice: bittrexResp.Limit,
		Status:      bittrexResp.Status,
	}
}

// Helper functions
func sha256Hash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

func hmacSha512(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
