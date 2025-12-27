package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Email           string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash    string         `gorm:"not null;size:255" json:"-"`
	FullName        string         `gorm:"size:255" json:"full_name"`
	Phone           string         `gorm:"size:50" json:"phone"`
	ChatID          string         `gorm:"size:100" json:"chat_id"` // Telegram Chat ID
	Status          string         `gorm:"size:50;default:active" json:"status"`
	SubscriptionEnd *time.Time     `json:"subscription_end"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	ExchangeKeys   []ExchangeKey   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	TradingConfigs []TradingConfig `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Orders         []Order         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Transactions   []Transaction   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	UserSignals    []UserSignal    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

type ExchangeKey struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index:idx_user_exchange,unique" json:"user_id"`
	Exchange     string         `gorm:"not null;size:50;index:idx_user_exchange,unique" json:"exchange"`
	TradingMode  string         `gorm:"not null;size:20;default:'spot'" json:"trading_mode"` // spot, futures
	APIKey       string         `gorm:"not null;size:255" json:"api_key"`
	APISecret    string         `gorm:"not null;size:255" json:"-"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	ListenKey    string         `gorm:"size:255" json:"-"` // WebSocket listen key
	ListenKeyExp *time.Time     `json:"-"`                 // Listen key expiration
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

type TradingConfig struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	UserID              uint           `gorm:"not null;index" json:"user_id"`
	Name                string         `gorm:"size:100" json:"name"` // Bot name
	Exchange            string         `gorm:"not null;size:50" json:"exchange"`
	Symbol              string         `gorm:"not null;size:50" json:"symbol"`
	Amount              float64        `gorm:"type:decimal(10,2)" json:"amount"`
	TradingMode         string         `gorm:"size:20;default:'spot'" json:"trading_mode"`    // spot, futures, margin
	Leverage            int            `gorm:"default:1" json:"leverage"`                     // Leverage for futures/margin trading (1-125)
	MarginMode          string         `gorm:"size:20;default:'ISOLATED'" json:"margin_mode"` // ISOLATED, CROSSED (for futures)
	APIKey              string         `gorm:"size:255" json:"-"`                             // Not exposed in JSON for security
	APISecret           string         `gorm:"size:255" json:"-"`                             // Not exposed in JSON for security
	StopLossPercent     float64        `gorm:"type:decimal(10,2)" json:"stop_loss_percent"`
	TakeProfitPercent   float64        `gorm:"type:decimal(10,2)" json:"take_profit_percent"`
	TrailingStopPercent float64        `gorm:"type:decimal(10,2);default:0" json:"trailing_stop_percent"` // Trailing stop for futures
	EnableTrailingStop  bool           `gorm:"default:false" json:"enable_trailing_stop"`                 // Enable/disable trailing stop
	ActivationPrice     float64        `gorm:"type:decimal(20,8);default:0" json:"activation_price"`      // Activation price for trailing stop
	CallbackRate        float64        `gorm:"type:decimal(10,2);default:1" json:"callback_rate"`         // Callback rate for trailing stop (0.1-5%)
	IsDefault           bool           `gorm:"default:false" json:"is_default"`                           // Only one default bot per user
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

type Order struct {
	ID               uint    `gorm:"primaryKey" json:"id"`
	UserID           uint    `gorm:"not null;index" json:"user_id"`
	ExchangeKeyID    uint    `gorm:"index" json:"exchange_key_id"` // Link to ExchangeKey (API Key)
	BotConfigID      uint    `gorm:"index" json:"bot_config_id"`   // Link to TradingConfig
	Exchange         string  `gorm:"not null;size:50" json:"exchange"`
	Symbol           string  `gorm:"not null;size:50" json:"symbol"`
	OrderID          string  `gorm:"size:255;index" json:"order_id"`  // Exchange's order ID
	ClientOrderID    string  `gorm:"size:255" json:"client_order_id"` // Our generated order ID
	Side             string  `gorm:"not null;size:10" json:"side"`
	Type             string  `gorm:"not null;size:20" json:"type"`
	Quantity         float64 `gorm:"type:decimal(20,8)" json:"quantity"`
	Price            float64 `gorm:"type:decimal(20,8)" json:"price"`
	FilledPrice      float64 `gorm:"type:decimal(20,8)" json:"filled_price"`
	FilledQuantity   float64 `gorm:"type:decimal(20,8)" json:"filled_quantity"` // Executed quantity
	CurrentPrice     float64 `gorm:"type:decimal(20,8)" json:"current_price"`   // Current market price from exchange
	Status           string  `gorm:"size:50;default:pending" json:"status"`
	TradingMode      string  `gorm:"size:20;default:spot" json:"trading_mode"` // spot, futures, margin
	Leverage         int     `gorm:"default:1" json:"leverage"`
	StopLossPrice    float64 `gorm:"type:decimal(20,8)" json:"stop_loss_price"`
	TakeProfitPrice  float64 `gorm:"type:decimal(20,8)" json:"take_profit_price"`
	AlgoIDStopLoss   string  `gorm:"size:100" json:"algo_id_stop_loss"`   // Binance Algo Order ID for Stop Loss
	AlgoIDTakeProfit string  `gorm:"size:100" json:"algo_id_take_profit"` // Binance Algo Order ID for Take Profit
	PnL              float64 `gorm:"type:decimal(20,8)" json:"pnl"`
	PnLPercent       float64 `gorm:"type:decimal(10,2)" json:"pnl_percent"`

	// Position Info (for Futures) - Not storing position_amt and mark_price as they change constantly
	PositionSide     string  `gorm:"size:20" json:"position_side"`                // LONG/SHORT/BOTH
	LiquidationPrice float64 `gorm:"type:decimal(20,8)" json:"liquidation_price"` // Liquidation price
	MarginType       string  `gorm:"size:20" json:"margin_type"`                  // isolated/cross
	IsolatedMargin   float64 `gorm:"type:decimal(20,8)" json:"isolated_margin"`   // Margin for isolated mode

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

type Transaction struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Amount      float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Type        string         `gorm:"not null;size:50" json:"type"`
	Status      string         `gorm:"size:50;default:pending" json:"status"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

