package controllers

import (
	"net/http"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// GetSystemStatus - Lấy trạng thái hệ thống
func GetSystemStatus(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connection
		sqlDB, err := services.DB.DB()
		dbStatus := "healthy"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "unhealthy"
		}

		// Check Redis connection
		redisStatus := "healthy"
		if services.Redis != nil {
			if err := services.Redis.Ping(c).Err(); err != nil {
				redisStatus = "unhealthy"
			}
		} else {
			redisStatus = "disabled"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "running",
			"timestamp": time.Now(),
			"services": gin.H{
				"database": dbStatus,
				"redis":    redisStatus,
			},
			"uptime": "5d 12h 30m", // TODO: Calculate actual uptime
		})
	}
}

// GetTradingMetrics - Lấy metrics trading
func GetTradingMetrics(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		// Lấy thống kê orders
		var totalOrders, successOrders, failedOrders int64
		services.DB.Model(&models.Order{}).Where("user_id = ?", userID).Count(&totalOrders)
		services.DB.Model(&models.Order{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&successOrders)
		services.DB.Model(&models.Order{}).Where("user_id = ? AND status = ?", userID, "failed").Count(&failedOrders)

		// Tính win rate
		var winRate float64
		if totalOrders > 0 {
			winRate = float64(successOrders) / float64(totalOrders) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"metrics": gin.H{
				"total_orders":   totalOrders,
				"success_orders": successOrders,
				"failed_orders":  failedOrders,
				"win_rate":       winRate,
			},
			"timestamp": time.Now(),
		})
	}
}

// GetActivePositions - Lấy các vị thế đang mở
func GetActivePositions(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		// TODO: Lấy active positions từ exchange API
		positions := []gin.H{
			{
				"symbol":        "BTCUSDT",
				"side":          "LONG",
				"size":          0.5,
				"entry_price":   42000.00,
				"current_price": 43500.00,
				"pnl":           750.00,
				"pnl_percent":   3.57,
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"positions": positions,
			"total":     len(positions),
		})
	}
}

// GetPerformanceStats - Lấy thống kê hiệu suất
func GetPerformanceStats(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		// Get date range from query
		period := c.DefaultQuery("period", "7d") // 7d, 30d, 90d, 1y

		// TODO: Calculate performance stats
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"period":  period,
			"stats": gin.H{
				"total_trades":         150,
				"winning_trades":       95,
				"losing_trades":        55,
				"win_rate":             63.33,
				"total_profit":         12500.00,
				"total_loss":           -3200.00,
				"net_profit":           9300.00,
				"avg_profit_per_trade": 62.00,
				"largest_win":          850.00,
				"largest_loss":         -420.00,
				"profit_factor":        3.91,
				"sharpe_ratio":         1.85,
			},
			"timestamp": time.Now(),
		})
	}
}

// GetBotStatus - Lấy trạng thái bot trading
func GetBotStatus(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		// Lấy trading configs
		var activeBots int64
		services.DB.Model(&models.TradingConfig{}).
			Where("user_id = ? AND is_active = ?", userID, true).
			Count(&activeBots)

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"bots": gin.H{
				"active":     activeBots,
				"status":     "running",
				"last_check": time.Now(),
			},
		})
	}
}

// GetAlerts - Lấy danh sách alerts
func GetAlerts(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		// TODO: Lấy alerts từ database
		alerts := []gin.H{
			{
				"id":        1,
				"type":      "price_alert",
				"message":   "BTC reached $45,000",
				"severity":  "info",
				"timestamp": time.Now().Add(-10 * time.Minute),
				"read":      false,
			},
			{
				"id":        2,
				"type":      "order_filled",
				"message":   "Order #12345 filled successfully",
				"severity":  "success",
				"timestamp": time.Now().Add(-25 * time.Minute),
				"read":      false,
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":      userID,
			"alerts":       alerts,
			"unread_count": 2,
		})
	}
}

// MarkAlertRead - Đánh dấu alert đã đọc
func MarkAlertRead(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		alertID := c.Param("id")

		// TODO: Update alert status in database
		c.JSON(http.StatusOK, gin.H{
			"message":  "Alert marked as read",
			"alert_id": alertID,
		})
	}
}
