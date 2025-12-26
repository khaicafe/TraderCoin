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
		log.Printf("\n🔷 ===== CREATE BOT CONFIG - START =====")

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			log.Printf("❌ Step 1: User authentication failed - no user_id in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		log.Printf("✅ Step 1: User authenticated - UserID: %v", userID)

		// Verify user exists
		var user models.User
		if err := services.DB.First(&user, userID).Error; err != nil {
			log.Printf("❌ Step 2: User verification failed - User %v not found in database: %v", userID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("✅ Step 2: User verified - Email: %s, Name: %s", user.Email, user.FullName)

		// Bind request data
		log.Printf("📝 Step 3: Parsing request body...")
		var input struct {
			Name                  string                   `json:"name" binding:"required"`
			Symbol                string                   `json:"symbol" binding:"required"`
			Exchange              string                   `json:"exchange" binding:"required"`
			Amount                float64                  `json:"amount"`
			TradingMode           string                   `json:"trading_mode"`
			Leverage              int                      `json:"leverage"`
			MarginMode            string                   `json:"margin_mode"` // ISOLATED or CROSSED
			APIKey                string                   `json:"api_key"`
			APISecret             string                   `json:"api_secret"`
			StopLossPercent       float64                  `json:"stop_loss_percent" binding:"gte=0,lte=100"`    // Optional, 0 = không dùng SL
			TakeProfitPercent     float64                  `json:"take_profit_percent" binding:"gte=0,lte=1000"` // Optional, 0 = không dùng TP
			TrailingStopPercent   float64                  `json:"trailing_stop_percent"`
			EnableTrailingStop    bool                     `json:"enable_trailing_stop"` // Enable/disable trailing stop
			ActivationPrice       float64                  `json:"activation_price"`     // Activation price for trailing stop
			CallbackRate          float64                  `json:"callback_rate"`        // Callback rate for trailing stop
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
			log.Printf("❌ Step 3: JSON binding failed - %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("✅ Step 3: Request body parsed successfully")
		log.Printf("   📊 Config details:")
		log.Printf("      - Name: %s", input.Name)
		log.Printf("      - Symbol: %s", input.Symbol)
		log.Printf("      - Exchange: %s", input.Exchange)
		log.Printf("      - Amount: %.8f", input.Amount)
		log.Printf("      - Trading Mode: %s", input.TradingMode)
		log.Printf("      - Leverage: %d", input.Leverage)
		log.Printf("      - Margin Mode: %s", input.MarginMode)
		log.Printf("      - Stop Loss %%: %.2f", input.StopLossPercent)
		log.Printf("      - Take Profit %%: %.2f", input.TakeProfitPercent)
		log.Printf("      - Trailing Stop %%: %.2f", input.TrailingStopPercent)
		log.Printf("      - Enable Trailing Stop: %t", input.EnableTrailingStop)
		log.Printf("      - Activation Price: %.8f", input.ActivationPrice)
		log.Printf("      - Callback Rate: %.2f", input.CallbackRate)
		log.Printf("      - API Key provided: %t", input.APIKey != "")
		log.Printf("      - API Secret provided: %t", input.APISecret != "")

		// Validate exchange
		log.Printf("🔍 Step 4: Validating exchange...")
		if input.Exchange != "binance" && input.Exchange != "bittrex" {
			log.Printf("❌ Step 4: Invalid exchange '%s' - must be 'binance' or 'bittrex'", input.Exchange)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange. Must be 'binance' or 'bittrex'"})
			return
		}
		log.Printf("✅ Step 4: Exchange '%s' validated", input.Exchange)

		// Validate trading mode if provided
		log.Printf("🔍 Step 5: Validating trading mode...")
		if input.TradingMode != "" {
			if input.TradingMode != "spot" && input.TradingMode != "futures" && input.TradingMode != "margin" {
				log.Printf("❌ Step 5: Invalid trading mode '%s' - must be 'spot', 'futures', or 'margin'", input.TradingMode)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trading mode. Must be 'spot', 'futures', or 'margin'"})
				return
			}
			log.Printf("✅ Step 5: Trading mode '%s' validated", input.TradingMode)
		} else {
			input.TradingMode = "spot" // Default to spot
			log.Printf("✅ Step 5: Trading mode not provided, defaulting to 'spot'")
		}

		// Validate leverage
		log.Printf("🔍 Step 6: Validating leverage...")
		originalLeverage := input.Leverage
		if input.Leverage < 1 || input.Leverage > 125 {
			input.Leverage = 1 // Default to 1x
			log.Printf("⚠️  Step 6: Leverage %d out of range (1-125), defaulting to 1x", originalLeverage)
		} else {
			log.Printf("✅ Step 6: Leverage %dx validated", input.Leverage)
		}

		// Validate margin mode
		log.Printf("🔍 Step 6a: Validating margin mode...")
		if input.MarginMode == "" {
			input.MarginMode = "ISOLATED" // Default to ISOLATED
			log.Printf("⚠️  Step 6a: Margin mode not provided, defaulting to 'ISOLATED'")
		} else if input.MarginMode != "ISOLATED" && input.MarginMode != "CROSSED" {
			input.MarginMode = "ISOLATED"
			log.Printf("⚠️  Step 6a: Invalid margin mode '%s', defaulting to 'ISOLATED'", input.MarginMode)
		} else {
			log.Printf("✅ Step 6a: Margin mode '%s' validated", input.MarginMode)
		}

		// Validate callback rate
		log.Printf("🔍 Step 6b: Validating callback rate...")
		if input.CallbackRate < 0.1 || input.CallbackRate > 5 {
			input.CallbackRate = 1.0 // Default to 1%
			log.Printf("⚠️  Step 6b: Callback rate out of range (0.1-5), defaulting to 1.0%%")
		} else {
			log.Printf("✅ Step 6b: Callback rate %.2f%% validated", input.CallbackRate)
		}

		// Validate activation price
		log.Printf("🔍 Step 6c: Validating activation price...")
		if input.ActivationPrice < 0 {
			input.ActivationPrice = 0 // Default to 0 (activate immediately)
			log.Printf("⚠️  Step 6c: Negative activation price, defaulting to 0 (activate immediately)")
		} else {
			log.Printf("✅ Step 6c: Activation price %.8f validated", input.ActivationPrice)
		}

		// Encrypt API credentials if provided
		log.Printf("🔐 Step 7: Encrypting API credentials...")
		var encryptedAPIKey, encryptedAPISecret string
		var err error
		if input.APIKey != "" {
			encryptedAPIKey, err = utils.EncryptString(input.APIKey)
			if err != nil {
				log.Printf("❌ Step 7: Failed to encrypt API key: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API credentials"})
				return
			}
			log.Printf("✅ Step 7a: API Key encrypted successfully (length: %d)", len(encryptedAPIKey))
		} else {
			log.Printf("⚠️  Step 7a: No API Key provided")
		}
		if input.APISecret != "" {
			encryptedAPISecret, err = utils.EncryptString(input.APISecret)
			if err != nil {
				log.Printf("❌ Step 7b: Failed to encrypt API secret: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API credentials"})
				return
			}
			log.Printf("✅ Step 7b: API Secret encrypted successfully (length: %d)", len(encryptedAPISecret))
		} else {
			log.Printf("⚠️  Step 7b: No API Secret provided")
		}

		// Create trading config
		log.Printf("💾 Step 8: Creating bot config in database...")
		config := models.TradingConfig{
			Name:                input.Name,
			UserID:              user.ID,
			Symbol:              input.Symbol,
			Exchange:            input.Exchange,
			Amount:              input.Amount,
			TradingMode:         input.TradingMode,
			Leverage:            input.Leverage,
			MarginMode:          input.MarginMode,
			APIKey:              encryptedAPIKey,
			APISecret:           encryptedAPISecret,
			StopLossPercent:     input.StopLossPercent,
			TakeProfitPercent:   input.TakeProfitPercent,
			TrailingStopPercent: input.TrailingStopPercent,
			EnableTrailingStop:  input.EnableTrailingStop,
			ActivationPrice:     input.ActivationPrice,
			CallbackRate:        input.CallbackRate,
			IsActive:            true, // Active by default
		}

		if err := services.DB.Create(&config).Error; err != nil {
			log.Printf("❌ Step 8: Database creation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
			return
		}

		log.Printf("✅ Step 8: Bot config created successfully in database")
		log.Printf("   📋 Created config details:")
		log.Printf("      - Config ID: %d", config.ID)
		log.Printf("      - User ID: %d", config.UserID)
		log.Printf("      - Symbol: %s", config.Symbol)
		log.Printf("      - Exchange: %s", config.Exchange)
		log.Printf("      - Trading Mode: %s", config.TradingMode)
		log.Printf("      - Leverage: %dx", config.Leverage)
		log.Printf("      - Margin Mode: %s", config.MarginMode)
		log.Printf("      - Amount: %.8f", config.Amount)
		log.Printf("      - Stop Loss: %.2f%%", config.StopLossPercent)
		log.Printf("      - Take Profit: %.2f%%", config.TakeProfitPercent)
		log.Printf("      - Trailing Stop: %.2f%%", config.TrailingStopPercent)
		log.Printf("      - Enable Trailing Stop: %t", config.EnableTrailingStop)
		log.Printf("      - Activation Price: %.8f", config.ActivationPrice)
		log.Printf("      - Callback Rate: %.2f%%", config.CallbackRate)
		log.Printf("      - Is Active: %t", config.IsActive)
		log.Printf("      - Created At: %s", config.CreatedAt)

		log.Printf("🎉 Step 9: Sending success response to client")
		log.Printf("🔷 ===== CREATE BOT CONFIG - SUCCESS =====\n")

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
			Symbol              *string  `json:"symbol"`
			Exchange            *string  `json:"exchange"`
			Amount              *float64 `json:"amount"`
			TradingMode         *string  `json:"trading_mode"`
			Leverage            *int     `json:"leverage"`
			MarginMode          *string  `json:"margin_mode"`
			APIKey              *string  `json:"api_key"`
			APISecret           *string  `json:"api_secret"`
			StopLossPercent     *float64 `json:"stop_loss_percent"`
			TakeProfitPercent   *float64 `json:"take_profit_percent"`
			TrailingStopPercent *float64 `json:"trailing_stop_percent"`
			EnableTrailingStop  *bool    `json:"enable_trailing_stop"`
			ActivationPrice     *float64 `json:"activation_price"`
			CallbackRate        *float64 `json:"callback_rate"`
			IsActive            *bool    `json:"is_active"`
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
		if input.MarginMode != nil {
			if *input.MarginMode != "ISOLATED" && *input.MarginMode != "CROSSED" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Margin mode must be 'ISOLATED' or 'CROSSED'"})
				return
			}
			config.MarginMode = *input.MarginMode
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
		if input.TrailingStopPercent != nil {
			if *input.TrailingStopPercent < 0 || *input.TrailingStopPercent > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Trailing stop must be between 0 and 100"})
				return
			}
			config.TrailingStopPercent = *input.TrailingStopPercent
		}
		if input.EnableTrailingStop != nil {
			config.EnableTrailingStop = *input.EnableTrailingStop
		}
		if input.ActivationPrice != nil {
			if *input.ActivationPrice < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Activation price must be greater than or equal to 0"})
				return
			}
			config.ActivationPrice = *input.ActivationPrice
		}
		if input.CallbackRate != nil {
			if *input.CallbackRate < 0.1 || *input.CallbackRate > 5 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Callback rate must be between 0.1 and 5"})
				return
			}
			config.CallbackRate = *input.CallbackRate
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
