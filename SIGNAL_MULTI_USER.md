# Signal Architecture - Multi-User Support

## Vấn đề cũ

**SAI THIẾT KẾ:** Signal có 1 status chung cho tất cả user

- Signal.status = "pending/executed/failed" → Chỉ 1 trạng thái chung
- User A đặt lệnh → Signal.status = "executed"
- User B, C, D không thể đặt lệnh nữa vì signal đã executed

## Kiến trúc mới - ĐÚNG

### 1. TradingSignal Table (signals)

**Chứa thông tin SIGNAL CHUNG** - Không có status cụ thể của user

```go
type TradingSignal struct {
    ID            uint
    Symbol        string      // BTCUSDT
    Action        string      // buy, sell, close
    Price         float64
    StopLoss      float64
    TakeProfit    float64
    Message       string
    Strategy      string
    WebhookPrefix string
    ReceivedAt    time.Time
    // KHÔNG CÒN: Status, ExecutedByUserID
}
```

**Đặc điểm:**

- 1 signal = 1 record trong DB
- TẤT CẢ user đều thấy signal này
- Signal KHÔNG có status - vì mỗi user có status khác nhau

### 2. UserSignalHistory Table (user_signal_histories)

**Chứa TRẠNG THÁI RIÊNG** của từng user với signal

```go
type UserSignalHistory struct {
    ID                 uint
    UserSignalConfigID uint        // Link to user config
    UserID             uint        // User nào đặt lệnh
    SignalID           string      // Link to TradingSignal
    Symbol             string
    Side               string      // BUY, SELL
    SignalPrice        float64
    Amount             float64
    Status             string      // pending, executed, failed, skipped
    OrderID            string      // Exchange order ID
    ExecutedPrice      float64
    ErrorMessage       string
    ReceivedAt         time.Time
    ExecutedAt         *time.Time
}
```

**Đặc điểm:**

- 1 signal có thể có NHIỀU records trong history (mỗi user 1 record)
- Mỗi user có status RIÊNG: pending/executed/failed/skipped
- User A executed, User B vẫn pending, User C failed

## Flow hoạt động

### 1. Signal đến từ TradingView Webhook

```
POST /api/v1/signals/webhook/abc123
{
  "symbol": "BTCUSDT",
  "action": "buy",
  "price": 42000,
  "stop_loss": 41500,
  "take_profit": 43000
}
```

**Backend xử lý:**

1. Tạo 1 record trong `TradingSignal` (signal CHUNG)
2. Broadcast WebSocket → Tất cả user nhận thông báo
3. **KHÔNG** tạo UserSignalHistory (chưa có user nào action)

### 2. User A đặt lệnh

```
POST /api/v1/signals/:id/my-execute
{
  "bot_config_id": 123
}
Headers: Authorization: Bearer {USER_A_TOKEN}
```

**Backend xử lý:**

1. Kiểm tra User A đã có history với signal này chưa
2. Đặt lệnh qua exchange
3. Tạo record trong `UserSignalHistory`:
   ```go
   {
     UserID: A,
     SignalID: signal.ID,
     Status: "executed",
     OrderID: "binance_order_789"
   }
   ```

**Kết quả:**

- Signal vẫn ở table `signals` (không đổi)
- User A có history với status = "executed"
- User B, C, D chưa có history → vẫn thấy signal là "pending"

### 3. User B đặt lệnh (sau User A)

```
POST /api/v1/signals/:id/my-execute
Headers: Authorization: Bearer {USER_B_TOKEN}
```

**Backend xử lý:**

1. Kiểm tra User B chưa có history → OK
2. Đặt lệnh qua exchange
3. Tạo record mới trong `UserSignalHistory`:
   ```go
   {
     UserID: B,
     SignalID: signal.ID,
     Status: "executed",
     OrderID: "binance_order_790"
   }
   ```

**Kết quả:**

- Signal vẫn nguyên
- User A: executed với order #789
- User B: executed với order #790
- User C, D: vẫn pending

### 4. User C đặt lệnh nhưng FAILED

```
POST /api/v1/signals/:id/my-execute
```

**Backend xử lý:**

1. Đặt lệnh qua exchange → LỖI (insufficient balance)
2. Tạo record trong `UserSignalHistory`:
   ```go
   {
     UserID: C,
     SignalID: signal.ID,
     Status: "failed",
     ErrorMessage: "Insufficient balance"
   }
   ```

**Kết quả:**

- User A: executed ✅
- User B: executed ✅
- User C: failed ❌
- User D: pending (chưa làm gì)

## API Endpoints

### GET /api/v1/signals/my-signals

**Lấy danh sách signal với STATUS RIÊNG của user hiện tại**

**Response:**

