package database

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// SeedData creates sample admin and user accounts for testing
func SeedData(db *sql.DB) error {
	log.Println("Starting to seed sample data...")

	// Check if admin already exists
	var adminCount int
	err := db.QueryRow("SELECT COUNT(*) FROM admins WHERE email = ?", "admin@tradercoin.com").Scan(&adminCount)
	if err != nil {
		return fmt.Errorf("failed to check admin: %v", err)
	}

	if adminCount == 0 {
		// Create admin account
		adminPassword := "admin123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %v", err)
		}

		_, err = db.Exec(`
			INSERT INTO admins (email, password_hash, full_name, role)
			VALUES (?, ?, ?, ?)
		`, "admin@tradercoin.com", string(hashedPassword), "System Administrator", "admin")

		if err != nil {
			return fmt.Errorf("failed to create admin: %v", err)
		}
		log.Println("✅ Created admin account: admin@tradercoin.com / admin123")
	} else {
		log.Println("ℹ️  Admin account already exists")
	}

	// Check if user already exists
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "user@example.com").Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check user: %v", err)
	}

	if userCount == 0 {
		// Create sample user account
		userPassword := "user123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash user password: %v", err)
		}

		_, err = db.Exec(`
			INSERT INTO users (email, password_hash, full_name, phone, status, subscription_end)
			VALUES (?, ?, ?, ?, ?, datetime('now', '+30 days'))
		`, "user@example.com", string(hashedPassword), "John Doe", "+1234567890", "active")

		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}
		log.Println("✅ Created user account: user@example.com / user123")
	} else {
		log.Println("ℹ️  User account already exists")
	}

	log.Println("✅ Sample data seeding completed!")
	return nil
}
