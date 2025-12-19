package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"
	tradingservice "tradercoin/backend/services"
	"tradercoin/backend/utils"

	"github.com/gin-gonic/gin"
)

// TradingViewWebhook handles incoming signals from TradingView
func TradingViewWebhook(services *services.Services, wsHub *services.WebSocketHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Optional URL path prefix to identify user webhook
		prefix := c.Param("prefix")
		var payload struct {
			Symbol     string  `json:"symbol" binding:"required"`
			Action     string  `json:"action" binding:"required"` // buy, sell, close
			Price      float64 `json:"price"`
			StopLoss   float64 `json:"stopLoss"`
			TakeProfit float64 `json:"takeProfit"`
			Message    string  `json:"message"`
			Timestamp  int64   `json:"timestamp"`
			Strategy   string  `json:"strategy"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			utils.LogError(fmt.Sprintf("❌ Invalid webhook payload: %v", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		utils.LogInfo(fmt.Sprintf("📡 TradingView Signal Received: %s %s @ %.2f",
			payload.Action, payload.Symbol, payload.Price))

		// Create signal record
		signal := models.TradingSignal{
			Symbol:        payload.Symbol,
			Action:        payload.Action,
			Price:         payload.Price,
			StopLoss:      payload.StopLoss,
			TakeProfit:    payload.TakeProfit,
			Message:       payload.Message,
			Strategy:      payload.Strategy,
			Status:        "pending",
			ReceivedAt:    time.Now(),
			WebhookPrefix: prefix,
		}

		if err := services.DB.Create(&signal).Error; err != nil {
			utils.LogError(fmt.Sprintf("❌ Failed to save signal: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save signal"})
			return
		}

		utils.LogInfo(fmt.Sprintf("✅ Signal saved with ID: %d", signal.ID))

		// 🔔 Broadcast signal_new event to all connected WebSocket clients
		utils.LogInfo(fmt.Sprintf("🔍 DEBUG: wsHub is nil? %v", wsHub == nil))
		if wsHub != nil {
			utils.LogInfo(fmt.Sprintf("📡 BEFORE BroadcastToAll for signal ID: %d", signal.ID))
			wsHub.BroadcastToAll(map[string]interface{}{
				"type": "signal_new",
				"data": map[string]interface{}{
					"signal_id":   signal.ID,
					"symbol":      signal.Symbol,
					"action":      signal.Action,
					"price":       signal.Price,
					"stop_loss":   signal.StopLoss,
					"take_profit": signal.TakeProfit,
					"strategy":    signal.Strategy,
					"message":     signal.Message,
					"received_at": signal.ReceivedAt,
				},
			})
			utils.LogInfo(fmt.Sprintf("📡 AFTER BroadcastToAll - Broadcasted signal_new event (ID: %d) to all WebSocket clients", signal.ID))
		} else {
			utils.LogError("❌ wsHub is NIL - cannot broadcast signal!")
		}

		c.JSON(http.StatusOK, gin.H{
			"status":         "received",
			"signal_id":      signal.ID,
			"message":        "Signal received and queued",
			"webhook_prefix": prefix,
		})
	}
}

// ListSignals returns all trading signals
func ListSignals(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var signals []models.TradingSignal

		// Query params
		status := c.Query("status")
		symbol := c.Query("symbol")
		prefix := c.Query("prefix")
		limitStr := c.DefaultQuery("limit", "50")
		limit, _ := strconv.Atoi(limitStr)
		sinceHoursStr := c.Query("since_hours")
		sinceTsStr := c.Query("since_ts") // optional: unix seconds or milliseconds from client 'now'

		query := services.DB.Order("received_at DESC")

		if status != "" {
			query = query.Where("status = ?", status)
		}
		if symbol != "" {
			query = query.Where("symbol = ?", symbol)
		}
		if prefix != "" {
			query = query.Where("webhook_prefix = ?", prefix)
		}

		if sinceTsStr != "" {
			// Accept unix seconds or milliseconds
			if ts, err := strconv.ParseInt(sinceTsStr, 10, 64); err == nil && ts > 0 {
				var cutoff time.Time
				// Heuristic: treat > 1e12 as milliseconds
				if ts > 1_000_000_000_000 {
					cutoff = time.UnixMilli(ts)
				} else {
					cutoff = time.Unix(ts, 0)
				}
				query = query.Where("received_at >= ?", cutoff)
			}
		} else if sinceHoursStr != "" {
			if hrs, err := strconv.ParseFloat(sinceHoursStr, 64); err == nil && hrs > 0 {
				cutoff := time.Now().Add(-time.Duration(hrs * float64(time.Hour)))
				query = query.Where("received_at >= ?", cutoff)
			}
		}

		if err := query.Limit(limit).Find(&signals).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch signals"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"signals": signals,
			"count":   len(signals),
		})
	}
}

// ListSignalsByPrefix returns trading signals filtered by webhook prefix via path param
func ListSignalsByPrefix(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var signals []models.TradingSignal

		// Path + query params
		prefix := c.Param("prefix")
		status := c.Query("status")
		symbol := c.Query("symbol")
		limitStr := c.DefaultQuery("limit", "50")
		limit, _ := strconv.Atoi(limitStr)
		sinceHoursStr := c.Query("since_hours")
		sinceTsStr := c.Query("since_ts") // optional: unix seconds or milliseconds from client 'now'

		query := services.DB.Order("received_at DESC")

		if prefix != "" {
			query = query.Where("webhook_prefix = ?", prefix)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if symbol != "" {
			query = query.Where("symbol = ?", symbol)
		}

		if sinceTsStr != "" {
			// Accept unix seconds or milliseconds
			if ts, err := strconv.ParseInt(sinceTsStr, 10, 64); err == nil && ts > 0 {
				var cutoff time.Time
				if ts > 1_000_000_000_000 {
					cutoff = time.UnixMilli(ts)
				} else {
					cutoff = time.Unix(ts, 0)
				}
				query = query.Where("received_at >= ?", cutoff)
			}
		} else if sinceHoursStr != "" {
			if hrs, err := strconv.ParseFloat(sinceHoursStr, 64); err == nil && hrs > 0 {
				cutoff := time.Now().Add(-time.Duration(hrs * float64(time.Hour)))
				query = query.Where("received_at >= ?", cutoff)
			}
		}

		if err := query.Limit(limit).Find(&signals).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch signals"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"signals": signals,
			"count":   len(signals),
		})
	}
}

// GetSignal returns a specific signal
func GetSignal(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		signalID := c.Param("id")

		var signal models.TradingSignal
		if err := services.DB.First(&signal, signalID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signal not found"})
			return
		}

		c.JSON(http.StatusOK, signal)
	}
}

// ExecuteSignal places an order based on a signal and bot config
func ExecuteSignal(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		signalIDStr := c.Param("id")
		signalID, err := strconv.Atoi(signalIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signal ID"})
			return
		}

		var payload struct {
			BotConfigID int `json:"bot_config_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Get signal
		var signal models.TradingSignal
		if err := services.DB.First(&signal, signalID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signal not found"})
			return
		}

		// Get bot config
		var config models.TradingConfig
		if err := services.DB.Where("id = ? AND user_id = ?", payload.BotConfigID, userID).First(&config).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot config not found"})
			return
		}

		// Validate signal is not already executed
		if signal.Status == "executed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Signal already executed"})
			return
		}

		utils.LogInfo(fmt.Sprintf("🎯 Executing signal %d with bot config %d", signalID, payload.BotConfigID))

		// Determine side from action
		side := "buy"
		if signal.Action == "sell" || signal.Action == "short" {
			side = "sell"
		}

		// Use signal price if available, otherwise use market price
		orderType := "market"
		var price float64
		if signal.Price > 0 {
			orderType = "limit"
			price = signal.Price
		}

		// Use config amount
		amount := config.Amount
		if amount <= 0 {
			signal.Status = "failed"
			signal.ErrorMessage = "Bot config amount must be greater than 0"
			signal.ExecutedAt = time.Now()
			services.DB.Save(&signal)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bot config amount must be greater than 0",
			})
			return
		}

		// Decrypt API credentials
		apiKey, err := utils.DecryptString(config.APIKey)
		if err != nil {
			signal.Status = "failed"
			signal.ErrorMessage = "Failed to decrypt API key"
			signal.ExecutedAt = time.Now()
			services.DB.Save(&signal)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to decrypt API credentials",
			})
			return
		}

		apiSecret, err := utils.DecryptString(config.APISecret)
		if err != nil {
			signal.Status = "failed"
			signal.ErrorMessage = "Failed to decrypt API secret"
			signal.ExecutedAt = time.Now()
			services.DB.Save(&signal)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to decrypt API credentials",
			})
			return
		}

		// Place order on exchange
		tradingService := tradingservice.NewTradingService(apiKey, apiSecret, config.Exchange)
		orderResult := tradingService.PlaceOrder(&config, side, orderType, signal.Symbol, amount, price)

		if !orderResult.Success {
			utils.LogError(fmt.Sprintf("❌ Failed to execute signal: %v", orderResult.Error))

			// Update signal status to failed
			signal.Status = "failed"
			signal.ErrorMessage = orderResult.Error
			signal.ExecutedAt = time.Now()
			services.DB.Save(&signal)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to place order",
				"details": orderResult.ErrorDetails,
			})
			return
		}

		// Calculate SL/TP prices
		var stopLoss, takeProfit float64
		filledPrice := orderResult.FilledPrice
		if filledPrice > 0 {
			// Use signal SL/TP if provided, otherwise use config
			if signal.StopLoss > 0 {
				stopLoss = signal.StopLoss
			} else if config.StopLossPercent > 0 {
				if side == "buy" {
					stopLoss = filledPrice * (1 - config.StopLossPercent/100)
				} else {
					stopLoss = filledPrice * (1 + config.StopLossPercent/100)
				}
			}

			if signal.TakeProfit > 0 {
				takeProfit = signal.TakeProfit
			} else if config.TakeProfitPercent > 0 {
				if side == "buy" {
					takeProfit = filledPrice * (1 + config.TakeProfitPercent/100)
				} else {
					takeProfit = filledPrice * (1 - config.TakeProfitPercent/100)
				}
			}
		}

		// Create order record
		order := models.Order{
			UserID:          userID.(uint),
			BotConfigID:     config.ID,
			Exchange:        config.Exchange,
			Symbol:          orderResult.Symbol,
			OrderID:         orderResult.OrderID,
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
			signal.Status = "failed"
			signal.ErrorMessage = fmt.Sprintf("Failed to create order record: %v", err)
			signal.ExecutedAt = time.Now()
			services.DB.Save(&signal)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create order record",
			})
			return
		}

		// Update signal status
		signal.Status = "executed"
		signal.OrderID = &order.ID
		signal.ExecutedAt = time.Now()
		services.DB.Save(&signal)

		utils.LogInfo(fmt.Sprintf("✅ Signal %d executed successfully, Order ID: %d", signalID, order.ID))

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"signal":  signal,
			"order":   order,
			"message": "Order placed successfully",
		})
	}
}

// UpdateSignalStatus updates the status of a signal (mark as ignored, etc.)
func UpdateSignalStatus(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		signalID := c.Param("id")

		var payload struct {
			Status string `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var signal models.TradingSignal
		if err := services.DB.First(&signal, signalID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signal not found"})
			return
		}

		signal.Status = payload.Status
		if err := services.DB.Save(&signal).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update signal"})
			return
		}

		c.JSON(http.StatusOK, signal)
	}
}

// DeleteSignal deletes a signal
func DeleteSignal(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		signalID := c.Param("id")

		if err := services.DB.Delete(&models.TradingSignal{}, signalID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete signal"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Signal deleted successfully"})
	}
}
