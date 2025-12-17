package controllers

import (
	"net/http"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// GetTradingConfigs - Lấy danh sách cấu hình trading (stop-loss, take-profit)
func GetTradingConfigs(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		var configs []models.TradingConfig
		if err := services.DB.Where("user_id = ?", userID).Find(&configs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		result := make([]map[string]interface{}, 0, len(configs))
		for _, config := range configs {
			result = append(result, map[string]interface{}{
				"id":                  config.ID,
				"exchange":            config.Exchange,
				"symbol":              config.Symbol,
				"stop_loss_percent":   config.StopLossPercent,
				"take_profit_percent": config.TakeProfitPercent,
				"is_active":           config.IsActive,
				"created_at":          config.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

// CreateTradingConfig - Tạo cấu hình trading mới
func CreateTradingConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		var input struct {
			Exchange          string  `json:"exchange" binding:"required"`
			Symbol            string  `json:"symbol" binding:"required"`
			StopLossPercent   float64 `json:"stop_loss_percent" binding:"required"`
			TakeProfitPercent float64 `json:"take_profit_percent" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate percentages
		if input.StopLossPercent <= 0 || input.StopLossPercent > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Stop loss percent must be between 0 and 100"})
			return
		}
		if input.TakeProfitPercent <= 0 || input.TakeProfitPercent > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Take profit percent must be between 0 and 1000"})
			return
		}

		config := models.TradingConfig{
			UserID:            userID.(uint),
			Exchange:          input.Exchange,
			Symbol:            input.Symbol,
			StopLossPercent:   input.StopLossPercent,
			TakeProfitPercent: input.TakeProfitPercent,
			IsActive:          true,
		}

		if err := services.DB.Create(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trading config"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":      config.ID,
			"message": "Trading config created successfully",
		})
	}
}

// UpdateTradingConfig - Cập nhật cấu hình trading
func UpdateTradingConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		configID := c.Param("id")

		var input struct {
			StopLossPercent   *float64 `json:"stop_loss_percent"`
			TakeProfitPercent *float64 `json:"take_profit_percent"`
			IsActive          *bool    `json:"is_active"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		var count int64
		services.DB.Model(&models.TradingConfig{}).
			Where("id = ? AND user_id = ?", configID, userID).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trading config not found"})
			return
		}

		// Build update map
		updates := make(map[string]interface{})
		if input.StopLossPercent != nil {
			if *input.StopLossPercent <= 0 || *input.StopLossPercent > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Stop loss percent must be between 0 and 100"})
				return
			}
			updates["stop_loss_percent"] = *input.StopLossPercent
		}
		if input.TakeProfitPercent != nil {
			if *input.TakeProfitPercent <= 0 || *input.TakeProfitPercent > 1000 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Take profit percent must be between 0 and 1000"})
				return
			}
			updates["take_profit_percent"] = *input.TakeProfitPercent
		}
		if input.IsActive != nil {
			updates["is_active"] = *input.IsActive
		}

		if len(updates) > 0 {
			if err := services.DB.Model(&models.TradingConfig{}).
				Where("id = ? AND user_id = ?", configID, userID).
				Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trading config"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Trading config updated successfully"})
	}
}

// DeleteTradingConfig - Xóa cấu hình trading
func DeleteTradingConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		configID := c.Param("id")

		result := services.DB.Where("id = ? AND user_id = ?", configID, userID).
			Delete(&models.TradingConfig{})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trading config"})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trading config not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Trading config deleted successfully"})
	}
}
