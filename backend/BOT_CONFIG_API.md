# Bot Configuration API - Updated

## Overview

Controller `config.go` đã được cập nhật để xử lý bot trading configurations theo logic tương tự FastAPI, với đầy đủ chức năng CRUD và validation.

---

## 📋 API Endpoints

Base URL: `http://localhost:8000/api/v1/config`

### 1. **Create Bot Configuration**

```http
POST /api/v1/config
```

**Authentication:** Required (JWT Token)

**Request Body:**

```json
{
  "name": "BTC Long Strategy",
  "symbol": "BTCUSDT",
  "exchange": "binance",
  "stop_loss_percent": 2.5,
  "take_profit_percent": 5.0,
  "tp_levels": [
    {"price": 45000, "percent": 50},
    {"price": 46000, "percent": 50}
  ],
  "enable_trailing": true,
  "trailing_type": "percentage",
  "trailing_percent": 1.5,
  "trading_mode": "live",
  "max_open_positions": 3,
  "enable_notifications": true
}
```

**Validation:**

- `name`: Required
- `symbol`: Required (e.g., BTCUSDT, ETHUSDT)
- `exchange`: Required, must be "binance" or "bittrex"
- `stop_loss_percent`: Required, 0-100%
- `take_profit_percent`: Required, 0-1000%

**Response (201 Created):**

