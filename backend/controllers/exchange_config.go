package controllers

import (
	"net/http"
	"strconv"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// ListExchangeConfigs returns all exchange API configurations
func ListExchangeConfigs(c *gin.Context) {
	svc := c.MustGet("services").(*services.Services)
	var configs []models.ExchangeAPIConfig

	// Optional filter: only active exchanges
	activeOnly := c.Query("active_only")
	query := svc.DB

	if activeOnly == "true" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("display_name").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exchange configurations"})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// GetExchangeConfig returns a specific exchange configuration
func GetExchangeConfig(c *gin.Context) {
	svc := c.MustGet("services").(*services.Services)
	exchangeName := c.Param("exchange")

	var config models.ExchangeAPIConfig
	if err := svc.DB.Where("exchange = ?", exchangeName).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exchange configuration not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CreateExchangeConfig creates a new exchange configuration (Admin only)
func CreateExchangeConfig(c *gin.Context) {
	svc := c.MustGet("services").(*services.Services)
	var config models.ExchangeAPIConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if exchange already exists
	var existingConfig models.ExchangeAPIConfig
	if err := svc.DB.Where("exchange = ?", config.Exchange).First(&existingConfig).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Exchange configuration already exists"})
		return
	}

	if err := svc.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exchange configuration"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// UpdateExchangeConfig updates an existing exchange configuration (Admin only)
func UpdateExchangeConfig(c *gin.Context) {
	svc := c.MustGet("services").(*services.Services)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var config models.ExchangeAPIConfig
	if err := svc.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exchange configuration not found"})
		return
	}

	var updateData models.ExchangeAPIConfig
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	config.DisplayName = updateData.DisplayName
	config.SpotAPIURL = updateData.SpotAPIURL
	config.SpotAPITestnetURL = updateData.SpotAPITestnetURL
	config.SpotWSURL = updateData.SpotWSURL
	config.SpotWSTestnetURL = updateData.SpotWSTestnetURL
	config.FuturesAPIURL = updateData.FuturesAPIURL
	config.FuturesAPITestnetURL = updateData.FuturesAPITestnetURL
	config.FuturesWSURL = updateData.FuturesWSURL
	config.FuturesWSTestnetURL = updateData.FuturesWSTestnetURL
	config.IsActive = updateData.IsActive
	config.SupportSpot = updateData.SupportSpot
	config.SupportFutures = updateData.SupportFutures
	config.SupportMargin = updateData.SupportMargin
	config.DefaultLeverage = updateData.DefaultLeverage
	config.MaxLeverage = updateData.MaxLeverage
	config.MinOrderSize = updateData.MinOrderSize
	config.MakerFee = updateData.MakerFee
	config.TakerFee = updateData.TakerFee
	config.Notes = updateData.Notes

	if err := svc.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exchange configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteExchangeConfig deletes an exchange configuration (Admin only)
func DeleteExchangeConfig(c *gin.Context) {
	svc := c.MustGet("services").(*services.Services)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := svc.DB.Delete(&models.ExchangeAPIConfig{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exchange configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange configuration deleted successfully"})
}

// ToggleExchangeStatus enables/disables an exchange (Admin only)
func ToggleExchangeStatus(c *gin.Context) {
	svc := c.MustGet("services").(*services.Services)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var config models.ExchangeAPIConfig
	if err := svc.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exchange configuration not found"})
		return
	}

	// Toggle status
	config.IsActive = !config.IsActive

	if err := svc.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exchange status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Exchange status updated successfully",
		"is_active": config.IsActive,
	})
}

// GetSupportedExchanges returns a simple list of supported exchanges (for dropdowns)
func GetSupportedExchanges(c *gin.Context) {
	svc := c.MustGet("services").(*services.Services)
	var configs []models.ExchangeAPIConfig

	if err := svc.DB.Where("is_active = ?", true).
		Select("exchange, display_name, support_spot, support_futures, support_margin").
		Order("display_name").
		Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exchanges"})
		return
	}

	// Format response for frontend dropdowns
	type ExchangeOption struct {
		Value          string `json:"value"`
		Label          string `json:"label"`
		SupportSpot    bool   `json:"support_spot"`
		SupportFutures bool   `json:"support_futures"`
		SupportMargin  bool   `json:"support_margin"`
	}

	options := make([]ExchangeOption, len(configs))
	for i, config := range configs {
		options[i] = ExchangeOption{
			Value:          config.Exchange,
			Label:          config.DisplayName,
			SupportSpot:    config.SupportSpot,
			SupportFutures: config.SupportFutures,
			SupportMargin:  config.SupportMargin,
		}
	}

	c.JSON(http.StatusOK, options)
}
