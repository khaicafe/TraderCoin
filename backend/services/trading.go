package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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

const (
	// MinNotionalUSDT is Binance Futures minimum notional value
	MinNotionalUSDT = 100.0
)

// OrderResult represents the result of placing an order
type OrderResult struct {
	Success          bool        `json:"success"`
	OrderID          string      `json:"order_id"`
	Symbol           string      `json:"symbol"`
	Side             string      `json:"side"`
	Type             string      `json:"type"`
	Quantity         float64     `json:"quantity"`
	Price            float64     `json:"price"`
	FilledPrice      float64     `json:"filled_price"`
	Status           string      `json:"status"`
	AlgoIDStopLoss   string      `json:"algo_id_stop_loss,omitempty"`
	AlgoIDTakeProfit string      `json:"algo_id_take_profit,omitempty"`
	StopLossPrice    float64     `json:"stop_loss_price,omitempty"`
	TakeProfitPrice  float64     `json:"take_profit_price,omitempty"`
	Error            string      `json:"error,omitempty"`
	ErrorDetails     interface{} `json:"error_details,omitempty"`
}

// NewTradingService creates a new trading service instance
func NewTradingService(apiKey, apiSecret, exchange string) *TradingService {
	return &TradingService{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Exchange:  exchange,
	}
}

// GetCurrentPrice lấy giá hiện tại của symbol từ Binance
func (ts *TradingService) GetCurrentPrice(config *models.TradingConfig, symbol string) (float64, error) {
	tradingMode := config.TradingMode
	if tradingMode == "" {
		tradingMode = "spot"
	}

	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	var apiURL string
	var endpoint string
	if tradingMode == "futures" {
		apiURL = adapter.FuturesAPIURL
		endpoint = "/fapi/v1/ticker/price"
	} else {
		apiURL = adapter.SpotAPIURL
		endpoint = "/api/v3/ticker/price"
	}

	fullURL := fmt.Sprintf("%s%s?symbol=%s", apiURL, endpoint, symbol)

	resp, err := http.Get(fullURL)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	var priceResp struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid price format: %w", err)
	}

	return price, nil
}

