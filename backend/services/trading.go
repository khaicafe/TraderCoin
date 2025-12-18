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
	"tradercoin/backend/config"
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
	// For now, default to production (not testnet)
	// TODO: Add testnet flag to TradingConfig or ExchangeKey model
	isTestnet := false
	tradingMode := config.TradingMode
	if tradingMode == "" {
		tradingMode = "spot" // Default to spot
	}

	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	var baseURL string
	var endpoint string
	if tradingMode == "futures" {
		baseURL = adapter.FuturesAPIURL
		endpoint = "/fapi/v1/order"
	} else {
		baseURL = adapter.SpotAPIURL
		endpoint = "/api/v3/order"
	}

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

	// Log raw response from exchange
	fmt.Printf("\n🟡 MAIN ORDER - Exchange Response:\n")
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n\n", string(body))

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

		fmt.Printf("❌ MAIN ORDER ERROR: %s\n\n", errorMsg)

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
		AvgPrice            string `json:"avgPrice"`
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
	} else if binanceResp.AvgPrice != "" {
		filledPrice, _ = strconv.ParseFloat(binanceResp.AvgPrice, 64)
	} else if binanceResp.Price != "" {
		filledPrice, _ = strconv.ParseFloat(binanceResp.Price, 64)
	}

	quantity, _ := strconv.ParseFloat(binanceResp.OrigQty, 64)
	orderPrice, _ := strconv.ParseFloat(binanceResp.Price, 64)

	// Log success details
	fmt.Printf("✅ MAIN ORDER PLACED:\n")
	fmt.Printf("   OrderID: %d\n", binanceResp.OrderID)
	fmt.Printf("   Symbol: %s\n", binanceResp.Symbol)
	fmt.Printf("   Type: %s\n", binanceResp.Type)
	fmt.Printf("   Side: %s\n", binanceResp.Side)
	fmt.Printf("   Quantity: %s\n", binanceResp.OrigQty)
	fmt.Printf("   Filled Price: %.8f\n", filledPrice)
	fmt.Printf("   Status: %s\n\n", binanceResp.Status)

	return OrderResult{
		Success:     true,
		OrderID:     strconv.FormatInt(binanceResp.OrderID, 10),
		Symbol:      binanceResp.Symbol,
		Side:        binanceResp.Side,
		Type:        binanceResp.Type,
		Quantity:    quantity,
		Price:       orderPrice,
		FilledPrice: filledPrice,
		Status:      strings.ToLower(binanceResp.Status), // Convert to lowercase: FILLED -> filled
	}
}

// placeBittrexOrder places an order on Bittrex
func (ts *TradingService) placeBittrexOrder(tradingConfig *models.TradingConfig, side, orderType, symbol string, amount, price float64) OrderResult {
	cfg := config.Load()
	baseURL := cfg.Exchanges.Bittrex.APIURL
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
		Status:      strings.ToLower(bittrexResp.Status), // Convert to lowercase
	}
}

