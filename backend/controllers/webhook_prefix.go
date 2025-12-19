package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

// GetWebhookPrefix returns the latest active webhook prefix for the authenticated user
func GetWebhookPrefix(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var wp models.WebhookPrefix
		if err := services.DB.Where("user_id = ? AND active = ?", userID, true).
			Order("created_at DESC").First(&wp).Error; err != nil {
			// Not found is OK; return empty
			c.JSON(http.StatusOK, gin.H{"prefix": "", "url": ""})
			return
		}

		base := c.Request.Host
		scheme := "http"
		if strings.HasPrefix(base, "localhost:") || strings.HasPrefix(base, "127.0.0.1:") {
			scheme = "http"
		} else if c.Request.TLS != nil {
			scheme = "https"
		}
		fullURL := fmt.Sprintf("%s://%s/api/v1/signals/webhook/tradingview/%s", scheme, base, wp.Prefix)

		c.JSON(http.StatusOK, gin.H{"prefix": wp.Prefix, "url": fullURL})
	}
}

// CreateWebhookPrefix creates a unique webhook prefix for the authenticated user
func CreateWebhookPrefix(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(uint)

		var input struct {
			Prefix string `json:"prefix"`
		}
		_ = c.ShouldBindJSON(&input)

		// Generate if not provided
		prefix := strings.TrimSpace(input.Prefix)
		if prefix == "" {
			b := make([]byte, 6)
			if _, err := rand.Read(b); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate prefix"})
				return
			}
			prefix = hex.EncodeToString(b)
		}

		// Normalize prefix
		prefix = strings.ToLower(prefix)
		if len(prefix) > 64 {
			prefix = prefix[:64]
		}

		// Ensure uniqueness
		var count int64
		services.DB.Model(&models.WebhookPrefix{}).Where("prefix = ?", prefix).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "prefix already exists"})
			return
		}

		wp := models.WebhookPrefix{
			UserID:    userID,
			Prefix:    prefix,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := services.DB.Create(&wp).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create prefix"})
			return
		}

		base := c.Request.Host
		scheme := "http"
		if strings.HasPrefix(base, "localhost:") || strings.HasPrefix(base, "127.0.0.1:") {
			scheme = "http"
		} else if c.Request.TLS != nil {
			scheme = "https"
		}
		fullURL := fmt.Sprintf("%s://%s/api/v1/signals/webhook/tradingview/%s", scheme, base, prefix)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Webhook prefix created",
			"prefix":  prefix,
			"url":     fullURL,
		})
	}
}
