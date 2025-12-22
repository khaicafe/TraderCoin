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
	TradingMode         string         `gorm:"size:20;default:'spot'" json:"trading_mode"` // spot, futures, margin
	Leverage            int            `gorm:"default:1" json:"leverage"`                  // Leverage for futures/margin trading (1-125)
	APIKey              string         `gorm:"size:255" json:"-"`                          // Not exposed in JSON for security
	APISecret           string         `gorm:"size:255" json:"-"`                          // Not exposed in JSON for security
	StopLossPercent     float64        `gorm:"type:decimal(10,2)" json:"stop_loss_percent"`
	TakeProfitPercent   float64        `gorm:"type:decimal(10,2)" json:"take_profit_percent"`
	TrailingStopPercent float64        `gorm:"type:decimal(10,2);default:0" json:"trailing_stop_percent"` // Trailing stop for futures
	IsDefault           bool           `gorm:"default:false" json:"is_default"`                           // Only one default bot per user
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

type Order struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	UserID           uint           `gorm:"not null;index" json:"user_id"`
	ExchangeKeyID    uint           `gorm:"index" json:"exchange_key_id"` // Link to ExchangeKey (API Key)
	BotConfigID      uint           `gorm:"index" json:"bot_config_id"`   // Link to TradingConfig
	Exchange         string         `gorm:"not null;size:50" json:"exchange"`
	Symbol           string         `gorm:"not null;size:50" json:"symbol"`
	OrderID          string         `gorm:"size:255;index" json:"order_id"`  // Exchange's order ID
	ClientOrderID    string         `gorm:"size:255" json:"client_order_id"` // Our generated order ID
	Side             string         `gorm:"not null;size:10" json:"side"`
	Type             string         `gorm:"not null;size:20" json:"type"`
	Quantity         float64        `gorm:"type:decimal(20,8)" json:"quantity"`
	Price            float64        `gorm:"type:decimal(20,8)" json:"price"`
	FilledPrice      float64        `gorm:"type:decimal(20,8)" json:"filled_price"`
	FilledQuantity   float64        `gorm:"type:decimal(20,8)" json:"filled_quantity"` // Executed quantity
	CurrentPrice     float64        `gorm:"type:decimal(20,8)" json:"current_price"`   // Current market price from exchange
	Status           string         `gorm:"size:50;default:pending" json:"status"`
	TradingMode      string         `gorm:"size:20;default:spot" json:"trading_mode"` // spot, futures, margin
	Leverage         int            `gorm:"default:1" json:"leverage"`
	StopLossPrice    float64        `gorm:"type:decimal(20,8)" json:"stop_loss_price"`
	TakeProfitPrice  float64        `gorm:"type:decimal(20,8)" json:"take_profit_price"`
	AlgoIDStopLoss   string         `gorm:"size:100" json:"algo_id_stop_loss"`   // Binance Algo Order ID for Stop Loss
	AlgoIDTakeProfit string         `gorm:"size:100" json:"algo_id_take_profit"` // Binance Algo Order ID for Take Profit
	PnL              float64        `gorm:"type:decimal(20,8)" json:"pnl"`
	PnLPercent       float64        `gorm:"type:decimal(10,2)" json:"pnl_percent"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

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
	Status        string    `gorm:"size:20;default:pending;index" json:"status"` // pending, executed, failed, ignored
	OrderID       *uint     `gorm:"index" json:"order_id"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
	WebhookPrefix string    `gorm:"size:64;index" json:"webhook_prefix"`
	ReceivedAt    time.Time `gorm:"not null;index" json:"received_at"`
	ExecutedAt    time.Time `json:"executed_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
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
