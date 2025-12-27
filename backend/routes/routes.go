package api

import (
	"tradercoin/backend/controllers"
	"tradercoin/backend/middleware"
	"tradercoin/backend/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, services *services.Services, wsHub *services.WebSocketHub) {
	// Inject services into context for all routes
	router.Use(func(c *gin.Context) {
		c.Set("services", services)
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// ============ AUTH ROUTES ============
		// Prefix: /api/v1/auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register(services))
			auth.POST("/login", controllers.Login(services))
			auth.POST("/refresh", controllers.RefreshToken(services))
		}

		// ============ USER ROUTES ============
		// Prefix: /api/v1/user
		user := v1.Group("/user")
		// user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", controllers.GetProfile(services))
			user.PUT("/profile", controllers.UpdateProfile(services))
		}

		// ============ CONFIG ROUTES ============
		// Prefix: /api/v1/config
		config := v1.Group("/config")
		config.Use(middleware.AuthMiddleware())
		{
			config.POST("", controllers.CreateBotConfig(services))                    // Create bot config
			config.GET("/list", controllers.ListBotConfigs(services))                 // List all bot configs
			config.GET("/:id", controllers.GetBotConfig(services))                    // Get single bot config
			config.PUT("/:id", controllers.UpdateBotConfig(services))                 // Update bot config
			config.PUT("/:id/set-default", controllers.SetDefaultBotConfig(services)) // Set as default bot
			config.DELETE("/:id", controllers.DeleteBotConfig(services))              // Delete bot config
		}

		// ============ EXCHANGE KEYS ROUTES ============
		// Prefix: /api/v1/keys
		keys := v1.Group("/keys")
		// keys.Use(middleware.AuthMiddleware())
		{
			keys.GET("", controllers.GetExchangeKeys(services))
			keys.POST("", controllers.AddExchangeKey(services))
			keys.PUT("/:id", controllers.UpdateExchangeKey(services))
			keys.DELETE("/:id", controllers.DeleteExchangeKey(services))
		}

		// ============ WEBHOOK ROUTES ============
		// Prefix: /api/v1/webhook
		webhook := v1.Group("/webhook")
		{
			webhook.POST("/binance", controllers.HandleBinanceWebhook(services))         // Binance webhook
			webhook.POST("/tradingview", controllers.HandleTradingViewWebhook(services)) // TradingView alerts
			webhook.POST("/price-alert", controllers.HandlePriceAlert(services))         // Price alerts
			webhook.GET("/logs", controllers.GetWebhookLogs(services))                   // Get webhook logs
			webhook.POST("/create", controllers.CreateWebhook(services))                 // Create webhook URL
		}

		// ============ ORDERS ROUTES ============
		// Prefix: /api/v1/orders
		orders := v1.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.GET("", controllers.GetOrders(services))                      // List all orders
			orders.GET("/history", controllers.GetOrderHistory(services))        // Get order history with filtering
			orders.GET("/completed", controllers.GetCompletedOrders(services))   // Get completed orders (filled/closed)
			orders.GET("/:id", controllers.GetOrder(services))                   // Get single order
			orders.POST("/close/:id", controllers.CloseOrdersBySymbol(services)) // Close all orders and position by symbol
		}

		// ============ MONITORING ROUTES ============
		// Prefix: /api/v1/monitoring
		monitoring := v1.Group("/monitoring")
		// monitoring.Use(middleware.AuthMiddleware())
		{
			monitoring.GET("/status", controllers.GetSystemStatus(services))          // System health status
			monitoring.GET("/metrics", controllers.GetTradingMetrics(services))       // Trading metrics
			monitoring.GET("/positions", controllers.GetActivePositions(services))    // Active positions
			monitoring.GET("/performance", controllers.GetPerformanceStats(services)) // Performance stats
			monitoring.GET("/bot-status", controllers.GetBotStatus(services))         // Bot status
			monitoring.GET("/alerts", controllers.GetAlerts(services))                // Get alerts
			monitoring.PUT("/alerts/:id/read", controllers.MarkAlertRead(services))   // Mark alert as read
		}

		// ============ TRADING ROUTES ============
		// Prefix: /api/v1/trading
		trading := v1.Group("/trading")
		trading.Use(middleware.AuthMiddleware())
		{
			// Order Management
			trading.POST("/place-order", controllers.PlaceOrderDirect(services))    // Place order directly
			trading.POST("/close-order/:id", controllers.CloseOrder(services))      // Close order
			trading.POST("/refresh-pnl/:id", controllers.RefreshPnL(services))      // Refresh PnL from exchange
			trading.GET("/symbols/:config_id", controllers.GetSymbols(services))    // Get symbols from exchange
			trading.GET("/check-order/:id", controllers.CheckOrderStatus(services)) // Check order status
			trading.GET("/account-info/:id", controllers.GetAccountInfo(services))  // Get account info from exchange

			// Testnet utilities
			trading.POST("/refill-testnet/:config_id", controllers.RefillTestnetBalance(services)) // Refill testnet balance

			// WebSocket endpoints
			trading.GET("/ws", controllers.ConnectWebSocket(services, wsHub))                     // WebSocket upgrade
			trading.POST("/listen-key/:exchange_key_id", controllers.CreateListenKey(services))   // Create listen key
			trading.PUT("/listen-key/:exchange_key_id", controllers.KeepAliveListenKey(services)) // Keep alive listen key

			// Legacy config routes (kept for backward compatibility)
			trading.GET("/configs", controllers.GetTradingConfigs(services))
			trading.POST("/configs", controllers.CreateTradingConfig(services))
			trading.PUT("/configs/:id", controllers.UpdateTradingConfig(services))
			trading.DELETE("/configs/:id", controllers.DeleteTradingConfig(services))
		}

		// ============ BINANCE API ROUTES ============
		// Prefix: /api/v1/binance
		binance := v1.Group("/binance")
		{
			binance.GET("/spot/symbols", controllers.GetBinanceSpotSymbols(services))
			binance.GET("/futures/symbols", controllers.GetBinanceFuturesSymbols(services))
		}

		// ============ BITTREX API ROUTES ============
		// Prefix: /api/v1/bittrex
		bittrex := v1.Group("/bittrex")
		{
			bittrex.GET("/symbols", controllers.GetBittrexSymbols(services))
		}

		// ============ TELEGRAM ROUTES ============
		// Prefix: /api/v1/telegram
		telegramController := controllers.NewTelegramController(services.DB)
		telegram := v1.Group("/telegram")
		telegram.Use(middleware.AuthMiddleware())
		{
			telegram.GET("/config", telegramController.GetTelegramConfig)                // Get user's Telegram config
			telegram.POST("/config", telegramController.CreateOrUpdateTelegramConfig)    // Create or update config
			telegram.DELETE("/config", telegramController.DeleteTelegramConfig)          // Delete config
			telegram.POST("/test-connection", telegramController.TestTelegramConnection) // Test connection
			telegram.POST("/test-message", telegramController.SendTestMessage)           // Send test message
			telegram.PATCH("/toggle", telegramController.ToggleTelegramNotifications)    // Enable/disable notifications
		}

		// ============ ADMIN ROUTES ============
		// Prefix: /api/v1/admin
		admin := v1.Group("/admin")
		// admin.Use(middleware.AdminAuthMiddleware())
		{
			admin.POST("/login", controllers.AdminLogin(services))
			admin.GET("/users", controllers.GetAllUsers(services))
			admin.PUT("/users/:id/status", controllers.UpdateUserStatus(services))
			admin.POST("/users/:id/suspend", controllers.SuspendUser(services))           // Khóa user
			admin.POST("/users/:id/activate", controllers.ActivateUser(services))         // Kích hoạt user
			admin.POST("/users/:id/extend", controllers.ExtendUserSubscription(services)) // Gia hạn subscription
			admin.GET("/transactions", controllers.GetAllTransactions(services))
			admin.GET("/statistics", controllers.GetStatistics(services))
			admin.GET("/orders", controllers.GetAllOrdersAdmin(services))                           // Get all orders from all users
			admin.GET("/signals", controllers.ListSignals(services))                                // Get all signals
			admin.GET("/signals/:id", controllers.GetSignal(services))                              // Get single signal
			admin.DELETE("/signals/:id", controllers.DeleteSignal(services))                        // Delete signal
			admin.GET("/logs", controllers.GetAllSystemLogs(services))                              // Get all system logs
			admin.GET("/telegram", telegramController.GetAllTelegramConfigs)                        // Get all Telegram configs
			admin.POST("/telegram", telegramController.AdminCreateTelegramConfig)                   // Create Telegram config for user
			admin.POST("/telegram/test-connection", telegramController.AdminTestTelegramConnection) // Test Telegram connection
			admin.POST("/telegram/start-listener", telegramController.StartCallbackListener)        // Start callback listener for buttons

			// Admin profile & settings - Require authentication
			adminAuth := admin.Group("")
			adminAuth.Use(middleware.AdminAuthMiddleware())
			{
				adminAuth.GET("/profile", controllers.GetAdminProfile(services))      // Get admin profile
				adminAuth.PUT("/profile", controllers.UpdateAdminProfile(services))   // Update admin profile
				adminAuth.PUT("/password", controllers.ChangeAdminPassword(services)) // Change admin password
			}
		}

		// ============ TRADING SIGNALS ROUTES ============
		// Prefix: /api/v1/signals
		signals := v1.Group("/signals")
		{
			// Public webhook endpoint (no auth) for TradingView
			signals.POST("/webhook/tradingview", controllers.TradingViewWebhook(services, wsHub))
			// Prefixed webhook to identify user by unique prefix
			signals.POST("/webhook/:prefix", controllers.TradingViewWebhook(services, wsHub))

			// Authenticated endpoints
			signalsAuth := signals.Group("")
			signalsAuth.Use(middleware.AuthMiddleware())
			{
				signalsAuth.GET("", controllers.ListSignals(services))                   // List all signals
				signalsAuth.GET("/:id", controllers.GetSignal(services))                 // Get single signal
				signalsAuth.POST("/:id/execute", controllers.ExecuteSignal(services))    // Execute signal with bot config
				signalsAuth.PUT("/:id/status", controllers.UpdateSignalStatus(services)) // Update signal status
				signalsAuth.DELETE("/:id", controllers.DeleteSignal(services))           // Delete signal
				// Webhook prefix management
				signalsAuth.GET("/webhook/prefix", controllers.GetWebhookPrefix(services))     // Get latest active prefix
				signalsAuth.POST("/webhook/prefix", controllers.CreateWebhookPrefix(services)) // Create new prefix
			}
		}

		// ============ SYSTEM LOGS ROUTES ============
		// Prefix: /api/v1/logs
		logs := v1.Group("/logs")
		logs.Use(middleware.AuthMiddleware())
		{
			logs.GET("", controllers.GetSystemLogs(services))            // Get system logs
			logs.GET("/stats", controllers.GetSystemLogStats(services))  // Get log statistics
			logs.DELETE("/clear", controllers.ClearSystemLogs(services)) // Clear old logs
		}

		// ============ EXCHANGE CONFIG ROUTES ============
		// Prefix: /api/v1/exchanges
		exchanges := v1.Group("/exchanges")
		{
			// Public endpoints (for frontend dropdowns)
			exchanges.GET("/supported", controllers.GetSupportedExchanges) // Get active exchanges list
			exchanges.GET("", controllers.ListExchangeConfigs)             // List all exchange configs
			exchanges.GET("/:exchange", controllers.GetExchangeConfig)     // Get specific exchange config

			// Admin endpoints (TODO: Add AdminAuthMiddleware)
			exchangesAdmin := exchanges.Group("")
			exchangesAdmin.Use(middleware.AuthMiddleware())
			{
				exchangesAdmin.POST("", controllers.CreateExchangeConfig)             // Create exchange config
				exchangesAdmin.PUT("/:id", controllers.UpdateExchangeConfig)          // Update exchange config
				exchangesAdmin.DELETE("/:id", controllers.DeleteExchangeConfig)       // Delete exchange config
				exchangesAdmin.PATCH("/:id/toggle", controllers.ToggleExchangeStatus) // Toggle exchange status
			}
		}
	}
}
