package services

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Services struct {
	DB           *gorm.DB
	Redis        *redis.Client
	OrderMonitor *OrderMonitorService // Background worker for order status updates
}

// GetWebSocketUpgrader returns WebSocket upgrader with CORS settings
func (s *Services) GetWebSocketUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for now (configure properly in production)
			return true
		},
	}
}

// GetExchangeAdapter returns appropriate exchange adapter
func (s *Services) GetExchangeAdapter(exchange string, isTestnet bool) ExchangeAdapter {
	return GetExchangeAdapter(exchange, isTestnet)
}
