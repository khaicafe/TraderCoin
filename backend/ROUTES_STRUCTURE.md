# Backend API Routes Structure

## Overview

Routes organized similar to FastAPI structure with clear grouping and prefixes.

Base URL: `http://localhost:8000`

---

## 📋 Route Groups

### 1. **AUTH** - `/api/v1/auth`

Authentication and authorization endpoints

| Method | Endpoint    | Description       | Auth Required |
| ------ | ----------- | ----------------- | ------------- |
| POST   | `/register` | Register new user | No            |
| POST   | `/login`    | User login        | No            |
| POST   | `/refresh`  | Refresh JWT token | No            |

**Controller:** `controllers/auth.go`

---

### 2. **USER** - `/api/v1/user`

User profile management

| Method | Endpoint   | Description         | Auth Required |
| ------ | ---------- | ------------------- | ------------- |
| GET    | `/profile` | Get user profile    | Yes           |
| PUT    | `/profile` | Update user profile | Yes           |

**Controller:** `controllers/user.go`

---

### 3. **CONFIG** - `/api/v1/config` ⭐ NEW

System and exchange configuration

| Method | Endpoint              | Description                         | Auth Required |
| ------ | --------------------- | ----------------------------------- | ------------- |
| GET    | `/`                   | Get user config summary             | Yes           |
| PUT    | `/system`             | Update system settings (admin only) | Admin         |
| GET    | `/exchange/:exchange` | Get exchange details                | Yes           |

**Controller:** `controllers/config.go`

**Features:**

- User configuration summary (active configs, limits, status)
- System-wide settings (max stop-loss/take-profit, maintenance mode)
- Exchange-specific configuration (supported pairs, features)

---

### 4. **KEYS** - `/api/v1/keys`

Exchange API key management

| Method | Endpoint | Description           | Auth Required |
| ------ | -------- | --------------------- | ------------- |
| GET    | `/`      | Get all exchange keys | Yes           |
| POST   | `/`      | Add new exchange key  | Yes           |
| PUT    | `/:id`   | Update exchange key   | Yes           |
| DELETE | `/:id`   | Delete exchange key   | Yes           |

**Controller:** `controllers/exchange_key.go`

**Supported Exchanges:** Binance, Bittrex

---

### 5. **WEBHOOK** - `/api/v1/webhook` ⭐ NEW

Webhook handling for external integrations

| Method | Endpoint       | Description               | Auth Required |
| ------ | -------------- | ------------------------- | ------------- |
| POST   | `/binance`     | Handle Binance webhooks   | No            |
| POST   | `/tradingview` | Handle TradingView alerts | No            |
| POST   | `/price-alert` | Handle price alerts       | No            |
| GET    | `/logs`        | Get webhook logs          | Yes           |
| POST   | `/create`      | Generate webhook URL      | Yes           |

**Controller:** `controllers/webhook.go`

**Features:**

- Binance order updates, position changes
- TradingView strategy signals
- Custom price alerts
- Webhook activity tracking

---

### 6. **ORDERS** - `/api/v1/orders`

Order history and tracking

| Method | Endpoint | Description                   | Auth Required |
| ------ | -------- | ----------------------------- | ------------- |
| GET    | `/`      | Get all orders (with filters) | Yes           |
| GET    | `/:id`   | Get single order details      | Yes           |

**Controller:** `controllers/order.go`

**Query Parameters:**

- `status`: filter by order status
- `exchange`: filter by exchange
- `symbol`: filter by trading pair
- `limit`: pagination limit

---

### 7. **MONITORING** - `/api/v1/monitoring` ⭐ NEW

System monitoring and metrics

| Method | Endpoint           | Description            | Auth Required |
| ------ | ------------------ | ---------------------- | ------------- |
| GET    | `/status`          | System health status   | Yes           |
| GET    | `/metrics`         | Trading metrics        | Yes           |
| GET    | `/positions`       | Active positions       | Yes           |
| GET    | `/performance`     | Performance statistics | Yes           |
| GET    | `/bot-status`      | Bot status             | Yes           |
| GET    | `/alerts`          | Get alerts             | Yes           |
| PUT    | `/alerts/:id/read` | Mark alert as read     | Yes           |

**Controller:** `controllers/monitoring.go`

**Features:**

