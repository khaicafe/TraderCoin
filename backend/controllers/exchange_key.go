package controllers

import (
	"net/http"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// GetExchangeKeys - Lấy danh sách API keys của các sàn
func GetExchangeKeys(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		var keys []models.ExchangeKey
		if err := services.DB.Where("user_id = ?", userID).Find(&keys).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		result := make([]map[string]interface{}, 0, len(keys))
		for _, key := range keys {
			// Mask API key for security
			maskedKey := key.APIKey
			if len(maskedKey) > 10 {
				maskedKey = maskedKey[:10] + "..."
			}

			result = append(result, map[string]interface{}{
				"id":         key.ID,
				"exchange":   key.Exchange,
				"api_key":    maskedKey,
				"is_active":  key.IsActive,
				"created_at": key.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

// AddExchangeKey - Thêm API key mới cho sàn
func AddExchangeKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		var input struct {
			Exchange  string `json:"exchange" binding:"required"`
			APIKey    string `json:"api_key" binding:"required"`
			APISecret string `json:"api_secret" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate exchange name
		if input.Exchange != "binance" && input.Exchange != "bittrex" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exchange name. Supported: binance, bittrex"})
			return
		}

		exchangeKey := models.ExchangeKey{
			UserID:    userID.(uint),
			Exchange:  input.Exchange,
			APIKey:    input.APIKey,
			APISecret: input.APISecret,
			IsActive:  true,
		}

		if err := services.DB.Create(&exchangeKey).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add exchange key"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":      exchangeKey.ID,
			"message": "Exchange key added successfully",
		})
	}
}

// UpdateExchangeKey - Cập nhật API key của sàn
func UpdateExchangeKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		keyID := c.Param("id")

		var input struct {
			APIKey    string `json:"api_key"`
			APISecret string `json:"api_secret"`
			IsActive  *bool  `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		var count int64
		services.DB.Model(&models.ExchangeKey{}).
			Where("id = ? AND user_id = ?", keyID, userID).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exchange key not found"})
			return
		}

		// Build update map
		updates := make(map[string]interface{})
		if input.APIKey != "" {
			updates["api_key"] = input.APIKey
		}
		if input.APISecret != "" {
			updates["api_secret"] = input.APISecret
		}
		if input.IsActive != nil {
			updates["is_active"] = *input.IsActive
		}

		if len(updates) > 0 {
			if err := services.DB.Model(&models.ExchangeKey{}).
				Where("id = ? AND user_id = ?", keyID, userID).
				Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exchange key"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Exchange key updated successfully"})
	}
}

// DeleteExchangeKey - Xóa API key của sàn
func DeleteExchangeKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		keyID := c.Param("id")

		result := services.DB.Where("id = ? AND user_id = ?", keyID, userID).
			Delete(&models.ExchangeKey{})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exchange key"})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exchange key not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Exchange key deleted successfully"})
	}
}