// GetMarkPrice lấy mark price cho Futures (dùng để validate SL/TP)
func (ts *TradingService) GetMarkPrice(symbol string) (float64, error) {
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	fullURL := fmt.Sprintf("%s/fapi/v1/premiumIndex?symbol=%s", adapter.FuturesAPIURL, symbol)

	resp, err := http.Get(fullURL)
	if err != nil {
		return 0, fmt.Errorf("failed to get mark price: %w", err)
	}
	defer resp.Body.Close()

	var markPriceResp struct {
		Symbol    string `json:"symbol"`
		MarkPrice string `json:"markPrice"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &markPriceResp); err != nil {
		return 0, fmt.Errorf("failed to parse mark price: %w", err)
	}

	markPrice, err := strconv.ParseFloat(markPriceResp.MarkPrice, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid mark price format: %w", err)
	}

	return markPrice, nil
}

// FormatPriceByTickSize làm tròn giá theo tickSize của symbol
// DOGEUSDT: tickSize=0.00001 → 5 decimals
// ETHUSDT: tickSize=0.01 → 2 decimals
// BTCUSDT: tickSize=0.1 → 1 decimal
func (ts *TradingService) FormatPriceByTickSize(symbol string, price float64) string {
	// Map common symbols to their tick sizes
	tickSizeMap := map[string]float64{
		"BTCUSDT":   0.1,
		"ETHUSDT":   0.01,
		"BNBUSDT":   0.01,
		"DOGEUSDT":  0.00001,
		"ADAUSDT":   0.00001,
		"XRPUSDT":   0.0001,
		"SOLUSDT":   0.001,
		"DOTUSDT":   0.001,
		"MATICUSDT": 0.0001,
		"SHIBUSDT":  0.00000001,
	}

	tickSize, exists := tickSizeMap[symbol]
	if !exists {
		// Default: 8 decimals for unknown symbols
		return fmt.Sprintf("%.8f", price)
	}

	// Calculate precision from tickSize
	precision := 0
	temp := tickSize
	for temp < 1 {
		precision++
		temp *= 10
	}

	// Round to tick size
	rounded := math.Round(price/tickSize) * tickSize

	// Format with correct precision
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, rounded)
}

// ValidateNotional kiểm tra minimum notional cho Binance Futures
func (ts *TradingService) ValidateNotional(config *models.TradingConfig, symbol string, quantity, price float64) error {
	if config.TradingMode != "futures" {
		return nil // Spot không cần validate notional
	}

	// Nếu là MARKET order hoặc chưa có price, lấy giá hiện tại
	if price == 0 {
		currentPrice, err := ts.GetCurrentPrice(config, symbol)
		if err != nil {
			return fmt.Errorf("failed to get current price: %w", err)
		}
		price = currentPrice
	}

	// Tính notional value
	notional := quantity * price

	// Kiểm tra minimum
	// if notional < MinNotionalUSDT {
	// 	minQuantity := MinNotionalUSDT / price
	// 	return fmt.Errorf(
	// 		"Order value ($%.2f) is below Binance Futures minimum ($%.2f USD). "+
	// 			"Please increase quantity to at least %.8f %s (at current price $%.2f)",
	// 		notional, MinNotionalUSDT, minQuantity, symbol, price,
	// 	)
	// }

	fmt.Printf("✅ NOTIONAL VALIDATION PASSED:\n")
	fmt.Printf("   Quantity: %.8f\n", quantity)
	fmt.Printf("   Price: $%.2f\n", price)
	fmt.Printf("   Notional: $%.2f (minimum: $%.2f)\n\n", notional, MinNotionalUSDT)

	return nil
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
		endpoint = "/fapi/v1/order" // Production endpoint

		// Validate minimum notional cho Futures
		// if err := ts.ValidateNotional(config, symbol, amount, price); err != nil {
		// 	return OrderResult{
		// 		Success: false,
		// 		Error:   err.Error(),
		// 	}
		// }

		////////// STEP 3: Pre-cleanup - Cancel all orders and close position
		if err := ts.CancelAllOrdersAndPosition(config, symbol); err != nil {
			fmt.Printf("⚠️  Warning: Cleanup had issues: %v\n", err)
		}

		time.Sleep(500 * time.Millisecond) // small delay to ensure state settles

		////////// STEP 1: Set Margin Mode (ISOLATED or CROSSED) - only once per symbol
		marginMode := config.MarginMode
		if marginMode == "" {
			marginMode = "ISOLATED" // Default to ISOLATED if not set
		}
		if err := ts.SetMarginType(config, symbol, marginMode); err != nil {
			fmt.Printf("⚠️  Warning: Failed to set margin type to %s: %v\n", marginMode, err)
			// Don't fail here, continue with order placement
		}

		////////// STEP 2: Set Leverage - only once per symbol
		if config.Leverage > 0 {
			if err := ts.SetLeverage(config, symbol, config.Leverage); err != nil {
				fmt.Printf("⚠️  Warning: Failed to set leverage: %v\n", err)
				// Don't fail here, continue with order placement
			}
		}

		// return OrderResult{
		// 	Success: false,
		// 	Error:   err.Error(),
		// }

		// small delay to ensure state settles
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

	// For Futures: add leverage if configured
	if tradingMode == "futures" && config.Leverage > 0 {
		params.Set("leverage", strconv.Itoa(config.Leverage))
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
	fmt.Printf("fullURL Code: %s\n", fullURL)

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
	fmt.Printf("   Status: %s\n", binanceResp.Status)
	fmt.Printf("   Trading Mode: %s\n", tradingMode)
	fmt.Printf("   Stop Loss %%: %.2f\n", config.StopLossPercent)
	fmt.Printf("   Take Profit %%: %.2f\n\n", config.TakeProfitPercent)

	// time.Sleep(3 * time.Second) // small delay to ensure state settles

	//////////// Đặt TP/SL tự động nếu có cấu hình trong bot (chỉ cho Futures) //////////
	var algoIDStopLoss, algoIDTakeProfit string
	if tradingMode == "futures" {
		algoIDStopLoss, algoIDTakeProfit = ts.placeAutoTPSL(config, symbol, binanceResp, binanceSide, binanceType, filledPrice, orderPrice, quantity)
	}

	//////////// Đặt trailing stop nếu có cấu hình trong bot (chỉ cho Futures) //////////
	if tradingMode == "futures" && config.CallbackRate > 0 {
		fmt.Printf("📊 Placing TRAILING STOP:\n")
		fmt.Printf("   Callback Rate: %.2f%%\n", config.CallbackRate)
		fmt.Printf("   Activation Price %%: %.2f%%\n", config.ActivationPrice)
		fmt.Printf("   Side: %s\n\n", binanceSide)

		ts.PlaceTrailingStopOrder(config, symbol, quantity, binanceSide, filledPrice, orderPrice)
	}

	return OrderResult{
		Success:          true,
		OrderID:          strconv.FormatInt(binanceResp.OrderID, 10),
		Symbol:           binanceResp.Symbol,
		Side:             binanceResp.Side,
		Type:             binanceResp.Type,
		Quantity:         quantity,
		Price:            orderPrice,
		FilledPrice:      filledPrice,
		Status:           strings.ToLower(binanceResp.Status), // Convert to lowercase: FILLED -> filled
		AlgoIDStopLoss:   algoIDStopLoss,
		AlgoIDTakeProfit: algoIDTakeProfit,
	}
}

// placeAutoTPSL tự động đặt TP/SL cho Futures orders và return Algo IDs
func (ts *TradingService) placeAutoTPSL(
	config *models.TradingConfig,
	symbol string,
	binanceResp struct {
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
	},
	binanceSide string,
	binanceType string,
	filledPrice float64,
	orderPrice float64,
	quantity float64,
) (algoIDStopLoss string, algoIDTakeProfit string) {
	statusFilled := "❌"
	shouldPlaceTPSL := false

	// Accept both FILLED and NEW status for MARKET orders
	if binanceResp.Status == "FILLED" {
		statusFilled = "✅"
		shouldPlaceTPSL = true
	} else if binanceResp.Status == "NEW" && binanceType == "MARKET" {
		statusFilled = "⚠️  (NEW but MARKET order, will place TP/SL)"
		shouldPlaceTPSL = true

		// For MARKET orders with NEW status, wait a bit for fill
		fmt.Printf("⏳ Waiting 2 seconds for MARKET order to fill...\n")
		time.Sleep(2 * time.Second)

		// Check order status again
		// statusResult := ts.CheckOrderStatus(config, strconv.FormatInt(binanceResp.OrderID, 10), symbol, "")
		// if statusResult.Success && statusResult.Status == "filled" {
		// 	fmt.Printf("✅ Order now FILLED after check\n")
		// 	// Update filled price from status check
		// 	if statusResult.AvgPrice > 0 {
		// 		filledPrice = statusResult.AvgPrice
		// 	}
		// }
	}

	slEnabled := "❌"
	if config.StopLossPercent > 0 {
		slEnabled = "✅"
	}

	tpEnabled := "❌"
	if config.TakeProfitPercent > 0 {
		tpEnabled = "✅"
	}

	fmt.Printf("🔍 CHECKING TP/SL CONDITIONS:\n")
	fmt.Printf("   Trading Mode: futures (required: futures) ✅\n")
	fmt.Printf("   Order Status: %s (required: FILLED) %s\n", binanceResp.Status, statusFilled)
	fmt.Printf("   Stop Loss %%: %.2f %s\n", config.StopLossPercent, slEnabled)
	fmt.Printf("   Take Profit %%: %.2f %s\n\n", config.TakeProfitPercent, tpEnabled)

	if !shouldPlaceTPSL {
		fmt.Printf("⚠️  Skipping TP/SL placement (conditions not met)\n\n")
		return "", ""
	}

	// Sử dụng filled price hoặc order price để tính TP/SL
	entryPrice := filledPrice
	if entryPrice == 0 {
		entryPrice = orderPrice
	}

	// Nếu vẫn không có entry price, lấy giá hiện tại
	if entryPrice == 0 {
		currentPrice, err := ts.GetCurrentPrice(config, symbol)
		if err == nil {
			entryPrice = currentPrice
			fmt.Printf("⚠️  Using current price as entry: %.8f\n", entryPrice)
		} else {
			fmt.Printf("❌ Cannot determine entry price, skipping TP/SL\n\n")
			return "", ""
		}
	}

	// Determine close side (opposite of entry side)
	closeSide := "SELL"
	if binanceSide == "SELL" {
		closeSide = "BUY"
	}

	// Place Stop Loss if configured
	fmt.Printf("\n🔍 DEBUG BEFORE SL: StopLossPercent=%.2f, TakeProfitPercent=%.2f\n", config.StopLossPercent, config.TakeProfitPercent)
	if config.StopLossPercent > 0 {
		var stopLossPrice float64
		// ⭐ Dựa vào POSITION type, không phải binanceSide
		// LONG position (BUY to open): Stop Loss BELOW entry (sell when price drops)
		// SHORT position (SELL to open): Stop Loss ABOVE entry (buy when price rises)
		if binanceSide == "BUY" {
			// LONG position: SL below entry
			stopLossPrice = entryPrice * (1 - config.StopLossPercent/100)
		} else {
			// SHORT position: SL above entry
			stopLossPrice = entryPrice * (1 + config.StopLossPercent/100)
		}

		// Validate: Stop Loss không được trigger ngay lập tức
		currentMarkPrice, err := ts.GetMarkPrice(symbol)
		if err != nil {
			fmt.Printf("⚠️  Cannot get current mark price for validation: %v\n", err)
		} else {
			// Kiểm tra xem SL có trigger ngay không
			var wouldTrigger bool
			var reason string

			if binanceSide == "BUY" {
				// LONG: SL triggers when price <= stopLossPrice
				if currentMarkPrice <= stopLossPrice {
					wouldTrigger = true
					reason = fmt.Sprintf("LONG position: Current price %.8f <= SL price %.8f", currentMarkPrice, stopLossPrice)
				}
			} else {
				// SHORT: SL triggers when price >= stopLossPrice
				if currentMarkPrice >= stopLossPrice {
					wouldTrigger = true
					reason = fmt.Sprintf("SHORT position: Current price %.8f >= SL price %.8f", currentMarkPrice, stopLossPrice)
				}
			}

			if wouldTrigger {
				errMsg := fmt.Sprintf("❌ STOP LOSS VALIDATION FAILED: %s. Order would immediately trigger!", reason)
				fmt.Printf("\n%s\n", errMsg)
				fmt.Printf("   Entry Price: %.8f\n", entryPrice)
				fmt.Printf("   Current Mark Price: %.8f\n", currentMarkPrice)
				fmt.Printf("   Stop Loss Price: %.8f\n", stopLossPrice)
				fmt.Printf("   Stop Loss %%: %.2f%%\n", config.StopLossPercent)
				fmt.Printf("   Position Type: %s\n", binanceSide)

				// Return empty algo IDs (validation failed, don't place orders)
				return "", ""
			}
		}

		fmt.Printf("📊 Placing STOP LOSS:\n")
		fmt.Printf("   Entry Price: %.8f\n", entryPrice)
		fmt.Printf("   Stop Loss %%: %.2f%%\n", config.StopLossPercent)
		fmt.Printf("   Stop Loss Price: %.8f\n", stopLossPrice)
		fmt.Printf("   Side: %s\n\n", closeSide)

		slResult := ts.PlaceStopLossOrder(config, symbol, stopLossPrice, quantity, closeSide)
		if !slResult.Success {
			fmt.Printf("⚠️  Failed to place Stop Loss: %s\n\n", slResult.Error)
		} else {
			algoIDStopLoss = slResult.OrderID
			fmt.Printf("✅ algoID Stop Loss: %s\n\n", slResult.OrderID)
		}
	}

	// Place Take Profit if configured
	fmt.Printf("🔍 DEBUG: TakeProfitPercent = %.2f (should be > 0 to place TP)\n", config.TakeProfitPercent)
	if config.TakeProfitPercent > 0 {
		var takeProfitPrice float64
		if binanceSide == "BUY" {
			// LONG position: TP above entry
			takeProfitPrice = entryPrice * (1 + config.TakeProfitPercent/100)
		} else {
			// SHORT position: TP below entry
			takeProfitPrice = entryPrice * (1 - config.TakeProfitPercent/100)
		}

		fmt.Printf("📊 Placing TAKE PROFIT:\n")
		fmt.Printf("   Entry Price: %.8f\n", entryPrice)
		fmt.Printf("   Take Profit %%: %.2f%%\n", config.TakeProfitPercent)
		fmt.Printf("   Take Profit Price: %.8f\n", takeProfitPrice)
		fmt.Printf("   Side: %s\n\n", closeSide)

		tpResult := ts.PlaceTakeProfitOrder(config, symbol, takeProfitPrice, quantity, closeSide)
		if !tpResult.Success {
			fmt.Printf("⚠️  Failed to place Take Profit: %s\n\n", tpResult.Error)
		} else {
			algoIDTakeProfit = tpResult.OrderID
			fmt.Printf("✅ algoID Take Profit: %s\n\n", tpResult.OrderID)
		}
	}

	return algoIDStopLoss, algoIDTakeProfit
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

	// isTestnet := false
	tradingMode := config.TradingMode
	if tradingMode == "" {
		tradingMode = "spot"
	}

	// For Futures: use Algo Order API (closePosition endpoint)
	if tradingMode == "futures" {
		return ts.PlaceAlgoStopLoss(config, symbol, stopPrice, strings.ToUpper(side), "LONG")
	}

	return OrderResult{
		Success: false,
		Error:   "Stop loss order placement not implemented for this trading mode",
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

	// For Futures: use Algo Order API (closePosition endpoint)
	if tradingMode == "futures" {
		return ts.PlaceAlgoTakeProfit(config, symbol, takeProfitPrice, side, "LONG")
	}

	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	var baseURL string
	var endpoint string
	// Spot: Use TAKE_PROFIT_LIMIT
	baseURL = adapter.SpotAPIURL
	endpoint = "/api/v3/order"

	// Prepare parameters
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))
	params.Set("type", "TAKE_PROFIT_LIMIT")
	params.Set("quantity", fmt.Sprintf("%.8f", quantity))
	params.Set("stopPrice", fmt.Sprintf("%.8f", takeProfitPrice))
	params.Set("price", fmt.Sprintf("%.8f", takeProfitPrice*1.01)) // Slightly higher to ensure execution
	params.Set("timeInForce", "GTC")

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

// PlaceTrailingStopOrder places a trailing stop order on Binance Futures
func (ts *TradingService) PlaceTrailingStopOrder(
	config *models.TradingConfig,
	symbol string,
	quantity float64, // số lượng cần đóng (bắt buộc cho Trailing Stop)
	side string, // side của vị thế mở: "BUY" (LONG) hoặc "SELL" (SHORT)
	filledPrice float64,
	orderPrice float64,
) OrderResult {

	if ts.Exchange != "binance" || config.TradingMode != "futures" {
		return OrderResult{Success: false, Error: "Trailing stop only supported on Binance Futures"}
	}

	// Lấy callback rate từ config
	callbackRate := config.CallbackRate
	if callbackRate <= 0 {
		callbackRate = 1.0 // Default 1%
	}

	// Tính activation price từ phần trăm trong config
	var activatePrice float64
	if config.ActivationPrice > 0 {
		// Sử dụng filled price hoặc order price làm entry
		entryPrice := filledPrice
		if entryPrice == 0 {
			entryPrice = orderPrice
		}

		// Tính activation price dựa trên phần trăm và side
		if strings.ToUpper(side) == "BUY" {
			// LONG position: activation price phía trên entry
			activatePrice = entryPrice * (1 + config.ActivationPrice/100)
		} else {
			// SHORT position: activation price phía dưới entry
			activatePrice = entryPrice * (1 - config.ActivationPrice/100)
		}
	}

	adapter := GetExchangeAdapter("binance", false).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/algoOrder"

	// Determine closing side
	closeSide := "SELL"                  // đóng LONG
	if strings.ToUpper(side) == "SELL" { // vị thế SHORT → đóng bằng BUY
		closeSide = "BUY"
	}

	params := url.Values{}
	params.Set("algoType", "CONDITIONAL") // ⭐ Bắt buộc và duy nhất: CONDITIONAL
	params.Set("symbol", symbol)
	params.Set("side", closeSide)
	params.Set("type", "TRAILING_STOP_MARKET") // type đúng

	params.Set("callbackRate", fmt.Sprintf("%.2f", callbackRate)) // 1.0 = 1%, min 0.1 max 10

	// Set activation price nếu có
	if activatePrice > 0 {
		params.Set("activatePrice", fmt.Sprintf("%.2f", activatePrice))
		fmt.Printf("   Calculated Activate Price: %.2f (from %.2f%% of entry)\n", activatePrice, config.ActivationPrice)
	}

	params.Set("quantity", fmt.Sprintf("%.8f", quantity)) // ⭐ Bắt buộc quantity cho TRAILING_STOP_MARKET
	params.Set("reduceOnly", "TRUE")                      // ⭐ Thêm dòng này
	params.Set("workingType", "MARK_PRICE")
	params.Set("priceProtect", "TRUE")

	// Nếu Hedge Mode → uncomment và set đúng LONG/SHORT
	// params.Set("positionSide", "LONG") // hoặc "SHORT"

	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	query := params.Encode()
	signature := ts.sign(query)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	// =======================
	// 🟣 DEBUG: REQUEST LOG
	// =======================
	fmt.Println("\n========== BINANCE TRAILING STOP REQUEST ==========")
	fmt.Println("URL:", baseURL+endpoint)
	fmt.Println("METHOD: POST")
	fmt.Println("QUERY:", params.Encode())
	fmt.Println("PARAMS:")
	for k, v := range params {
		fmt.Printf("  %s = %v\n", k, v)
	}
	fmt.Println("==================================================")

	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// =======================
	// 🟣 DEBUG: RESPONSE LOG
	// =======================
	fmt.Println("\n========== BINANCE TRAILING STOP RESPONSE =========")
	fmt.Println("HTTP STATUS:", resp.StatusCode)
	fmt.Println("RAW BODY:", string(body))
	fmt.Println("==================================================")

	if resp.StatusCode != http.StatusOK {
		return OrderResult{
			Success: false,
			Error:   fmt.Sprintf("Trailing stop failed: %s", string(body)),
		}
	}

	// Response trả về algoId (không phải orderId)
	var result struct {
		AlgoID int64 `json:"algoId"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return OrderResult{
			Success: false,
			Error:   "Failed to parse Binance response",
		}
	}

	fmt.Println("\n✅ TRAILING STOP PLACED SUCCESSFULLY")
	fmt.Println("AlgoID:", result.AlgoID)
	fmt.Println("==================================================")

	return OrderResult{
		Success: true,
		OrderID: strconv.FormatInt(result.AlgoID, 10), // dùng algoId để cancel/query sau
		Symbol:  symbol,
	}
}

