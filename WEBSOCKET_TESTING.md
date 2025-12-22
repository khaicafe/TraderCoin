# Testing WebSocket Order Updates

## Quick Test Steps

### 1. Start Backend

```bash
cd /Users/khaicafe/Develop/TraderCoin/Backend
go run .
```

### 2. Start Frontend

```bash
cd /Users/khaicafe/Develop/TraderCoin/frontend
npm run dev
```

### 3. Open Browser Console

- Navigate to: http://localhost:3000/orders
- Open DevTools (F12)
- Go to Console tab

### 4. Login and Check Logs

You should see:

```
Connecting to WebSocket: ws://localhost:8080/api/v1/trading/ws?token=HIDDEN&session_id=...
WebSocket connected
```

### 5. Place a Test Order

Use the frontend to place a Futures order with Stop Loss/Take Profit

### 6. Watch for Updates (every 5 seconds)

**Backend logs should show:**

```
🔍 ===== ORDER MONITOR - Checking X pending orders =====
📦 Loaded X bot configs
🔍 Order X (Futures): Status=NEW, IsRunning=true
   📊 Position Info:
      Symbol: ETHUSDT | Size: 0.007 LONG
      Entry Price: 2986.64 | Mark Price: 2986.64 | Liq.Price: 0.00
      PnL: 0.00 USDT (0.00%) | Margin: 20.91 USDT | Leverage: 1x
📤 WebSocket notification sent to user X for order Y (with position info)
```

**Frontend console should show:**

```
📡 Received order update: {type: 'order_update', data: {...}}
📦 Order update data: {order_id: 123, symbol: 'ETHUSDT', status: 'NEW', has_position: true}
✅ Updated order 123 with status=NEW, position data included
```

### 7. Verify Table Updates

Check the Orders table for:

- ✅ Position column shows "LONG 0.007" with leverage "1x"
- ✅ Liq Price shows liquidation price (if > 0)
- ✅ PnL updates from position data
- ✅ ROI shows percentage
- ✅ Values update every 5 seconds (pulse animation on active orders)

## Message Structure

**Backend sends:**

```json
{
  "type": "order_update",
  "data": {
    "order_id": 123,
    "symbol": "ETHUSDT",
    "side": "BUY",
    "status": "NEW",
    "trading_mode": "futures",
    "position": {
      "symbol": "ETHUSDT",
      "position_amt": 0.007,
      "position_side": "LONG",
      "entry_price": 2986.64,
      "mark_price": 2986.64,
      "liquidation_price": 0,
      "unrealized_profit": 0,
      "pnl_percent": 0,
      "leverage": 1,
      "margin_type": "isolated",
      "isolated_margin": 20.906448
    },
    "timestamp": 1703289968
  }
}
```

**Frontend processes:**

- Extracts `data.order_id` to find order in array
- Updates order status from `data.status`
- Converts position numbers to strings for display
- Updates PnL from `data.position.unrealized_profit`
- Triggers React state update → table re-renders

## Troubleshooting

### WebSocket Not Connecting

- Check backend is running on port 8080
- Check token in localStorage: `localStorage.getItem('token')`
- Check browser console for connection errors

### No Messages Received

- Check you have active orders (not 'closed' status)
- Wait 5 seconds for order monitor cycle
- Check backend logs for "ORDER MONITOR" messages

### Table Not Updating

- Open browser console and check for "📡 Received order update"
- If receiving messages but table not updating → check order IDs match
- If no position data → check order is Futures mode and has active position

### Position Shows "-"

- Position only shows for Futures orders
- Position must have `positionAmt != 0`
- Check backend logs for "Position Info" output
