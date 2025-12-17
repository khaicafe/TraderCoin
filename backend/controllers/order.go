package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrderResponse represents the API response for an order with additional fields
type OrderResponse struct {
	models.Order
	BotConfigName string `json:"bot_config_name"`
}

// GetOrderHistory - Lấy danh sách order history với filtering
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
			query = query.Where("status = ?", status)
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
			log.Printf("Error fetching order history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order history"})
			return
		}

		// Build response with bot_config_name
		result := make([]OrderResponse, 0, len(orders))
		for _, order := range orders {
			botConfigName := getBotConfigName(services.DB, order)
			result = append(result, OrderResponse{
				Order:         order,
				BotConfigName: botConfigName,
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
			botConfigName := getBotConfigName(services.DB, order)
			result = append(result, OrderResponse{
				Order:         order,
				BotConfigName: botConfigName,
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
		botConfigName := getBotConfigName(services.DB, order)

		c.JSON(http.StatusOK, OrderResponse{
			Order:         order,
			BotConfigName: botConfigName,
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
			botConfigName := getBotConfigName(services.DB, order)
			result = append(result, OrderResponse{
				Order:         order,
				BotConfigName: botConfigName,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

// getBotConfigName - Helper function để lấy bot config name
func getBotConfigName(db *gorm.DB, order models.Order) string {
	// If no bot config ID, return default
	if order.BotConfigID == 0 {
		return order.Exchange + " - " + order.Symbol
	}

	var config models.TradingConfig
	err := db.Where("id = ?", order.BotConfigID).First(&config).Error
	if err != nil {
		// Fallback if bot config not found
		return order.Exchange + " - " + order.Symbol
	}

	// Return name if exists, otherwise format: "Exchange - Symbol"
	if config.Name != "" {
		return config.Name
	}

	// Capitalize first letter of exchange
	exchange := config.Exchange
	if len(exchange) > 0 && exchange[0] >= 'a' && exchange[0] <= 'z' {
		exchange = string(exchange[0]-32) + exchange[1:]
	}

	return exchange + " - " + config.Symbol
}