// CheckOrderStatus checks order status on exchange
func (ts *TradingService) CheckOrderStatus(config *models.TradingConfig, exchangeOrderID string, symbol string, algoIDStopLoss string) OrderStatusResult {
	switch ts.Exchange {
	case "binance":
		return ts.checkBinanceOrderStatus(config, exchangeOrderID, symbol, algoIDStopLoss)
	case "bittrex":
		return ts.checkBittrexOrderStatus(config, exchangeOrderID, symbol)
	default:
		return OrderStatusResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported exchange: %s", ts.Exchange),
		}
	}
}

// OrderStatusResult represents the result of checking order status
type OrderStatusResult struct {
	Success     bool    `json:"success"`
	Error       string  `json:"error,omitempty"`
	OrderID     string  `json:"order_id"`
	Symbol      string  `json:"symbol"`
	Status      string  `json:"status"` // filled, new, canceled, ...
	Filled      float64 `json:"filled_qty"`
	Remaining   float64 `json:"remaining_qty"`
	AvgPrice    float64 `json:"avg_price"`
	IsRunning   bool    `json:"is_running"`             // true nếu lệnh hoặc Algo Order liên quan đang chạy
	RunningType string  `json:"running_type,omitempty"` // "NORMAL", "ALGO", hoặc ""
	AlgoStatus  string  `json:"algo_status,omitempty"`
	AlgoType    string  `json:"algo_type,omitempty"`
	OrigQty     float64 `json:"orig_qty,omitempty"`
	Side        string  `json:"side,omitempty"`
}

