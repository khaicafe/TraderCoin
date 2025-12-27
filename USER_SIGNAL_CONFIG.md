# User Signal Configuration

## Tổng quan

Hệ thống User Signal Config cho phép mỗi user cấu hình cách xử lý signal trading riêng biệt. Mỗi user có thể bật/tắt, tùy chỉnh tham số và theo dõi lịch sử đặt lệnh từ signal.

## Database Models

### 1. UserSignalConfig

Lưu cấu hình signal trading cho từng user.

**Fields:**

- `user_id` - ID của user (unique, mỗi user 1 config)
- `is_enabled` - Bật/tắt signal trading
- `auto_trade` - Tự động đặt lệnh khi nhận signal
- `default_amount` - Số lượng mặc định cho mỗi lệnh
- `max_daily_trades` - Giới hạn số lệnh tối đa mỗi ngày
- `allowed_symbols` - Các cặp được phép (BTCUSDT,ETHUSDT)
- `stop_loss_percent` - Stop loss mặc định (%)
- `take_profit_percent` - Take profit mặc định (%)

**Ví dụ:**

```json
{
  "id": 1,
  "user_id": 5,
  "is_enabled": true,
  "auto_trade": false,
  "default_amount": 10.0,
  "max_daily_trades": 10,
  "allowed_symbols": "BTCUSDT,ETHUSDT,BNBUSDT",
  "stop_loss_percent": 2.0,
  "take_profit_percent": 5.0,
  "created_at": "2025-12-27T10:00:00Z",
  "updated_at": "2025-12-27T10:00:00Z"
}
```

### 2. UserSignalHistory

Lưu lịch sử xử lý signal cho từng user.

**Fields:**

- `user_signal_config_id` - ID của config
- `user_id` - ID của user
- `signal_id` - ID signal từ hệ thống gốc
- `symbol` - Cặp giao dịch (BTCUSDT)
- `side` - BUY hoặc SELL
- `signal_price` - Giá từ signal
- `amount` - Số lượng đặt lệnh
- `status` - pending, executed, failed, skipped
- `order_id` - ID lệnh trên exchange
- `executed_price` - Giá thực tế thực hiện
- `error_message` - Lỗi nếu failed
- `signal_data` - Dữ liệu signal đầy đủ (JSON)
- `received_at` - Thời gian nhận signal
- `executed_at` - Thời gian thực hiện lệnh

**Status Values:**

- `pending` - Chờ xử lý
- `executed` - Đã đặt lệnh thành công
- `failed` - Thất bại (lỗi API, không đủ balance, etc)
- `skipped` - Bỏ qua (không nằm trong allowed_symbols, vượt max_daily_trades, etc)

## API Endpoints

### 1. GET /api/v1/user-signals/config

Lấy cấu hình signal của user đang đăng nhập.

**Headers:**

```
Authorization: Bearer {JWT_TOKEN}
```

**Response:**

```json
{
  "id": 1,
  "user_id": 5,
  "is_enabled": true,
  "auto_trade": false,
  "default_amount": 10.0,
  "max_daily_trades": 10,
  "allowed_symbols": "BTCUSDT,ETHUSDT",
  "stop_loss_percent": 2.0,
  "take_profit_percent": 5.0,
  "created_at": "2025-12-27T10:00:00Z",
  "updated_at": "2025-12-27T10:00:00Z"
}
```

### 2. PUT /api/v1/user-signals/config

Cập nhật cấu hình signal.

**Headers:**

