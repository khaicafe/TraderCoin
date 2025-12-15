package models

import "time"

type User struct {
	ID              int       `json:"id"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	FullName        string    `json:"full_name"`
	Phone           string    `json:"phone"`
	Status          string    `json:"status"`
	SubscriptionEnd time.Time `json:"subscription_end"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ExchangeKey struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Exchange  string    `json:"exchange"`
	APIKey    string    `json:"api_key"`
	APISecret string    `json:"-"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TradingConfig struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	Exchange          string    `json:"exchange"`
	Symbol            string    `json:"symbol"`
	StopLossPercent   float64   `json:"stop_loss_percent"`
	TakeProfitPercent float64   `json:"take_profit_percent"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"`
	OrderID   string    `json:"order_id"`
	Side      string    `json:"side"`
	Type      string    `json:"type"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Admin struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
