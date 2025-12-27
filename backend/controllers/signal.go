package controllers

import (
	"fmt"
	"log"
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

		// Create signal record (NO STATUS - shared by all users)
		signal := models.TradingSignal{
			Symbol:        payload.Symbol,
			Action:        payload.Action,
			Price:         payload.Price,
			StopLoss:      payload.StopLoss,
			TakeProfit:    payload.TakeProfit,
			Message:       payload.Message,
			Strategy:      payload.Strategy,
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

// ListSignals returns all trading signals with user-specific status
func ListSignals(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Response struct with signal + user status
		type SignalWithStatus struct {
			models.TradingSignal
			Status           string     `json:"status"`              // From user_signals
			OrderID          *uint      `json:"order_id"`            // From user_signals
			ExecutedByUserID *uint      `json:"executed_by_user_id"` // From user_signals.user_id
			ExecutedAt       *time.Time `json:"executed_at"`         // From user_signals
			ErrorMessage     string     `json:"error_message"`       // From user_signals
		}

		// Query params
		status := c.Query("status")
		symbol := c.Query("symbol")
		prefix := c.Query("prefix")
		limitStr := c.DefaultQuery("limit", "50")
		limit, _ := strconv.Atoi(limitStr)
		sinceHoursStr := c.Query("since_hours")
		sinceTsStr := c.Query("since_ts")

		// Base query: LEFT JOIN to get all signals + user status if exists
		query := services.DB.Table("trading_signals").
			Select(`trading_signals.*, 
				COALESCE(user_signals.status, 'pending') as status,
				user_signals.order_id,
				user_signals.user_id as executed_by_user_id,
				user_signals.executed_at,
				user_signals.error_msg as error_message`).
			Joins("LEFT JOIN user_signals ON user_signals.signal_id = trading_signals.id AND user_signals.user_id = ?", userID).
			Order("trading_signals.received_at DESC")

		if status != "" {
			query = query.Where("COALESCE(user_signals.status, 'pending') = ?", status)
		}
		if symbol != "" {
			query = query.Where("trading_signals.symbol = ?", symbol)
		}
		if prefix != "" {
			query = query.Where("trading_signals.webhook_prefix = ?", prefix)
		}

		if sinceTsStr != "" {
			if ts, err := strconv.ParseInt(sinceTsStr, 10, 64); err == nil && ts > 0 {
				var cutoff time.Time
				if ts > 1_000_000_000_000 {
					cutoff = time.UnixMilli(ts)
				} else {
					cutoff = time.Unix(ts, 0)
				}
				query = query.Where("trading_signals.received_at >= ?", cutoff)
			}
		} else if sinceHoursStr != "" {
			if hrs, err := strconv.ParseFloat(sinceHoursStr, 64); err == nil && hrs > 0 {
				cutoff := time.Now().Add(-time.Duration(hrs * float64(time.Hour)))
				query = query.Where("trading_signals.received_at >= ?", cutoff)
			}
		}

		var signals []SignalWithStatus
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

		// Check if user already executed this signal
		var existingUserSignal models.UserSignal
		if err := services.DB.Where("user_id = ? AND signal_id = ?", userID, signalID).First(&existingUserSignal).Error; err == nil {
			// User already has a record for this signal
			if existingUserSignal.Status == "executed" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "You already executed this signal"})
				return
			}
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
		// if signal.Price > 0 {
		// 	orderType = "limit"
		// 	price = signal.Price
		// }

		// Use config amount
		amount := config.Amount
		if amount <= 0 {
			// Create failed UserSignal record
			now := time.Now()
			userSignal := models.UserSignal{
				UserID:      userID.(uint),
				SignalID:    uint(signalID),
				Status:      "failed",
				BotConfigID: &config.ID,
				ExecutedAt:  &now,
				ErrorMsg:    "Bot config amount must be greater than 0",
			}
			services.DB.Create(&userSignal)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bot config amount must be greater than 0",
			})
			return
		}

		////////////// Decrypt API credentials //////////////
		apiKey, apiSecret, err := GetDecryptedAPICredentials(&config)
		if err != nil {
			log.Printf("Failed to decrypt API credentials: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API credentials"})
			return
		}

		// Place order on exchange (LIVE MODE)
		utils.LogInfo(fmt.Sprintf("🔍 DEBUG PlaceOrder params: side=%s, orderType=%s, symbol=%s, amount=%.8f, price=%.8f",
			side, orderType, signal.Symbol, amount, price))

		tradingService := tradingservice.NewTradingService(apiKey, apiSecret, config.Exchange, services.DB, userID.(uint))
		orderResult := tradingService.PlaceOrder(&config, side, orderType, signal.Symbol, amount, price)

		if !orderResult.Success {
			utils.LogError(fmt.Sprintf("❌ Failed to execute signal: %v", orderResult.Error))

			// Create failed UserSignal record
			now := time.Now()
			userSignal := models.UserSignal{
				UserID:      userID.(uint),
				SignalID:    uint(signalID),
				Status:      "failed",
				BotConfigID: &config.ID,
				ExecutedAt:  &now,
				ErrorMsg:    orderResult.Error,
			}
			services.DB.Create(&userSignal)

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

		// Create order record using Algo IDs from orderResult
		// Service has already placed SL/TP orders and returned their IDs
		order := models.Order{
			UserID:           userID.(uint),
			BotConfigID:      config.ID,
			Exchange:         config.Exchange,
			Symbol:           orderResult.Symbol,
			OrderID:          orderResult.OrderID, // Exchange order ID
			Side:             orderResult.Side,
			Type:             orderResult.Type,
			Quantity:         orderResult.Quantity,
			Price:            orderResult.Price,
			FilledPrice:      orderResult.FilledPrice,
			Status:           orderResult.Status,
			TradingMode:      config.TradingMode,
			Leverage:         config.Leverage,
			StopLossPrice:    stopLoss,
			TakeProfitPrice:  takeProfit,
			AlgoIDStopLoss:   orderResult.AlgoIDStopLoss,   // Use from service
			AlgoIDTakeProfit: orderResult.AlgoIDTakeProfit, // Use from service
			PnL:              0,
			PnLPercent:       0,
		}

		if err := services.DB.Create(&order).Error; err != nil {
			now := time.Now()
			userSignal := models.UserSignal{
				UserID:      userID.(uint),
				SignalID:    uint(signalID),
				Status:      "failed",
				BotConfigID: &config.ID,
				ExecutedAt:  &now,
				ErrorMsg:    fmt.Sprintf("Failed to create order record: %v", err),
			}
			services.DB.Create(&userSignal)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create order record",
			})
			return
		}

		//////////////////////////////////////////////////////////////////////

		// Create UserSignal record for this user
		now := time.Now()
		userSignal := models.UserSignal{
			UserID:      userID.(uint),
			SignalID:    uint(signalID),
			Status:      "executed",
			BotConfigID: &config.ID,
			OrderID:     &order.ID,
			ExecutedAt:  &now,
		}
		if err := services.DB.Create(&userSignal).Error; err != nil {
			utils.LogError(fmt.Sprintf("❌ Failed to create UserSignal: %v", err))
		}

		utils.LogInfo(fmt.Sprintf("✅ Signal %d executed successfully by user %d, Order ID: %d", signalID, userID, order.ID))

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"signal":  signal,
			"order":   order,
			"message": "Order placed successfully",
		})
	}
}

// UpdateSignalStatus updates the status of a signal for current user (mark as ignored, etc.)
func UpdateSignalStatus(services *services.Services) gin.HandlerFunc {
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
			Status string `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Check if signal exists
		var signal models.TradingSignal
		if err := services.DB.First(&signal, signalID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signal not found"})
			return
		}

		// Create or update UserSignal record for this user
		var userSignal models.UserSignal
		result := services.DB.Where("user_id = ? AND signal_id = ?", userID, signalID).First(&userSignal)

		if result.Error != nil {
			// Create new UserSignal
			userSignal = models.UserSignal{
				UserID:   userID.(uint),
				SignalID: uint(signalID),
				Status:   payload.Status,
			}
			if err := services.DB.Create(&userSignal).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user signal"})
				return
			}
		} else {
			// Update existing UserSignal
			userSignal.Status = payload.Status
			if err := services.DB.Save(&userSignal).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user signal"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"signal":      signal,
			"user_status": userSignal.Status,
		})
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
