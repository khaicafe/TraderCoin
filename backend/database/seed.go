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

	log.Println("✅ Sample data seeding completed!")
	return nil
}
