package main

import (
	"fmt"
	"log"
	"tradercoin/backend/config"
	"tradercoin/backend/database"
	"tradercoin/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("🔧 Starting migration to user_signals architecture...")

	// Step 1: Create user_signals table if not exists
	log.Println("📝 Step 1: Creating user_signals table...")
	if err := db.AutoMigrate(&models.UserSignal{}); err != nil {
		log.Fatalf("Failed to create user_signals table: %v", err)
	}
	log.Println("✅ user_signals table created")

	// Step 2: Migrate existing data from trading_signals to user_signals
	log.Println("📝 Step 2: Migrating existing signal data...")

	// Get all signals with executed status
	var signals []struct {
		ID               uint
		Status           string
		OrderID          *uint
		ExecutedByUserID *uint
		ExecutedAt       *string
		ErrorMessage     string
	}

	if err := db.Table("trading_signals").
		Select("id, status, order_id, executed_by_user_id, executed_at, error_message").
		Where("status != ?", "pending").
		Or("executed_by_user_id IS NOT NULL").
		Find(&signals).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to query existing signals: %v", err)
	} else {
		log.Printf("Found %d signals with user actions to migrate", len(signals))

		// Create UserSignal records for each
		for _, signal := range signals {
			if signal.ExecutedByUserID != nil {
				userSignal := models.UserSignal{
					UserID:   *signal.ExecutedByUserID,
					SignalID: signal.ID,
					Status:   signal.Status,
					OrderID:  signal.OrderID,
					ErrorMsg: signal.ErrorMessage,
				}

				// Create the user signal record
				if err := db.Create(&userSignal).Error; err != nil {
					log.Printf("⚠️  Warning: Failed to create UserSignal for signal %d: %v", signal.ID, err)
				} else {
					log.Printf("✅ Migrated signal %d (user %d, status: %s)", signal.ID, *signal.ExecutedByUserID, signal.Status)
				}
			}
		}
	}

	// Step 3: Drop old columns from trading_signals
	log.Println("📝 Step 3: Cleaning up trading_signals table...")

	// SQLite doesn't support DROP COLUMN directly, we need to recreate the table
	log.Println("⚠️  SQLite detected - will need to recreate table")

	// Check if old columns exist
	var columnExists bool
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('trading_signals') WHERE name = 'status'").Scan(&columnExists)

	if columnExists {
		log.Println("📝 Recreating trading_signals table without user-specific columns...")

		// 1. Rename old table
		if err := db.Exec("ALTER TABLE trading_signals RENAME TO trading_signals_old").Error; err != nil {
			log.Fatalf("Failed to rename table: %v", err)
		}

		// Drop existing indexes to avoid conflicts
		db.Exec("DROP INDEX IF EXISTS idx_trading_signals_symbol")
		db.Exec("DROP INDEX IF EXISTS idx_trading_signals_status")
		db.Exec("DROP INDEX IF EXISTS idx_trading_signals_order_id")
		db.Exec("DROP INDEX IF EXISTS idx_trading_signals_executed_by_user_id")
		db.Exec("DROP INDEX IF EXISTS idx_trading_signals_webhook_prefix")
		db.Exec("DROP INDEX IF EXISTS idx_trading_signals_received_at")

		// 2. Create new table with correct schema
		if err := db.AutoMigrate(&models.TradingSignal{}); err != nil {
			log.Fatalf("Failed to create new trading_signals table: %v", err)
		}

		// 3. Copy data (only the columns we want to keep)
		if err := db.Exec(`
			INSERT INTO trading_signals 
			(id, symbol, action, price, stop_loss, take_profit, message, strategy, webhook_prefix, received_at, raw_payload, created_at, updated_at)
			SELECT 
			id, symbol, action, price, stop_loss, take_profit, message, strategy, webhook_prefix, received_at, 
			COALESCE(raw_payload, ''), created_at, updated_at
			FROM trading_signals_old
		`).Error; err != nil {
			log.Fatalf("Failed to copy data: %v", err)
		}

		// 4. Drop old table
		if err := db.Exec("DROP TABLE trading_signals_old").Error; err != nil {
			log.Fatalf("Failed to drop old table: %v", err)
		}

		log.Println("✅ Table recreated successfully")
	} else {
		log.Println("ℹ️  Table already migrated, skipping recreation")
	}

	// Step 4: Run all migrations to ensure consistency
	log.Println("📝 Step 4: Running full migration...")
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("📊 Summary:")
	fmt.Println("   - user_signals table created")
	fmt.Println("   - Existing signal data migrated")
	fmt.Println("   - trading_signals table cleaned up")
	fmt.Println("   - Database schema updated")
}