type Admin struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null;size:255" json:"-"`
	FullName     string         `gorm:"size:255" json:"full_name"`
	Role         string         `gorm:"size:50;default:admin" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type TradingSignal struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Symbol        string    `gorm:"not null;size:50;index" json:"symbol"`
	Action        string    `gorm:"not null;size:20" json:"action"` // buy, sell, close
	Price         float64   `gorm:"type:decimal(20,8)" json:"price"`
	StopLoss      float64   `gorm:"type:decimal(20,8)" json:"stop_loss"`
	TakeProfit    float64   `gorm:"type:decimal(20,8)" json:"take_profit"`
	Message       string    `gorm:"type:text" json:"message"`
	Strategy      string    `gorm:"size:100" json:"strategy"`
	WebhookPrefix string    `gorm:"size:64;index" json:"webhook_prefix"`
	ReceivedAt    time.Time `gorm:"not null;index" json:"received_at"`
	RawPayload    string    `gorm:"type:text" json:"raw_payload"` // Store original webhook JSON
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	UserSignals []UserSignal `gorm:"foreignKey:SignalID;constraint:OnDelete:CASCADE" json:"user_signals,omitempty"`
}

// UserSignal tracks each user's interaction with a signal (many-to-many with status)
type UserSignal struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"not null;index:idx_user_signal,unique" json:"user_id"`
	SignalID    uint       `gorm:"not null;index:idx_user_signal,unique;index" json:"signal_id"`
	Status      string     `gorm:"size:20;default:pending;index" json:"status"` // pending, executed, failed, ignored
	BotConfigID *uint      `gorm:"index" json:"bot_config_id"`                  // Which bot config was used
	OrderID     *uint      `gorm:"index" json:"order_id"`                       // Link to order if executed
	ExecutedAt  *time.Time `json:"executed_at"`
	ErrorMsg    string     `gorm:"type:text" json:"error_msg"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	User   User           `gorm:"foreignKey:UserID" json:"-"`
	Signal TradingSignal  `gorm:"foreignKey:SignalID" json:"signal,omitempty"`
	Order  *Order         `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Config *TradingConfig `gorm:"foreignKey:BotConfigID" json:"config,omitempty"`
}

