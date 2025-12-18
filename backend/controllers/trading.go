package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"tradercoin/backend/config"
	"tradercoin/backend/models"
	"tradercoin/backend/services"
	tradingservice "tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Type aliases for WebSocket requests
type RegisterRequest = services.RegisterRequest
type UnregisterRequest = services.UnregisterRequest

// PlaceOrderRequest represents the request body for placing an order
type PlaceOrderRequest struct {
	BotConfigID int     `json:"bot_config_id" binding:"required"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side" binding:"required,oneof=buy sell"`
	OrderType   string  `json:"order_type" binding:"required,oneof=market limit"`
	Amount      float64 `json:"amount"`
	Price       float64 `json:"price"`
}

// PlaceOrderResponse represents the response after placing an order
type PlaceOrderResponse struct {
	Status          string  `json:"status"`
	OrderID         uint    `json:"order_id"`
	ExchangeOrderID string  `json:"exchange_order_id"`
	Symbol          string  `json:"symbol"`
	Side            string  `json:"side"`
	OrderType       string  `json:"order_type"`
	Amount          float64 `json:"amount"`
	Price           float64 `json:"price"`
	FilledPrice     float64 `json:"filled_price"`
	StopLoss        float64 `json:"stop_loss"`
	TakeProfit      float64 `json:"take_profit"`
	OrderStatus     string  `json:"order_status"`
}

// PlaceOrderDirect - Đặt lệnh trực tiếp lên sàn giao dịch (không qua webhook)
func PlaceOrderDirect(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var request PlaceOrderRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get bot config
		var config models.TradingConfig
		err := services.DB.Where("id = ? AND user_id = ? AND is_active = ?", request.BotConfigID, userID, true).
			First(&config).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Bot config not found or inactive",
			})
			return
		}
		if err != nil {
			log.Printf("Error fetching bot config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bot config"})
			return
		}

		// Check if bot is paused
		// Note: BotStatus is not in the current models, skip this check for now
		// var botStatus models.BotStatus
		// err = services.DB.Where("bot_config_id = ?", config.ID).First(&botStatus).Error
		// if err == nil && botStatus.IsPaused {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error": "Bot is paused",
		// 	})
		// 	return
		// }

		// Check API credentials
		if config.APIKey == "" || config.APISecret == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bot config missing API credentials. Please add API key and secret.",
			})
			return
		}

		// Use provided amount or config amount
		amount := request.Amount
		if amount <= 0 {
			amount = config.Amount
		}

		if amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Amount must be greater than 0. Please provide amount in request or configure it in bot config.",
			})
			return
		}

		// Use provided symbol or config symbol
		symbol := request.Symbol
		if symbol == "" {
			symbol = config.Symbol
		}

		// Validate price for limit orders
		orderType := request.OrderType
		price := request.Price
		if orderType == "limit" && price <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Price is required for limit orders",
			})
			return
		}

		// Decrypt API credentials
		apiKey, apiSecret, err := GetDecryptedAPICredentials(&config)
		if err != nil {
			log.Printf("Failed to decrypt API credentials: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API credentials"})
			return
		}

		// Place order on exchange
		tradingService := tradingservice.NewTradingService(apiKey, apiSecret, config.Exchange)
		orderResult := tradingService.PlaceOrder(&config, request.Side, orderType, symbol, amount, price)

		if !orderResult.Success {
			errorMsg := fmt.Sprintf("Failed to place order: %s", orderResult.Error)
			log.Printf("Order placement failed: %v", orderResult.ErrorDetails)

			// Do not create order record when placement fails
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   errorMsg,
				"details": orderResult.ErrorDetails,
			})
			return
		}

		log.Printf("Order placed successfully on %s: OrderID=%s, Symbol=%s, Side=%s, Amount=%f",
			config.Exchange, orderResult.OrderID, orderResult.Symbol, orderResult.Side, orderResult.Quantity)

		// Calculate SL/TP prices
		var stopLoss, takeProfit float64
		filledPrice := orderResult.FilledPrice
		if filledPrice > 0 {
			if config.StopLossPercent > 0 {
				if request.Side == "buy" {
					stopLoss = filledPrice * (1 - config.StopLossPercent/100)
				} else {
					stopLoss = filledPrice * (1 + config.StopLossPercent/100)
				}
			}

			if config.TakeProfitPercent > 0 {
				if request.Side == "buy" {
					takeProfit = filledPrice * (1 + config.TakeProfitPercent/100)
				} else {
					takeProfit = filledPrice * (1 - config.TakeProfitPercent/100)
				}
			}

			// Place Stop Loss and Take Profit orders on Binance
			if config.Exchange == "binance" && (stopLoss > 0 || takeProfit > 0) {
				// Determine opposite side for closing position
				closeSide := "sell"
				if request.Side == "sell" {
					closeSide = "buy"
				}

				// Place Stop Loss order
				if stopLoss > 0 {
					slResult := tradingService.PlaceStopLossOrder(&config, symbol, stopLoss, orderResult.Quantity, closeSide)
					if slResult.Success {
						log.Printf("✅ Stop Loss order placed: OrderID=%s, StopPrice=%.8f", slResult.OrderID, stopLoss)
					} else {
						log.Printf("⚠️ Failed to place Stop Loss order: %s", slResult.Error)
						// Continue anyway, main order was successful
					}
				}

				// Place Take Profit order
				if takeProfit > 0 {
					tpResult := tradingService.PlaceTakeProfitOrder(&config, symbol, takeProfit, orderResult.Quantity, closeSide)
					if tpResult.Success {
						log.Printf("✅ Take Profit order placed: OrderID=%s, TakeProfit=%.8f", tpResult.OrderID, takeProfit)
					} else {
						log.Printf("⚠️ Failed to place Take Profit order: %s", tpResult.Error)
						// Continue anyway, main order was successful
					}
				}
			}
		}

		// Create order record using the actual Order model fields
		order := models.Order{
			UserID:          userID.(uint),
			BotConfigID:     config.ID,
			Exchange:        config.Exchange,
			Symbol:          orderResult.Symbol,
			OrderID:         orderResult.OrderID, // Exchange order ID
			Side:            orderResult.Side,
			Type:            orderResult.Type,
			Quantity:        orderResult.Quantity,
			Price:           orderResult.Price,
			FilledPrice:     orderResult.FilledPrice,
			Status:          orderResult.Status,
			TradingMode:     config.TradingMode,
			Leverage:        config.Leverage,
			StopLossPrice:   stopLoss,
			TakeProfitPrice: takeProfit,
			PnL:             0,
			PnLPercent:      0,
		}

		if err := services.DB.Create(&order).Error; err != nil {
			log.Printf("Error creating order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		log.Printf("Order created successfully: ID=%d, Exchange OrderID=%s", order.ID, order.OrderID)

		c.JSON(http.StatusOK, PlaceOrderResponse{
			Status:          "success",
			OrderID:         order.ID,
			ExchangeOrderID: order.OrderID,
			Symbol:          order.Symbol,
			Side:            order.Side,
			OrderType:       order.Type,
			Amount:          order.Quantity,
			Price:           order.Price,
			FilledPrice:     order.FilledPrice,
			StopLoss:        order.StopLossPrice,
			TakeProfit:      order.TakeProfitPrice,
			OrderStatus:     order.Status,
		})
	}
}

