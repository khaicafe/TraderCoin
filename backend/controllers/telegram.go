package controllers

import (
	"net/http"
	"strconv"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TelegramController struct {
	db              *gorm.DB
	telegramService *services.TelegramService
}

func NewTelegramController(db *gorm.DB) *TelegramController {
	return &TelegramController{
		db:              db,
		telegramService: services.NewTelegramService(db),
	}
}

// GetTelegramConfig returns the user's Telegram configuration
func (tc *TelegramController) GetTelegramConfig(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var config models.TelegramConfig
	// Chỉ có duy nhất 1 config, lấy row đầu tiên
	if err := tc.db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Telegram configuration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CreateOrUpdateTelegramConfig creates or updates the user's Telegram configuration
func (tc *TelegramController) CreateOrUpdateTelegramConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		BotToken  string `json:"bot_token" binding:"required"`
		ChatID    string `json:"chat_id" binding:"required"`
		BotName   string `json:"bot_name"`
		IsEnabled bool   `json:"is_enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if config exists (chỉ có 1 row duy nhất)
	var config models.TelegramConfig
	err := tc.db.First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// Create new config
		config = models.TelegramConfig{
			UserID:    userID.(uint),
			BotToken:  input.BotToken,
			ChatID:    input.ChatID,
			BotName:   input.BotName,
			IsEnabled: input.IsEnabled,
		}

		if err := tc.db.Create(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Telegram configuration created successfully",
			"config":  config,
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configuration"})
		return
	}

	// Update existing config
	config.BotToken = input.BotToken
	config.ChatID = input.ChatID
	config.BotName = input.BotName
	config.IsEnabled = input.IsEnabled

	if err := tc.db.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Telegram configuration updated successfully",
		"config":  config,
	})
}

// DeleteTelegramConfig deletes the user's Telegram configuration
func (tc *TelegramController) DeleteTelegramConfig(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Xóa row duy nhất
	if err := tc.db.Delete(&models.TelegramConfig{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Telegram configuration deleted successfully"})
}

// TestTelegramConnection tests the Telegram bot connection
func (tc *TelegramController) TestTelegramConnection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		BotToken string `json:"bot_token" binding:"required"`
		ChatID   string `json:"chat_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Test the connection
	err := tc.telegramService.TestConnection(input.BotToken, input.ChatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to connect to Telegram",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Telegram connection test successful",
		"user_id": userID,
	})
}

// SendTestMessage sends a test message to the user's Telegram
func (tc *TelegramController) SendTestMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	testMessage := "🚀 <b>Test Message</b>\n\nThis is a test notification from TraderCoin Bot!"

	err := tc.telegramService.SendMessageToUser(userID.(uint), testMessage, "HTML")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to send test message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test message sent successfully"})
}

// ToggleTelegramNotifications enables or disables Telegram notifications
func (tc *TelegramController) ToggleTelegramNotifications(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		IsEnabled bool `json:"is_enabled" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lấy config duy nhất
	var config models.TelegramConfig
	if err := tc.db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Telegram configuration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configuration"})
		return
	}

	config.IsEnabled = input.IsEnabled
	if err := tc.db.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	status := "disabled"
	if input.IsEnabled {
		status = "enabled"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Telegram notifications " + status + " successfully",
		"config":  config,
	})
}

// Admin: Get all Telegram configurations
func (tc *TelegramController) GetAllTelegramConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var configs []models.TelegramConfig
	var total int64

	// Get total count
	tc.db.Model(&models.TelegramConfig{}).Count(&total)

	// Get configs with pagination
	if err := tc.db.Preload("User").Offset(offset).Limit(limit).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  configs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Admin: Create Telegram configuration for a user
func (tc *TelegramController) AdminCreateTelegramConfig(c *gin.Context) {
	var input struct {
		UserID    uint   `json:"user_id" binding:"required"`
		BotToken  string `json:"bot_token" binding:"required"`
		ChatID    string `json:"chat_id" binding:"required"`
		BotName   string `json:"bot_name"`
		IsEnabled bool   `json:"is_enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if config already exists (chỉ có 1 row duy nhất)
	var existingConfig models.TelegramConfig
	err := tc.db.First(&existingConfig).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telegram configuration already exists"})
		return
	}

	// Create new config
	config := models.TelegramConfig{
		UserID:    input.UserID,
		BotToken:  input.BotToken,
		ChatID:    input.ChatID,
		BotName:   input.BotName,
		IsEnabled: input.IsEnabled,
	}

	if err := tc.db.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
		return
	}

	// Preload User data before returning
	tc.db.Preload("User").First(&config, config.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Telegram configuration created successfully",
		"config":  config,
	})
}

// Admin: Update Telegram configuration
func (tc *TelegramController) AdminUpdateTelegramConfig(c *gin.Context) {
	configID := c.Param("id")

	var input struct {
		BotToken  string `json:"bot_token" binding:"required"`
		ChatID    string `json:"chat_id" binding:"required"`
		BotName   string `json:"bot_name"`
		IsEnabled bool   `json:"is_enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing config
	var config models.TelegramConfig
	if err := tc.db.First(&config, configID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Telegram configuration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configuration"})
		return
	}

	// Update config
	config.BotToken = input.BotToken
	config.ChatID = input.ChatID
	config.BotName = input.BotName
	config.IsEnabled = input.IsEnabled

	if err := tc.db.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	// Preload User data before returning
	tc.db.Preload("User").First(&config, config.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Telegram configuration updated successfully",
		"config":  config,
	})
}

// Admin: Test Telegram connection (no user_id required)
func (tc *TelegramController) AdminTestTelegramConnection(c *gin.Context) {
	var input struct {
		BotToken string `json:"bot_token" binding:"required"`
		ChatID   string `json:"chat_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Test the connection
	err := tc.telegramService.TestConnection(input.BotToken, input.ChatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to connect to Telegram",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Telegram connection test successful",
	})
}

// Admin: Start callback listener for Telegram buttons
func (tc *TelegramController) StartCallbackListener(c *gin.Context) {
	var input struct {
		BotToken string `json:"bot_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start callback listener in background
	go func() {
		if err := tc.telegramService.HandleUpdates(input.BotToken); err != nil {
			// Log error but don't stop the goroutine
			// You can add logging here
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Callback listener started successfully",
		"info":    "Bot is now listening for button clicks in the background",
	})
}
