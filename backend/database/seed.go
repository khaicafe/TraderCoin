package database

import (
	"log"
	"time"
	"tradercoin/backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData creates sample admin and user accounts for testing
func SeedData(db *gorm.DB) error {
	log.Println("Starting to seed sample data...")

	// Check if admin already exists
	var adminCount int64
	db.Model(&models.Admin{}).Where("email = ?", "admin@tradercoin.com").Count(&adminCount)

	if adminCount == 0 {
		// Create admin account
		adminPassword := "admin123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash admin password: %v", err)
			return err
		}

		admin := models.Admin{
			Email:        "admin@tradercoin.com",
			PasswordHash: string(hashedPassword),
			FullName:     "System Administrator",
			Role:         "admin",
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Failed to create admin: %v", err)
			return err
		}
		log.Println("✅ Created admin account: admin@tradercoin.com / admin123")
	} else {
		log.Println("ℹ️  Admin account already exists")
	}

	// Check if user already exists
	var userCount int64
	db.Model(&models.User{}).Where("email = ?", "user@example.com").Count(&userCount)

	if userCount == 0 {
		// Create sample user account
		userPassword := "user123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash user password: %v", err)
			return err
		}

		subscriptionEnd := time.Now().AddDate(0, 0, 30) // 30 days from now
		user := models.User{
			Email:           "user@example.com",
			PasswordHash:    string(hashedPassword),
			FullName:        "John Doe",
			Phone:           "+1234567890",
			Status:          "active",
			SubscriptionEnd: &subscriptionEnd,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			return err
		}
		log.Println("✅ Created user account: user@example.com / user123")
	} else {
		log.Println("ℹ️  User account already exists")
	}

	// Seed Exchange API Configurations
	if err := SeedExchangeConfigs(db); err != nil {
		log.Printf("Failed to seed exchange configs: %v", err)
		return err
	}

	// Seed sample transactions
	if err := SeedTransactions(db); err != nil {
		log.Printf("Failed to seed transactions: %v", err)
		return err
	}

	log.Println("✅ Sample data seeding completed!")
	return nil
}

// SeedExchangeConfigs creates default exchange API configurations
func SeedExchangeConfigs(db *gorm.DB) error {
	log.Println("Seeding exchange API configurations...")

	exchanges := []models.ExchangeAPIConfig{
		{
			Exchange:             "binance",
			DisplayName:          "Binance",
			SpotAPIURL:           "https://api.binance.com",
			SpotAPITestnetURL:    "https://testnet.binance.vision",
			SpotWSURL:            "wss://stream.binance.com:9443/ws",
			SpotWSTestnetURL:     "wss://stream.testnet.binance.vision/ws",
			FuturesAPIURL:        "https://fapi.binance.com",
			FuturesAPITestnetURL: "https://testnet.binancefuture.com",
			FuturesWSURL:         "wss://fstream.binance.com/ws",
			FuturesWSTestnetURL:  "wss://stream.binancefuture.com/ws",
			IsActive:             true,
			SupportSpot:          true,
			SupportFutures:       true,
			SupportMargin:        true,
			DefaultLeverage:      1,
			MaxLeverage:          125,
			MinOrderSize:         0.0001,
			MakerFee:             0.001, // 0.1%
			TakerFee:             0.001, // 0.1%
			Notes:                "World's largest cryptocurrency exchange by trading volume",
		},
		{
			Exchange:             "bingx",
			DisplayName:          "BingX",
			SpotAPIURL:           "https://open-api.bingx.com",
			SpotAPITestnetURL:    "https://open-api-vst.bingx.com",
			SpotWSURL:            "wss://open-api-ws.bingx.com/market",
			SpotWSTestnetURL:     "wss://open-api-vst.bingx.com/market",
			FuturesAPIURL:        "https://open-api.bingx.com",
			FuturesAPITestnetURL: "https://open-api-vst.bingx.com",
			FuturesWSURL:         "wss://open-api-swap.bingx.com/swap-market",
			FuturesWSTestnetURL:  "wss://open-api-vst.bingx.com/swap-market",
			IsActive:             true,
			SupportSpot:          true,
			SupportFutures:       true,
			SupportMargin:        true,
			DefaultLeverage:      1,
			MaxLeverage:          150,
			MinOrderSize:         0.0001,
			MakerFee:             0.0002, // 0.02%
			TakerFee:             0.0004, // 0.04%
			Notes:                "Global crypto exchange with copy trading and derivatives",
		},
	}

	for _, exchange := range exchanges {
		// Check if exchange already exists
		var count int64
		db.Model(&models.ExchangeAPIConfig{}).Where("exchange = ?", exchange.Exchange).Count(&count)

		if count == 0 {
			if err := db.Create(&exchange).Error; err != nil {
				log.Printf("❌ Failed to create exchange config for %s: %v", exchange.DisplayName, err)
				return err
			}
			log.Printf("✅ Created exchange config: %s", exchange.DisplayName)
		} else {
			log.Printf("ℹ️  Exchange config already exists: %s", exchange.DisplayName)
		}
	}

	log.Println("✅ Exchange API configurations seeding completed!")
	return nil
}

// SeedTransactions creates sample transaction data
func SeedTransactions(db *gorm.DB) error {
	log.Println("Seeding sample transactions...")

	// Check if transactions already exist
	var txCount int64
	db.Model(&models.Transaction{}).Count(&txCount)
	if txCount > 0 {
		log.Println("ℹ️  Transactions already exist")
		return nil
	}

	// Get existing users
	var users []models.User
	db.Find(&users)

	if len(users) == 0 {
		log.Println("⚠️  No users found, skipping transaction seeding")
		return nil
	}

	// Sample transactions for each user
	transactions := []models.Transaction{
		// User 1 transactions
		{
			UserID:      users[0].ID,
			Amount:      1000.00,
			Type:        "deposit",
			Status:      "completed",
			Description: "Initial deposit via Bank Transfer",
		},
		{
			UserID:      users[0].ID,
			Amount:      500.00,
			Type:        "deposit",
			Status:      "completed",
			Description: "Deposit via Credit Card",
		},
		{
			UserID:      users[0].ID,
			Amount:      -150.00,
			Type:        "withdrawal",
			Status:      "completed",
			Description: "Withdrawal to bank account",
		},
		{
			UserID:      users[0].ID,
			Amount:      200.00,
			Type:        "deposit",
			Status:      "pending",
			Description: "Pending deposit via PayPal",
		},
		{
			UserID:      users[0].ID,
			Amount:      -50.00,
			Type:        "payment",
			Status:      "completed",
			Description: "Subscription payment - Premium Plan",
		},
	}

	// Add more transactions if there are more users
	if len(users) > 1 {
		for i := 1; i < len(users); i++ {
			transactions = append(transactions, []models.Transaction{
				{
					UserID:      users[i].ID,
					Amount:      750.00,
					Type:        "deposit",
					Status:      "completed",
					Description: "Initial deposit",
				},
				{
					UserID:      users[i].ID,
					Amount:      -100.00,
					Type:        "withdrawal",
					Status:      "completed",
					Description: "Withdrawal",
				},
			}...)
		}
	}

	for _, tx := range transactions {
		if err := db.Create(&tx).Error; err != nil {
			log.Printf("❌ Failed to create transaction: %v", err)
			return err
		}
	}

	log.Printf("✅ Created %d sample transactions", len(transactions))
	return nil
}