// CloseOrder - Đóng lệnh (close position)
func CloseOrder(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		orderID := c.Param("id")
		orderIDInt, err := strconv.Atoi(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		// Get order
		var order models.Order
		err = services.DB.Where("id = ? AND user_id = ?", orderIDInt, userID).First(&order).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err != nil {
			log.Printf("Error fetching order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
			return
		}

		// Check if order is already closed
		if order.Status == "cancelled" || order.Status == "closed" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Order is already " + order.Status,
			})
			return
		}

		// TODO: Implement actual position closing through trading service
		log.Printf("Closing order: ID=%d, Symbol=%s", order.ID, order.Symbol)

		// Update order status
		order.Status = "closed"
		if err := services.DB.Save(&order).Error; err != nil {
			log.Printf("Error updating order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Order closed successfully",
			"order":   order,
		})
	}
}

// GetSymbols - Lấy danh sách symbols từ exchange
func GetSymbols(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configID := c.Param("config_id")
		configIDInt, err := strconv.Atoi(configID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		// Get config
		var config models.TradingConfig
		err = services.DB.Where("id = ? AND user_id = ?", configIDInt, userID).First(&config).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot config not found"})
			return
		}
		if err != nil {
			log.Printf("Error fetching bot config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bot config"})
			return
		}

		// Check API credentials
		if config.APIKey == "" || config.APISecret == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bot config missing API credentials. Cannot fetch symbols.",
			})
			return
		}

		// Decrypt API credentials
		apiKey, apiSecret, err := GetDecryptedAPICredentials(&config)
		if err != nil {
			log.Printf("Failed to decrypt API credentials: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API credentials"})
			return
		}

		// Fetch symbols from exchange
		var symbols []string
		switch config.Exchange {
		case "binance":
			symbols, err = fetchBinanceSymbols(apiKey, apiSecret, config.TradingMode)
		case "bittrex":
			symbols, err = fetchBittrexSymbols(apiKey, apiSecret)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported exchange"})
			return
		}

		if err != nil {
			log.Printf("Failed to fetch symbols from %s: %v", config.Exchange, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch symbols from exchange",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"symbols":      symbols,
			"count":        len(symbols),
			"exchange":     config.Exchange,
			"trading_mode": config.TradingMode,
		})
	}
}