```json
{
  "message": "Bot configuration created successfully",
  "config": {
    "id": 1,
    "user_id": 123,
    "symbol": "BTCUSDT",
    "exchange": "binance",
    "stop_loss_percent": 2.5,
    "take_profit_percent": 5.0,
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses:**

- `400 Bad Request`: Invalid input or validation error
- `401 Unauthorized`: No authentication token
- `404 Not Found`: User not found
- `500 Internal Server Error`: Database error

---

### 2. **List Bot Configurations**

```http
GET /api/v1/config?skip=0&limit=100
```

**Authentication:** Required

**Query Parameters:**

- `skip`: Pagination offset (default: 0)
- `limit`: Max results (default: 100)

**Response (200 OK):**

```json
{
  "configs": [
    {
      "id": 1,
      "user_id": 123,
      "symbol": "BTCUSDT",
      "exchange": "binance",
      "stop_loss_percent": 2.5,
      "take_profit_percent": 5.0,
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "user_id": 123,
      "symbol": "ETHUSDT",
      "exchange": "binance",
      "stop_loss_percent": 3.0,
      "take_profit_percent": 6.0,
      "is_active": false,
      "created_at": "2024-01-14T09:20:00Z",
      "updated_at": "2024-01-14T09:20:00Z"
    }
  ],
  "total": 2
}
```

**Features:**

- Results ordered by ID descending (newest first)
- Filters by authenticated user automatically
- Pagination support

---

### 3. **Get Single Bot Configuration**

```http
GET /api/v1/config/:id
```

**Authentication:** Required

**Path Parameters:**

- `id`: Configuration ID

**Response (200 OK):**

```json
{
  "id": 1,
  "user_id": 123,
  "symbol": "BTCUSDT",
  "exchange": "binance",
  "stop_loss_percent": 2.5,
  "take_profit_percent": 5.0,
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**

- `401 Unauthorized`: No authentication token
- `404 Not Found`: Configuration not found or doesn't belong to user

---

### 4. **Update Bot Configuration**

```http
PUT /api/v1/config/:id
```

**Authentication:** Required

**Path Parameters:**

- `id`: Configuration ID

**Request Body (all fields optional):**

```json
{
  "symbol": "ETHUSDT",
  "exchange": "binance",
  "stop_loss_percent": 3.0,
  "take_profit_percent": 6.0,
  "is_active": false
}
```

**Validation:**

- `exchange`: Must be "binance" or "bittrex"
- `stop_loss_percent`: 0-100%
- `take_profit_percent`: 0-1000%

**Response (200 OK):**

```json
{
  "message": "Bot configuration updated successfully",
  "config": {
    "id": 1,
    "user_id": 123,
    "symbol": "ETHUSDT",
    "exchange": "binance",
    "stop_loss_percent": 3.0,
    "take_profit_percent": 6.0,
    "is_active": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:45:00Z"
  }
}
```

**Features:**

- Only updates provided fields (partial update)
- Validates each field independently
- Preserves unchanged fields

---

### 5. **Delete Bot Configuration**

```http
DELETE /api/v1/config/:id
```

**Authentication:** Required

**Path Parameters:**

- `id`: Configuration ID

**Response (204 No Content):**
Empty response body

**Error Responses:**

- `400 Bad Request`: Cannot delete config with associated orders

```json
{
  "error": "Cannot delete bot config: orders are associated with this config. Please delete or reassign orders first.",
  "orders_count": 5
}
```

- `401 Unauthorized`: No authentication token
- `404 Not Found`: Configuration not found

**Safety Feature:**

- Prevents deletion if orders exist
- Returns count of associated orders
- Protects data integrity

---

## 🔧 Implementation Details

### Controller Functions

```go
// CreateBotConfig - Tạo bot configuration mới
func CreateBotConfig(services *services.Services) gin.HandlerFunc

// ListBotConfigs - Lấy danh sách tất cả bot configurations
func ListBotConfigs(services *services.Services) gin.HandlerFunc

// GetBotConfig - Lấy bot configuration cụ thể
func GetBotConfig(services *services.Services) gin.HandlerFunc

// UpdateBotConfig - Cập nhật bot configuration
func UpdateBotConfig(services *services.Services) gin.HandlerFunc

// DeleteBotConfig - Xóa bot configuration
func DeleteBotConfig(services *services.Services) gin.HandlerFunc
```

### Validation Rules

| Field               | Rule                           | Error Message                                      |
| ------------------- | ------------------------------ | -------------------------------------------------- |
| exchange            | Must be "binance" or "bittrex" | "Invalid exchange. Must be 'binance' or 'bittrex'" |
| stop_loss_percent   | 0-100                          | "Stop loss must be between 0 and 100"              |
| take_profit_percent | 0-1000                         | "Take profit must be between 0 and 1000"           |

### Authentication

All endpoints require JWT authentication token in header:

```http
Authorization: Bearer <token>
```

User ID is extracted from JWT token context.

---

## 📊 Database Schema

```sql
CREATE TABLE trading_configs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    exchange VARCHAR(50) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    stop_loss_percent DECIMAL(10,2),
    take_profit_percent DECIMAL(10,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    INDEX idx_trading_configs_user_id (user_id),
    INDEX idx_trading_configs_deleted_at (deleted_at)
);
```

---

## 🧪 Example Usage

### Create Config

```bash
curl -X POST http://localhost:8000/api/v1/config \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "BTC Strategy",
    "symbol": "BTCUSDT",
    "exchange": "binance",
    "stop_loss_percent": 2.5,
    "take_profit_percent": 5.0
  }'
```

### List Configs

```bash
curl -X GET "http://localhost:8000/api/v1/config?skip=0&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get Config

```bash
curl -X GET http://localhost:8000/api/v1/config/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update Config

```bash
curl -X PUT http://localhost:8000/api/v1/config/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stop_loss_percent": 3.0,
    "is_active": false
  }'
```

### Delete Config

```bash
curl -X DELETE http://localhost:8000/api/v1/config/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🔐 Security Features

1. **Authentication Required**: All endpoints require valid JWT token
2. **User Isolation**: Users can only access their own configurations
3. **Input Validation**: All inputs validated before processing
4. **SQL Injection Protection**: GORM prevents SQL injection
5. **Referential Integrity**: Cannot delete configs with active orders

---

## 🚀 Changes from Previous Version

### Before (Old config.go)

- Generic system configuration management
- Admin-focused endpoints
- Exchange configuration lookup

### After (New config.go)

- Bot-specific trading configurations
- User-focused CRUD operations
- Order relationship validation
- Pagination support
- Partial update capability
- Enhanced error handling
- Logging for all operations

---

## 📝 Logging

All operations are logged:

```
Bot config created: 1 - BTC Strategy for user 123
Bot config updated: 1
Bot config deleted: 1
Found 5 configs for user 123
Error listing configs: database connection failed
```

---

## ⚠️ Important Notes

1. **Soft Delete**: Uses GORM's `DeletedAt` for soft deletion
2. **Timestamps**: Automatically managed by GORM
3. **Foreign Keys**: Maintains relationship with users and orders
4. **Active by Default**: New configs have `is_active: true`
5. **Order Protection**: Cannot delete configs with associated orders

---

## 🔄 Migration Path

If upgrading from old version:

1. Update routes.go with new function names
2. Test all endpoints with authentication
3. Verify user isolation works correctly
4. Check order deletion protection
5. Update frontend to use new API structure

---

## 🎯 Next Steps

- [ ] Add batch operations (enable/disable multiple configs)
- [ ] Add filtering by exchange, symbol, active status
- [ ] Add sorting options
- [ ] Implement rate limiting
- [ ] Add comprehensive unit tests
- [ ] Add API documentation (Swagger)