// WebhookPrefix associates a unique webhook prefix with a user
type WebhookPrefix struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Prefix    string         `gorm:"uniqueIndex;size:64;not null" json:"prefix"`
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// SystemLog stores system activity logs for user actions
type SystemLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Level     string    `gorm:"size:20;not null;default:'INFO'" json:"level"` // SUCCESS, INFO, WARNING, ERROR
	Action    string    `gorm:"size:100;not null" json:"action"`              // Action description
	Symbol    string    `gorm:"size:50;index" json:"symbol"`                  // Trading symbol if applicable
	Exchange  string    `gorm:"size:50" json:"exchange"`                      // Exchange name
	OrderID   *uint     `gorm:"index" json:"order_id"`                        // Related order ID if applicable
	Price     float64   `gorm:"type:decimal(20,8)" json:"price"`              // Price if applicable
	Amount    float64   `gorm:"type:decimal(20,8)" json:"amount"`             // Amount if applicable
	Message   string    `gorm:"type:text;not null" json:"message"`            // Log message
	Details   string    `gorm:"type:text" json:"details"`                     // Additional details (JSON string)
	IPAddress string    `gorm:"size:45" json:"ip_address"`                    // Client IP
	UserAgent string    `gorm:"size:255" json:"user_agent"`                   // User agent
	CreatedAt time.Time `gorm:"index" json:"created_at"`

	// Relationships
	User  User   `gorm:"foreignKey:UserID" json:"-"`
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// ExchangeAPIConfig stores API endpoint configurations for different exchanges
type ExchangeAPIConfig struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	Exchange             string         `gorm:"uniqueIndex;not null;size:50" json:"exchange"`       // binance, bingx, bittrex, okx, bybit, etc.
	DisplayName          string         `gorm:"size:100" json:"display_name"`                       // Display name (e.g., "Binance", "BingX")
	SpotAPIURL           string         `gorm:"size:255;not null" json:"spot_api_url"`              // Spot REST API base URL
	SpotAPITestnetURL    string         `gorm:"size:255" json:"spot_api_testnet_url"`               // Spot Testnet REST API URL
	SpotWSURL            string         `gorm:"size:255" json:"spot_ws_url"`                        // Spot WebSocket stream URL
	SpotWSTestnetURL     string         `gorm:"size:255" json:"spot_ws_testnet_url"`                // Spot Testnet WebSocket URL
	FuturesAPIURL        string         `gorm:"size:255" json:"futures_api_url"`                    // Futures REST API URL
	FuturesAPITestnetURL string         `gorm:"size:255" json:"futures_api_testnet_url"`            // Futures Testnet REST API URL
	FuturesWSURL         string         `gorm:"size:255" json:"futures_ws_url"`                     // Futures WebSocket URL
	FuturesWSTestnetURL  string         `gorm:"size:255" json:"futures_ws_testnet_url"`             // Futures Testnet WebSocket URL
	IsActive             bool           `gorm:"default:true" json:"is_active"`                      // Enable/disable exchange
	SupportSpot          bool           `gorm:"default:true" json:"support_spot"`                   // Support spot trading
	SupportFutures       bool           `gorm:"default:false" json:"support_futures"`               // Support futures trading
	SupportMargin        bool           `gorm:"default:false" json:"support_margin"`                // Support margin trading
	DefaultLeverage      int            `gorm:"default:1" json:"default_leverage"`                  // Default leverage
	MaxLeverage          int            `gorm:"default:1" json:"max_leverage"`                      // Max leverage (e.g., 125 for Binance, 150 for BingX)
	MinOrderSize         float64        `gorm:"type:decimal(20,8);default:0" json:"min_order_size"` // Minimum order size
	MakerFee             float64        `gorm:"type:decimal(10,4);default:0.001" json:"maker_fee"`  // Maker fee (0.1% = 0.001)
	TakerFee             float64        `gorm:"type:decimal(10,4);default:0.001" json:"taker_fee"`  // Taker fee (0.1% = 0.001)
	Notes                string         `gorm:"type:text" json:"notes"`                             // Additional notes
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

// TelegramConfig stores Telegram bot configuration
type TelegramConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;uniqueIndex" json:"user_id"` // One config per user
	BotToken  string         `gorm:"not null;size:255" json:"bot_token"`  // Telegram bot token
	ChatID    string         `gorm:"not null;size:100" json:"chat_id"`    // Telegram chat ID
	BotName   string         `gorm:"size:100" json:"bot_name"`            // Bot name for display
	IsEnabled bool           `gorm:"default:true" json:"is_enabled"`      // Enable/disable notifications
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}