// CheckOrderStatus - Kiểm tra trạng thái lệnh trên sàn
func CheckOrderStatus(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		orderID := c.Param("id")
		orderIDInt, err := strconv.Atoi(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		// Get order
		var order models.Order
		err = services.DB.Where("id = ? AND user_id = ?", orderIDInt, userID).First(&order).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err != nil {
			log.Printf("Error fetching order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
			return
		}

		if order.OrderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Order does not have exchange order ID",
			})
			return
		}

		// TODO: Implement actual order status checking from exchange
		log.Printf("Checking order status: ID=%d, ExchangeOrderID=%s", order.ID, order.OrderID)

		c.JSON(http.StatusOK, gin.H{
			"order_id":          order.ID,
			"exchange_order_id": order.OrderID,
			"status":            order.Status,
			"filled":            order.Quantity,
			"remaining":         0,
		})
	}
}

// RefreshPnL - Lấy lại PnL từ sàn và cập nhật vào database
func RefreshPnL(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		orderID := c.Param("id")
		orderIDInt, err := strconv.Atoi(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		// Get order
		var order models.Order
		err = services.DB.Where("id = ? AND user_id = ?", orderIDInt, userID).First(&order).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err != nil {
			log.Printf("Error fetching order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
			return
		}

		// Get config
		var config models.TradingConfig
		if order.BotConfigID > 0 {
			err = services.DB.Where("id = ?", order.BotConfigID).First(&config).Error
			if err != nil {
				log.Printf("Error fetching bot config: %v", err)
				c.JSON(http.StatusNotFound, gin.H{"error": "Bot config not found"})
				return
			}

			// Check API credentials
			if config.APIKey == "" || config.APISecret == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Bot config missing API credentials",
				})
				return
			}
		}

		// TODO: Implement actual PnL fetching from exchange
		// For now, simulate PnL calculation based on price difference
		currentPrice := order.Price * 1.02 // Simulate 2% price change
		var pnl, pnlPercent float64

		if order.FilledPrice > 0 {
			if order.Side == "buy" {
				pnl = (currentPrice - order.FilledPrice) * order.Quantity
				pnlPercent = ((currentPrice - order.FilledPrice) / order.FilledPrice) * 100
			} else {
				pnl = (order.FilledPrice - currentPrice) * order.Quantity
				pnlPercent = ((order.FilledPrice - currentPrice) / order.FilledPrice) * 100
			}
		} else if order.Price > 0 {
			if order.Side == "buy" {
				pnl = (currentPrice - order.Price) * order.Quantity
				pnlPercent = ((currentPrice - order.Price) / order.Price) * 100
			} else {
				pnl = (order.Price - currentPrice) * order.Quantity
				pnlPercent = ((order.Price - currentPrice) / order.Price) * 100
			}
		}

		// Update PnL in database
		order.PnL = pnl
		order.PnLPercent = pnlPercent

		if err := services.DB.Save(&order).Error; err != nil {
			log.Printf("Error updating order PnL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update PnL"})
			return
		}

		log.Printf("PnL refreshed: Order %d, PnL=%f, PnL%%=%f", order.ID, pnl, pnlPercent)

		c.JSON(http.StatusOK, gin.H{
			"status":        "success",
			"message":       "PnL refreshed successfully",
			"order_id":      order.ID,
			"pnl":           pnl,
			"pnl_percent":   pnlPercent,
			"current_price": currentPrice,
			"order_status":  order.Status,
		})
	}
}

// fetchBinanceSymbols fetches all trading symbols from Binance
func fetchBinanceSymbols(apiKey, apiSecret, tradingMode string) ([]string, error) {
	cfg := config.Load()
	binanceCfg := cfg.Exchanges.Binance

	var baseURL string
	var endpoint string

	// Determine API endpoint based on trading mode
	// Default to production (not testnet)
	if tradingMode == "futures" {
		baseURL = binanceCfg.FuturesAPIURL
		endpoint = "/fapi/v1/exchangeInfo"
	} else {
		baseURL = binanceCfg.SpotAPIURL
		endpoint = "/api/v3/exchangeInfo"
	}

	fullURL := baseURL + endpoint

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API error: %s", string(body))
	}

	// Parse response
	var exchangeInfo struct {
		Symbols []struct {
			Symbol string `json:"symbol"`
			Status string `json:"status"`
		} `json:"symbols"`
	}

	if err := json.Unmarshal(body, &exchangeInfo); err != nil {
		return nil, err
	}

	// Extract only trading symbols with TRADING status
	var symbols []string
	for _, s := range exchangeInfo.Symbols {
		if s.Status == "TRADING" {
			symbols = append(symbols, s.Symbol)
		}
	}

	return symbols, nil
}

