# WebSocket Order Update Message Format

## Message Type: `order_update`

Khi có order được update (status change hoặc position info update), backend sẽ gửi message qua WebSocket với format sau:

### Message Structure

```json
{
  "type": "order_update",
  "data": {
    "order_id": 123,
    "timestamp": 1703289600,
    "symbol": "ETHUSDT",
    "side": "buy",
    "status": "filled",
    "trading_mode": "futures",
    "position": {
      "symbol": "ETHUSDT",
      "position_amt": 0.007,
      "position_side": "LONG",
      "entry_price": 3030.16,
      "mark_price": 3003.37,
      "liquidation_price": 2969.56,
      "unrealized_profit": -0.19,
      "pnl_percent": -0.93,
      "leverage": 1,
      "margin_type": "ISOLATED",
      "isolated_margin": 21.2
    }
  }
}
```

## Field Descriptions

### Base Order Data (Always Present)

| Field          | Type   | Description                                   |
| -------------- | ------ | --------------------------------------------- |
| `order_id`     | number | ID của order trong database                   |
| `timestamp`    | number | Unix timestamp khi message được gửi           |
| `symbol`       | string | Trading pair (e.g., "ETHUSDT")                |
| `side`         | string | "buy" hoặc "sell"                             |
| `status`       | string | Order status: "new", "filled", "closed", etc. |
| `trading_mode` | string | "spot" hoặc "futures"                         |

### Position Data (Only for Futures with Active Position)

| Field                        | Type   | Description                                    |
| ---------------------------- | ------ | ---------------------------------------------- |
| `position.symbol`            | string | Trading pair                                   |
| `position.position_amt`      | number | Kích thước position (âm = SHORT, dương = LONG) |
| `position.position_side`     | string | "LONG", "SHORT", hoặc "BOTH"                   |
| `position.entry_price`       | number | Giá entry của position                         |
| `position.mark_price`        | number | Giá mark hiện tại                              |
| `position.liquidation_price` | number | Giá thanh lý                                   |
| `position.unrealized_profit` | number | PnL chưa thực hiện (USDT)                      |
| `position.pnl_percent`       | number | PnL % so với entry                             |
| `position.leverage`          | number | Đòn bẩy sử dụng                                |
| `position.margin_type`       | string | "ISOLATED" hoặc "CROSS"                        |
| `position.isolated_margin`   | number | Margin cho isolated position                   |

## Client Implementation Example (React/TypeScript)

```typescript
interface OrderUpdateMessage {
  type: 'order_update';
  data: {
    order_id: number;
    timestamp: number;
    symbol: string;
    side: 'buy' | 'sell';
    status: string;
    trading_mode: 'spot' | 'futures';
    position?: {
      symbol: string;
      position_amt: number;
      position_side: 'LONG' | 'SHORT' | 'BOTH';
      entry_price: number;
      mark_price: number;
      liquidation_price: number;
      unrealized_profit: number;
      pnl_percent: number;
      leverage: number;
      margin_type: 'ISOLATED' | 'CROSS';
      isolated_margin: number;
    };
  };
}

// WebSocket handler
ws.onmessage = (event) => {
  const message: OrderUpdateMessage = JSON.parse(event.data);

  if (message.type === 'order_update') {
    const {data} = message;

    // Update order in table
    updateOrderInTable(data.order_id, {
      status: data.status,
      symbol: data.symbol,
      side: data.side,
    });

    // If has position info, update position columns
    if (data.position) {
      updatePositionInfo(data.order_id, {
        size: data.position.position_amt,
        side: data.position.position_side,
        entryPrice: data.position.entry_price,
        markPrice: data.position.mark_price,
        liqPrice: data.position.liquidation_price,
        pnl: data.position.unrealized_profit,
        pnlPercent: data.position.pnl_percent,
        leverage: data.position.leverage,
        margin: data.position.isolated_margin,
      });
    }
  }
};
```

## Update Frequency

- **Futures Orders (Active Position)**: Update mỗi 5 giây với thông tin position real-time
- **Spot Orders / Status Changes**: Chỉ update khi có thay đổi status

## Notes

- Field `position` chỉ có khi:
  - `trading_mode` = "futures"
  - Position vẫn còn active (position_amt ≠ 0)
  - Order status đang running (IsRunning = true)
- Khi position đóng hoặc order closed:
  - `status` sẽ thành "closed"
  - Field `position` sẽ không có trong message