```json
{
  "signals": [
    {
      // Signal data (chung)
      "id": 1,
      "symbol": "BTCUSDT",
      "action": "buy",
      "price": 42000,

      // User-specific data
      "user_status": "executed", // Status của user hiện tại
      "user_order_id": "order_789", // Order ID của user này
      "user_executed_at": "2025-12-27 10:30:00",
      "has_user_executed": true // User đã đặt lệnh chưa
    },
    {
      "id": 2,
      "symbol": "ETHUSDT",
      "action": "sell",
      "price": 2200,

      "user_status": "pending", // User chưa đặt lệnh
      "has_user_executed": false
    }
  ]
}
```

### POST /api/v1/signals/:id/my-execute

**Đặt lệnh cho signal (chỉ cho user hiện tại)**

**Request:**

```json
{
  "bot_config_id": 123,
  "test_mode": false
}
```

**Logic:**

1. Check: User đã có history với signal này chưa?
   - Có → Return 409 "Bạn đã đặt lệnh rồi"
   - Chưa → Continue
2. Đặt lệnh qua exchange
3. Tạo UserSignalHistory
4. Return success

## Frontend Changes

### Old (SAI)

```typescript
// Signal có 1 status chung
interface TradingSignal {
  id: number;
  symbol: string;
  status: string; // ❌ SAI - chỉ 1 status cho tất cả user
  executed_by_user_id?: number;
}
```

### New (ĐÚNG)

```typescript
// Signal có status riêng cho từng user
interface SignalWithUserStatus {
  // Signal data (chung)
  id: number;
  symbol: string;
  action: string;
  price: number;

  // User-specific status
  user_status: string; // pending/executed/failed/skipped
  user_order_id: string;
  user_executed_at: string;
  user_error_message: string;
  has_user_executed: boolean;
}
```

### API Call Changes

**OLD:**

```typescript
// GET /api/v1/signals
const signals = await listSignals();
// → Trả về signal với 1 status chung ❌
```

**NEW:**

```typescript
// GET /api/v1/signals/my-signals
const signals = await listMySignals();
// → Trả về signal với status RIÊNG của user ✅
```

**OLD:**

```typescript
// POST /api/v1/signals/:id/execute
await executeSignal(signalId, {bot_config_id: 123});
// → Cập nhật signal status chung ❌
```

**NEW:**

```typescript
// POST /api/v1/signals/:id/my-execute
await executeMySignal(signalId, {bot_config_id: 123});
// → Tạo history riêng cho user ✅
```

## UI Display Logic

### Status Column

```tsx
{
  /* Hiển thị STATUS của USER hiện tại */
}
<span className={getStatusBadge(signal.user_status)}>
  {signal.user_status.toUpperCase()}
</span>;

{
  /* Nếu signal chưa được user này execute */
}
{
  !signal.has_user_executed && (
    <span className="text-yellow-600">Chưa đặt lệnh</span>
  );
}

{
  /* Nếu đã execute */
}
{
  signal.has_user_executed && (
    <div>
      <span className="text-green-600">✅ Đã đặt</span>
      <div className="text-xs">Order #{signal.user_order_id}</div>
    </div>
  );
}
```

### Actions Column

```tsx
{
  /* Chỉ hiển thị button nếu user chưa đặt lệnh */
}
{
  !signal.has_user_executed ? (
    <button onClick={() => executeSignal(signal.id)}>Đặt lệnh</button>
  ) : (
    <div className="text-gray-500">Đã đặt lúc {signal.user_executed_at}</div>
  );
}
```

## So sánh Old vs New

| Aspect            | OLD (SAI)                         | NEW (ĐÚNG)                   |
| ----------------- | --------------------------------- | ---------------------------- |
| **Signal Record** | 1 signal, 1 status                | 1 signal, KHÔNG có status    |
| **User Status**   | Lưu trong Signal.status           | Lưu trong UserSignalHistory  |
| **Multi-user**    | ❌ User A execute → signal locked | ✅ Mỗi user có history riêng |
| **Conflict**      | User B không thể đặt lệnh nữa     | User B vẫn đặt lệnh được     |
| **Tracking**      | Không biết ai đã đặt lệnh         | Track đầy đủ từng user       |

## Migration Plan

1. ✅ Tạo models: `UserSignalConfig`, `UserSignalHistory`
2. ✅ Tạo controllers: `signal_user.go`
3. ✅ Tạo routes: `/my-signals`, `/my-execute`
4. ⏳ Frontend: Update service để dùng API mới
5. ⏳ Frontend: Update UI hiển thị user status
6. ⏳ Migrate data cũ (nếu cần)

## Benefits

1. **True Multi-user**: Nhiều user cùng trade 1 signal
2. **Individual Tracking**: Biết chính xác ai đã làm gì
3. **Isolated Status**: Lỗi của User A không ảnh hưởng User B
4. **History**: Lưu vết đầy đủ từng user
5. **Analytics**: Phân tích performance từng user với signals
