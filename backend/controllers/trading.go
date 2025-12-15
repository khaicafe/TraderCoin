package controllers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"
	"tradercoin/backend/models"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-super-secret-jwt-key-change-this-in-production")

// Auth handlers
func Register(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
			FullName string `json:"full_name" binding:"required"`
			Phone    string `json:"phone"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if user exists
		var exists int
		err := services.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", input.Email).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if exists > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Create user
		result, err := services.DB.Exec(`
			INSERT INTO users (email, password_hash, full_name, phone, status, subscription_end)
			VALUES (?, ?, ?, ?, 'active', datetime('now', '+30 days'))
		`, input.Email, string(hashedPassword), input.FullName, input.Phone)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		userID, _ := result.LastInsertId()

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"email":   input.Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"token":   tokenString,
			"user": gin.H{
				"id":        userID,
				"email":     input.Email,
				"full_name": input.FullName,
			},
		})
	}
}

func Login(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Get user from database
		var user models.User
		err := services.DB.QueryRow(`
			SELECT id, email, password_hash, full_name, phone, status, subscription_end
			FROM users WHERE email = ?
		`, input.Email).Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
			&user.Phone, &user.Status, &user.SubscriptionEnd,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Check if user is suspended
		if user.Status == "suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
			return
		}

		// Verify password
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   tokenString,
			"user": gin.H{
				"id":               user.ID,
				"email":            user.Email,
				"full_name":        user.FullName,
				"phone":            user.Phone,
				"status":           user.Status,
				"subscription_end": user.SubscriptionEnd,
			},
		})
	}
}

func RefreshToken(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			return
		}

		// Remove "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Generate new token
		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": claims["user_id"],
			"email":   claims["email"],
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})

		newTokenString, err := newToken.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": newTokenString,
		})
	}
}

// User handlers
func GetProfile(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// For now, without middleware, get from query param or default to 1
			userID = 1
		}

		var user models.User
		err := services.DB.QueryRow(`
			SELECT id, email, full_name, phone, status, subscription_end, created_at
			FROM users WHERE id = ?
		`, userID).Scan(
			&user.ID, &user.Email, &user.FullName, &user.Phone,
			&user.Status, &user.SubscriptionEnd, &user.CreatedAt,
		)

		if err == sql.ErrNoRows {
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

func UpdateProfile(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
		}

		var input struct {
			FullName string `json:"full_name"`
			Phone    string `json:"phone"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := services.DB.Exec(`
			UPDATE users SET full_name = ?, phone = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, input.FullName, input.Phone, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}

// Exchange keys handlers
func GetExchangeKeys(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
		}

		rows, err := services.DB.Query(`
			SELECT id, exchange, api_key, api_secret, is_active, created_at
			FROM exchange_keys WHERE user_id = ?
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var keys []map[string]interface{}
		for rows.Next() {
			var key models.ExchangeKey
			err := rows.Scan(
				&key.ID, &key.Exchange, &key.APIKey,
				&key.APISecret, &key.IsActive, &key.CreatedAt,
			)
			if err != nil {
				continue
			}

			// Mask API key for security
			maskedKey := key.APIKey
			if len(maskedKey) > 10 {
				maskedKey = maskedKey[:10] + "..."
			}

			keys = append(keys, map[string]interface{}{
				"id":         key.ID,
				"exchange":   key.Exchange,
				"api_key":    maskedKey,
				"is_active":  key.IsActive,
				"created_at": key.CreatedAt,
			})
		}

		if keys == nil {
			keys = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, keys)
	}
}

func AddExchangeKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
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

		result, err := services.DB.Exec(`
			INSERT INTO exchange_keys (user_id, exchange, api_key, api_secret, is_active)
			VALUES (?, ?, ?, ?, true)
		`, userID, input.Exchange, input.APIKey, input.APISecret)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add exchange key"})
			return
		}

		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{
			"id":      id,
			"message": "Exchange key added successfully",
		})
	}
}

func UpdateExchangeKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
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
		var count int
		err := services.DB.QueryRow(`
			SELECT COUNT(*) FROM exchange_keys WHERE id = ? AND user_id = ?
		`, keyID, userID).Scan(&count)

		if err != nil || count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exchange key not found"})
			return
		}

		// Build update query
		query := "UPDATE exchange_keys SET updated_at = CURRENT_TIMESTAMP"
		args := []interface{}{}

		if input.APIKey != "" {
			query += ", api_key = ?"
			args = append(args, input.APIKey)
		}
		if input.APISecret != "" {
			query += ", api_secret = ?"
			args = append(args, input.APISecret)
		}
		if input.IsActive != nil {
			query += ", is_active = ?"
			args = append(args, *input.IsActive)
		}

		query += " WHERE id = ? AND user_id = ?"
		args = append(args, keyID, userID)

		_, err = services.DB.Exec(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exchange key"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Exchange key updated successfully"})
	}
}

func DeleteExchangeKey(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
		}

		keyID := c.Param("id")

		result, err := services.DB.Exec(`
			DELETE FROM exchange_keys WHERE id = ? AND user_id = ?
		`, keyID, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exchange key"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exchange key not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Exchange key deleted successfully"})
	}
}

// Trading config handlers
func GetTradingConfigs(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
		}

		rows, err := services.DB.Query(`
			SELECT id, exchange, symbol, stop_loss_percent, take_profit_percent, is_active, created_at
			FROM trading_configs WHERE user_id = ?
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var configs []map[string]interface{}
		for rows.Next() {
			var config models.TradingConfig
			err := rows.Scan(
				&config.ID, &config.Exchange, &config.Symbol,
				&config.StopLossPercent, &config.TakeProfitPercent,
				&config.IsActive, &config.CreatedAt,
			)
			if err != nil {
				continue
			}

			configs = append(configs, map[string]interface{}{
				"id":                  config.ID,
				"exchange":            config.Exchange,
				"symbol":              config.Symbol,
				"stop_loss_percent":   config.StopLossPercent,
				"take_profit_percent": config.TakeProfitPercent,
				"is_active":           config.IsActive,
				"created_at":          config.CreatedAt,
			})
		}

		if configs == nil {
			configs = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, configs)
	}
}

func CreateTradingConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
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

		result, err := services.DB.Exec(`
			INSERT INTO trading_configs (user_id, exchange, symbol, stop_loss_percent, take_profit_percent, is_active)
			VALUES (?, ?, ?, ?, ?, true)
		`, userID, input.Exchange, input.Symbol, input.StopLossPercent, input.TakeProfitPercent)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trading config"})
			return
		}

		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{
			"id":      id,
			"message": "Trading config created successfully",
		})
	}
}

func UpdateTradingConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
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
		var count int
		err := services.DB.QueryRow(`
			SELECT COUNT(*) FROM trading_configs WHERE id = ? AND user_id = ?
		`, configID, userID).Scan(&count)

		if err != nil || count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trading config not found"})
			return
		}

		// Build update query
		query := "UPDATE trading_configs SET updated_at = CURRENT_TIMESTAMP"
		args := []interface{}{}

		if input.StopLossPercent != nil {
			if *input.StopLossPercent <= 0 || *input.StopLossPercent > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Stop loss percent must be between 0 and 100"})
				return
			}
			query += ", stop_loss_percent = ?"
			args = append(args, *input.StopLossPercent)
		}
		if input.TakeProfitPercent != nil {
			if *input.TakeProfitPercent <= 0 || *input.TakeProfitPercent > 1000 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Take profit percent must be between 0 and 1000"})
				return
			}
			query += ", take_profit_percent = ?"
			args = append(args, *input.TakeProfitPercent)
		}
		if input.IsActive != nil {
			query += ", is_active = ?"
			args = append(args, *input.IsActive)
		}

		query += " WHERE id = ? AND user_id = ?"
		args = append(args, configID, userID)

		_, err = services.DB.Exec(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trading config"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Trading config updated successfully"})
	}
}

func DeleteTradingConfig(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
		}

		configID := c.Param("id")

		result, err := services.DB.Exec(`
			DELETE FROM trading_configs WHERE id = ? AND user_id = ?
		`, configID, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trading config"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trading config not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Trading config deleted successfully"})
	}
}

// Orders handlers
func GetOrders(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
		}

		// Get query parameters for filtering
		exchange := c.Query("exchange")
		symbol := c.Query("symbol")
		status := c.Query("status")

		// Build query
		query := `SELECT id, exchange, symbol, order_id, side, type, quantity, price, status, created_at
				  FROM orders WHERE user_id = ?`
		args := []interface{}{userID}

		if exchange != "" {
			query += " AND exchange = ?"
			args = append(args, exchange)
		}
		if symbol != "" {
			query += " AND symbol = ?"
			args = append(args, symbol)
		}
		if status != "" {
			query += " AND status = ?"
			args = append(args, status)
		}

		query += " ORDER BY created_at DESC"

		rows, err := services.DB.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var orders []map[string]interface{}
		for rows.Next() {
			var order models.Order
			err := rows.Scan(
				&order.ID, &order.Exchange, &order.Symbol, &order.OrderID,
				&order.Side, &order.Type, &order.Quantity, &order.Price,
				&order.Status, &order.CreatedAt,
			)
			if err != nil {
				continue
			}

			orders = append(orders, map[string]interface{}{
				"id":         order.ID,
				"exchange":   order.Exchange,
				"symbol":     order.Symbol,
				"order_id":   order.OrderID,
				"side":       order.Side,
				"type":       order.Type,
				"quantity":   order.Quantity,
				"price":      order.Price,
				"status":     order.Status,
				"created_at": order.CreatedAt,
			})
		}

		if orders == nil {
			orders = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, orders)
	}
}

func GetOrder(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = 1
		}

		orderID := c.Param("id")

		var order models.Order
		err := services.DB.QueryRow(`
			SELECT id, exchange, symbol, order_id, side, type, quantity, price, status, created_at, updated_at
			FROM orders WHERE id = ? AND user_id = ?
		`, orderID, userID).Scan(
			&order.ID, &order.Exchange, &order.Symbol, &order.OrderID,
			&order.Side, &order.Type, &order.Quantity, &order.Price,
			&order.Status, &order.CreatedAt, &order.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         order.ID,
			"exchange":   order.Exchange,
			"symbol":     order.Symbol,
			"order_id":   order.OrderID,
			"side":       order.Side,
			"type":       order.Type,
			"quantity":   order.Quantity,
			"price":      order.Price,
			"status":     order.Status,
			"created_at": order.CreatedAt,
			"updated_at": order.UpdatedAt,
		})
	}
}

// Admin handlers
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
		err := services.DB.QueryRow(`
			SELECT id, email, password_hash, full_name, role, created_at
			FROM admins WHERE email = ?
		`, input.Email).Scan(
			&admin.ID, &admin.Email, &admin.PasswordHash,
			&admin.FullName, &admin.Role, &admin.CreatedAt,
		)

		if err == sql.ErrNoRows {
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

		tokenString, err := token.SignedString([]byte(jwtSecret))
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

func GetAllUsers(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for filtering
		status := c.Query("status")
		search := c.Query("search")

		// Build query
		query := `SELECT id, email, full_name, phone, status, subscription_end, created_at
				  FROM users WHERE 1=1`
		args := []interface{}{}

		if status != "" {
			query += " AND status = ?"
			args = append(args, status)
		}
		if search != "" {
			query += " AND (email LIKE ? OR full_name LIKE ?)"
			searchPattern := "%" + search + "%"
			args = append(args, searchPattern, searchPattern)
		}

		query += " ORDER BY created_at DESC"

		rows, err := services.DB.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var user models.User
			err := rows.Scan(
				&user.ID, &user.Email, &user.FullName, &user.Phone,
				&user.Status, &user.SubscriptionEnd, &user.CreatedAt,
			)
			if err != nil {
				continue
			}

			users = append(users, map[string]interface{}{
				"id":               user.ID,
				"email":            user.Email,
				"full_name":        user.FullName,
				"phone":            user.Phone,
				"status":           user.Status,
				"subscription_end": user.SubscriptionEnd,
				"created_at":       user.CreatedAt,
			})
		}

		if users == nil {
			users = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, users)
	}
}

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

		result, err := services.DB.Exec(`
			UPDATE users SET status = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, input.Status, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
	}
}

