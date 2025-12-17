package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"tradercoin/backend/config"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// OrderUpdate represents a real-time order update from exchange
type OrderUpdate struct {
	UserID        uint    `json:"user_id"`
	ExchangeKeyID uint    `json:"exchange_key_id"`
	Exchange      string  `json:"exchange"`
	TradingMode   string  `json:"trading_mode"`
	OrderID       string  `json:"order_id"`
	ClientOrderID string  `json:"client_order_id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	Price         float64 `json:"price"`
	Quantity      float64 `json:"quantity"`
	ExecutedQty   float64 `json:"executed_qty"`
	ExecutedPrice float64 `json:"executed_price"`
	CurrentPrice  float64 `json:"current_price"` // Current market price
	UpdateTime    int64   `json:"update_time"`
}

// ExchangeConnection represents a WebSocket connection to an exchange
type ExchangeConnection struct {
	ExchangeKeyID uint
	UserID        uint
	Exchange      string
	TradingMode   string
	ListenKey     string
	Conn          *websocket.Conn

	// Map of session IDs to user's browser WebSocket connections
	UserTabs map[string]*websocket.Conn
	mu       sync.RWMutex

	// Keep-alive ticker
	keepAliveTicker *time.Ticker
	done            chan bool
}

// RegisterRequest for registering a user connection
type RegisterRequest struct {
	UserID        uint
	ExchangeKeyID uint
	Exchange      string
	TradingMode   string
	ListenKey     string
	SessionID     string
	UserConn      *websocket.Conn
}

// UnregisterRequest for unregistering a user connection
type UnregisterRequest struct {
	UserID        uint
	ExchangeKeyID uint
	SessionID     string
}

// BroadcastMessage for broadcasting to users
type BroadcastMessage struct {
	UserID uint
	Type   string
	Data   interface{}
}

// WebSocketHub manages all exchange WebSocket connections
type WebSocketHub struct {
	// Map: "exchange_tradingMode_exchangeKeyID" → ExchangeConnection
	ExchangeConns map[string]*ExchangeConnection

	// Map: userID → list of sessionIDs
	UserSessions map[uint]map[string]bool

	// Channels
	Register   chan *RegisterRequest
	Unregister chan *UnregisterRequest
	Broadcast  chan *BroadcastMessage

	// Database
	DB *gorm.DB

	// Config
	Config *config.Config

	mu sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(db *gorm.DB, cfg *config.Config) *WebSocketHub {
	return &WebSocketHub{
		ExchangeConns: make(map[string]*ExchangeConnection),
		UserSessions:  make(map[uint]map[string]bool),
		Register:      make(chan *RegisterRequest, 100),
		Unregister:    make(chan *UnregisterRequest, 100),
		Broadcast:     make(chan *BroadcastMessage, 1000),
		DB:            db,
		Config:        cfg,
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	log.Println("WebSocket Hub started")

	for {
		select {
		case req := <-h.Register:
			h.handleRegister(req)

		case req := <-h.Unregister:
			h.handleUnregister(req)

		case msg := <-h.Broadcast:
			h.handleBroadcast(msg)
		}
	}
}

// handleRegister registers a new user connection
func (h *WebSocketHub) handleRegister(req *RegisterRequest) {
	log.Printf("Registering user %d, session %s for %s %s (key %d)",
		req.UserID, req.SessionID, req.Exchange, req.TradingMode, req.ExchangeKeyID)

	// Create connection key
	connKey := fmt.Sprintf("%s_%s_%d", req.Exchange, req.TradingMode, req.ExchangeKeyID)

	h.mu.Lock()

	// Track user session
	if h.UserSessions[req.UserID] == nil {
		h.UserSessions[req.UserID] = make(map[string]bool)
	}
	h.UserSessions[req.UserID][req.SessionID] = true

	h.mu.Unlock()

	// Check if exchange connection exists
	h.mu.RLock()
	exchConn, exists := h.ExchangeConns[connKey]
	h.mu.RUnlock()

	if !exists {
		// Create new exchange connection
		wsURL := h.getExchangeWSURL(req.Exchange, req.TradingMode, req.ListenKey)
		if wsURL == "" {
			log.Printf("Unsupported exchange: %s", req.Exchange)
			return
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			log.Printf("Failed to connect to %s: %v", req.Exchange, err)
			return
		}

		exchConn = &ExchangeConnection{
			ExchangeKeyID: req.ExchangeKeyID,
			UserID:        req.UserID,
			Exchange:      req.Exchange,
			TradingMode:   req.TradingMode,
			ListenKey:     req.ListenKey,
			Conn:          conn,
			UserTabs:      make(map[string]*websocket.Conn),
			done:          make(chan bool),
		}

		h.mu.Lock()
		h.ExchangeConns[connKey] = exchConn
		h.mu.Unlock()

		// Start listening to exchange
		go h.listenToExchange(exchConn, connKey)

		// Start keep-alive for listen key
		go h.keepAliveListenKey(exchConn)

		log.Printf("Created new exchange connection: %s", connKey)
	}

	// Add user's browser connection
	exchConn.mu.Lock()
	exchConn.UserTabs[req.SessionID] = req.UserConn
	exchConn.mu.Unlock()

	log.Printf("User %d now has %d tabs connected to %s",
		req.UserID, len(exchConn.UserTabs), connKey)
}

// handleUnregister unregisters a user connection
func (h *WebSocketHub) handleUnregister(req *UnregisterRequest) {
	log.Printf("Unregistering user %d, session %s (key %d)",
		req.UserID, req.SessionID, req.ExchangeKeyID)

	connKey := fmt.Sprintf("binance_spot_%d", req.ExchangeKeyID) // TODO: support other exchanges

	h.mu.Lock()

	// Remove session from user sessions
	if sessions, ok := h.UserSessions[req.UserID]; ok {
		delete(sessions, req.SessionID)
		if len(sessions) == 0 {
			delete(h.UserSessions, req.UserID)
		}
	}

	h.mu.Unlock()

	// Remove from exchange connection
	h.mu.RLock()
	exchConn, exists := h.ExchangeConns[connKey]
	h.mu.RUnlock()

	if exists {
		exchConn.mu.Lock()
		delete(exchConn.UserTabs, req.SessionID)
		tabCount := len(exchConn.UserTabs)
		exchConn.mu.Unlock()

		log.Printf("User %d now has %d tabs connected to %s",
			req.UserID, tabCount, connKey)

		// If no more tabs, close exchange connection
		if tabCount == 0 {
			log.Printf("No more tabs for %s, closing exchange connection", connKey)
			exchConn.done <- true
			exchConn.Conn.Close()

			h.mu.Lock()
			delete(h.ExchangeConns, connKey)
			h.mu.Unlock()
		}
	}
}

// handleBroadcast sends message to user's connections
func (h *WebSocketHub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	sessions := h.UserSessions[msg.UserID]
	h.mu.RUnlock()

	if sessions == nil {
		return
	}

	// Find all exchange connections for this user
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, exchConn := range h.ExchangeConns {
		if exchConn.UserID != msg.UserID {
			continue
		}

		exchConn.mu.RLock()
		for sessionID, userConn := range exchConn.UserTabs {
			if !sessions[sessionID] {
				continue
			}

			err := userConn.WriteJSON(map[string]interface{}{
				"type": msg.Type,
				"data": msg.Data,
			})
			if err != nil {
				log.Printf("Failed to send to user %d session %s: %v",
					msg.UserID, sessionID, err)
			}
		}
		exchConn.mu.RUnlock()
	}
}

// listenToExchange listens to exchange WebSocket messages
func (h *WebSocketHub) listenToExchange(exchConn *ExchangeConnection, connKey string) {
	defer func() {
		log.Printf("Stopped listening to %s", connKey)
	}()

	for {
		select {
		case <-exchConn.done:
			return

		default:
			var message map[string]interface{}
			err := exchConn.Conn.ReadJSON(&message)
			if err != nil {
				log.Printf("Error reading from %s: %v", exchConn.Exchange, err)
				return
			}

			// Process exchange message
			h.processExchangeMessage(exchConn, message)
		}
	}
}

// processExchangeMessage processes a message from exchange
func (h *WebSocketHub) processExchangeMessage(
	exchConn *ExchangeConnection,
	message map[string]interface{},
) {
	var orderUpdate *OrderUpdate

	switch exchConn.Exchange {
	case "binance":
		orderUpdate = h.parseBinanceMessage(exchConn, message)
	case "okx":
		orderUpdate = h.parseOKXMessage(exchConn, message)
	case "bybit":
		orderUpdate = h.parseBybitMessage(exchConn, message)
	default:
		log.Printf("Unsupported exchange: %s", exchConn.Exchange)
		return
	}

	if orderUpdate == nil {
		return
	}

	// Update order in database
	h.updateOrderInDB(orderUpdate)

	// Broadcast to user
	h.Broadcast <- &BroadcastMessage{
		UserID: exchConn.UserID,
		Type:   "order_update",
		Data:   orderUpdate,
	}
}

// parseBinanceMessage parses Binance WebSocket message
func (h *WebSocketHub) parseBinanceMessage(
	exchConn *ExchangeConnection,
	message map[string]interface{},
) *OrderUpdate {
	eventType, ok := message["e"].(string)
	if !ok || eventType != "executionReport" {
		return nil
	}

	symbol := getStringValue(message, "s")

	// Create order update
	update := &OrderUpdate{
		UserID:        exchConn.UserID,
		ExchangeKeyID: exchConn.ExchangeKeyID,
		Exchange:      exchConn.Exchange,
		TradingMode:   exchConn.TradingMode,
		OrderID:       fmt.Sprintf("%v", message["i"]),
		ClientOrderID: getStringValue(message, "c"),
		Symbol:        symbol,
		Side:          getStringValue(message, "S"),
		Type:          getStringValue(message, "o"),
		Status:        getStringValue(message, "X"),
		Price:         getFloatValue(message, "p"),
		Quantity:      getFloatValue(message, "q"),
		ExecutedQty:   getFloatValue(message, "z"),
		ExecutedPrice: getFloatValue(message, "L"),
		CurrentPrice:  0, // Will be fetched from API
		UpdateTime:    getInt64Value(message, "E"),
	}

	// Fetch current market price from Binance API (synchronously)
	update.CurrentPrice = h.fetchCurrentMarketPrice(symbol, exchConn.TradingMode)

	return update
}

// fetchCurrentMarketPrice fetches current market price from Binance API
func (h *WebSocketHub) fetchCurrentMarketPrice(symbol, tradingMode string) float64 {
	if symbol == "" {
		return 0
	}

	// Get Binance config
	binanceCfg := h.Config.Exchanges.Binance

	// Determine API endpoint based on trading mode
	var apiURL string
	if tradingMode == "futures" {
		apiURL = fmt.Sprintf("%s%s?symbol=%s",
			binanceCfg.FuturesAPIURL,
			binanceCfg.FuturesTickerAPI,
			symbol)
	} else {
		apiURL = fmt.Sprintf("%s%s?symbol=%s",
			binanceCfg.SpotAPIURL,
			binanceCfg.SpotTickerAPI,
			symbol)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Make request
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("Error fetching current price for %s: %v", symbol, err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch price for %s: status %d", symbol, resp.StatusCode)
		return 0
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading price response for %s: %v", symbol, err)
		return 0
	}

	// Parse response
	var priceData struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.Unmarshal(body, &priceData); err != nil {
		log.Printf("Error parsing price data for %s: %v", symbol, err)
		return 0
	}

	// Convert price string to float64
	var price float64
	fmt.Sscanf(priceData.Price, "%f", &price)

	log.Printf("✓ Fetched current market price for %s: %.8f", symbol, price)
	return price
}

// parseOKXMessage parses OKX WebSocket message
func (h *WebSocketHub) parseOKXMessage(
	exchConn *ExchangeConnection,
	message map[string]interface{},
) *OrderUpdate {
	// TODO: Implement OKX message parsing
	return nil
}

// parseBybitMessage parses Bybit WebSocket message
func (h *WebSocketHub) parseBybitMessage(
	exchConn *ExchangeConnection,
	message map[string]interface{},
) *OrderUpdate {
	// TODO: Implement Bybit message parsing
	return nil
}

// updateOrderInDB updates order in database
func (h *WebSocketHub) updateOrderInDB(update *OrderUpdate) {
	var order struct {
		ID             uint
		Status         string
		FilledQuantity float64
		FilledPrice    float64
		UpdatedAt      time.Time
	}

	// Find order by exchange order ID
	result := h.DB.Model(&order).
		Where("user_id = ? AND exchange_key_id = ? AND order_id = ?",
			update.UserID, update.ExchangeKeyID, update.OrderID).
		Updates(map[string]interface{}{
			"status":          update.Status,
			"filled_quantity": update.ExecutedQty,
			"filled_price":    update.ExecutedPrice,
			"current_price":   update.CurrentPrice,
			"updated_at":      time.Now(),
		})

	if result.Error != nil {
		log.Printf("Failed to update order in DB: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("Updated order %s: status=%s, filled=%f",
			update.OrderID, update.Status, update.ExecutedQty)
	}
}

// keepAliveListenKey keeps the listen key alive
func (h *WebSocketHub) keepAliveListenKey(exchConn *ExchangeConnection) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Implement keep-alive logic for each exchange
			log.Printf("Keep-alive for %s listen key %s",
				exchConn.Exchange, exchConn.ListenKey)

		case <-exchConn.done:
			return
		}
	}
}

// getExchangeWSURL returns WebSocket URL for exchange using adapters
func (h *WebSocketHub) getExchangeWSURL(exchange, tradingMode, listenKey string) string {
	// Get appropriate adapter based on exchange
	adapter := GetExchangeAdapter(exchange, false) // Use production by default
	if adapter == nil {
		return ""
	}

	return adapter.GetWSURL(tradingMode, listenKey)
}

// Helper functions
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case string:
			var f float64
			fmt.Sscanf(v, "%f", &f)
			return f
		}
	}
	return 0
}

func getInt64Value(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		}
	}
	return 0
}
