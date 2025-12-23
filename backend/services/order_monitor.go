package services

import (
	"log"
	"strings"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/utils"

	"gorm.io/gorm"
)

// OrderMonitorService monitors pending orders and updates their status
type OrderMonitorService struct {
	DB             *gorm.DB
	WebSocketHub   *WebSocketHub
	tickerInterval time.Duration
	stopChan       chan bool
}

// NewOrderMonitorService creates a new order monitor service
func NewOrderMonitorService(db *gorm.DB, wsHub *WebSocketHub) *OrderMonitorService {
	return &OrderMonitorService{
		DB:             db,
		WebSocketHub:   wsHub,
		tickerInterval: 5 * time.Second, // Check every 5 seconds
		stopChan:       make(chan bool),
	}
}

// Start begins the background monitoring process
func (oms *OrderMonitorService) Start() {
	log.Println("🔄 Order Monitor Service started - checking every 5 seconds")

	ticker := time.NewTicker(oms.tickerInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				oms.checkPendingOrders()
			case <-oms.stopChan:
				ticker.Stop()
				log.Println("⏹️  Order Monitor Service stopped")
				return
			}
		}
	}()
}

// Stop stops the monitoring service
func (oms *OrderMonitorService) Stop() {
	oms.stopChan <- true
}

// checkPendingOrders checks all pending orders from exchange
func (oms *OrderMonitorService) checkPendingOrders() {
	// Query orders to monitor:
	// - Spot: new, pending, partially_filled
	// - Futures: all except 'closed' (including 'filled' because position is still open)
	var orders []models.Order
	err := oms.DB.Where(
		"(LOWER(trading_mode) IN (?, ?) AND LOWER(status) != ?) OR "+
			"((trading_mode IS NULL OR LOWER(trading_mode) = ?) AND LOWER(status) IN (?, ?, ?))",
		"futures", "future", "closed", // Futures: monitor all except closed
		"spot", "new", "pending", "partially_filled", // Spot: only monitor pending statuses
	).Preload("User"). // Load user info
				Find(&orders).Error

	if err != nil {
		log.Printf("❌ Failed to query pending orders: %v", err)
		return
	}

	if len(orders) == 0 {
		log.Println("📊 No pending orders to check")
		return
	}

	log.Printf("\n🔍 ===== ORDER MONITOR - Checking %d pending orders =====", len(orders))

	// Group orders by bot_config_id to batch load configs
	configIDs := make(map[uint]bool)
	for _, order := range orders {
		if order.BotConfigID > 0 {
			configIDs[order.BotConfigID] = true
		}
	}

	// Load all bot configs in one query
	configIDList := make([]uint, 0, len(configIDs))
	for id := range configIDs {
		configIDList = append(configIDList, id)
	}

	var configs []models.TradingConfig
	configMap := make(map[uint]models.TradingConfig)
	if len(configIDList) > 0 {
		oms.DB.Where("id IN ?", configIDList).Find(&configs)
		for _, config := range configs {
			configMap[config.ID] = config
		}
	}

	log.Printf("📦 Loaded %d bot configs", len(configMap))

	// Check each order
	updatedCount := 0
	errorCount := 0

	for _, order := range orders {
		// Get bot config
		config, exists := configMap[order.BotConfigID]
		if !exists {
			log.Printf("⚠️  Order %d: Bot config %d not found", order.ID, order.BotConfigID)
			continue
		}

		// Skip if no order ID
		if order.OrderID == "" {
			continue
		}

		// Decrypt API credentials
		apiKey, err := utils.DecryptString(config.APIKey)
		if err != nil {
			log.Printf("⚠️  Order %d: Failed to decrypt API key: %v", order.ID, err)
			errorCount++
			continue
		}

		apiSecret, err := utils.DecryptString(config.APISecret)
		if err != nil {
			log.Printf("⚠️  Order %d: Failed to decrypt API secret: %v", order.ID, err)
			errorCount++
			continue
		}

		// Check order status from exchange
		tradingService := NewTradingService(apiKey, apiSecret, order.Exchange)
		statusResult := tradingService.CheckOrderStatus(&config, order.OrderID, order.Symbol, order.AlgoIDStopLoss)

		if !statusResult.Success {
			log.Printf("⚠️  Order %d: Failed to check status - %s", order.ID, statusResult.Error)
			errorCount++
			continue
		}

		// Check if status changed
		oldStatus := order.Status
		newStatus := statusResult.Status
		newStatusLower := strings.ToLower(newStatus)
		oldStatusLower := strings.ToLower(oldStatus)

		// Variable to store position info for notification
		var positionInfo *FuturesPositionInfo

		// For Futures: Check if order/position is still running
		if strings.ToLower(order.TradingMode) == "futures" || strings.ToLower(order.TradingMode) == "future" {
			if statusResult.IsRunning {
				// Order hoặc Algo Order vẫn đang chạy - Get position info
				log.Printf("🔍 Order %d: Calling GetFuturesPosition for symbol=%s", order.ID, order.Symbol)
				position, err := tradingService.GetFuturesPosition(order.Symbol)
				if err != nil {
					log.Printf("⚠️  Order %d: Failed to get position info: %v", order.ID, err)
				} else if position != nil {
					positionInfo = position // Save for WebSocket notification
					log.Printf("🔍 Order %d (Futures): Status=%s, IsRunning=true, Type=%s, AlgoStatus=%s",
						order.ID, statusResult.Status, statusResult.RunningType, statusResult.AlgoStatus)
					log.Printf("   📊 Position Info:")
					log.Printf("      Symbol: %s | Size: %.3f %s",
						position.Symbol, position.PositionAmt, position.PositionSide)
					log.Printf("      Entry Price: %.2f | Mark Price: %.2f | Liq.Price: %.2f",
						position.EntryPrice, position.MarkPrice, position.LiquidationPrice)
					log.Printf("      PnL: %.2f USDT (%.2f%%) | Margin: %.2f USDT | Leverage: %dx",
						position.UnrealizedProfit, position.PnlPercent, position.IsolatedMargin, position.Leverage)

					// Update database with position info (entry price, leverage, PnL)
					updateFields := make(map[string]interface{})
					if order.FilledPrice == 0 && position.EntryPrice > 0 {
						order.FilledPrice = position.EntryPrice
						order.Price = position.EntryPrice // Also set Price if not set
						updateFields["filled_price"] = position.EntryPrice
						updateFields["price"] = position.EntryPrice
						log.Printf("   📝 Updated FilledPrice from position: %.2f", position.EntryPrice)
					}
					if order.Leverage == 0 && position.Leverage > 0 {
						order.Leverage = position.Leverage
						updateFields["leverage"] = position.Leverage
						log.Printf("   📝 Updated Leverage from position: %dx", position.Leverage)
					}
					if len(updateFields) > 0 {
						if err := oms.DB.Model(&models.Order{}).Where("id = ?", order.ID).Updates(updateFields).Error; err != nil {
							log.Printf("⚠️  Order %d: Failed to update position info: %v", order.ID, err)
						} else {
							log.Printf("✅ Order %d: Updated FilledPrice and Leverage in database", order.ID)
						}
					}

					// Send WebSocket update with position info (even if status not changed)
					oms.notifyOrderUpdate(order.UserID, order.ID, &order, positionInfo)
				} else {
					log.Printf("⚠️  Order %d: GetFuturesPosition returned nil (no position or positionAmt=0)", order.ID)
					log.Printf("🔍 Order %d (Futures): Status=%s, IsRunning=true, Type=%s, AlgoStatus=%s (No position)",
						order.ID, statusResult.Status, statusResult.RunningType, statusResult.AlgoStatus)
				}
				// Không update gì, order vẫn active
				continue
			} else {
				// Order và Algo Order đã không còn chạy → Close position
				log.Printf("🔍 Order %d (Futures): Status=%s, IsRunning=false → Setting to CLOSED",
					order.ID, statusResult.Status)
				newStatus = "closed"
				newStatusLower = "closed"
			}
		} else {
			// For Spot: Just log status
			log.Printf("🔍 Order %d (Spot): Status=%s", order.ID, statusResult.Status)
		}

		if newStatusLower != oldStatusLower {
			// Update order in database
			order.Status = newStatus

			// Update filled price and quantity for filled orders
			if newStatusLower == "filled" && order.TradingMode == "spot" {
				if statusResult.AvgPrice > 0 {
					order.FilledPrice = statusResult.AvgPrice
				}
				order.FilledQuantity = statusResult.Filled

				log.Printf("✅ Order %d: %s → %s (Filled Price: %.8f, Qty: %.8f)",
					order.ID, oldStatus, newStatus, order.FilledPrice, order.FilledQuantity)
			} else {
				log.Printf("✅ Order %d: %s → %s", order.ID, oldStatus, newStatus)
			}

			// Validate order ID before update
			if order.ID == 0 {
				log.Printf("⚠️  Order has invalid ID (0), skipping update. OrderID: %s", order.OrderID)
				errorCount++
				continue
			}

			// Save to database using primary key ID
			updateData := map[string]interface{}{
				"status":          newStatus,
				"filled_price":    order.FilledPrice,
				"filled_quantity": order.FilledQuantity,
			}
			if err := oms.DB.Model(&models.Order{}).Where("id = ?", order.ID).Updates(updateData).Error; err != nil {
				log.Printf("❌ Order %d (OrderID: %s): Failed to update in DB: %v", order.ID, order.OrderID, err)
				errorCount++
				continue
			}

			updatedCount++

			// Send WebSocket notification to user (with position info if available)
			oms.notifyOrderUpdate(order.UserID, order.ID, &order, positionInfo)
		}
	}

	log.Printf("🔷 ===== ORDER MONITOR - Complete: %d updated, %d errors =====\n", updatedCount, errorCount)
}