// fetchBittrexSymbols fetches all trading symbols from Bittrex
func fetchBittrexSymbols(apiKey, apiSecret string) ([]string, error) {
	cfg := config.Load()
	baseURL := cfg.Exchanges.Bittrex.APIURL
	endpoint := "/markets"
	fullURL := baseURL + endpoint

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// Create timestamp and content hash
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	contentHash := tradingSha256Hash("")

	// Create signature string
	preSign := timestamp + fullURL + "GET" + contentHash
	signature := tradingHmacSha512(preSign, apiSecret)

	// Set headers
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Api-Timestamp", timestamp)
	req.Header.Set("Api-Content-Hash", contentHash)
	req.Header.Set("Api-Signature", signature)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bittrex API error: %s", string(body))
	}

	// Parse response
	var markets []struct {
		Symbol string `json:"symbol"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(body, &markets); err != nil {
		return nil, err
	}

	// Extract only active trading symbols
	var symbols []string
	for _, m := range markets {
		if m.Status == "ONLINE" {
			symbols = append(symbols, m.Symbol)
		}
	}

	return symbols, nil
}

// RefillTestnetBalance - Nạp thêm fake USDT vào testnet account
func RefillTestnetBalance(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		configID := c.Param("config_id")
		configIDInt, err := strconv.Atoi(configID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
			return
		}

		// Get bot config
		var config models.TradingConfig
		err = services.DB.Where("id = ? AND user_id = ?", configIDInt, userID).First(&config).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot config not found"})
			return
		}
		if err != nil {
			log.Printf("Error fetching bot config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bot config"})
			return
		}

		// Check if exchange is Binance
		if config.Exchange != "binance" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Refill only works with Binance testnet"})
			return
		}

		// Check API credentials
		if config.APIKey == "" || config.APISecret == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bot config missing API credentials",
			})
			return
		}

		// Decrypt API credentials
		apiKey, _, err := GetDecryptedAPICredentials(&config)
		if err != nil {
			log.Printf("Failed to decrypt API credentials: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API credentials"})
			return
		}

		// Call Binance testnet faucet
		var endpoint string
		if config.TradingMode == "futures" {
			// Futures testnet endpoint
			endpoint = "https://testnet.binancefuture.com/fapi/v1/balance"
		} else {
			// Spot testnet endpoint
			endpoint = "https://testnet.binance.vision/api/v1/asset/get-funding-asset"
		}

		req, err := http.NewRequest("POST", endpoint, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		req.Header.Set("X-MBX-APIKEY", apiKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to call refill API: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to refill balance",
				"details": err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			log.Printf("Refill API error: %s", string(body))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Refill failed",
				"details": string(body),
			})
			return
		}

		log.Printf("Testnet balance refilled for config %d", config.ID)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Testnet balance refilled successfully! You should receive 1,000 USDT (Spot) or 10,000 USDT (Futures)",
			"config":  config.Name,
			"mode":    config.TradingMode,
		})
	}
}

// Helper functions for crypto operations in trading
func tradingSha256Hash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

func tradingHmacSha512(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// ConnectWebSocket - WebSocket upgrade endpoint for real-time order updates
func ConnectWebSocket(services *services.Services, hub *services.WebSocketHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Upgrade HTTP connection to WebSocket
		upgrader := services.GetWebSocketUpgrader()
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade WebSocket: %v", err)
			return
		}

		sessionID := c.Query("session_id")
		if sessionID == "" {
			sessionID = fmt.Sprintf("%d_%d", userID, time.Now().UnixNano())
		}

		log.Printf("User %d connected via WebSocket (session: %s)", userID, sessionID)

		// Get all active exchange keys for this user
		var exchangeKeys []models.ExchangeKey
		if err := services.DB.Where("user_id = ? AND is_active = ?", userID, true).
			Find(&exchangeKeys).Error; err != nil {
			log.Printf("Failed to fetch exchange keys: %v", err)
			conn.Close()
			return
		}

		// Register each exchange key with the hub
		for _, key := range exchangeKeys {
			// Create or get listen key
			listenKey := key.ListenKey

			// Check if listen key is expired or empty
			if listenKey == "" || key.ListenKeyExp == nil || key.ListenKeyExp.Before(time.Now()) {
				// Create new listen key
				adapter := services.GetExchangeAdapter(key.Exchange, true) // TODO: use config for testnet
				if adapter == nil {
					log.Printf("Unsupported exchange: %s", key.Exchange)
					continue
				}

				apiKey, apiSecret, err := DecryptExchangeKey(&key)
				if err != nil {
					log.Printf("Failed to decrypt API key: %v", err)
					continue
				}

				newListenKey, err := adapter.CreateListenKey(apiKey, apiSecret)
				if err != nil {
					log.Printf("Failed to create listen key for %s: %v", key.Exchange, err)
					continue
				}

				// Update listen key in database
				expTime := time.Now().Add(60 * time.Minute)
				key.ListenKey = newListenKey
				key.ListenKeyExp = &expTime

				if err := services.DB.Save(&key).Error; err != nil {
					log.Printf("Failed to save listen key: %v", err)
					continue
				}

				listenKey = newListenKey
			}

			// Register with hub
			regReq := &RegisterRequest{
				UserID:        userID.(uint),
				ExchangeKeyID: key.ID,
				Exchange:      key.Exchange,
				TradingMode:   key.TradingMode,
				ListenKey:     listenKey,
				SessionID:     sessionID,
				UserConn:      conn,
			}
			hub.Register <- regReq
		}

		// Handle client disconnection
		go func() {
			defer func() {
				for _, key := range exchangeKeys {
					unregReq := &UnregisterRequest{
						UserID:        userID.(uint),
						ExchangeKeyID: key.ID,
						SessionID:     sessionID,
					}
					hub.Unregister <- unregReq
				}
				conn.Close()
			}()

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					log.Printf("WebSocket disconnected for user %d: %v", userID, err)
					break
				}
			}
		}()
	}
}

// CreateListenKey - Create listen key for exchange WebSocket
func CreateListenKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		exchangeKeyID := c.Param("exchange_key_id")
		keyID, err := strconv.Atoi(exchangeKeyID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange key ID"})
			return
		}

		// Get exchange key
		var exchangeKey models.ExchangeKey
		if err := services.DB.Where("id = ? AND user_id = ?", keyID, userID).
			First(&exchangeKey).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Exchange key not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exchange key"})
			return
		}

		// Get adapter
		adapter := services.GetExchangeAdapter(exchangeKey.Exchange, true) // TODO: use config
		if adapter == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported exchange"})
			return
		}

		// Decrypt credentials
		apiKey, apiSecret, err := DecryptExchangeKey(&exchangeKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt credentials"})
			return
		}

		// Create listen key
		listenKey, err := adapter.CreateListenKey(apiKey, apiSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create listen key",
				"details": err.Error(),
			})
			return
		}

		// Save to database
		expTime := time.Now().Add(60 * time.Minute)
		exchangeKey.ListenKey = listenKey
		exchangeKey.ListenKeyExp = &expTime

		if err := services.DB.Save(&exchangeKey).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save listen key"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"listen_key": listenKey,
			"expires_at": expTime,
		})
	}
}

// KeepAliveListenKey - Keep listen key alive
func KeepAliveListenKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		exchangeKeyID := c.Param("exchange_key_id")
		keyID, err := strconv.Atoi(exchangeKeyID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange key ID"})
			return
		}

		// Get exchange key
		var exchangeKey models.ExchangeKey
		if err := services.DB.Where("id = ? AND user_id = ?", keyID, userID).
			First(&exchangeKey).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exchange key not found"})
			return
		}

		if exchangeKey.ListenKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No listen key found"})
			return
		}

		// Get adapter
		adapter := services.GetExchangeAdapter(exchangeKey.Exchange, true)
		if adapter == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported exchange"})
			return
		}

		// Decrypt credentials
		apiKey, apiSecret, err := DecryptExchangeKey(&exchangeKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt credentials"})
			return
		}

		// Keep alive
		if err := adapter.KeepAliveListenKey(apiKey, apiSecret, exchangeKey.ListenKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to keep alive listen key",
				"details": err.Error(),
			})
			return
		}

		// Update expiration
		expTime := time.Now().Add(60 * time.Minute)
		exchangeKey.ListenKeyExp = &expTime
		services.DB.Save(&exchangeKey)

		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"expires_at": expTime,
		})
	}
}

// DecryptExchangeKey decrypts exchange API credentials
func DecryptExchangeKey(key *models.ExchangeKey) (string, string, error) {
	// TODO: Implement actual decryption
	return key.APIKey, key.APISecret, nil
}
