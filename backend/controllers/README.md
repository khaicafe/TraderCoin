# Controllers Structure

## 📁 Cấu trúc Controllers đã tách riêng

Các file controller đã được tách riêng biệt theo chức năng để dễ bảo trì và mở rộng:

```
Backend/controllers/
├── auth.go              # 🔐 Authentication (Register, Login, RefreshToken)
├── user.go              # 👤 User Management (Profile)
├── exchange_key.go      # 🔑 Exchange API Keys (Binance, Bittrex)
├── trading_config.go    # ⚙️ Trading Configurations (Stop-loss, Take-profit)
├── order.go             # 📊 Orders (History, Details)
├── admin.go             # 👨‍💼 Admin Management (Users, Transactions, Statistics)
├── binance.go           # 🌐 Binance API Integration
├── config.go            # ⚙️ System & Exchange Configuration ⭐ NEW
├── webhook.go           # 🔗 Webhook Handlers (Binance, TradingView) ⭐ NEW
├── monitoring.go        # 📈 Monitoring & Metrics ⭐ NEW
├── utils.go             # 🛠️ Shared utilities (JWT Secret)
└── trading.go.old       # 📦 Backup của file cũ
```

---

## 📋 Chi tiết từng file

### 1. **auth.go** - Authentication Controllers

**Chức năng:** Xử lý đăng ký, đăng nhập, refresh token

**Functions:**

- `Register()` - Đăng ký tài khoản user mới
- `Login()` - Đăng nhập user
- `RefreshToken()` - Làm mới JWT token

**Request Examples:**

```bash
# Register
POST /api/v1/auth/register
{
  "email": "user@example.com",
  "password": "password123",
  "full_name": "John Doe",
  "phone": "+1234567890"
}

# Login
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "password123"
}

# Refresh Token
POST /api/v1/auth/refresh
Authorization: Bearer {token}
```

---

### 2. **user.go** - User Profile Controllers

**Chức năng:** Quản lý thông tin profile user

**Functions:**

- `GetProfile()` - Lấy thông tin profile user
- `UpdateProfile()` - Cập nhật thông tin profile

**Request Examples:**

```bash
# Get Profile
GET /api/v1/user/profile
Authorization: Bearer {token}

# Update Profile
PUT /api/v1/user/profile
{
  "full_name": "John Doe Updated",
  "phone": "+9876543210"
}
```

---

### 3. **exchange_key.go** - Exchange API Keys Controllers

**Chức năng:** Quản lý API keys của các sàn giao dịch

**Functions:**

- `GetExchangeKeys()` - Lấy danh sách API keys
- `AddExchangeKey()` - Thêm API key mới
- `UpdateExchangeKey()` - Cập nhật API key
- `DeleteExchangeKey()` - Xóa API key

**Request Examples:**

```bash
# Get All Keys
GET /api/v1/keys

# Add New Key
POST /api/v1/keys
{
  "exchange": "binance",
  "api_key": "your-api-key",
  "api_secret": "your-api-secret"
}

# Update Key
PUT /api/v1/keys/:id
{
  "is_active": false
}

# Delete Key
DELETE /api/v1/keys/:id
```

**Supported Exchanges:**

- `binance` - Binance
- `bittrex` - Bittrex

---

### 4. **trading_config.go** - Trading Configuration Controllers

**Chức năng:** Quản lý cấu hình stop-loss và take-profit

**Functions:**

- `GetTradingConfigs()` - Lấy danh sách cấu hình
- `CreateTradingConfig()` - Tạo cấu hình mới
- `UpdateTradingConfig()` - Cập nhật cấu hình
- `DeleteTradingConfig()` - Xóa cấu hình

**Request Examples:**

```bash
# Get All Configs
GET /api/v1/trading/configs

# Create Config
POST /api/v1/trading/configs
{
  "exchange": "binance",
  "symbol": "BTCUSDT",
  "stop_loss_percent": 5.0,
  "take_profit_percent": 10.0
}

# Update Config
PUT /api/v1/trading/configs/:id
{
  "stop_loss_percent": 3.0,
  "is_active": true
}

# Delete Config
DELETE /api/v1/trading/configs/:id
```

**Validation:**

- Stop Loss: 0-100%
- Take Profit: 0-1000%