```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Body:**

```json
{
  "is_enabled": true,
  "auto_trade": true,
  "default_amount": 20.0,
  "max_daily_trades": 15,
  "allowed_symbols": "BTCUSDT,ETHUSDT,BNBUSDT",
  "stop_loss_percent": 3.0,
  "take_profit_percent": 7.0
}
```

**Response:**

```json
{
  "id": 1,
  "user_id": 5,
  "is_enabled": true,
  "auto_trade": true,
  "default_amount": 20.0,
  "max_daily_trades": 15,
  "allowed_symbols": "BTCUSDT,ETHUSDT,BNBUSDT",
  "stop_loss_percent": 3.0,
  "take_profit_percent": 7.0,
  "created_at": "2025-12-27T10:00:00Z",
  "updated_at": "2025-12-27T10:30:00Z"
}
```

### 3. GET /api/v1/user-signals/history

Lấy lịch sử signal của user.

**Headers:**

```
Authorization: Bearer {JWT_TOKEN}
```

**Query Parameters:**

- `page` - Số trang (default: 1)
- `limit` - Số record mỗi trang (default: 50)
- `status` - Lọc theo status: pending, executed, failed, skipped

**Example:**

```
GET /api/v1/user-signals/history?page=1&limit=20&status=executed
```

**Response:**

```json
{
  "data": [
    {
      "id": 101,
      "user_signal_config_id": 1,
      "user_id": 5,
      "signal_id": "signal_12345",
      "symbol": "BTCUSDT",
      "side": "BUY",
      "signal_price": 42000.5,
      "amount": 10.0,
      "status": "executed",
      "order_id": "binance_order_789",
      "executed_price": 42001.2,
      "error_message": "",
      "received_at": "2025-12-27T14:00:00Z",
      "executed_at": "2025-12-27T14:00:05Z",
      "created_at": "2025-12-27T14:00:00Z"
    },
    {
      "id": 102,
      "user_signal_config_id": 1,
      "user_id": 5,
      "signal_id": "signal_12346",
      "symbol": "ETHUSDT",
      "side": "SELL",
      "signal_price": 2200.0,
      "amount": 10.0,
      "status": "failed",
      "order_id": "",
      "executed_price": 0,
      "error_message": "Insufficient balance",
      "received_at": "2025-12-27T15:00:00Z",
      "executed_at": null,
      "created_at": "2025-12-27T15:00:00Z"
    }
  ],
  "total": 25,
  "page": "1",
  "limit": "20"
}
```

### 4. GET /api/v1/user-signals/stats

Thống kê signal của user.

**Headers:**

```
Authorization: Bearer {JWT_TOKEN}
```

**Response:**

```json
{
  "total_signals": 100,
  "executed_signals": 75,
  "failed_signals": 10,
  "skipped_signals": 10,
  "pending_signals": 5
}
```

## Use Cases

### 1. User mới enable signal trading

```javascript
// 1. Lấy config hiện tại
const config = await fetch('/api/v1/user-signals/config', {
  headers: {Authorization: `Bearer ${token}`},
}).then((r) => r.json());

// 2. Enable và cấu hình
await fetch('/api/v1/user-signals/config', {
  method: 'PUT',
  headers: {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    is_enabled: true,
    auto_trade: false, // Manual confirmation
    default_amount: 10.0,
    max_daily_trades: 10,
    allowed_symbols: 'BTCUSDT,ETHUSDT',
    stop_loss_percent: 2.0,
    take_profit_percent: 5.0,
  }),
});
```

### 2. Xem lịch sử signal

```javascript
// Lấy tất cả signal đã executed
const history = await fetch('/api/v1/user-signals/history?status=executed', {
  headers: {Authorization: `Bearer ${token}`},
}).then((r) => r.json());

console.log(`Đã thực hiện ${history.total} signal thành công`);
```

### 3. Dashboard signal stats

```javascript
// Lấy thống kê
const stats = await fetch('/api/v1/user-signals/stats', {
  headers: {Authorization: `Bearer ${token}`},
}).then((r) => r.json());

const successRate = (
  (stats.executed_signals / stats.total_signals) *
  100
).toFixed(2);
console.log(`Tỷ lệ thành công: ${successRate}%`);
```

## Integration với Signal Webhook

Khi có signal mới đến từ webhook:

1. **Kiểm tra user config**: Tìm tất cả user có `is_enabled = true`
2. **Validate**: Kiểm tra `allowed_symbols`, `max_daily_trades`
3. **Create history record**: Tạo record với status `pending`
4. **Execute trade** (nếu `auto_trade = true`):
   - Đặt lệnh qua exchange API
   - Update history với `order_id`, `executed_price`, `executed_at`
   - Status = `executed` hoặc `failed`
5. **Send notification**: Gửi thông báo qua Telegram

## Frontend Integration

Ở frontend, cần tạo:

1. **Signal Settings Page** (`/app/signals/settings/page.tsx`):

   - Toggle enable/disable
   - Form cấu hình parameters
   - List allowed symbols với add/remove

2. **Signal History Page** (`/app/signals/history/page.tsx`):

   - Table hiển thị lịch sử
   - Filter theo status
   - Chi tiết từng signal

3. **Signal Stats Component**:
   - Cards hiển thị thống kê
   - Chart success rate
   - Daily/Weekly performance

## Migration

Chạy backend sẽ tự động tạo tables:

```bash
cd Backend
./backend
```

Tables sẽ được tạo:

- `user_signal_configs`
- `user_signal_histories`

## Notes

- Mỗi user chỉ có 1 config duy nhất (unique constraint trên `user_id`)
- Config được tạo tự động với giá trị mặc định khi user lần đầu GET config
- History lưu vô thời hạn để phân tích performance
- `signal_data` field lưu full JSON để debug sau này
