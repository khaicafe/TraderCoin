# Trading Mode & Additional Fields Update

## Overview

Added support for trading mode (spot/futures/margin) and other important fields to bot configuration.

## Changes Made

### 1. Backend - Model Update

**File:** `Backend/models/models.go`

Added 4 new fields to `TradingConfig` struct:

```go
Amount       float64 `gorm:"type:decimal(10,2)" json:"amount"`
TradingMode  string  `gorm:"size:20;default:'spot'" json:"trading_mode"`
APIKey       string  `gorm:"size:255" json:"-"`  // Hidden from JSON
APISecret    string  `gorm:"size:255" json:"-"`  // Hidden from JSON
```

### 2. Backend - Controller Update

**File:** `Backend/controllers/config.go`

#### CreateBotConfig

- Added `Amount`, `TradingMode`, `APIKey`, `APISecret` to input struct
- Added validation for trading_mode (must be: spot, futures, or margin)
- Default trading_mode to "spot" if not provided
- Assigned new fields to model before DB create

#### UpdateBotConfig

- Added optional fields: `Amount`, `TradingMode`, `APIKey`, `APISecret`
- Added validation for trading_mode on update
- Update logic handles all new fields properly

### 3. Frontend - Service Interface Update

**File:** `frontend/services/botConfigService.ts`

Updated interfaces:

```typescript
export interface BotConfig {
  // ... existing fields
  amount?: number;
  trading_mode?: string;
  // Note: APIKey and APISecret not returned for security
}

export interface BotConfigCreate {
  // ... existing fields
  amount?: number;
  trading_mode?: string;
  api_key?: string;
  api_secret?: string;
}

export interface BotConfigUpdate {
  // ... existing fields
  amount?: number;
  trading_mode?: string;
  api_key?: string;
  api_secret?: string;
}
```

### 4. Frontend - UI Update

**File:** `frontend/app/bot-configs/page.tsx`

#### Form State

Added to `initialFormData`:

```typescript
trading_mode: 'spot',  // Default value
```

#### Modal Form

Added new field between Exchange and Amount:

- **Trading Mode**: Dropdown select with options:
  - Spot (default)
  - Futures
  - Margin
- Help text: "Chọn loại giao dịch (Spot, Futures, hoặc Margin)"

#### Submit Handler

Updated `handleSubmit` to include:

```typescript
trading_mode: formData.trading_mode,
amount: formData.amount ? parseFloat(formData.amount) : undefined,
api_key: formData.api_key || undefined,
api_secret: formData.api_secret || undefined,
```

## Validation Rules

### Trading Mode

- **Values**: `spot`, `futures`, `margin`
- **Default**: `spot`
- **Required**: No (defaults to spot if not provided)

### Amount

- **Type**: Decimal (10,2)
- **Required**: No
- **Description**: Amount in USDT per trade

### API Credentials

- **APIKey**: Max 255 characters, hidden from JSON response
- **APISecret**: Max 255 characters, hidden from JSON response
- **Security**: Never returned in API responses

## Database Migration

The changes will auto-migrate when you restart the backend:

```bash
cd Backend
./tradercoin
```

This will add the new columns to the `trading_configs` table:

- `amount` (decimal)
- `trading_mode` (varchar(20), default 'spot')
- `api_key` (varchar(255))
- `api_secret` (varchar(255))

## Testing

### Create Bot Config

```bash
curl -X POST http://localhost:8080/api/v1/config \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My BTC Bot",
    "symbol": "BTCUSDT",
    "exchange": "binance",
    "trading_mode": "futures",
    "amount": 100.50,
    "api_key": "your_api_key",
    "api_secret": "your_api_secret",
    "stop_loss_percent": 5.0,
    "take_profit_percent": 10.0
  }'
```

### Update Bot Config

```bash
curl -X PUT http://localhost:8080/api/v1/config/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "trading_mode": "spot",
    "amount": 200.00
  }'
```

## Security Notes

1. **API Credentials**:

   - Never logged or returned in API responses
   - Stored with `json:"-"` tag to prevent JSON serialization
   - Consider encrypting in database for production

2. **Validation**:
   - Trading mode validated against allowed values
   - Exchange validated (binance/bittrex only)
   - Stop loss: 0-100%
   - Take profit: 0-1000%

## Next Steps

1. ✅ Model updated with new fields
2. ✅ Controller handles new fields
3. ✅ Frontend UI includes trading mode dropdown
4. ✅ Service interfaces updated
5. ⏳ Run backend to trigger auto-migration
6. ⏳ Test full CRUD flow with all fields
7. ⏳ Consider encryption for API credentials

## Files Changed

1. `Backend/models/models.go` - Added fields to TradingConfig
2. `Backend/controllers/config.go` - Updated Create & Update handlers
3. `frontend/services/botConfigService.ts` - Updated interfaces
4. `frontend/app/bot-configs/page.tsx` - Added trading_mode field to form

## Build Status

- ✅ Backend builds successfully
- ✅ Frontend has no TypeScript errors
- ✅ Ready for testing
