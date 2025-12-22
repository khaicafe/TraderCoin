# Client WebSocket Implementation - Order Position Updates

## Overview

Frontend client nhận realtime position updates từ backend WebSocket và cập nhật table trực tiếp mà không cần refetch toàn bộ danh sách orders.

## Changes Made

### 1. Order Interface Enhancement

**File:** `frontend/services/orderService.ts`

Added position data fields to Order interface:

```typescript
export interface Order {
  // ... existing fields

  // Position data (for futures orders)
  position?: {
    position_amt?: string;
    entry_price?: string;
    mark_price?: string;
    liquidation_price?: string;
    unrealized_profit?: string;
    pnl_percent?: string;
    leverage?: string;
    margin_type?: string;
    isolated_margin?: string;
    position_side?: string;
  };
}
```

### 2. WebSocket Message Handler

**File:** `frontend/app/orders/page.tsx`

#### Added OrderUpdateMessage Interface

```typescript
interface OrderUpdateMessage {
  type: string;
  order_id: number;
  user_id: number;
  symbol: string;
  status: string;
  order: {
    id: number;
    symbol: string;
    side: string;
    quantity: number;
    price: number;
    status: string;
    trading_mode?: string;
  };
  position?: {
    position_amt: string;
    entry_price: string;
    mark_price: string;
    liquidation_price: string;
    unrealized_profit: string;
    pnl_percent: string;
    leverage: string;
    margin_type: string;
    isolated_margin: string;
    position_side: string;
  };
  timestamp: string;
}
```

#### Enhanced onMessage Handler

**Before:**

```typescript
const unsubscribeOrderUpdates = websocketService.onMessage((message) => {
  if (message.type === 'order_update') {
    refreshOrdersLight(); // Full refetch
  }
});
```

**After:**

```typescript
const unsubscribeOrderUpdates = websocketService.onMessage((message) => {
  if (message.type === 'order_update') {
    const updateMsg = message as OrderUpdateMessage;
    console.log('📡 Received order update:', updateMsg);

    // Update specific order in state with position data
    setOrders((prevOrders) => {
      const orderIndex = prevOrders.findIndex(
        (o) => o.id === updateMsg.order_id,
      );

      if (orderIndex === -1) {
        // Order not found, do full refresh
        console.log(
          `Order ${updateMsg.order_id} not found in current list, refreshing...`,
        );
        refreshOrdersLight();
        return prevOrders;
      }

      // Update order with new data including position info
      const updatedOrders = [...prevOrders];
      updatedOrders[orderIndex] = {
        ...updatedOrders[orderIndex],
        status: updateMsg.status,
        position: updateMsg.position,
        // Update PnL from position if available
        ...(updateMsg.position && {
          pnl: parseFloat(updateMsg.position.unrealized_profit || '0'),
          pnl_percent: parseFloat(updateMsg.position.pnl_percent || '0'),
        }),
      };

      console.log(`✅ Updated order ${updateMsg.order_id} with position data`);
      return updatedOrders;
    });
  }
});
```

### 3. Table UI Enhancements

#### Added New Columns

- **Position Column**: Shows LONG/SHORT position type, amount, and leverage
- **Liq Price Column**: Shows liquidation price for futures positions

#### Table Headers

```tsx
<th>Position</th>     {/* NEW */}
<th>Liq Price</th>    {/* NEW */}
<th>SL / TP</th>
<th>Status</th>
<th>PnL</th>
<th>ROI</th>
```

#### Position Display Cell

```tsx
<td className="px-6 py-4 text-sm">
  {order.position ? (
    <div className="space-y-1">
      <div
        className={`text-xs font-semibold ${
          parseFloat(order.position.position_amt || '0') > 0
            ? 'text-green-600' // LONG
            : parseFloat(order.position.position_amt || '0') < 0
            ? 'text-red-600' // SHORT
            : 'text-gray-500'
        }`}>
        {parseFloat(order.position.position_amt || '0') > 0
          ? 'LONG'
          : parseFloat(order.position.position_amt || '0') < 0
          ? 'SHORT'
          : 'NONE'}{' '}
        {Math.abs(parseFloat(order.position.position_amt || '0'))}
      </div>
      {order.position.leverage && (
        <div className="text-xs text-purple-600 font-medium">
          {order.position.leverage}x
        </div>
      )}
    </div>
  ) : (
    <span className="text-gray-400">-</span>
  )}
</td>
```