func GetAllTransactions(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for filtering
		userID := c.Query("user_id")
		txType := c.Query("type")
		status := c.Query("status")

		// Build query
		query := `SELECT t.id, t.user_id, t.amount, t.type, t.status, t.description, t.created_at, u.email
				  FROM transactions t
				  LEFT JOIN users u ON t.user_id = u.id
				  WHERE 1=1`
		args := []interface{}{}

		if userID != "" {
			query += " AND t.user_id = ?"
			args = append(args, userID)
		}
		if txType != "" {
			query += " AND t.type = ?"
			args = append(args, txType)
		}
		if status != "" {
			query += " AND t.status = ?"
			args = append(args, status)
		}

		query += " ORDER BY t.created_at DESC"

		rows, err := services.DB.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var transactions []map[string]interface{}
		for rows.Next() {
			var tx models.Transaction
			var userEmail string
			err := rows.Scan(
				&tx.ID, &tx.UserID, &tx.Amount, &tx.Type,
				&tx.Status, &tx.Description, &tx.CreatedAt, &userEmail,
			)
			if err != nil {
				continue
			}

			transactions = append(transactions, map[string]interface{}{
				"id":          tx.ID,
				"user_id":     tx.UserID,
				"user_email":  userEmail,
				"amount":      tx.Amount,
				"type":        tx.Type,
				"status":      tx.Status,
				"description": tx.Description,
				"created_at":  tx.CreatedAt,
			})
		}

		if transactions == nil {
			transactions = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, transactions)
	}
}

