package controllers

import (
	"net/http"
	"strconv"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// GetSystemLogs returns system logs for the authenticated user
func GetSystemLogs(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Query parameters
		level := c.Query("level")   // SUCCESS, INFO, WARNING, ERROR
		symbol := c.Query("symbol") // Filter by symbol
		action := c.Query("action") // Filter by action
		limitStr := c.DefaultQuery("limit", "50")
		offsetStr := c.DefaultQuery("offset", "0")
		hoursStr := c.Query("hours") // Filter by hours ago (e.g., 24 for last 24 hours)

		limit, _ := strconv.Atoi(limitStr)
		if limit < 1 || limit > 500 {
			limit = 50
		}

		offset, _ := strconv.Atoi(offsetStr)
		if offset < 0 {
			offset = 0
		}

		// Build query
		query := services.DB.Where("user_id = ?", userID).Order("created_at DESC")

		// Apply filters
		if level != "" {
			query = query.Where("level = ?", level)
		}
		if symbol != "" {
			query = query.Where("symbol = ?", symbol)
		}
		if action != "" {
			query = query.Where("action LIKE ?", "%"+action+"%")
		}
		if hoursStr != "" {
			if hours, err := strconv.ParseFloat(hoursStr, 64); err == nil && hours > 0 {
				cutoff := time.Now().Add(-time.Duration(hours * float64(time.Hour)))
				query = query.Where("created_at >= ?", cutoff)
			}
		}

		// Get total count
		var total int64
		query.Model(&models.SystemLog{}).Count(&total)

		// Get logs
		var logs []models.SystemLog
		if err := query.Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"logs":   logs,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
	}
}

// GetSystemLogStats returns statistics about system logs
func GetSystemLogStats(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		hoursStr := c.DefaultQuery("hours", "24")
		hours, _ := strconv.ParseFloat(hoursStr, 64)
		cutoff := time.Now().Add(-time.Duration(hours * float64(time.Hour)))

		// Count by level
		var stats struct {
			Success int64 `json:"success"`
			Info    int64 `json:"info"`
			Warning int64 `json:"warning"`
			Error   int64 `json:"error"`
			Total   int64 `json:"total"`
		}

		services.DB.Model(&models.SystemLog{}).
			Where("user_id = ? AND created_at >= ? AND level = ?", userID, cutoff, "SUCCESS").
			Count(&stats.Success)

		services.DB.Model(&models.SystemLog{}).
			Where("user_id = ? AND created_at >= ? AND level = ?", userID, cutoff, "INFO").
			Count(&stats.Info)

		services.DB.Model(&models.SystemLog{}).
			Where("user_id = ? AND created_at >= ? AND level = ?", userID, cutoff, "WARNING").
			Count(&stats.Warning)

		services.DB.Model(&models.SystemLog{}).
			Where("user_id = ? AND created_at >= ? AND level = ?", userID, cutoff, "ERROR").
			Count(&stats.Error)

		stats.Total = stats.Success + stats.Info + stats.Warning + stats.Error

		c.JSON(http.StatusOK, stats)
	}
}

// ClearSystemLogs clears old system logs (optional, for maintenance)
func ClearSystemLogs(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		daysStr := c.DefaultQuery("days", "30")
		days, _ := strconv.Atoi(daysStr)
		if days < 1 {
			days = 30
		}

		cutoff := time.Now().AddDate(0, 0, -days)

		result := services.DB.Where("user_id = ? AND created_at < ?", userID, cutoff).
			Delete(&models.SystemLog{})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Logs cleared successfully",
			"deleted": result.RowsAffected,
		})
	}
}