- System health (database, Redis status)
- Trading metrics (total orders, win rate)
- Active positions (real-time P&L)
- Performance stats (profit factor, Sharpe ratio)
- Bot status (active bots count)
- Alert management (price alerts, order fills)

---

### 8. **TRADING** - `/api/v1/trading`

Trading configuration (stop-loss, take-profit)

| Method | Endpoint       | Description             | Auth Required |
| ------ | -------------- | ----------------------- | ------------- |
| GET    | `/configs`     | Get all trading configs | Yes           |
| POST   | `/configs`     | Create new config       | Yes           |
| PUT    | `/configs/:id` | Update config           | Yes           |
| DELETE | `/configs/:id` | Delete config           | Yes           |

**Controller:** `controllers/trading_config.go`

**Validation:**

- Stop Loss: 0-100%
- Take Profit: 0-1000%

---

### 9. **BINANCE** - `/api/v1/binance`

Binance API integration

| Method | Endpoint           | Description                 | Auth Required |
| ------ | ------------------ | --------------------------- | ------------- |
| GET    | `/futures/symbols` | Get Binance Futures symbols | No            |

**Controller:** `controllers/binance.go`

**Note:** Returns USDT perpetual pairs only

---

### 10. **ADMIN** - `/api/v1/admin`

Admin management panel

| Method | Endpoint            | Description          | Auth Required |
| ------ | ------------------- | -------------------- | ------------- |
| POST   | `/login`            | Admin login          | No            |
| GET    | `/users`            | Get all users        | Admin         |
| PUT    | `/users/:id/status` | Update user status   | Admin         |
| GET    | `/transactions`     | Get all transactions | Admin         |
| GET    | `/statistics`       | Get dashboard stats  | Admin         |

**Controller:** `controllers/admin.go`

**Features:**

- User management (suspend, activate)
- Transaction history
- Dashboard statistics
- User search and filtering

---

## 🔒 Authentication

### JWT Token

All protected routes require JWT token in Authorization header:

```http
Authorization: Bearer <token>
```

### Admin Routes

Admin routes require admin JWT token with elevated permissions.

---

## 📝 Example Requests

### 1. User Login

```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 2. Get Trading Metrics

```bash
curl -X GET http://localhost:8000/api/v1/monitoring/metrics \
  -H "Authorization: Bearer <token>"
```

### 3. Create Webhook

```bash
curl -X POST http://localhost:8000/api/v1/webhook/create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "binance",
    "enabled": true
  }'
```

### 4. Get System Status

```bash
curl -X GET http://localhost:8000/api/v1/monitoring/status \
  -H "Authorization: Bearer <token>"
```

---

## 🎯 Route Organization Benefits

1. **Clear Structure**: Similar to FastAPI's `include_router()` pattern
2. **Easy Navigation**: Routes grouped by functionality
3. **Scalable**: Easy to add new routes to existing groups
4. **Maintainable**: Each group has its own controller file
5. **Self-Documenting**: Route prefixes indicate purpose

---

## 🚀 New Controllers Added

1. **config.go** (113 lines)

   - User configuration management
   - System settings (admin)
   - Exchange configuration

2. **webhook.go** (147 lines)

   - Binance webhook handling
   - TradingView alert processing
   - Price alert management
   - Webhook logging

3. **monitoring.go** (175 lines)
   - System health monitoring
   - Trading performance metrics
   - Active position tracking
   - Alert management
   - Bot status monitoring

---

## 📊 Route Count Summary

| Group      | Routes | Controller         |
| ---------- | ------ | ------------------ |
| Auth       | 3      | auth.go            |
| User       | 2      | user.go            |
| Config     | 3      | config.go ⭐       |
| Keys       | 4      | exchange_key.go    |
| Webhook    | 5      | webhook.go ⭐      |
| Orders     | 2      | order.go           |
| Monitoring | 7      | monitoring.go ⭐   |
| Trading    | 4      | trading_config.go  |
| Binance    | 1      | binance.go         |
| Admin      | 5      | admin.go           |
| **Total**  | **36** | **10 controllers** |

---

## 🔄 Migration Notes

- All existing routes remain backward compatible
- No breaking changes to existing endpoints
- New routes follow same pattern as existing ones
- Authentication middleware commented out (ready to enable)
