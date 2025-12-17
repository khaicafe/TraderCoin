package controllers

import (
	"log"
	"net/http"
	"strconv"
	"tradercoin/backend/models"
	"tradercoin/backend/services"
	"tradercoin/backend/utils"

	"github.com/gin-gonic/gin"
)

// CreateBotConfig - Tạo bot configuration mới
func CreateBotConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Verify user exists
		var user models.User
		if err := services.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Bind request data
		var input struct {
			Name                  string                   `json:"name" binding:"required"`
			Symbol                string                   `json:"symbol" binding:"required"`
			Exchange              string                   `json:"exchange" binding:"required"`
			Amount                float64                  `json:"amount"`
			TradingMode           string                   `json:"trading_mode"`
			Leverage              int                      `json:"leverage"`
			APIKey                string                   `json:"api_key"`
			APISecret             string                   `json:"api_secret"`
			StopLossPercent       float64                  `json:"stop_loss_percent" binding:"required,gte=0,lte=100"`
			TakeProfitPercent     float64                  `json:"take_profit_percent" binding:"required,gte=0,lte=1000"`
			TPLevels              []map[string]interface{} `json:"tp_levels"`
			EnableTrailing        bool                     `json:"enable_trailing"`
			TrailingType          string                   `json:"trailing_type"`
			TrailingPercent       *float64                 `json:"trailing_percent"`
			TrailingATRMultiplier *float64                 `json:"trailing_atr_multiplier"`
			TrailingATRPeriod     *int                     `json:"trailing_atr_period"`
			IPWhitelist           []string                 `json:"ip_whitelist"`
			MaxOpenPositions      int                      `json:"max_open_positions"`
			EnableNotifications   bool                     `json:"enable_notifications"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate exchange
		if input.Exchange != "binance" && input.Exchange != "bittrex" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange. Must be 'binance' or 'bittrex'"})
			return
		}

		// Validate trading mode if provided
		if input.TradingMode != "" {
			if input.TradingMode != "spot" && input.TradingMode != "futures" && input.TradingMode != "margin" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trading mode. Must be 'spot', 'futures', or 'margin'"})
				return
			}
		} else {
			input.TradingMode = "spot" // Default to spot
		}

		// Validate leverage
		if input.Leverage < 1 || input.Leverage > 125 {
			input.Leverage = 1 // Default to 1x
		}

		// Encrypt API credentials if provided
		var encryptedAPIKey, encryptedAPISecret string
		var err error
		if input.APIKey != "" {
			encryptedAPIKey, err = utils.EncryptString(input.APIKey)
			if err != nil {
				log.Printf("Failed to encrypt API key: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API credentials"})
				return
			}
		}
		if input.APISecret != "" {
			encryptedAPISecret, err = utils.EncryptString(input.APISecret)
			if err != nil {
				log.Printf("Failed to encrypt API secret: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API credentials"})
				return
			}
		}

		// Create trading config
		config := models.TradingConfig{
			UserID:            user.ID,
			Symbol:            input.Symbol,
			Exchange:          input.Exchange,
			Amount:            input.Amount,
			TradingMode:       input.TradingMode,
			Leverage:          input.Leverage,
			APIKey:            encryptedAPIKey,
			APISecret:         encryptedAPISecret,
			StopLossPercent:   input.StopLossPercent,
			TakeProfitPercent: input.TakeProfitPercent,
			IsActive:          true, // Active by default
		}

		if err := services.DB.Create(&config).Error; err != nil {
			log.Printf("Failed to create bot config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
			return
		}

		log.Printf("Bot config created: %d - %s for user %d", config.ID, input.Name, user.ID)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Bot configuration created successfully",
			"config":  config,
		})
	}
}

// ListBotConfigs - Lấy danh sách tất cả bot configurations
func ListBotConfigs(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get pagination params
		skip := 0
		limit := 100
		if skipStr := c.Query("skip"); skipStr != "" {
			skip, _ = strconv.Atoi(skipStr)
		}
		if limitStr := c.Query("limit"); limitStr != "" {
			limit, _ = strconv.Atoi(limitStr)
		}

		// Query configs with ordering
		var configs []models.TradingConfig
		if err := services.DB.Where("user_id = ?", userID).
			Order("id DESC").
			Offset(skip).
			Limit(limit).
			Find(&configs).Error; err != nil {
			log.Printf("Error listing configs: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
			return
		}

		log.Printf("Found %d configs for user %v", len(configs), userID)

		c.JSON(http.StatusOK, gin.H{
			"configs": configs,
			"total":   len(configs),
		})
	}
}

// GetBotConfig - Lấy bot configuration cụ thể
func GetBotConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get config ID from URL
		configID := c.Param("id")

		// Get user ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Query config
		var config models.TradingConfig
		if err := services.DB.Where("id = ? AND user_id = ?", configID, userID).
			First(&config).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}

		c.JSON(http.StatusOK, config)
	}
}

// UpdateBotConfig - Cập nhật bot configuration
func UpdateBotConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get config ID
		configID := c.Param("id")

		// Get user ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Find config
		var config models.TradingConfig
		if err := services.DB.Where("id = ? AND user_id = ?", configID, userID).
			First(&config).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}

		// Bind update data
		var input struct {
			Symbol            *string  `json:"symbol"`
			Exchange          *string  `json:"exchange"`
			Amount            *float64 `json:"amount"`
			TradingMode       *string  `json:"trading_mode"`
			Leverage          *int     `json:"leverage"`
			APIKey            *string  `json:"api_key"`
			APISecret         *string  `json:"api_secret"`
			StopLossPercent   *float64 `json:"stop_loss_percent"`
			TakeProfitPercent *float64 `json:"take_profit_percent"`
			IsActive          *bool    `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update fields if provided
		if input.Symbol != nil {
			config.Symbol = *input.Symbol
		}
		if input.Exchange != nil {
			if *input.Exchange != "binance" && *input.Exchange != "bittrex" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange"})
				return
			}
			config.Exchange = *input.Exchange
		}
		if input.Amount != nil {
			config.Amount = *input.Amount
		}
		if input.TradingMode != nil {
			if *input.TradingMode != "spot" && *input.TradingMode != "futures" && *input.TradingMode != "margin" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trading mode. Must be 'spot', 'futures', or 'margin'"})
				return
			}
			config.TradingMode = *input.TradingMode
		}
		if input.Leverage != nil {
			if *input.Leverage < 1 || *input.Leverage > 125 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Leverage must be between 1 and 125"})
				return
			}
			config.Leverage = *input.Leverage
		}
		if input.APIKey != nil {
			encryptedKey, err := utils.EncryptString(*input.APIKey)
			if err != nil {
				log.Printf("Failed to encrypt API key: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
				return
			}
			config.APIKey = encryptedKey
		}
		if input.APISecret != nil {
			encryptedSecret, err := utils.EncryptString(*input.APISecret)
			if err != nil {
				log.Printf("Failed to encrypt API secret: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API secret"})
				return
			}
			config.APISecret = encryptedSecret
		}
		if input.StopLossPercent != nil {
			if *input.StopLossPercent < 0 || *input.StopLossPercent > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Stop loss must be between 0 and 100"})
				return
			}
			config.StopLossPercent = *input.StopLossPercent
		}
		if input.TakeProfitPercent != nil {
			if *input.TakeProfitPercent < 0 || *input.TakeProfitPercent > 1000 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Take profit must be between 0 and 1000"})
				return
			}
			config.TakeProfitPercent = *input.TakeProfitPercent
		}
		if input.IsActive != nil {
			config.IsActive = *input.IsActive
		}

		// Save updates
		if err := services.DB.Save(&config).Error; err != nil {
			log.Printf("Failed to update config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
			return
		}

		log.Printf("Bot config updated: %d", config.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Bot configuration updated successfully",
			"config":  config,
		})
	}
}