func GetStatistics(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get total users count
		var totalUsers int
		services.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)

		// Get active users count
		var activeUsers int
		services.DB.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&activeUsers)

		// Get suspended users count
		var suspendedUsers int
		services.DB.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'suspended'").Scan(&suspendedUsers)

		// Get total orders count
		var totalOrders int
		services.DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&totalOrders)

		// Get total transactions and sum
		var totalTransactions int
		var totalRevenue float64
		services.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM transactions WHERE status = 'completed'").Scan(&totalTransactions, &totalRevenue)

		// Get total trading configs
		var totalConfigs int
		services.DB.QueryRow("SELECT COUNT(*) FROM trading_configs").Scan(&totalConfigs)

		// Get active trading configs
		var activeConfigs int
		services.DB.QueryRow("SELECT COUNT(*) FROM trading_configs WHERE is_active = true").Scan(&activeConfigs)

		// Get exchange keys count
		var totalKeys int
		services.DB.QueryRow("SELECT COUNT(*) FROM exchange_keys").Scan(&totalKeys)

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

// Binance API handlers
func GetBinanceFuturesSymbols(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Call Binance Futures API to get exchange info
		resp, err := http.Get("https://fapi.binance.com/fapi/v1/exchangeInfo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Binance data"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Binance API error"})
			return
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Parse JSON
		var exchangeInfo struct {
			Symbols []struct {
				Symbol            string `json:"symbol"`
				Pair              string `json:"pair"`
				ContractType      string `json:"contractType"`
				Status            string `json:"status"`
				BaseAsset         string `json:"baseAsset"`
				QuoteAsset        string `json:"quoteAsset"`
				MarginAsset       string `json:"marginAsset"`
				PricePrecision    int    `json:"pricePrecision"`
				QuantityPrecision int    `json:"quantityPrecision"`
			} `json:"symbols"`
		}

		if err := json.Unmarshal(body, &exchangeInfo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		// Filter only PERPETUAL contracts with TRADING status and USDT quote
		var activeSymbols []map[string]interface{}
		for _, symbol := range exchangeInfo.Symbols {
			if symbol.Status == "TRADING" &&
				symbol.ContractType == "PERPETUAL" &&
				symbol.QuoteAsset == "USDT" {
				activeSymbols = append(activeSymbols, map[string]interface{}{
					"symbol":             symbol.Symbol,
					"pair":               symbol.Pair,
					"base_asset":         symbol.BaseAsset,
					"quote_asset":        symbol.QuoteAsset,
					"price_precision":    symbol.PricePrecision,
					"quantity_precision": symbol.QuantityPrecision,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   len(activeSymbols),
			"symbols": activeSymbols,
		})
	}
}
