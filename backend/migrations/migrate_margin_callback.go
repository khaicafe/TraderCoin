package migrations
package main

import (
	"fmt"
	"log"
	"tradercoin/backend/config"
	"tradercoin/backend/database"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔄 Running database migration for margin_mode and callback_rate fields...")

	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Run migration
	if err := migrateMarginCallbackFields(db); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	fmt.Println("✅ Migration completed successfully!")
}

func migrateMarginCallbackFields(db *gorm.DB) error {
	fmt.Println("📝 Adding margin_mode and callback_rate columns to trading_configs...")

	// Check if columns already exist
	var columnExists bool
	
	// Check margin_mode
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='trading_configs' 
			AND column_name='margin_mode'
		);
	`).Scan(&columnExists).Error
	
	if err != nil {
		return fmt.Errorf("failed to check margin_mode column: %w", err)
	}

	if !columnExists {
		fmt.Println("   Adding margin_mode column...")
		if err := db.Exec(`
			ALTER TABLE trading_configs 
			ADD COLUMN margin_mode VARCHAR(20) DEFAULT 'ISOLATED';
		`).Error; err != nil {
			return fmt.Errorf("failed to add margin_mode column: %w", err)
		}
		fmt.Println("   ✅ margin_mode column added")
	} else {
		fmt.Println("   ℹ️  margin_mode column already exists")
	}

	// Check callback_rate
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='trading_configs' 
			AND column_name='callback_rate'
		);
	`).Scan(&columnExists).Error
	
	if err != nil {
		return fmt.Errorf("failed to check callback_rate column: %w", err)
	}

	if !columnExists {
		fmt.Println("   Adding callback_rate column...")
		if err := db.Exec(`
			ALTER TABLE trading_configs 
			ADD COLUMN callback_rate DECIMAL(10,2) DEFAULT 1.0;
		`).Error; err != nil {
			return fmt.Errorf("failed to add callback_rate column: %w", err)
		}
		fmt.Println("   ✅ callback_rate column added")
	} else {
		fmt.Println("   ℹ️  callback_rate column already exists")
	}

	// Add comments
	fmt.Println("   Adding column comments...")
	if err := db.Exec(`
		COMMENT ON COLUMN trading_configs.margin_mode IS 'Margin mode for futures trading: ISOLATED or CROSSED';
	`).Error; err != nil {
		log.Printf("⚠️  Warning: failed to add comment to margin_mode: %v", err)
	}

	if err := db.Exec(`
		COMMENT ON COLUMN trading_configs.callback_rate IS 'Callback rate for trailing stop (0.1-5%)';
	`).Error; err != nil {
		log.Printf("⚠️  Warning: failed to add comment to callback_rate: %v", err)
	}

	return nil
}
