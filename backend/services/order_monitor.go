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
	// Query all orders with pending status (new, pending, partially_filled)
	var orders []models.Order
	err := oms.DB.Where("LOWER(status) IN ?", []string{"new", "pending", "partially_filled"}).
		Preload("User"). // Load user info
		Find(&orders).Error

	if err != nil {
		log.Printf("❌ Failed to query pending orders: %v", err)
		return
	}

	if len(orders) == 0 {
		// log.Println("📊 No pending orders to check")
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
		statusResult := tradingService.CheckOrderStatus(&config, order.OrderID, order.Symbol)

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

		if newStatusLower != oldStatusLower {
			// Update order in database
			order.Status = newStatus

			// Update filled price and quantity for filled orders
			if newStatusLower == "filled" {
				if statusResult.AvgPrice > 0 {
					order.FilledPrice = statusResult.AvgPrice
				}
				order.FilledQuantity = statusResult.Filled

				log.Printf("✅ Order %d: %s → %s (Filled Price: %.8f, Qty: %.8f)",
					order.ID, oldStatus, newStatus, order.FilledPrice, order.FilledQuantity)
			} else {
				log.Printf("✅ Order %d: %s → %s", order.ID, oldStatus, newStatus)
			}

			// Save to database
			if err := oms.DB.Save(&order).Error; err != nil {
				log.Printf("❌ Order %d: Failed to update in DB: %v", order.ID, err)
				errorCount++
				continue
			}

			updatedCount++

			// Send WebSocket notification to user
			oms.notifyOrderUpdate(order.UserID, order.ID)
		}
	}

	log.Printf("🔷 ===== ORDER MONITOR - Complete: %d updated, %d errors =====\n", updatedCount, errorCount)
}

// notifyOrderUpdate sends WebSocket notification to user
func (oms *OrderMonitorService) notifyOrderUpdate(userID uint, orderID uint) {
	if oms.WebSocketHub == nil {
		return
	}

	message := WebSocketMessage{
		Type: "order_update",
		Data: map[string]interface{}{
			"order_id":  orderID,
			"timestamp": time.Now().Unix(),
		},
	}

	oms.WebSocketHub.BroadcastToUser(userID, message)
	log.Printf("📤 WebSocket notification sent to user %d for order %d", userID, orderID)
}
