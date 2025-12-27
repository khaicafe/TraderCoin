package controllers

import (
	"net/http"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"
	"tradercoin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminLogin - Đăng nhập admin
func AdminLogin(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var admin models.Admin
		err := services.DB.Where("email = ?", input.Email).First(&admin).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate JWT token with admin role
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"admin_id": admin.ID,
			"email":    admin.Email,
			"role":     admin.Role,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString(utils.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"admin": gin.H{
				"id":        admin.ID,
				"email":     admin.Email,
				"full_name": admin.FullName,
				"role":      admin.Role,
			},
		})
	}
}

// GetAllUsers - Lấy danh sách tất cả users (Admin only)
func GetAllUsers(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for filtering
		status := c.Query("status")
		search := c.Query("search")

		// Build query
		query := services.DB.Model(&models.User{})

		if status != "" {
			query = query.Where("status = ?", status)
		}
		if search != "" {
			searchPattern := "%" + search + "%"
			query = query.Where("email LIKE ? OR full_name LIKE ?", searchPattern, searchPattern)
		}

		var users []models.User
		if err := query.Order("created_at desc").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		result := make([]map[string]interface{}, 0, len(users))
		for _, user := range users {
			// Calculate subscription status
			isSubscriptionActive := false
			daysRemaining := 0
			if user.SubscriptionEnd != nil {
				if user.SubscriptionEnd.After(time.Now()) {
					isSubscriptionActive = true
					daysRemaining = int(time.Until(*user.SubscriptionEnd).Hours() / 24)
				}
			}

			result = append(result, map[string]interface{}{
				"id":                     user.ID,
				"email":                  user.Email,
				"full_name":              user.FullName,
				"phone":                  user.Phone,
				"status":                 user.Status,
				"subscription_end":       user.SubscriptionEnd,
				"is_subscription_active": isSubscriptionActive,
				"days_remaining":         daysRemaining,
				"created_at":             user.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"users": result,
			"total": len(result),
		})
	}
}

// UpdateUserStatus - Cập nhật trạng thái user (khóa/mở khóa)
func UpdateUserStatus(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var input struct {
			Status string `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate status
		if input.Status != "active" && input.Status != "suspended" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'active' or 'suspended'"})
			return
		}

		result := services.DB.Model(&models.User{}).
			Where("id = ?", userID).
			Update("status", input.Status)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
	}
}

// GetAllTransactions - Lấy danh sách tất cả transactions (Admin only)
func GetAllTransactions(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for filtering
		userID := c.Query("user_id")
		txType := c.Query("type")
		status := c.Query("status")

		// Build query with join
		query := services.DB.Model(&models.Transaction{}).
			Select("transactions.*, users.email as user_email").
			Joins("LEFT JOIN users ON transactions.user_id = users.id")

		if userID != "" {
			query = query.Where("transactions.user_id = ?", userID)
		}
		if txType != "" {
			query = query.Where("transactions.type = ?", txType)
		}
		if status != "" {
			query = query.Where("transactions.status = ?", status)
		}

		var transactions []struct {
			models.Transaction
			UserEmail string
		}
		if err := query.Order("transactions.created_at desc").Find(&transactions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		result := make([]map[string]interface{}, 0, len(transactions))
		for _, tx := range transactions {
			result = append(result, map[string]interface{}{
				"id":          tx.ID,
				"user_id":     tx.UserID,
				"user_email":  tx.UserEmail,
				"amount":      tx.Amount,
				"type":        tx.Type,
				"status":      tx.Status,
				"description": tx.Description,
				"created_at":  tx.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetStatistics - Lấy thống kê tổng quan (Admin only)
func GetStatistics(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get total users count
		var totalUsers int64
		services.DB.Model(&models.User{}).Count(&totalUsers)

		// Get active users count
		var activeUsers int64
		services.DB.Model(&models.User{}).Where("status = 'active'").Count(&activeUsers)

		// Get suspended users count
		var suspendedUsers int64
		services.DB.Model(&models.User{}).Where("status = 'suspended'").Count(&suspendedUsers)

		// Get total orders count
		var totalOrders int64
		services.DB.Model(&models.Order{}).Count(&totalOrders)

		// Get total transactions and sum
		var totalTransactions int64
		var totalRevenue float64
		services.DB.Model(&models.Transaction{}).
			Where("status = 'completed'").
			Count(&totalTransactions)
		services.DB.Model(&models.Transaction{}).
			Where("status = 'completed'").
			Select("COALESCE(SUM(amount), 0)").
			Scan(&totalRevenue)

		// Get total trading configs
		var totalConfigs int64
		services.DB.Model(&models.TradingConfig{}).Count(&totalConfigs)

		// Get active trading configs
		var activeConfigs int64
		services.DB.Model(&models.TradingConfig{}).Where("is_active = ?", true).Count(&activeConfigs)

		// Get exchange keys count
		var totalKeys int64
		services.DB.Model(&models.ExchangeKey{}).Count(&totalKeys)

		c.JSON(http.StatusOK, gin.H{
			"users": gin.H{
				"total":     totalUsers,
				"active":    activeUsers,
				"suspended": suspendedUsers,
			},
			"orders": gin.H{
				"total": totalOrders,
			},
			"transactions": gin.H{
				"total":   totalTransactions,
				"revenue": totalRevenue,
			},
			"trading_configs": gin.H{
				"total":  totalConfigs,
				"active": activeConfigs,
			},
			"exchange_keys": gin.H{
				"total": totalKeys,
			},
		})
	}
}

// ExtendUserSubscription - Gia hạn subscription cho user (Admin only)
func ExtendUserSubscription(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var input struct {
			Days int `json:"days" binding:"required,min=1"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Days must be a positive number"})
			return
		}

		// Get user
		var user models.User
		if err := services.DB.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Calculate new subscription end date
		var newEndDate time.Time
		if user.SubscriptionEnd != nil && user.SubscriptionEnd.After(time.Now()) {
			// Extend from current end date if still valid
			newEndDate = user.SubscriptionEnd.AddDate(0, 0, input.Days)
		} else {
			// Start from now if expired or no subscription
			newEndDate = time.Now().AddDate(0, 0, input.Days)
		}

		// Update user subscription
		user.SubscriptionEnd = &newEndDate
		if err := services.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extend subscription"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":          "Subscription extended successfully",
			"subscription_end": newEndDate,
			"extended_days":    input.Days,
		})
	}
}

// SuspendUser - Khóa user (Admin only)
func SuspendUser(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var input struct {
			Reason string `json:"reason"`
		}

		c.ShouldBindJSON(&input)

		// Get user
		var user models.User
		if err := services.DB.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Update status
		user.Status = "suspended"
		if err := services.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to suspend user"})
			return
		}

		// Log action
		utils.LogInfo("Admin suspended user: " + user.Email)
		if input.Reason != "" {
			utils.LogInfo("Reason: " + input.Reason)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User suspended successfully",
			"user": gin.H{
				"id":     user.ID,
				"email":  user.Email,
				"status": user.Status,
			},
		})
	}
}

// ActivateUser - Kích hoạt lại user (Admin only)
func ActivateUser(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		// Get user
		var user models.User
		if err := services.DB.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Update status
		user.Status = "active"
		if err := services.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate user"})
			return
		}

		// Log action
		utils.LogInfo("Admin activated user: " + user.Email)

		c.JSON(http.StatusOK, gin.H{
			"message": "User activated successfully",
			"user": gin.H{
				"id":     user.ID,
				"email":  user.Email,
				"status": user.Status,
			},
		})
	}
}
