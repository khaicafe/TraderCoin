package controllers

import (
	"net/http"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetProfile - Lấy thông tin profile user
func GetProfile(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
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
			"id":         user.ID,
			"username":   user.Email, // Sử dụng email làm username
			"email":      user.Email,
			"full_name":  user.FullName,
			"phone":      user.Phone,
			"chat_id":    user.ChatID,
			"is_active":  user.Status == "active",
			"created_at": user.CreatedAt,
		})
	}
}

// UpdateProfile - Cập nhật thông tin profile user
func UpdateProfile(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var input struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			FullName string `json:"full_name"`
			Phone    string `json:"phone"`
			ChatID   string `json:"chat_id"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Prepare update map
		updates := make(map[string]interface{})

		if input.Email != "" {
			// Check if email is already taken by another user
			var count int64
			services.DB.Model(&models.User{}).
				Where("email = ? AND id != ?", input.Email, userID).
				Count(&count)
			if count > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
				return
			}
			updates["email"] = input.Email
		}

		if input.FullName != "" {
			updates["full_name"] = input.FullName
		}

		if input.Phone != "" {
			updates["phone"] = input.Phone
		}

		// Always update ChatID even if empty (user might want to clear it)
		updates["chat_id"] = input.ChatID

		// Update user
		err := services.DB.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(updates).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		// Get updated user data
		var user models.User
		if err := services.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Email, // Sử dụng email làm username
			"email":      user.Email,
			"full_name":  user.FullName,
			"phone":      user.Phone,
			"chat_id":    user.ChatID,
			"is_active":  user.Status == "active",
			"created_at": user.CreatedAt,
		})
	}
}

// ChangePassword - Đổi mật khẩu user
func ChangePassword(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var input struct {
			CurrentPassword string `json:"current_password" binding:"required"`
			NewPassword     string `json:"new_password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		// Get user từ database
		var user models.User
		if err := services.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Verify current password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}

		// Update password trong database
		if err := services.DB.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}