// PlaceStopLossOrder places a stop loss order on Binance
func (ts *TradingService) PlaceStopLossOrder(config *models.TradingConfig, symbol string, stopPrice float64, quantity float64, side string) OrderResult {
	if ts.Exchange != "binance" {
		return OrderResult{
			Success: false,
			Error:   "Stop loss orders only supported on Binance",
		}
	}

	isTestnet := false
	tradingMode := config.TradingMode
	if tradingMode == "" {
		tradingMode = "spot"
	}

	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	var baseURL string
	var endpoint string
	if tradingMode == "futures" {
		baseURL = adapter.FuturesAPIURL
		endpoint = "/fapi/v1/order"
	} else {
		// Spot doesn't support STOP_MARKET, use STOP_LOSS_LIMIT
		baseURL = adapter.SpotAPIURL
		endpoint = "/api/v3/order"
	}

	// Prepare parameters
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))

	if tradingMode == "futures" {
		// Futures: Use STOP_MARKET with closePosition
		params.Set("type", "STOP_MARKET")
		params.Set("stopPrice", fmt.Sprintf("%.8f", stopPrice))
		params.Set("closePosition", "true") // Close entire position
	} else {
		// Spot: Use STOP_LOSS_LIMIT
		params.Set("type", "STOP_LOSS_LIMIT")
		params.Set("quantity", fmt.Sprintf("%.8f", quantity))
		params.Set("stopPrice", fmt.Sprintf("%.8f", stopPrice))
		params.Set("price", fmt.Sprintf("%.8f", stopPrice*0.99)) // Slightly lower to ensure execution
		params.Set("timeInForce", "GTC")
	}

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
			Error:        "Failed to create stop loss order request",
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
			Error:        "Failed to place stop loss order",
			ErrorDetails: err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to read stop loss response",
			ErrorDetails: err.Error(),
		}
	}

	// Log raw response from exchange
	fmt.Printf("\n🔵 STOP LOSS ORDER - Exchange Response:\n")
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n\n", string(body))

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)

		errorMsg := fmt.Sprintf("Stop loss order failed (status %d)", resp.StatusCode)
		if msg, ok := errorResp["msg"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, msg)
		}

		fmt.Printf("❌ STOP LOSS ERROR: %s\n", errorMsg)
		fmt.Printf("Error Details: %+v\n\n", errorResp)

		return OrderResult{
			Success:      false,
			Error:        errorMsg,
			ErrorDetails: errorResp,
		}
	}

	var binanceResp struct {
		OrderID   int64  `json:"orderId"`
		Symbol    string `json:"symbol"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		Side      string `json:"side"`
		StopPrice string `json:"stopPrice"`
	}

	json.Unmarshal(body, &binanceResp)

	// Log success details
	fmt.Printf("✅ STOP LOSS ORDER PLACED:\n")
	fmt.Printf("   OrderID: %d\n", binanceResp.OrderID)
	fmt.Printf("   Symbol: %s\n", binanceResp.Symbol)
	fmt.Printf("   Type: %s\n", binanceResp.Type)
	fmt.Printf("   Side: %s\n", binanceResp.Side)
	fmt.Printf("   Stop Price: %s\n", binanceResp.StopPrice)
	fmt.Printf("   Status: %s\n\n", binanceResp.Status)

	return OrderResult{
		Success: true,
		OrderID: strconv.FormatInt(binanceResp.OrderID, 10),
		Symbol:  binanceResp.Symbol,
		Status:  strings.ToLower(binanceResp.Status), // Convert to lowercase
	}
}

// PlaceTakeProfitOrder places a take profit order on Binance
func (ts *TradingService) PlaceTakeProfitOrder(config *models.TradingConfig, symbol string, takeProfitPrice float64, quantity float64, side string) OrderResult {
	if ts.Exchange != "binance" {
		return OrderResult{
			Success: false,
			Error:   "Take profit orders only supported on Binance",
		}
	}

	isTestnet := false
	tradingMode := config.TradingMode
	if tradingMode == "" {
		tradingMode = "spot"
	}

	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	var baseURL string
	var endpoint string
	if tradingMode == "futures" {
		baseURL = adapter.FuturesAPIURL
		endpoint = "/fapi/v1/order"
	} else {
		baseURL = adapter.SpotAPIURL
		endpoint = "/api/v3/order"
	}

	// Prepare parameters
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))

	if tradingMode == "futures" {
		// Futures: Use TAKE_PROFIT_MARKET with closePosition
		params.Set("type", "TAKE_PROFIT_MARKET")
		params.Set("stopPrice", fmt.Sprintf("%.8f", takeProfitPrice))
		params.Set("closePosition", "true") // Close entire position
	} else {
		// Spot: Use TAKE_PROFIT_LIMIT
		params.Set("type", "TAKE_PROFIT_LIMIT")
		params.Set("quantity", fmt.Sprintf("%.8f", quantity))
		params.Set("stopPrice", fmt.Sprintf("%.8f", takeProfitPrice))
		params.Set("price", fmt.Sprintf("%.8f", takeProfitPrice*1.01)) // Slightly higher to ensure execution
		params.Set("timeInForce", "GTC")
	}

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
			Error:        "Failed to create take profit order request",
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
			Error:        "Failed to place take profit order",
			ErrorDetails: err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OrderResult{
			Success:      false,
			Error:        "Failed to read take profit response",
			ErrorDetails: err.Error(),
		}
	}

	// Log raw response from exchange
	fmt.Printf("\n🟢 TAKE PROFIT ORDER - Exchange Response:\n")
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n\n", string(body))

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)

		errorMsg := fmt.Sprintf("Take profit order failed (status %d)", resp.StatusCode)
		if msg, ok := errorResp["msg"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, msg)
		}

		fmt.Printf("❌ TAKE PROFIT ERROR: %s\n", errorMsg)
		fmt.Printf("Error Details: %+v\n\n", errorResp)

		return OrderResult{
			Success:      false,
			Error:        errorMsg,
			ErrorDetails: errorResp,
		}
	}

	var binanceResp struct {
		OrderID   int64  `json:"orderId"`
		Symbol    string `json:"symbol"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		Side      string `json:"side"`
		StopPrice string `json:"stopPrice"`
	}

	json.Unmarshal(body, &binanceResp)

	// Log success details
	fmt.Printf("✅ TAKE PROFIT ORDER PLACED:\n")
	fmt.Printf("   OrderID: %d\n", binanceResp.OrderID)
	fmt.Printf("   Symbol: %s\n", binanceResp.Symbol)
	fmt.Printf("   Type: %s\n", binanceResp.Type)
	fmt.Printf("   Side: %s\n", binanceResp.Side)
	fmt.Printf("   Take Profit Price: %s\n", binanceResp.StopPrice)
	fmt.Printf("   Status: %s\n\n", binanceResp.Status)

	return OrderResult{
		Success: true,
		OrderID: strconv.FormatInt(binanceResp.OrderID, 10),
		Symbol:  binanceResp.Symbol,
		Status:  strings.ToLower(binanceResp.Status), // Convert to lowercase
	}
}