---

### 5. **order.go** - Orders Controllers

**Chức năng:** Xem lịch sử và chi tiết orders

**Functions:**

- `GetOrders()` - Lấy danh sách orders với filter
- `GetOrder()` - Lấy chi tiết 1 order

**Request Examples:**

```bash
# Get All Orders
GET /api/v1/orders

# Filter Orders
GET /api/v1/orders?exchange=binance&symbol=BTCUSDT&status=completed

# Get Order Detail
GET /api/v1/orders/:id
```

**Query Parameters:**

- `exchange` - Filter by exchange (binance, bittrex)
- `symbol` - Filter by symbol (BTCUSDT, ETHUSDT, etc.)
- `status` - Filter by status (pending, completed, cancelled)

---

### 6. **admin.go** - Admin Management Controllers

**Chức năng:** Quản lý users, transactions, thống kê (Admin only)

**Functions:**

- `AdminLogin()` - Đăng nhập admin
- `GetAllUsers()` - Lấy danh sách users
- `UpdateUserStatus()` - Cập nhật trạng thái user (khóa/mở)
- `GetAllTransactions()` - Lấy danh sách transactions
- `GetStatistics()` - Lấy thống kê tổng quan

**Request Examples:**

```bash
# Admin Login
POST /api/v1/admin/login
{
  "email": "admin@tradercoin.com",
  "password": "admin123"
}

# Get All Users
GET /api/v1/admin/users?status=active&search=john

# Update User Status
PUT /api/v1/admin/users/:id/status
{
  "status": "suspended"
}

# Get Transactions
GET /api/v1/admin/transactions?type=deposit&status=completed

# Get Statistics
GET /api/v1/admin/statistics
```

**User Status:**

- `active` - Tài khoản hoạt động
- `suspended` - Tài khoản bị khóa

---

### 7. **binance.go** - Binance API Integration

**Chức năng:** Tích hợp với Binance Futures API

**Functions:**

- `GetBinanceFuturesSymbols()` - Lấy danh sách symbols từ Binance

**Request Examples:**

```bash
# Get Binance Futures Symbols
GET /api/v1/binance/futures/symbols
```

**Response:**

```json
{
  "total": 200,
  "symbols": [
    {
      "symbol": "BTCUSDT",
      "pair": "BTCUSDT",
      "base_asset": "BTC",
      "quote_asset": "USDT",
      "price_precision": 2,
      "quantity_precision": 3
    }
  ]
}
```

**Filter:** Chỉ lấy PERPETUAL contracts, TRADING status, USDT quote

---

### 8. **utils.go** - Shared Utilities

**Chức năng:** Chứa các biến và hàm dùng chung

**Variables:**

- `JWTSecret` - JWT secret key (shared across all controllers)

**Usage:**

```go
import "tradercoin/backend/controllers"

// Sử dụng JWT Secret
token.SignedString(controllers.JWTSecret)
```

---

## 🔄 Migration từ file cũ

### File cũ: `trading.go` (1035 dòng)

Tất cả functions trong 1 file duy nhất

### File mới: Tách thành 8 files

- **auth.go** (203 dòng) - Authentication
- **user.go** (79 dòng) - User profile
- **exchange_key.go** (182 dòng) - Exchange keys
- **trading_config.go** (189 dòng) - Trading configs
- **order.go** (98 dòng) - Orders
- **admin.go** (259 dòng) - Admin functions
- **binance.go** (75 dòng) - Binance API
- **utils.go** (5 dòng) - Shared utilities

**Total:** 1090 dòng (chia thành 8 files)

---

## ✅ Lợi ích của cấu trúc mới

### 1. **Dễ tìm kiếm và bảo trì**

- Mỗi file có chức năng rõ ràng
- Không phải scroll qua 1000+ dòng code
- Dễ dàng locate bug

### 2. **Phân công công việc**

- Developer A làm authentication → `auth.go`
- Developer B làm admin → `admin.go`
- Không conflict khi merge code

### 3. **Testing dễ dàng hơn**

- Test riêng từng module
- Mock dependencies độc lập
- Unit test rõ ràng hơn

### 4. **Mở rộng linh hoạt**