// notifyOrderUpdate sends WebSocket notification to user with position info
func (oms *OrderMonitorService) notifyOrderUpdate(userID uint, orderID uint, order *models.Order, position *FuturesPositionInfo) {
	if oms.WebSocketHub == nil {
		return
	}

	// Base order data
	data := map[string]interface{}{
		"order_id":     orderID,
		"timestamp":    time.Now().Unix(),
		"symbol":       order.Symbol,
		"side":         order.Side,
		"status":       order.Status,
		"trading_mode": order.TradingMode,
	}

	// Add position info if available (for futures)
	if position != nil {
		data["position"] = map[string]interface{}{
			"symbol":            position.Symbol,
			"position_amt":      position.PositionAmt,
			"position_side":     position.PositionSide,
			"entry_price":       position.EntryPrice,
			"mark_price":        position.MarkPrice,
			"liquidation_price": position.LiquidationPrice,
			"unrealized_profit": position.UnrealizedProfit,
			"pnl_percent":       position.PnlPercent,
			"leverage":          position.Leverage,
			"margin_type":       position.MarginType,
			"isolated_margin":   position.IsolatedMargin,
		}
	}

	message := WebSocketMessage{
		Type: "order_update",
		Data: data,
	}

	oms.WebSocketHub.BroadcastToUser(userID, message)

	if position != nil {
		log.Printf("📤 WebSocket notification sent to user %d for order %d (with position info)", userID, orderID)
	} else {
		log.Printf("📤 WebSocket notification sent to user %d for order %d", userID, orderID)
	}
}