// OrderStatusResult represents the result of checking order status
type OrderStatusResult struct {
	Success   bool    `json:"success"`
	OrderID   string  `json:"order_id"`
	Status    string  `json:"status"`
	Filled    float64 `json:"filled"`
	Remaining float64 `json:"remaining"`
	AvgPrice  float64 `json:"avg_price"`
	Error     string  `json:"error,omitempty"`
}

// CheckOrderStatus checks order status on exchange
func (ts *TradingService) CheckOrderStatus(config *models.TradingConfig, exchangeOrderID string, symbol string) OrderStatusResult {
	switch ts.Exchange {
	case "binance":
		return ts.checkBinanceOrderStatus(config, exchangeOrderID, symbol)
	case "bittrex":
		return ts.checkBittrexOrderStatus(config, exchangeOrderID, symbol)
	default:
		return OrderStatusResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported exchange: %s", ts.Exchange),
		}
	}
}

// checkBinanceOrderStatus checks order status on Binance
func (ts *TradingService) checkBinanceOrderStatus(config *models.TradingConfig, exchangeOrderID string, symbol string) OrderStatusResult {
	isTestnet := false
	tradingMode := config.TradingMode
	if tradingMode == "" {
		tradingMode = "spot"
	}

	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	var baseURL string
	var endpoint string
	if tradingMode == "futures" {
		baseURL = adapter.FuturesAPIURL
		endpoint = "/fapi/v1/order"
	} else {
		baseURL = adapter.SpotAPIURL
		endpoint = "/api/v3/order"
	}

	// Prepare parameters
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", exchangeOrderID)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))

	// Sign request
	signature := ts.sign(params.Encode())
	params.Set("signature", signature)

	// Make request
	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return OrderStatusResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OrderStatusResult{
			Success: false,
			Error:   fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)

		errorMsg := fmt.Sprintf("Binance API error (status %d)", resp.StatusCode)
		if msg, ok := errorResp["msg"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, msg)
		}

		return OrderStatusResult{
			Success: false,
			Error:   errorMsg,
		}
	}

	// Parse response
	var binanceResp struct {
		OrderID             int64  `json:"orderId"`
		Symbol              string `json:"symbol"`
		Status              string `json:"status"`
		OrigQty             string `json:"origQty"`
		ExecutedQty         string `json:"executedQty"`
		AvgPrice            string `json:"avgPrice"`
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	}

	if err := json.Unmarshal(body, &binanceResp); err != nil {
		return OrderStatusResult{
			Success: false,
			Error:   "Failed to parse response",
		}
	}

	origQty, _ := strconv.ParseFloat(binanceResp.OrigQty, 64)
	executedQty, _ := strconv.ParseFloat(binanceResp.ExecutedQty, 64)
	avgPrice, _ := strconv.ParseFloat(binanceResp.AvgPrice, 64)
	remaining := origQty - executedQty

	return OrderStatusResult{
		Success:   true,
		OrderID:   strconv.FormatInt(binanceResp.OrderID, 10),
		Status:    strings.ToLower(binanceResp.Status), // Convert to lowercase: FILLED -> filled
		Filled:    executedQty,
		Remaining: remaining,
		AvgPrice:  avgPrice,
	}
}

// checkBittrexOrderStatus checks order status on Bittrex
func (ts *TradingService) checkBittrexOrderStatus(config *models.TradingConfig, exchangeOrderID string, symbol string) OrderStatusResult {
	// TODO: Implement Bittrex order status check
	return OrderStatusResult{
		Success: false,
		Error:   "Bittrex order status check not implemented yet",
	}
}

// sign creates HMAC SHA256 signature
func (ts *TradingService) sign(message string) string {
	h := hmac.New(sha256.New, []byte(ts.APISecret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
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