// RunningOrderStatus represents the status of a running order (normal or algo)
type RunningOrderStatus struct {
	IsRunning   bool    `json:"is_running"`       // true nếu lệnh đang mở/active
	Type        string  `json:"type"`             // "NORMAL" hoặc "ALGO" hoặc "UNKNOWN"
	Status      string  `json:"status,omitempty"` // NEW, PARTIALLY_FILLED, WORKING...
	Symbol      string  `json:"symbol,omitempty"`
	Side        string  `json:"side,omitempty"`
	OrigQty     float64 `json:"orig_qty,omitempty"`
	ExecutedQty float64 `json:"executed_qty,omitempty"`
	AvgPrice    float64 `json:"avg_price,omitempty"`
	AlgoType    string  `json:"algo_type,omitempty"` // TRAILING_STOP_MARKET, STOP_MARKET...
}

func (ts *TradingService) checkBinanceOrderStatus(
	config *models.TradingConfig,
	exchangeOrderID string,
	symbol string,
	algoIDStopLoss string, // Thêm tham số để check Algo Order (nếu có)
) OrderStatusResult {

	isTestnet := false
	tradingMode := config.TradingMode
	if tradingMode == "" {
		tradingMode = "spot"
	}

	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	var baseURL, endpoint string
	if tradingMode == "futures" {
		baseURL = adapter.FuturesAPIURL
		endpoint = "/fapi/v1/order"
		// fmt.Printf("\n🔍 CHECK FUTURES ORDER STATUS - Request:\n")
		// fmt.Printf("   Symbol: %s\n", symbol)
		// fmt.Printf("   OrderID: %s\n", exchangeOrderID)
	} else {
		baseURL = adapter.SpotAPIURL
		endpoint = "/api/v3/order"
	}

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", exchangeOrderID)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	signature := ts.sign(params.Encode())
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return OrderStatusResult{Success: false, Error: fmt.Sprintf("Request creation failed: %v", err)}
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OrderStatusResult{Success: false, Error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if tradingMode == "futures" {
		// fmt.Printf("🔍 CHECK FUTURES ORDER STATUS - Response:\n")
		// fmt.Printf("   Status Code: %d\n", resp.StatusCode)
		// fmt.Printf("   Response Body: %s\n\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)
		errorMsg := fmt.Sprintf("Binance API error (status %d)", resp.StatusCode)
		if msg, ok := errorResp["msg"].(string); ok {
			errorMsg += ": " + msg
		}
		return OrderStatusResult{Success: false, Error: errorMsg}
	}

	// Parse normal order response
	var binanceResp struct {
		OrderID     int64  `json:"orderId"`
		Symbol      string `json:"symbol"`
		Status      string `json:"status"`
		OrigQty     string `json:"origQty"`
		ExecutedQty string `json:"executedQty"`
		AvgPrice    string `json:"avgPrice"`
		Side        string `json:"side"`
		Type        string `json:"type"`
	}

	if err := json.Unmarshal(body, &binanceResp); err != nil {
		return OrderStatusResult{Success: false, Error: "Failed to parse response"}
	}

	origQty, _ := strconv.ParseFloat(binanceResp.OrigQty, 64)
	executedQty, _ := strconv.ParseFloat(binanceResp.ExecutedQty, 64)
	avgPrice, _ := strconv.ParseFloat(binanceResp.AvgPrice, 64)
	remaining := origQty - executedQty

	finalStatus := strings.ToLower(binanceResp.Status)
	isNormalRunning := finalStatus == "new" || finalStatus == "partially_filled"

	// Khởi tạo result cơ bản
	result := OrderStatusResult{
		Success:   true,
		OrderID:   strconv.FormatInt(binanceResp.OrderID, 10),
		Symbol:    binanceResp.Symbol,
		Status:    finalStatus,
		Filled:    executedQty,
		Remaining: remaining,
		AvgPrice:  avgPrice,
		IsRunning: isNormalRunning,
	}

	if isNormalRunning {
		result.RunningType = "NORMAL"
		if tradingMode == "futures" {
			fmt.Printf("✅ Order %s đang chạy (NORMAL): %s\n\n", exchangeOrderID, finalStatus)
		}
		return result
	}

	// Nếu lệnh thường đã FILLED/CANCELED → check Algo Order (Stop Loss/Trailing Stop) nếu có algoIDStopLoss
	if tradingMode == "futures" && algoIDStopLoss != "" {
		// Parse algoIDStopLoss từ string sang int64
		algoID, err := strconv.ParseInt(algoIDStopLoss, 10, 64)
		if err == nil {
			isRunning, status, err := ts.CheckFuturesAlgoOrderStatus(symbol, algoID)
			if err == nil && isRunning {
				result.IsRunning = true
				result.RunningType = "ALGO"
				result.AlgoStatus = status
				fmt.Printf("✅ Order %s đã FILLED nhưng ALGO ORDER (Stop Loss) vẫn đang chạy: Status=%s\n\n",
					exchangeOrderID, status)
				return result
			}
		}
	}

	result.IsRunning = false
	return result
}

// CheckFuturesAlgoOrderStatus checks if an algo order (Trailing Stop, Conditional SL/TP) is still active
func (ts *TradingService) CheckFuturesAlgoOrderStatus(symbol string, algoId int64) (bool, string, error) {
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	endpoint := "/fapi/v1/openAlgoOrders"
	baseURL := adapter.FuturesAPIURL

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	signature := ts.sign(params.Encode())
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return false, "", err
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse open algo orders
	var algoOrders []struct {
		AlgoId       int64  `json:"algoId"`
		Symbol       string `json:"symbol"`
		Side         string `json:"side"`
		TotalQty     string `json:"totalQty"`
		ExecutedQty  string `json:"executedQty"`
		ExecutedAmt  string `json:"executedAmt"`
		AvgPrice     string `json:"avgPrice"`
		ClientAlgoId string `json:"clientAlgoId"`
		BookTime     int64  `json:"bookTime"`
		AlgoStatus   string `json:"algoStatus"`
		AlgoType     string `json:"algoType"`
		UrgencyType  string `json:"urgencyType"`
		TimeInForce  string `json:"timeInForce"`
		PositionSide string `json:"positionSide"`
		ReduceOnly   bool   `json:"reduceOnly"`
	}

	if err := json.Unmarshal(body, &algoOrders); err != nil {
		return false, "", err
	}

	// Check if algoId exists in open orders
	for _, order := range algoOrders {
		if order.AlgoId == algoId && order.Symbol == symbol {
			// fmt.Printf("🔵 Algo Order still ACTIVE:\n")
			// fmt.Printf("   AlgoId: %d\n", order.AlgoId)
			// fmt.Printf("   Symbol: %s\n", order.Symbol)
			// fmt.Printf("   Type: %s\n", order.AlgoType)
			// fmt.Printf("   Status: %s\n", order.AlgoStatus)
			// fmt.Printf("   Side: %s\n", order.Side)
			// fmt.Printf("   Total Qty: %s\n", order.TotalQty)
			// fmt.Printf("   Executed Qty: %s\n\n", order.ExecutedQty)
			return true, order.AlgoStatus, nil
		}
	}

	// Not found in open orders → algo order finished/cancelled/triggered
	fmt.Printf("⚪ Algo Order %d not in open orders → finished/cancelled/triggered\n\n", algoId)
	return false, "finished", nil
}

// checkBittrexOrderStatus checks order status on Bittrex
func (ts *TradingService) checkBittrexOrderStatus(config *models.TradingConfig, exchangeOrderID string, symbol string) OrderStatusResult {
	// TODO: Implement Bittrex order status check
	return OrderStatusResult{
		Success: false,
		Error:   "Bittrex order status check not implemented yet",
	}
}

// FuturesPosition represents a Binance Futures position
type FuturesPosition struct {
	Symbol           string  `json:"symbol"`
	PositionAmt      float64 `json:"positionAmt"`
	EntryPrice       float64 `json:"entryPrice"`
	BreakEvenPrice   float64 `json:"breakEvenPrice"`
	MarkPrice        float64 `json:"markPrice"`
	UnrealizedProfit float64 `json:"unRealizedProfit"`
	LiquidationPrice float64 `json:"liquidationPrice"`
	Leverage         int     `json:"leverage"`
	MarginType       string  `json:"marginType"`
	IsolatedMargin   float64 `json:"isolatedMargin"`
	PositionSide     string  `json:"positionSide"`
	NotionalValue    float64 `json:"notional"`
	IsolatedWallet   float64 `json:"isolatedWallet"`
	UpdateTime       int64   `json:"updateTime"`
}

// FuturesPositionResult represents the result of getting futures positions
type FuturesPositionResult struct {
	Success   bool              `json:"success"`
	Positions []FuturesPosition `json:"positions,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// GetFuturesPositions gets all futures positions from Binance
func (ts *TradingService) GetFuturesPositions(symbol string) FuturesPositionResult {
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)

	endpoint := "/fapi/v2/positionRisk"
	baseURL := adapter.FuturesAPIURL

	// Prepare parameters
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	// Sign request
	signature := ts.sign(params.Encode())
	params.Set("signature", signature)

	// Make request
	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return FuturesPositionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return FuturesPositionResult{
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

		return FuturesPositionResult{
			Success: false,
			Error:   errorMsg,
		}
	}

	// Parse response - Binance returns array of positions
	var binancePositions []struct {
		Symbol           string `json:"symbol"`
		PositionAmt      string `json:"positionAmt"`
		EntryPrice       string `json:"entryPrice"`
		BreakEvenPrice   string `json:"breakEvenPrice"`
		MarkPrice        string `json:"markPrice"`
		UnRealizedProfit string `json:"unRealizedProfit"`
		LiquidationPrice string `json:"liquidationPrice"`
		Leverage         string `json:"leverage"`
		MarginType       string `json:"marginType"`
		IsolatedMargin   string `json:"isolatedMargin"`
		PositionSide     string `json:"positionSide"`
		Notional         string `json:"notional"`
		IsolatedWallet   string `json:"isolatedWallet"`
		UpdateTime       int64  `json:"updateTime"`
	}

	if err := json.Unmarshal(body, &binancePositions); err != nil {
		return FuturesPositionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse response: %v", err),
		}
	}

	// Convert to FuturesPosition structs and filter non-zero positions
	positions := make([]FuturesPosition, 0)
	for _, pos := range binancePositions {
		posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)

		// Only include positions with non-zero amount
		if posAmt == 0 {
			continue
		}

		entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
		breakEvenPrice, _ := strconv.ParseFloat(pos.BreakEvenPrice, 64)
		markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
		unrealizedProfit, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
		liquidationPrice, _ := strconv.ParseFloat(pos.LiquidationPrice, 64)
		leverage, _ := strconv.Atoi(pos.Leverage)
		isolatedMargin, _ := strconv.ParseFloat(pos.IsolatedMargin, 64)
		notional, _ := strconv.ParseFloat(pos.Notional, 64)
		isolatedWallet, _ := strconv.ParseFloat(pos.IsolatedWallet, 64)

		positions = append(positions, FuturesPosition{
			Symbol:           pos.Symbol,
			PositionAmt:      posAmt,
			EntryPrice:       entryPrice,
			BreakEvenPrice:   breakEvenPrice,
			MarkPrice:        markPrice,
			UnrealizedProfit: unrealizedProfit,
			LiquidationPrice: liquidationPrice,
			Leverage:         leverage,
			MarginType:       pos.MarginType,
			IsolatedMargin:   isolatedMargin,
			PositionSide:     pos.PositionSide,
			NotionalValue:    notional,
			IsolatedWallet:   isolatedWallet,
			UpdateTime:       pos.UpdateTime,
		})
	}

	return FuturesPositionResult{
		Success:   true,
		Positions: positions,
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

// isFuturesHedgeMode queries account to detect if hedge mode (dualSidePosition) is enabled
func (ts *TradingService) isFuturesHedgeMode(config *models.TradingConfig) (bool, error) {
	if config.TradingMode != "futures" {
		return false, nil
	}
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/account"

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")
	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("account query failed (status %d)", resp.StatusCode)
	}

	var account struct {
		DualSidePosition bool `json:"dualSidePosition"`
	}
	if err := json.Unmarshal(body, &account); err != nil {
		return false, err
	}
	return account.DualSidePosition, nil
}

// CancelAllOpenOrders cancels all open orders for a symbol (Futures)
func (ts *TradingService) CancelAllOpenOrders(config *models.TradingConfig, symbol string) error {
	if config.TradingMode != "futures" {
		return nil
	}

	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/allOpenOrders"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")
	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.Unmarshal(body, &errorResp)
		msg := fmt.Sprintf("cancel all open orders failed (status %d)", resp.StatusCode)
		if m, ok := errorResp["msg"].(string); ok {
			msg = fmt.Sprintf("%s: %s", msg, m)
		}
		return errors.New(msg)
	}

	fmt.Printf("🧹 Cancelled all open orders for %s\n", symbol)
	return nil
}

func (ts *TradingService) GetOpenAlgoOrders(symbol string) ([]int64, error) {
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/openOrders"

	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	// Debug: Log request details
	fmt.Printf("\n🔵 GET OPEN TRAILING STOP ORDERS - Request Details:\n")
	fmt.Printf("   Endpoint: %s\n", endpoint)
	if symbol != "" {
		fmt.Printf("   Symbol: %s\n", symbol)
	} else {
		fmt.Printf("   Symbol: ALL\n")
	}
	fmt.Printf("   Filter Type: TRAILING_STOP_MARKET\n\n")

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Printf("   ❌ Request creation failed: %v\n\n", err)
		return nil, fmt.Errorf("failed to create open orders request: %w", err)
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Request failed: %v\n\n", err)
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Debug: Log response
	fmt.Printf("🔵 GET OPEN TRAILING STOP ORDERS - Exchange Response:\n")
	fmt.Printf("   Status Code: %d\n", resp.StatusCode)
	fmt.Printf("   Response Body: %s\n\n", string(body))

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.Unmarshal(body, &errorResp)
		errorMsg := fmt.Sprintf("get open orders failed (status %d)", resp.StatusCode)
		if m, ok := errorResp["msg"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, m)
		}
		if code, ok := errorResp["code"].(float64); ok {
			errorMsg = fmt.Sprintf("%s [Code: %.0f]", errorMsg, code)
		}
		fmt.Printf("   ❌ Error: %s\n\n", errorMsg)
		return nil, errors.New(errorMsg)
	}

	var orders []map[string]interface{}
	if err := json.Unmarshal(body, &orders); err != nil {
		fmt.Printf("   ❌ JSON parsing failed: %v\n\n", err)
		return nil, fmt.Errorf("failed to parse open orders response: %w", err)
	}

	// Filter TRAILING_STOP_MARKET orders and extract orderId
	var algoIds []int64
	trailingStopCount := 0
	for _, order := range orders {
		if orderType, ok := order["type"].(string); ok && orderType == "TRAILING_STOP_MARKET" {
			// Use orderId field (Binance returns orderId for standard orders)
			if orderId, ok := order["orderId"].(float64); ok {
				algoIds = append(algoIds, int64(orderId))
				trailingStopCount++
				fmt.Printf("   ✅ Found Trailing Stop Order ID: %.0f\n", orderId)
			}
		}
	}

	fmt.Printf("\n   📊 Summary:\n")
	fmt.Printf("      Total Orders Returned: %d\n", len(orders))
	fmt.Printf("      Trailing Stop Orders Found: %d\n", trailingStopCount)
	fmt.Printf("      Order IDs: %v\n\n", algoIds)

	return algoIds, nil
}

func (ts *TradingService) CancelAllTrailingStops(symbol string) error {
	adapter := GetExchangeAdapter("binance", false).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL

	// ⭐ Bước 1: Lấy open Algo Orders (endpoint đúng)
	openEndpoint := "/fapi/v1/openAlgoOrders" // ⭐ FIX: openAlgoOrders

	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol) // filter theo symbol, khuyến nghị
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	query := params.Encode()
	signature := ts.sign(query)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, openEndpoint, params.Encode())

	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Get open algo orders failed: %s", string(body))
	}

	var openOrders []struct {
		AlgoID    int64  `json:"algoId"`
		OrderType string `json:"orderType"` // TRAILING_STOP_MARKET, STOP_MARKET, etc.
		Symbol    string `json:"symbol"`
	}
	if err := json.Unmarshal(body, &openOrders); err != nil {
		return err
	}

	if len(openOrders) == 0 {
		fmt.Println("Không có Algo Order (Trailing Stop) nào đang mở trên", symbol)
		return nil
	}

	// ⭐ Bước 2: Loop hủy (chỉ hủy TRAILING_STOP_MARKET nếu muốn filter)
	cancelEndpoint := "/fapi/v1/algoOrder"
	for _, order := range openOrders {
		if order.OrderType != "TRAILING_STOP_MARKET" {
			continue // bỏ qua nếu không phải Trailing Stop (tùy chọn)
		}

		cancelParams := url.Values{}
		cancelParams.Set("algoId", strconv.FormatInt(order.AlgoID, 10))

		cancelParams.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		cancelParams.Set("recvWindow", "5000")

		cancelQuery := cancelParams.Encode()
		cancelSig := ts.sign(cancelQuery)
		cancelParams.Set("signature", cancelSig)

		cancelURL := fmt.Sprintf("%s%s?%s", baseURL, cancelEndpoint, cancelParams.Encode())

		cancelReq, _ := http.NewRequest("DELETE", cancelURL, nil)
		cancelReq.Header.Set("X-MBX-APIKEY", ts.APIKey)

		cancelResp, _ := http.DefaultClient.Do(cancelReq)
		cancelBody, _ := io.ReadAll(cancelResp.Body)
		cancelResp.Body.Close()

		if cancelResp.StatusCode == 200 {
			fmt.Printf("✅ Hủy thành công Trailing Stop algoId=%d trên %s\n", order.AlgoID, order.Symbol)
		} else {
			fmt.Printf("❌ Hủy thất bại algoId=%d: %s\n", order.AlgoID, string(cancelBody))
		}

		time.Sleep(100 * time.Millisecond) // tránh rate limit
	}

	return nil
}

// CloseFuturesPositionMarket tries to close existing position with MARKET closePosition; falls back to reduceOnly MARKET
func (ts *TradingService) CloseFuturesPositionMarket(config *models.TradingConfig, symbol string) OrderResult {
	if config.TradingMode != "futures" {
		return OrderResult{Success: true}
	}

	pos, err := ts.getFuturesPositionInfo(config, symbol)
	if err != nil {
		return OrderResult{Success: false, Error: fmt.Sprintf("position query failed: %v", err)}
	}
	if pos.Quantity == 0 {
		return OrderResult{Success: false, Error: "no-position"}
	}

	// Attempt MARKET closePosition (with reduceOnly per doc)
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/order"

	oppositeSide := "SELL"
	if pos.Side == "SHORT" {
		oppositeSide = "BUY"
	}

	hedge, _ := ts.isFuturesHedgeMode(config)

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", oppositeSide)
	if hedge {
		// include positionSide ONLY in Hedge Mode
		if pos.Side == "LONG" {
			params.Set("positionSide", "LONG")
		} else {
			params.Set("positionSide", "SHORT")
		}
	}
	params.Set("type", "MARKET")
	params.Set("closePosition", "true")
	params.Set("reduceOnly", "true")
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("signature", ts.sign(params.Encode()))

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Closed position via MARKET closePosition for %s\n", symbol)
		return OrderResult{Success: true}
	}

	// Fallback: reduceOnly MARKET with quantity (no closePosition)
	fmt.Printf("⚠️  MARKET closePosition not supported/status %d, fallback to reduceOnly MARKET\n", resp.StatusCode)

	params = url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", oppositeSide)
	params.Set("type", "MARKET")
	params.Set("quantity", fmt.Sprintf("%.8f", pos.Quantity))
	params.Set("reduceOnly", "true")
	if hedge {
		if pos.Side == "LONG" {
			params.Set("positionSide", "LONG")
		} else {
			params.Set("positionSide", "SHORT")
		}
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("signature", ts.sign(params.Encode()))

	fullURL = fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err = http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	resp, err = client.Do(req)
	if err != nil {
		return OrderResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.Unmarshal(body, &errorResp)
		msg := fmt.Sprintf("close position fallback failed (status %d)", resp.StatusCode)
		if m, ok := errorResp["msg"].(string); ok {
			msg = fmt.Sprintf("%s: %s", msg, m)
		}
		return OrderResult{Success: false, Error: msg, ErrorDetails: errorResp}
	}

	fmt.Printf("✅ Closed position via reduceOnly MARKET for %s (qty %.8f)\n", symbol, pos.Quantity)
	return OrderResult{Success: true}
}

// futuresPosition holds simplified position info
type futuresPosition struct {
	Quantity float64 // absolute position size
	Side     string  // LONG or SHORT
}

// getFuturesPositionInfo retrieves current position size/side for the symbol
func (ts *TradingService) getFuturesPositionInfo(config *models.TradingConfig, symbol string) (futuresPosition, error) {
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v2/positionRisk"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")
	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return futuresPosition{}, err
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return futuresPosition{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.Unmarshal(body, &errorResp)
		return futuresPosition{}, fmt.Errorf("positionRisk failed (status %d)", resp.StatusCode)
	}

	// Response is an array of positions
	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil {
		return futuresPosition{}, err
	}
	for _, it := range arr {
		sym, _ := it["symbol"].(string)
		if sym != symbol {
			continue
		}
		posAmtStr, _ := it["positionAmt"].(string)
		posAmt, _ := strconv.ParseFloat(posAmtStr, 64)
		side := "LONG"
		if posAmt < 0 {
			side = "SHORT"
		}
		return futuresPosition{Quantity: math.Abs(posAmt), Side: side}, nil
	}
	return futuresPosition{Quantity: 0, Side: "LONG"}, nil
}

// getAllFuturesPositions retrieves all futures positions for the account
func (ts *TradingService) getAllFuturesPositions(config *models.TradingConfig) (map[string]futuresPosition, error) {
	if config.TradingMode != "futures" {
		return map[string]futuresPosition{}, nil
	}

	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v2/positionRisk"

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")
	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("positionRisk failed (status %d)", resp.StatusCode)
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, err
	}

	positions := make(map[string]futuresPosition)
	for _, it := range arr {
		sym, _ := it["symbol"].(string)
		posAmtStr, _ := it["positionAmt"].(string)
		posAmt, _ := strconv.ParseFloat(posAmtStr, 64)
		side := "LONG"
		if posAmt < 0 {
			side = "SHORT"
		}
		positions[sym] = futuresPosition{Quantity: math.Abs(posAmt), Side: side}
	}
	return positions, nil
}

// listFuturesOpenOrderSymbols lists unique symbols that currently have open orders
func (ts *TradingService) listFuturesOpenOrderSymbols(config *models.TradingConfig) ([]string, error) {
	if config.TradingMode != "futures" {
		return []string{}, nil
	}
	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/openOrders"

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")
	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openOrders failed (status %d)", resp.StatusCode)
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, err
	}
	set := make(map[string]struct{})
	for _, it := range arr {
		sym, _ := it["symbol"].(string)
		if sym != "" {
			set[sym] = struct{}{}
		}
	}
	symbols := make([]string, 0, len(set))
	for s := range set {
		symbols = append(symbols, s)
	}
	return symbols, nil
}

// CancelAllOpenOrdersForAllSymbols cancels all open orders across all futures symbols
func (ts *TradingService) CancelAllOpenOrdersForAllSymbols(config *models.TradingConfig) error {
	if config.TradingMode != "futures" {
		return nil
	}
	symbols, err := ts.listFuturesOpenOrderSymbols(config)
	if err != nil {
		return err
	}
	var firstErr error
	for _, sym := range symbols {
		if e := ts.CancelAllOpenOrders(config, sym); e != nil && firstErr == nil {
			firstErr = e
		}
	}
	return firstErr
}

// CloseAllFuturesPositionsMarket closes all non-zero futures positions across all symbols
func (ts *TradingService) CloseAllFuturesPositionsMarket(config *models.TradingConfig) error {
	if config.TradingMode != "futures" {
		return nil
	}
	positions, err := ts.getAllFuturesPositions(config)
	if err != nil {
		return err
	}
	var firstErr error
	for sym, pos := range positions {
		if pos.Quantity == 0 {
			continue
		}
		res := ts.CloseFuturesPositionMarket(config, sym)
		if !res.Success && res.Error != "no-position" && firstErr == nil {
			firstErr = errors.New(res.Error)
		}
	}
	return firstErr
}

// SetMarginType sets margin mode for a symbol (ISOLATED or CROSSED)
// Only call once per symbol unless changing margin type
func (ts *TradingService) SetMarginType(config *models.TradingConfig, symbol, marginType string) error {
	if config.TradingMode != "futures" {
		return nil
	}

	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/marginType"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("marginType", strings.ToUpper(marginType)) // ISOLATED or CROSSED
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create margin type request: %w", err)
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set margin type: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.Unmarshal(body, &errorResp)
		errorMsg := fmt.Sprintf("set margin type failed (status %d)", resp.StatusCode)
		if m, ok := errorResp["msg"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, m)
		}
		if code, ok := errorResp["code"].(float64); ok {
			errorMsg = fmt.Sprintf("%s [Code: %.0f]", errorMsg, code)
		}
		// Code -4046 means margin type already set, which is fine
		if code, ok := errorResp["code"].(float64); ok && code == -4046 {
			fmt.Printf("⚠️  Margin type already set for %s (code -4046)\n", symbol)
			return nil // Not an error
		}
		return errors.New(errorMsg)
	}

	fmt.Printf("✅ Set margin type to %s for %s\n", marginType, symbol)
	return nil
}

// SetLeverage sets leverage for a symbol (1-125)
// Only call once per symbol unless changing leverage
func (ts *TradingService) SetLeverage(config *models.TradingConfig, symbol string, leverage int) error {
	if config.TradingMode != "futures" {
		return nil
	}

	if leverage < 1 || leverage > 125 {
		return fmt.Errorf("leverage must be between 1 and 125, got %d", leverage)
	}

	isTestnet := false
	adapter := GetExchangeAdapter("binance", isTestnet).(*BinanceAdapter)
	baseURL := adapter.FuturesAPIURL
	endpoint := "/fapi/v1/leverage"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("leverage", strconv.Itoa(leverage))
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	// Sign BEFORE adding signature to the URL
	queryString := params.Encode()
	signature := ts.sign(queryString)
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create leverage request: %w", err)
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set leverage: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.Unmarshal(body, &errorResp)
		errorMsg := fmt.Sprintf("set leverage failed (status %d)", resp.StatusCode)
		if m, ok := errorResp["msg"].(string); ok {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, m)
		}
		if code, ok := errorResp["code"].(float64); ok {
			errorMsg = fmt.Sprintf("%s [Code: %.0f]", errorMsg, code)
		}
		return errors.New(errorMsg)
	}

	fmt.Printf("✅ Set leverage to %d for %s\n", leverage, symbol)
	return nil
}

// FuturesPositionInfo represents position information from Binance Futures
type FuturesPositionInfo struct {
	Symbol           string  `json:"symbol"`
	PositionAmt      float64 `json:"position_amt"`      // Position size
	EntryPrice       float64 `json:"entry_price"`       // Entry price
	MarkPrice        float64 `json:"mark_price"`        // Current mark price
	UnrealizedProfit float64 `json:"unrealized_profit"` // Unrealized PnL
	LiquidationPrice float64 `json:"liquidation_price"` // Liquidation price
	Leverage         int     `json:"leverage"`          // Leverage
	MarginType       string  `json:"margin_type"`       // ISOLATED or CROSS
	Isolated         bool    `json:"isolated"`          // true if isolated margin, false if cross margin
	IsolatedMargin   float64 `json:"isolated_margin"`   // Margin for isolated position
	PositionSide     string  `json:"position_side"`     // BOTH, LONG, or SHORT
	PnlPercent       float64 `json:"pnl_percent"`       // PnL percentage
}

// GetFuturesPosition gets position information for a symbol
func (ts *TradingService) GetFuturesPosition(symbol string) (*FuturesPositionInfo, error) {
	adapter := GetExchangeAdapter("binance", false).(*BinanceAdapter)

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	signature := ts.sign(params.Encode())
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s/fapi/v2/positionRisk?%s", adapter.FuturesAPIURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", ts.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 🔍 LOG RESPONSE JSON TỪ BINANCE
	log.Printf("\n🔍 ===== BINANCE POSITION RISK API RESPONSE =====")
	log.Printf("URL: %s", fullURL)
	log.Printf("Status Code: %d", resp.StatusCode)
	log.Printf("Raw JSON Response:\n%s", string(body))
	log.Printf("================================================\n")

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.Unmarshal(body, &errorResp)
		return nil, fmt.Errorf("get position failed (status %d): %s", resp.StatusCode, string(body))
	}

	var positions []struct {
		Symbol           string `json:"symbol"`
		PositionAmt      string `json:"positionAmt"`
		EntryPrice       string `json:"entryPrice"`
		MarkPrice        string `json:"markPrice"`
		UnRealizedProfit string `json:"unRealizedProfit"`
		LiquidationPrice string `json:"liquidationPrice"`
		Leverage         string `json:"leverage"`
		MarginType       string `json:"marginType"`
		Isolated         bool   `json:"isolated"`
		IsolatedMargin   string `json:"isolatedMargin"`
		PositionSide     string `json:"positionSide"`
	}

	if err := json.Unmarshal(body, &positions); err != nil {
		return nil, fmt.Errorf("failed to parse positions: %v, body: %s", err, string(body))
	}

	// Debug: Log all positions
	fmt.Printf("🔍 GetFuturesPosition(%s): Found %d positions from Binance\n", symbol, len(positions))
	for i, pos := range positions {
		fmt.Printf("   [%d] Symbol=%s, PositionAmt=%s, Side=%s, Entry=%s, Mark=%s\n",
			i, pos.Symbol, pos.PositionAmt, pos.PositionSide, pos.EntryPrice, pos.MarkPrice)
	}

	// Find position for this symbol
	for _, pos := range positions {
		if pos.Symbol == symbol {
			posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)

			fmt.Printf("✅ Found position for %s: PositionAmt=%.8f\n", symbol, posAmt)

			// Skip if no position
			if posAmt == 0 {
				fmt.Printf("⚠️  Position amount is 0, returning nil\n")
				return nil, nil
			}

			entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
			markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
			unrealizedPnl, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
			liqPrice, _ := strconv.ParseFloat(pos.LiquidationPrice, 64)
			leverage, _ := strconv.Atoi(pos.Leverage)
			isolatedMargin, _ := strconv.ParseFloat(pos.IsolatedMargin, 64)

			// Calculate PnL percentage
			pnlPercent := 0.0
			if entryPrice > 0 {
				pnlPercent = (unrealizedPnl / (math.Abs(posAmt) * entryPrice)) * 100
			}

			fmt.Printf("📊 Position Details: Entry=%.2f, Mark=%.2f, PnL=%.2f (%.2f%%), Leverage=%dx\n",
				entryPrice, markPrice, unrealizedPnl, pnlPercent, leverage)

			return &FuturesPositionInfo{
				Symbol:           pos.Symbol,
				PositionAmt:      posAmt,
				EntryPrice:       entryPrice,
				MarkPrice:        markPrice,
				UnrealizedProfit: unrealizedPnl,
				LiquidationPrice: liqPrice,
				Leverage:         leverage,
				MarginType:       pos.MarginType,
				Isolated:         pos.Isolated,
				IsolatedMargin:   isolatedMargin,
				PositionSide:     pos.PositionSide,
				PnlPercent:       pnlPercent,
			}, nil
		}
	}

	fmt.Printf("❌ No position found for symbol %s\n", symbol)
	return nil, nil
}

// CancelAllOrdersAndPosition cancels all orders and closes position for a symbol
func (ts *TradingService) CancelAllOrdersAndPosition(config *models.TradingConfig, symbol string) error {
	fmt.Printf("🔄 Starting cancellation process for %s\n", symbol)

	// Step 1: Close any existing position
	closeRes := ts.CloseFuturesPositionMarket(config, symbol)
	if closeRes.Success {
		fmt.Printf("✅ Closed existing position for %s\n", symbol)
	} else if closeRes.Error != "no-position" { // no-position is not an error
		fmt.Printf("⚠️  Failed to close existing position for %s: %s\n", symbol, closeRes.Error)
	}

	// Step 2: Cancel all open orders
	cleanupErr := ts.CancelAllOpenOrders(config, symbol)
	if cleanupErr != nil {
		fmt.Printf("⚠️  Failed to cancel open orders for %s: %v\n", symbol, cleanupErr)
	} else {
		fmt.Printf("✅ Canceled all open orders for %s\n", symbol)
	}

	// Step 3: Cancel any existing Trailing Stop (ALGO) orders
	err := ts.CancelAllTrailingStops(symbol)
	if err != nil {
		fmt.Printf("⚠️  Failed to cancel trailing stops for %s: %v\n", symbol, err)
	} else {
		fmt.Printf("✅ Canceled trailing stops for %s\n", symbol)
	}

	fmt.Printf("✅ Cancellation process completed for %s\n", symbol)
	return nil
}
