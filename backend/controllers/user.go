package controllers

import (
	"net/http"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetProfile - Lấy thông tin profile user
func GetProfile(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// For now, without middleware, get from query param or default to 1
			userID = uint(1)
		}

		var user models.User
		err := services.DB.First(&user, userID).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":               user.ID,
			"email":            user.Email,
			"full_name":        user.FullName,
			"phone":            user.Phone,
			"status":           user.Status,
			"subscription_end": user.SubscriptionEnd,
			"created_at":       user.CreatedAt,
		})
	}
}

// UpdateProfile - Cập nhật thông tin profile user
func UpdateProfile(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = uint(1)
		}

		var input struct {
			FullName string `json:"full_name"`
			Phone    string `json:"phone"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := services.DB.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"full_name": input.FullName,
				"phone":     input.Phone,
			}).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}