// DeleteBotConfig - Xóa bot configuration
func DeleteBotConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get config ID
		configID := c.Param("id")

		// Get user ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Find config
		var config models.TradingConfig
		if err := services.DB.Where("id = ? AND user_id = ?", configID, userID).
			First(&config).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}

		// Check if there are orders associated with this config
		var ordersCount int64
		services.DB.Model(&models.Order{}).Where("trading_config_id = ?", configID).Count(&ordersCount)

		if ordersCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        "Cannot delete bot config: orders are associated with this config. Please delete or reassign orders first.",
				"orders_count": ordersCount,
			})
			return
		}

		// Delete config
		if err := services.DB.Delete(&config).Error; err != nil {
			log.Printf("Failed to delete config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
			return
		}

		log.Printf("Bot config deleted: %d", config.ID)

		c.JSON(http.StatusNoContent, nil)
	}
}

// SetDefaultBotConfig - Set bot configuration làm default
func SetDefaultBotConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get config ID
		configID := c.Param("id")

		// Get user ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Find config
		var config models.TradingConfig
		if err := services.DB.Where("id = ? AND user_id = ?", configID, userID).
			First(&config).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}

		// Start transaction
		tx := services.DB.Begin()

		// Unset all other default configs for this user
		if err := tx.Model(&models.TradingConfig{}).
			Where("user_id = ? AND id != ?", userID, configID).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to unset other defaults: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default configuration"})
			return
		}

		// Set this config as default
		config.IsDefault = true
		if err := tx.Save(&config).Error; err != nil {
			tx.Rollback()
			log.Printf("Failed to set default: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default configuration"})
			return
		}

		// Commit transaction
		tx.Commit()

		log.Printf("Bot config %d set as default for user %v", config.ID, userID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Bot configuration set as default successfully",
			"config":  config,
		})
	}
}