- Thêm exchange mới → Tạo file mới (VD: `okex.go`)
- Thêm chức năng mới → Không ảnh hưởng file cũ
- Refactor từng phần

### 5. **Code Review hiệu quả**

- Review từng file nhỏ
- Dễ phát hiện lỗi
- Comment cụ thể theo chức năng

---

## 🚀 Cách sử dụng

### Build project

```bash
cd Backend
go build -o tradercoin
```

### Run server

```bash
# SQLite (Default)
./tradercoin

# PostgreSQL
DB_TYPE=postgresql DB_HOST=localhost DB_PORT=5432 \
DB_USER=tradercoin DB_PASSWORD=tradercoin123 \
DB_NAME=tradercoin_db DB_SSLMODE=disable \
./tradercoin
```

### Import trong routes

```go
import (
    "tradercoin/backend/controllers"
)

// Authentication
router.POST("/auth/register", controllers.Register(services))
router.POST("/auth/login", controllers.Login(services))

// User
router.GET("/user/profile", controllers.GetProfile(services))

// Exchange Keys
router.GET("/keys", controllers.GetExchangeKeys(services))
router.POST("/keys", controllers.AddExchangeKey(services))

// Trading Configs
router.GET("/trading/configs", controllers.GetTradingConfigs(services))

// Orders
router.GET("/orders", controllers.GetOrders(services))

// Admin
router.POST("/admin/login", controllers.AdminLogin(services))
router.GET("/admin/users", controllers.GetAllUsers(services))

// Binance
router.GET("/binance/futures/symbols", controllers.GetBinanceFuturesSymbols(services))
```

---

## 📝 Notes

- File `trading.go.old` là backup của file cũ, có thể xóa sau khi test xong
- JWT Secret hiện đang hardcode trong `utils.go`, nên move vào environment variable
- Tất cả functions đều return `gin.HandlerFunc` để dễ dàng sử dụng với Gin router
- GORM được sử dụng cho tất cả database operations
- Tất cả responses đều follow format JSON chuẩn

---

## 🔧 TODO

- [ ] Move JWT Secret to environment variable
- [ ] Add middleware authentication cho protected routes
- [ ] Add rate limiting per controller
- [ ] Add request validation middleware
- [ ] Add logging cho từng controller action
- [ ] Add Swagger documentation
- [ ] Add unit tests cho từng controller

---

**Updated:** December 16, 2025  
**Status:** ✅ Production Ready

// 🔄 Update order status from exchange before returning
for i := range orders {
order := &orders[i]

    		// Only check orders that are not in final state (new, pending, partially_filled)
    		statusLower := strings.ToLower(order.Status)
    		if statusLower == "new" || statusLower == "pending" || statusLower == "partially_filled" {
    			// Get bot config to retrieve API credentials
    			if order.BotConfigID > 0 {
    				var config models.TradingConfig
    				if err := services.DB.Where("id = ?", order.BotConfigID).First(&config).Error; err == nil {
    					// Decrypt API credentials
    					apiKey, apiSecret, err := GetDecryptedAPICredentials(&config)
    					if err == nil && order.OrderID != "" {
    						// Check order status from exchange
    						tradingService := tradingservice.NewTradingService(apiKey, apiSecret, order.Exchange)
    						statusResult := tradingService.CheckOrderStatus(&config, order.OrderID, order.Symbol)

    						if statusResult.Success {
    							// Update order in database if status changed
    							if statusResult.Status != order.Status {
    								log.Printf("🔄 Updating order %d status: %s -> %s", order.ID, order.Status, statusResult.Status)
    								order.Status = statusResult.Status

    								// Update filled price if available
    								if statusResult.AvgPrice > 0 {
    									order.FilledPrice = statusResult.AvgPrice
    								}

    								// Update filled quantity
    								order.FilledQuantity = statusResult.Filled

    								// Save to database
    								if err := services.DB.Save(order).Error; err != nil {
    									log.Printf("⚠️  Failed to update order %d: %v", order.ID, err)
    								} else {
    									log.Printf("✅ Order %d updated successfully", order.ID)
    								}
    							}
    						} else {
    							log.Printf("⚠️  Failed to check order %d status: %s", order.ID, statusResult.Error)
    						}
    					}
    				}
    			}
    		}
    	}