#### Liquidation Price Cell

```tsx
<td className="px-6 py-4 text-sm">
  {order.position?.liquidation_price &&
  parseFloat(order.position.liquidation_price) > 0 ? (
    <span className="text-red-600 font-semibold text-xs">
      ${parseFloat(order.position.liquidation_price).toFixed(2)}
    </span>
  ) : (
    <span className="text-gray-400">-</span>
  )}
</td>
```

#### Enhanced PnL/ROI Display

PnL và ROI cells giờ ưu tiên hiển thị data từ position (nếu có) thay vì calculated values:

```tsx
// PnL Cell
{
  (() => {
    // Ưu tiên dùng position unrealized_profit nếu có
    const positionPnl = order.position?.unrealized_profit
      ? parseFloat(order.position.unrealized_profit)
      : null;
    const displayPnl = positionPnl !== null ? positionPnl : pnl;

    return displayPnl !== null ? (
      <span
        className={`font-semibold ${
          displayPnl >= 0 ? 'text-green-600' : 'text-red-600'
        } ${isOpen ? 'animate-pulse' : ''}`}>
        {displayPnl >= 0 ? '+' : ''}${displayPnl.toFixed(2)}
      </span>
    ) : (
      '-'
    );
  })();
}

// ROI Cell
{
  (() => {
    // Ưu tiên dùng position pnl_percent nếu có
    const positionRoi = order.position?.pnl_percent
      ? parseFloat(order.position.pnl_percent)
      : null;
    const displayRoi = positionRoi !== null ? positionRoi : roi;

    return displayRoi !== null ? (
      <span
        className={`font-semibold ${
          displayRoi >= 0 ? 'text-green-600' : 'text-red-600'
        } ${isOpen ? 'animate-pulse' : ''}`}>
        {displayRoi >= 0 ? '+' : ''}
        {displayRoi.toFixed(2)}%
      </span>
    ) : (
      '-'
    );
  })();
}
```

## Features

### Real-time Updates

- ✅ WebSocket handler cập nhật specific order trong state array
- ✅ Không refetch toàn bộ orders list (performance improvement)
- ✅ Log chi tiết để debug
- ✅ Fallback to full refresh nếu order không tìm thấy trong current list

### Position Information Display

- ✅ Position type (LONG/SHORT) với color coding
- ✅ Position amount
- ✅ Leverage
- ✅ Liquidation price
- ✅ Unrealized PnL from position data
- ✅ PnL percentage

### Visual Indicators

- 🟢 Green: LONG positions, positive PnL
- 🔴 Red: SHORT positions, negative PnL, liquidation price
- 🟣 Purple: Leverage indicator
- ⚡ Pulse animation: Active open positions

## Testing

### Console Logs

Monitor browser console for WebSocket messages:

```
📡 Received order update: {order_id: 123, position: {...}}
✅ Updated order 123 with position data
```

### Manual Testing

1. Open Orders page
2. Check WebSocket status indicator (green dot = connected)
3. Place a futures order with Stop Loss/Take Profit
4. Watch table auto-update with position info every 5 seconds
5. Verify:
   - Position column shows LONG/SHORT
   - Leverage displays correctly
   - Liquidation price appears
   - PnL updates in realtime
   - No full page refresh

## Backend Integration

Backend sends updates every 5 seconds via order_monitor service:

- Fetches Binance position info using `/fapi/v2/positionRisk`
- Includes position data in WebSocket message
- Only sends for active futures positions with `positionAmt != 0`

## Performance Benefits

**Before:**

- Full orders list refetch on every update
- Multiple API calls
- Slower UI updates

**After:**

- Direct state update for specific order
- Single WebSocket message
- Instant UI updates
- No API overhead

## Future Enhancements

- [ ] Add position side filter (LONG/SHORT/BOTH)
- [ ] Add liquidation price proximity warning
- [ ] Show margin usage
- [ ] Add position history chart
- [ ] Support multiple position modes (One-way vs Hedge)
