package utils

import (
	"encoding/json"
	"fmt"
	"time"
	"tradercoin/backend/models"

	"gorm.io/gorm"
)

// CreateSystemLog creates a system log entry in the database
func CreateSystemLog(db *gorm.DB, userID uint, level, action, message string, options map[string]interface{}) error {
	log := models.SystemLog{
		UserID:    userID,
		Level:     level,
		Action:    action,
		Message:   message,
		CreatedAt: time.Now(),
	}

	// Extract optional fields from map
	if symbol, ok := options["symbol"].(string); ok {
		log.Symbol = symbol
	}
	if exchange, ok := options["exchange"].(string); ok {
		log.Exchange = exchange
	}
	if orderID, ok := options["order_id"].(uint); ok {
		log.OrderID = &orderID
	}
	if price, ok := options["price"].(float64); ok {
		log.Price = price
	}
	if amount, ok := options["amount"].(float64); ok {
		log.Amount = amount
	}
	if ipAddress, ok := options["ip_address"].(string); ok {
		log.IPAddress = ipAddress
	}
	if userAgent, ok := options["user_agent"].(string); ok {
		log.UserAgent = userAgent
	}
	if details, ok := options["details"]; ok {
		if detailsJSON, err := json.Marshal(details); err == nil {
			log.Details = string(detailsJSON)
		}
	}

	if err := db.Create(&log).Error; err != nil {
		LogError(fmt.Sprintf("Failed to create system log: %v", err))
		return err
	}

	return nil
}
