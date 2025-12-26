# Margin Mode and Callback Rate Implementation

## Summary

Added two new fields to the bot configuration system:

1. **Margin Mode** - Allows users to select ISOLATED or CROSSED margin mode for futures trading
2. **Callback Rate** - Allows users to set the callback rate for trailing stop orders (0.1-5%)

## Changes Made

### 1. Backend Model (`Backend/models/models.go`)

Added two new fields to `TradingConfig` struct:

- `MarginMode` (string, default: "ISOLATED") - For setting margin type in futures trading
- `CallbackRate` (float64, default: 1.0) - For trailing stop callback rate

```go
MarginMode          string         `gorm:"size:20;default:'ISOLATED'" json:"margin_mode"` // ISOLATED, CROSSED (for futures)
CallbackRate        float64        `gorm:"type:decimal(10,2);default:1" json:"callback_rate"` // Callback rate for trailing stop (0.1-5%)
```

### 2. Backend Controller (`Backend/controllers/config.go`)

#### CreateBotConfig Function:

- Added `MarginMode` and `CallbackRate` fields to input struct
- Added validation for margin mode (must be "ISOLATED" or "CROSSED", defaults to "ISOLATED")
- Added validation for callback rate (must be between 0.1 and 5, defaults to 1.0)
- Updated logging to include these new fields
- Updated config creation to include the new fields

#### UpdateBotConfig Function:

- Added `MarginMode`, `CallbackRate`, and `TrailingStopPercent` fields to input struct
- Added validation when updating these fields
- Margin mode validation: must be "ISOLATED" or "CROSSED"
- Callback rate validation: must be between 0.1 and 5

### 3. Backend Trading Service (`Backend/services/trading.go`)

Updated `PlaceFuturesOrder` function to use the config's `MarginMode`:

```go
marginMode := config.MarginMode
if marginMode == "" {
    marginMode = "ISOLATED" // Default to ISOLATED if not set
}
if err := ts.SetMarginType(config, symbol, marginMode); err != nil {
    fmt.Printf("⚠️  Warning: Failed to set margin type to %s: %v\n", marginMode, err)
}
```

### 4. Frontend TypeScript Interfaces (`frontend/services/botConfigService.ts`)

Updated all interfaces to include the new fields:

**BotConfig Interface:**

```typescript
margin_mode?: string;
trailing_stop_percent?: number;
callback_rate?: number;
```

**BotConfigCreate Interface:**

```typescript
margin_mode?: string;
trailing_stop_percent?: number;
callback_rate?: number;
```

**BotConfigUpdate Interface:**

```typescript
margin_mode?: string;
trailing_stop_percent?: number;
callback_rate?: number;
```

### 5. Frontend Bot Config Page (`frontend/app/bot-configs/page.tsx`)

#### Added State Management:

- Added `marginModeSearch` and `showMarginModeDropdown` states
- Added `marginModes` array with ISOLATED and CROSSED options
- Updated `initialFormData` to include `margin_mode: 'ISOLATED'` and `callback_rate: '1'`

#### Updated Form Handling:

- Updated `handleOpenModal` to populate margin_mode and callback_rate from existing bot config
- Updated `handleCloseModal` to reset margin mode dropdown state
- Updated `handleSubmit` to include margin_mode and callback_rate in API calls
- Added `filteredMarginModes` for search functionality

#### Added UI Components:

A new section that appears only when trading mode is "futures":

- **Margin Mode Dropdown**: Searchable dropdown with ISOLATED and CROSSED options
- **Callback Rate Input**: Number input with validation (0.1-5%)

```tsx
{
  formData.trading_mode === 'futures' && (
    <div className="grid grid-cols-2 gap-4">
      {/* Margin Mode Dropdown */}
      {/* Callback Rate Input */}
    </div>
  );
}
```

### 6. Database Migration Files

#### SQL Migration (`Backend/migrations/add_margin_callback_fields.sql`)

Manual SQL migration file to add the new columns:

```sql
ALTER TABLE trading_configs
ADD COLUMN IF NOT EXISTS margin_mode VARCHAR(20) DEFAULT 'ISOLATED';

ALTER TABLE trading_configs
ADD COLUMN IF NOT EXISTS callback_rate DECIMAL(10,2) DEFAULT 1.0;
```

#### Go Migration (`Backend/migrations/migrate_margin_callback.go`)

Programmatic migration script that:

- Checks if columns already exist
- Adds columns if they don't exist
- Adds column comments for documentation
- Provides detailed logging of the migration process

## Usage

### For Users:

1. Navigate to Bot Configs page
2. Create or edit a bot configuration
3. Select "Futures" as trading mode
4. New fields will appear:
   - **Margin Mode**: Choose between Isolated or Crossed
   - **Callback Rate**: Set the callback rate for trailing stop (0.1-5%)

### For Developers:

1. The database schema will be automatically updated by GORM's AutoMigrate
2. Alternatively, run the SQL migration manually:
   ```bash
   psql -U your_user -d tradercoin_db -f Backend/migrations/add_margin_callback_fields.sql
   ```
3. Or run the Go migration:
   ```bash
   cd Backend/migrations
   go run migrate_margin_callback.go
   ```

## Technical Details

### Field Specifications:

**Margin Mode:**

- Type: String (VARCHAR 20)
- Values: "ISOLATED" or "CROSSED"
- Default: "ISOLATED"
- Used for: Setting futures margin type on Binance

**Callback Rate:**

- Type: Decimal (10,2)
- Range: 0.1 - 5.0
- Default: 1.0
- Unit: Percentage (%)
- Used for: Trailing stop loss callback rate on Binance Futures

### Validation:

- Backend validates margin mode must be one of: "ISOLATED", "CROSSED"
- Backend validates callback rate must be between 0.1 and 5.0
- Frontend enforces these constraints with input validation
- Both fields are only shown/used when trading mode is "futures"

### Default Values:

- Margin Mode: "ISOLATED" (safer for beginners)
- Callback Rate: 1.0% (balanced setting)

## API Changes

### POST /api/config (Create Bot Config)

New request body fields:

```json
{
  "margin_mode": "ISOLATED",
  "callback_rate": 1.0,
  "trailing_stop_percent": 0
}
```

### PUT /api/config/:id (Update Bot Config)

New request body fields:

```json
{
  "margin_mode": "CROSSED",
  "callback_rate": 2.5,
  "trailing_stop_percent": 5.0
}
```

### Response format now includes:

```json
{
  "id": 1,
  "margin_mode": "ISOLATED",
  "callback_rate": 1.0,
  "trailing_stop_percent": 0
}
```

## Testing Checklist

- [ ] Backend compiles without errors
- [ ] Frontend compiles without errors
- [ ] Database migration runs successfully
- [ ] Can create new bot config with margin mode and callback rate
- [ ] Can update existing bot config with new fields
- [ ] Margin mode correctly applied when placing futures orders
- [ ] Validation works for invalid margin mode values
- [ ] Validation works for callback rate outside 0.1-5 range
- [ ] Fields only show when trading mode is "futures"
- [ ] Existing bot configs still work (backward compatibility)

## Notes

- These fields only affect futures trading
- Margin mode setting is attempted before placing futures orders
- If setting margin mode fails, the system logs a warning but continues (non-blocking)
- Callback rate will be used for trailing stop orders in futures trading
- Both fields have sensible defaults for safety
