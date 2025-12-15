package api

import (
	"tradercoin/backend/controllers"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, services *services.Services) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register(services))
			auth.POST("/login", controllers.Login(services))
			auth.POST("/refresh", controllers.RefreshToken(services))
		}

		// User routes (protected)
		user := v1.Group("/user")
		// user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", controllers.GetProfile(services))
			user.PUT("/profile", controllers.UpdateProfile(services))
		}

		// Exchange keys routes
		keys := v1.Group("/keys")
		// keys.Use(middleware.AuthMiddleware())
		{
			keys.GET("", controllers.GetExchangeKeys(services))
			keys.POST("", controllers.AddExchangeKey(services))
			keys.PUT("/:id", controllers.UpdateExchangeKey(services))
			keys.DELETE("/:id", controllers.DeleteExchangeKey(services))
		}

		// Trading config routes
		trading := v1.Group("/trading")
		// trading.Use(middleware.AuthMiddleware())
		{
			trading.GET("/configs", controllers.GetTradingConfigs(services))
			trading.POST("/configs", controllers.CreateTradingConfig(services))
			trading.PUT("/configs/:id", controllers.UpdateTradingConfig(services))
			trading.DELETE("/configs/:id", controllers.DeleteTradingConfig(services))
		}

		// Orders routes
		orders := v1.Group("/orders")
		// orders.Use(middleware.AuthMiddleware())
		{
			orders.GET("", controllers.GetOrders(services))
			orders.GET("/:id", controllers.GetOrder(services))
		}

		// Binance API routes
		binance := v1.Group("/binance")
		{
			binance.GET("/futures/symbols", controllers.GetBinanceFuturesSymbols(services))
		}

		// Admin routes
		admin := v1.Group("/admin")
		// admin.Use(middleware.AdminAuthMiddleware())
		{
			admin.POST("/login", controllers.AdminLogin(services))
			admin.GET("/users", controllers.GetAllUsers(services))
			admin.PUT("/users/:id/status", controllers.UpdateUserStatus(services))
			admin.GET("/transactions", controllers.GetAllTransactions(services))
			admin.GET("/statistics", controllers.GetStatistics(services))
		}
	}
}
