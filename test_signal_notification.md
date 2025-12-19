# Test TradingView Signal Real-time Notification

## Vấn đề hiện tại

Signal được tạo thành công (ID: 3) nhưng WebSocket notification không được gửi đến frontend.

## Nguyên nhân

WebSocket broadcast chỉ gửi đến các **connected clients**. Nếu không có user nào đang kết nối WebSocket (tức là không có ai đang mở trang signals), broadcast sẽ không có effect.

## Cách test đúng:

### Bước 1: Start Backend

```bash
cd /Users/khaicafe/Develop/TraderCoin/backend
./tradercoin
```

### Bước 2: Start Frontend

```bash
cd /Users/khaicafe/Develop/TraderCoin/frontend
npm run dev
```

### Bước 3: Mở browser và login

1. Mở http://localhost:3000/login
2. Login với: `user@example.com` / `password123`
3. Navigate đến trang **Signals** (http://localhost:3000/signals)
4. **Quan trọng**: Để trang signals mở, không đóng tab
5. Kiểm tra WebSocket status indicator - phải là màu **xanh** (CONNECTED)

### Bước 4: Send webhook trong terminal khác

```bash
curl -X POST http://localhost:8080/api/v1/signals/webhook/tradingview \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "ETHUSDT",
    "action": "BUY",
    "price": 2250.50,
    "stopLoss": 2200.00,
    "takeProfit": 2350.00,
    "strategy": "WebSocket Test",
    "message": "Testing real-time notification"
  }'
```

### Bước 5: Kiểm tra kết quả

**Backend logs sẽ show:**

```
2025/12/18 15:35:32 📡 TradingView Signal Received: BUY ETHUSDT @ 2250.50
2025/12/18 15:35:32 ✅ Signal saved with ID: 4
2025/12/18 15:35:32 📡 Broadcasted message via connection binance_1_session_xxx (1 tabs)
2025/12/18 15:35:32 ✅ Broadcast successful: 1 messages sent to 1 users
2025/12/18 15:35:32 📡 Broadcasted signal_new event (ID: 4) to all WebSocket clients
```

**Frontend browser console sẽ show:**

```
📥 New signal notification received: {signal_id: 4, symbol: "ETHUSDT", action: "BUY", ...}
WebSocket message received: {type: "signal_new", data: {...}}
```

**Frontend UI sẽ:**

1. Hiện toast notification màu xanh ở góc trên:
   ```
   🔔 Signal mới từ TradingView!
   Symbol: ETHUSDT
   Action: BUY
   ```
2. Danh sách signals tự động refresh và show signal mới
3. Toast tự động biến mất sau 5 giây

## Debug nếu không nhận được notification:

### 1. Kiểm tra WebSocket connection status

- Mở trang signals
- Xem status indicator bên cạnh nút Refresh
- Phải là **"Real-time active"** với chấm màu xanh
- Nếu là "Disconnected" (đỏ) hoặc "Connecting" (vàng) → WebSocket chưa connect

### 2. Kiểm tra browser console

```javascript
// Mở DevTools Console (F12)
// Gõ command này:
console.log('WS State:', websocketService.getConnectionState());
// Phải return: "CONNECTED"
```

### 3. Kiểm tra backend logs

```bash
# Khi frontend connect, backend phải show:
2025/12/18 15:35:33 User 1 connected via WebSocket (session: session_xxx)
```

### 4. Test với multiple tabs

- Mở 2-3 tabs cùng trang signals
- Gửi webhook
- **Tất cả tabs** đều phải nhận notification đồng thời

## Expected Flow:

```
TradingView Alert
    ↓
POST /api/v1/signals/webhook/tradingview
    ↓
Backend: Create signal in DB (ID: 4)
    ↓
Backend: wsHub.BroadcastToAll({type: "signal_new", data: {...}})
    ↓
WebSocketHub: Send to ALL connected users
    ↓
Frontend (ALL tabs): Receive WebSocket message
    ↓
Frontend: Show toast + Auto-refresh signals list
    ↓
User sees new signal immediately! ✅
```

## Common Issues:

1. **"📭 No active WebSocket connections to broadcast to"**

   - Solution: Mở trang signals trong browser trước khi send webhook

2. **WebSocket status = "DISCONNECTED"**

   - Solution: Check backend đang chạy, check token hợp lệ, reload trang

3. **Toast không hiện nhưng list có signal mới**

   - Check browser console có lỗi không
   - Check console.log có message "📥 New signal notification received" không

4. **Signal được tạo nhưng không broadcast**
   - Check backend logs có dòng "📡 Broadcasted signal_new event" không
   - Nếu có nhưng không có dòng "✅ Broadcast successful" → Không có user nào connected
