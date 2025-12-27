# System Logs Implementation

## Overview

Implemented comprehensive system logging functionality to track all PlaceOrder operations and trading activities with detailed logging (SUCCESS, INFO, WARNING, ERROR levels).

## Changes Made

### 1. Database Model

**File**: `backend/models/models.go`

- Added `SystemLog` model with fields:
  - `ID`, `UserID`, `Level`, `Action`, `Symbol`, `Exchange`
  - `OrderID`, `Price`, `Amount`, `Message`, `Details`
  - `IPAddress`, `UserAgent`, `CreatedAt`
- Indexes on: `UserID`, `Symbol`, `OrderID`, `CreatedAt`

### 2. Utility Functions

**File**: `backend/utils/logger_db.go` (NEW)

- Created `CreateSystemLog()` function
- Accepts flexible `options` map for optional fields
- Marshals details to JSON for storage

**File**: `backend/utils/logger.go`

- Added log level constants:
  - `LogLevelSuccess = "SUCCESS"`
  - `LogLevelWarning = "WARNING"`

### 3. API Endpoints

**File**: `backend/controllers/system_log.go` (NEW)

- `GetSystemLogs(svc)` - Returns paginated logs with filters
  - Query params: `level`, `symbol`, `action`, `hours`, `page`, `limit`
- `GetSystemLogStats(svc)` - Returns log count by level
  - Query param: `hours` (default 24)
- `ClearSystemLogs(svc)` - Deletes logs older than N days
  - Query param: `days` (default 30)

### 4. Routes

**File**: `backend/routes/routes.go`

- Added logs group under `/api/v1/logs`
- All routes protected with `AuthMiddleware()`

```go
GET    /api/v1/logs              // List logs
GET    /api/v1/logs/stats        // Statistics
DELETE /api/v1/logs/clear        // Clear old logs
```

### 5. Trading Service Integration

**File**: `backend/services/trading.go`

#### Updated TradingService Struct:

```go
type TradingService struct {
    APIKey    string
    APISecret string
    Exchange  string
    DB        *gorm.DB  // Added
    UserID    uint      // Added
}
```

#### Added Logging Points:

1. **Order Initiation** (INFO):

   - When PlaceOrder starts
   - Message: "Initiating BUY MARKET order for BTCUSDT"

2. **Order Success** (SUCCESS):

   - When order is placed successfully
   - Message: "Successfully placed BUY MARKET order for BTCUSDT at $50000"
   - Includes: price, amount, order_id

3. **Order Failed** (ERROR):
   - When order placement fails
   - Message: "Failed to place BUY order for BTCUSDT: error details"
   - Includes: error code, error message

### 6. Controller Updates

**Updated Files**:

- `backend/controllers/trading.go`
- `backend/controllers/order.go`
- `backend/controllers/signal.go`
- `backend/services/order_monitor.go`

All `NewTradingService()` calls updated to pass `db` and `userID`:

```go
tradingService := services.NewTradingService(apiKey, apiSecret, config.Exchange, svc.DB, userID.(uint))
```

### 7. Database Migration

**File**: `backend/database/database.go`

- Added `&models.SystemLog{}` to auto-migration list

## Log Levels

| Level       | Usage                                                |
| ----------- | ---------------------------------------------------- |
| **SUCCESS** | Order placed successfully                            |
| **INFO**    | Order initiated, bot started                         |
| **WARNING** | High volatility detected, API rate limit approaching |
| **ERROR**   | Order failed, API error                              |

## API Usage Examples

### Get System Logs

```bash
GET /api/v1/logs?level=SUCCESS&hours=24&page=1&limit=20
```

Response:

```json
{
  "logs": [
    {
      "id": 1,
      "user_id": 1,
      "level": "SUCCESS",
      "action": "ORDER_EXECUTED",
      "symbol": "BTCUSDT",
      "exchange": "BINANCE",
      "order_id": 123456,
      "price": 50000.5,
      "amount": 0.01,
      "message": "Successfully placed BUY MARKET order for BTCUSDT at $50000.50",
      "created_at": "2025-12-27T02:40:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### Get Statistics

```bash
GET /api/v1/logs/stats?hours=24
```

Response:

```json
{
  "stats": [
    {"level": "SUCCESS", "count": 45},
    {"level": "INFO", "count": 120},
    {"level": "WARNING", "count": 8},
    {"level": "ERROR", "count": 2}
  ],
  "total": 175,
  "period_hours": 24
}
```

### Clear Old Logs

```bash
DELETE /api/v1/logs/clear?days=30
```

Response:

```json
{
  "success": true,
  "message": "Deleted 450 logs older than 30 days",
  "deleted_count": 450
}
```

## Frontend Integration (TODO)

Create a new page at `/app/logs/page.tsx` to display system logs:

1. **Stats Cards** showing SUCCESS, INFO, WARNING, ERROR counts
2. **Filter Controls** for level, symbol, time range
3. **Log Table** with columns:
   - Timestamp
   - Level (colored badge)
   - Action
   - Symbol
   - Exchange
   - Message
   - Details (expandable)
4. **Pagination** controls
5. **Export** functionality (CSV/JSON)

## Testing

1. Place an order via the frontend
2. Check backend console logs for system log creation
3. Query logs API:
   ```bash
   curl -H "Authorization: Bearer YOUR_TOKEN" \
        http://localhost:8080/api/v1/logs?limit=10
   ```
4. Verify logs are stored in database:
   ```sql
   SELECT * FROM system_logs ORDER BY created_at DESC LIMIT 10;
   ```

## Benefits

- **Comprehensive Tracking**: All trading actions logged with full context
- **Debugging**: Easy to trace order execution flow
- **Audit Trail**: Complete history of user trading activities
- **Analytics**: Statistics and patterns analysis
- **Troubleshooting**: Quick identification of errors and issues
- **User Transparency**: Users can see detailed history of their bot's actions

## Next Steps

1. ✅ Backend system logging infrastructure complete
2. ⏳ Create frontend page to display logs
3. ⏳ Add real-time log streaming via WebSocket
4. ⏳ Implement log export functionality
5. ⏳ Add email notifications for ERROR level logs
6. ⏳ Create automated cleanup job for old logs
