package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrderResponse represents the API response for an order with additional fields
type OrderResponse struct {
	models.Order
	BotConfigName     string  `json:"bot_config_name"`
	StopLossPercent   float64 `json:"stop_loss_percent,omitempty"`
	TakeProfitPercent float64 `json:"take_profit_percent,omitempty"`
}

// GetOrderHistory - Lấy danh sách order history với filtering
// Status updates are handled by background worker, not here
func GetOrderHistory(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get query parameters for filtering
		botConfigIDStr := c.Query("bot_config_id")
		symbol := c.Query("symbol")
		status := c.Query("status")
		side := c.Query("side")
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")
		limitStr := c.DefaultQuery("limit", "20") // Reduced default limit for better performance
		offsetStr := c.DefaultQuery("offset", "0")

		// Parse pagination
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			limit = 20
		}
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Build query
		query := services.DB.Where("user_id = ?", userID)

		// Apply filters
		if botConfigIDStr != "" {
			botConfigID, err := strconv.Atoi(botConfigIDStr)
			if err == nil {
				query = query.Where("bot_config_id = ?", botConfigID)
			}
		}

		if symbol != "" {
			query = query.Where("symbol = ?", symbol)
		}

		if status != "" {
			query = query.Where("LOWER(status) = ?", strings.ToLower(status))
		}

		if side != "" {
			query = query.Where("LOWER(side) = ?", strings.ToLower(side))
		}

		if startDateStr != "" {
			startDate, err := time.Parse(time.RFC3339, startDateStr)
			if err == nil {
				query = query.Where("created_at >= ?", startDate)
			}
		}

		if endDateStr != "" {
			endDate, err := time.Parse(time.RFC3339, endDateStr)
			if err == nil {
				query = query.Where("created_at <= ?", endDate)
			}
		}

		// Execute query with order by created_at desc (newest first)
		var orders []models.Order
		if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
			log.Printf("Error fetching order history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order history"})
			return
		}

		// Build response with bot_config_name
		// No status checking here - background worker handles it
		result := make([]OrderResponse, 0, len(orders))
		for _, order := range orders {
			botConfigName, stopLossPercent, takeProfitPercent := getBotConfigInfo(services.DB, order)
			result = append(result, OrderResponse{
				Order:             order,
				BotConfigName:     botConfigName,
				StopLossPercent:   stopLossPercent,
				TakeProfitPercent: takeProfitPercent,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetOrders - Lấy danh sách orders (list all)
func GetOrders(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get pagination params
		skipStr := c.DefaultQuery("skip", "0")
		limitStr := c.DefaultQuery("limit", "100")

		skip, err := strconv.Atoi(skipStr)
		if err != nil || skip < 0 {
			skip = 0
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			limit = 100
		}

		// Query orders
		var orders []models.Order
		if err := services.DB.Where("user_id = ?", userID).
			Order("created_at desc").
			Offset(skip).
			Limit(limit).
			Find(&orders).Error; err != nil {
			log.Printf("Error listing orders: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
			return
		}

		// Build response with bot_config_name
		result := make([]OrderResponse, 0, len(orders))
		for _, order := range orders {
			botConfigName, stopLossPercent, takeProfitPercent := getBotConfigInfo(services.DB, order)
			result = append(result, OrderResponse{
				Order:             order,
				BotConfigName:     botConfigName,
				StopLossPercent:   stopLossPercent,
				TakeProfitPercent: takeProfitPercent,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetOrder - Lấy chi tiết 1 order
func GetOrder(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		orderID := c.Param("id")

		var order models.Order
		err := services.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err != nil {
			log.Printf("Error fetching order %s: %v", orderID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
			return
		}

		// Get bot config name
		botConfigName, stopLossPercent, takeProfitPercent := getBotConfigInfo(services.DB, order)

		c.JSON(http.StatusOK, OrderResponse{
			Order:             order,
			BotConfigName:     botConfigName,
			StopLossPercent:   stopLossPercent,
			TakeProfitPercent: takeProfitPercent,
		})
	}
}

// GetCompletedOrders - Lấy danh sách lệnh đã hoàn thành (filled hoặc closed)
func GetCompletedOrders(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get query parameters for filtering
		botConfigIDStr := c.Query("bot_config_id")
		symbol := c.Query("symbol")
		side := c.Query("side")
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")
		limitStr := c.DefaultQuery("limit", "100")
		offsetStr := c.DefaultQuery("offset", "0")

		// Parse pagination
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			limit = 100
		}
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Build query - filter by completed statuses (filled or closed)
		query := services.DB.Where("user_id = ? AND status IN (?)", userID, []string{"filled", "closed"})

		// Apply filters
		if botConfigIDStr != "" {
			botConfigID, err := strconv.Atoi(botConfigIDStr)
			if err == nil {
				query = query.Where("bot_config_id = ?", botConfigID)
			}
		}

		if symbol != "" {
			query = query.Where("symbol = ?", symbol)
		}

		if side != "" {
			query = query.Where("LOWER(side) = ?", side)
		}

		if startDateStr != "" {
			startDate, err := time.Parse(time.RFC3339, startDateStr)
			if err == nil {
				query = query.Where("created_at >= ?", startDate)
			}
		}

		if endDateStr != "" {
			endDate, err := time.Parse(time.RFC3339, endDateStr)
			if err == nil {
				query = query.Where("created_at <= ?", endDate)
			}
		}

		// Execute query with order by created_at desc (newest first)
		var orders []models.Order
		if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
			log.Printf("Error fetching completed orders: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch completed orders"})
			return
		}

		// Build response with bot_config_name
		result := make([]OrderResponse, 0, len(orders))
		for _, order := range orders {
			botConfigName, stopLossPercent, takeProfitPercent := getBotConfigInfo(services.DB, order)
			result = append(result, OrderResponse{
				Order:             order,
				BotConfigName:     botConfigName,
				StopLossPercent:   stopLossPercent,
				TakeProfitPercent: takeProfitPercent,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

// getBotConfigName - Helper function để lấy bot config name
// getBotConfigInfo returns bot config name and SL/TP percentages
func getBotConfigInfo(db *gorm.DB, order models.Order) (string, float64, float64) {
	// If no bot config ID, return default
	if order.BotConfigID == 0 {
		return order.Exchange + " - " + order.Symbol, 0, 0
	}

	var config models.TradingConfig
	err := db.Where("id = ?", order.BotConfigID).First(&config).Error
	if err != nil {
		// Fallback if bot config not found
		return order.Exchange + " - " + order.Symbol, 0, 0
	}

	// Get name
	name := ""
	if config.Name != "" {
		name = config.Name
	} else {
		// Capitalize first letter of exchange
		exchange := config.Exchange
		if len(exchange) > 0 && exchange[0] >= 'a' && exchange[0] <= 'z' {
			exchange = string(exchange[0]-32) + exchange[1:]
		}
		name = exchange + " - " + config.Symbol
	}

	return name, config.StopLossPercent, config.TakeProfitPercent
}

// CloseOrdersBySymbol - Đóng tất cả lệnh và position của một symbol
func CloseOrdersBySymbol(svc *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get order ID from URL param
		orderIDStr := c.Param("id")
		orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		// Find the order
		var order models.Order
		if err := svc.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
				return
			}
			log.Printf("Error finding order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find order"})
			return
		}

		// Get the trading config for this order
		var config models.TradingConfig
		if err := svc.DB.Where("id = ?", order.BotConfigID).First(&config).Error; err != nil {
			log.Printf("Error finding trading config: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Trading config not found"})
			return
		}

		// Decrypt API credentials
		apiKey, apiSecret, err := GetDecryptedAPICredentials(&config)
		if err != nil {
			log.Printf("❌ Error decrypting API credentials: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API credentials"})
			return
		}

		// Debug logs
		log.Printf("🔑 Decrypted API Key length: %d (first 10 chars: %s...)",
			len(apiKey),
			func() string {
				if len(apiKey) > 10 {
					return apiKey[:10]
				}
				return apiKey
			}())
		log.Printf("🔐 Decrypted API Secret length: %d", len(apiSecret))

		// Create trading service with decrypted credentials
		tradingService := services.NewTradingService(apiKey, apiSecret, config.Exchange, svc.DB, userID.(uint))

		// Log details before calling cancellation
		log.Printf("🔴 CloseOrdersBySymbol - OrderID: %d, Symbol: %s, Exchange: %s, BotConfigID: %d",
			orderID, order.Symbol, config.Exchange, order.BotConfigID)

		// Call the cancellation function
		if err := tradingService.CancelAllOrdersAndPosition(&config, order.Symbol); err != nil {
			log.Printf("❌ Error canceling orders and position for symbol %s: %v", order.Symbol, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to cancel orders and close position",
				"details": err.Error(),
			})
			return
		}

		log.Printf("✅ Successfully closed all orders and position for symbol: %s", order.Symbol)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Successfully closed all orders and position for " + order.Symbol,
		})
	}
}
