# Spot vs Futures API Integration

## 📋 Tổng Quan

TraderCoin system hỗ trợ cả **Spot Trading** và **Futures Trading** trên sàn Binance. Mỗi loại trading mode sử dụng **API endpoint khác nhau** để lấy giá real-time.

## 🔗 API Endpoints

### Spot Trading

- **Testnet**: `https://testnet.binance.vision/api/v3/*`
- **Production**: `https://api.binance.com/api/v3/*`
- **Price Endpoint**: `/api/v3/ticker/24hr?symbol=BTCUSDT`

### Futures Trading

- **Testnet**: `https://testnet.binancefuture.com/fapi/v1/*`
- **Production**: `https://fapi.binance.com/fapi/v1/*`
- **Price Endpoint**: `/fapi/v1/ticker/24hr?symbol=BTCUSDT`

## 🎯 Implementation Details

### 1. Database Schema

Order model có các field quan trọng:

```typescript
interface Order {
  trading_mode?: string; // 'spot', 'futures', 'margin'
  leverage?: number; // 1-125x (chỉ cho Futures)
  symbol: string; // 'BTCUSDT', 'ETHUSDT', etc.
  // ... other fields
}
```

### 2. Frontend Price Fetching

File: `/frontend/app/orders/page.tsx`

```typescript
// Group orders by trading mode
const spotOrders = openOrders.filter(
  (o) => !o.trading_mode || o.trading_mode.toLowerCase() === 'spot',
);

const futuresOrders = openOrders.filter(
  (o) =>
    o.trading_mode?.toLowerCase() === 'futures' ||
    o.trading_mode?.toLowerCase() === 'future',
);

// Fetch Spot prices
for (const symbol of spotSymbols) {
  const response = await fetch(
    `https://testnet.binance.vision/api/v3/ticker/24hr?symbol=${symbol}`,
  );
  // Store with key: symbol (e.g., "BTCUSDT")
}

// Fetch Futures prices
for (const symbol of futuresSymbols) {
  const response = await fetch(
    `https://testnet.binancefuture.com/fapi/v1/ticker/24hr?symbol=${symbol}`,
  );
  // Store with key: symbol_FUTURES (e.g., "BTCUSDT_FUTURES")
}
```

### 3. Price Key Convention

Để phân biệt Spot và Futures của cùng 1 symbol, sử dụng naming convention:

- **Spot**: `BTCUSDT`
- **Futures**: `BTCUSDT_FUTURES`

### 4. Get Current Price Logic

```typescript
const getCurrentPriceForPnL = (order: Order): number | null => {
  const status = order.status?.toLowerCase();

  // Filled orders → use filled_price
  if (status === 'filled' || status === 'closed') {
    return order.filled_price || null;
  }

  // Open orders → use real-time price
  const isFutures =
    order.trading_mode?.toLowerCase() === 'futures' ||
    order.trading_mode?.toLowerCase() === 'future';

  const priceKey = isFutures ? `${order.symbol}_FUTURES` : order.symbol;

  // Priority: realtimePrices → current_price → null
  return realtimePrices[priceKey]?.price || order.current_price || null;
};
```

## 💰 PnL/ROI Calculations

### Spot Trading (No Leverage)

```typescript
// PnL
const pnl = (currentPrice - entryPrice) * quantity; // for BUY
const pnl = (entryPrice - currentPrice) * quantity; // for SELL

// ROI
const investment = entryPrice * quantity;
const roi = (pnl / investment) * 100;
```

### Futures Trading (With Leverage)

```typescript
const leverage = order.leverage || 1;

// PnL (MULTIPLIED by leverage)
const pnl = (currentPrice - entryPrice) * quantity * leverage; // for BUY
const pnl = (entryPrice - currentPrice) * quantity * leverage; // for SELL

// ROI (based on MARGIN, not full investment)
const margin = (entryPrice * quantity) / leverage;
const roi = (pnl / margin) * 100;
```

**⚠️ Important Notes:**

- Futures PnL is **AMPLIFIED** by leverage
- Futures ROI is calculated based on **margin** (capital used), not full position value
- Example: 10x leverage means you only need $1,000 to control a $10,000 position

## 🎨 UI Display

### Current Price Column

```tsx
// Check trading mode
const isFutures =
  order.trading_mode?.toLowerCase() === 'futures' ||
  order.trading_mode?.toLowerCase() === 'future';

const priceKey = isFutures ? `${order.symbol}_FUTURES` : order.symbol;

// Display with indicator for Futures
{
  realtimePrices[priceKey] && (
    <div>
      ${realtimePrices[priceKey].price.toFixed(5)}
      {isFutures && <span className="text-purple-600">📊</span>}
    </div>
  );
}
```

### Visual Indicators

- **📊 Purple chart emoji**: Futures order
- **🟢 Green + pulse**: Price increasing
- **🔴 Red + pulse**: Price decreasing
- **✓ Green "Filled"**: Order completed

## 🔄 Real-time Update Flow

```
1. User opens Orders page
   ↓
2. Load orders from backend
   ↓
3. Filter open orders (status = new/pending/open)
   ↓
4. Group by trading_mode:
   - spotOrders → fetch from testnet.binance.vision
   - futuresOrders → fetch from testnet.binancefuture.com
   ↓
5. Store in realtimePrices:
   - Spot: { "BTCUSDT": {...} }
   - Futures: { "BTCUSDT_FUTURES": {...} }
   ↓
6. Display in UI with correct key lookup
   ↓
7. Repeat every 5 seconds (setInterval)
```

## 🧪 Testing Checklist

- [ ] Place Spot BUY order → Verify price from `/api/v3/ticker/24hr`
- [ ] Place Spot SELL order → Verify PnL calculation (no leverage)
- [ ] Place Futures BUY order with 10x leverage → Verify price from `/fapi/v1/ticker/24hr`
- [ ] Place Futures SELL order with 5x leverage → Verify PnL multiplied by leverage
- [ ] Verify ROI calculation for Futures uses margin (investment / leverage)
- [ ] Check UI shows 📊 emoji for Futures orders
- [ ] Verify different symbols work (BTCUSDT, ETHUSDT, etc.)
- [ ] Test with mixed portfolio (some Spot + some Futures)
- [ ] Verify console logs show correct API being called
- [ ] Test error handling when API fails

## 🐛 Debugging

### Console Logs

Enable detailed logging:

```typescript
console.log(
  `📊 Fetching prices: ${spotSymbols.length} spot symbols, ${futuresSymbols.length} futures symbols`,
);
console.log(`✅ Spot ${symbol}: $${price}`);
console.log(`✅ Futures ${symbol}: $${price}`);
console.log(
  `🔍 Getting price for ${symbol} (${
    isFutures ? 'Futures' : 'Spot'
  }) - Key: ${priceKey}`,
);
```

### Common Issues

| Issue                      | Cause                                  | Solution                                                |
| -------------------------- | -------------------------------------- | ------------------------------------------------------- |
| No price for Futures order | Using wrong key (no `_FUTURES` suffix) | Check `priceKey` logic                                  |
| PnL too high on Futures    | Not dividing by leverage for ROI       | Use `margin = (entry * qty) / leverage`                 |
| API 404 error              | Using wrong endpoint                   | Verify Spot uses `/api/v3/*`, Futures uses `/fapi/v1/*` |
| Price not updating         | Filter logic excludes order            | Check `trading_mode` field in DB                        |

## 📚 Related Files

- `/frontend/app/orders/page.tsx` - Main Orders page with price fetching
- `/frontend/services/orderService.ts` - Order interface definition
- `/backend/models/models.go` - Order model with `trading_mode` field
- `/REALTIME_ORDER_FLOW.md` - Overall real-time monitoring flow

## 🚀 Future Enhancements

- [ ] Add WebSocket support for Futures API (faster updates)
- [ ] Implement batch price fetching (multiple symbols in 1 request)
- [ ] Add Futures-specific metrics (funding rate, open interest)
- [ ] Support for different Futures types (USDⓈ-M, COIN-M)
- [ ] Liquidation price calculator for Futures
- [ ] Risk management warnings based on leverage

## 📖 References

- [Binance Spot API Docs](https://binance-docs.github.io/apidocs/spot/en/)
- [Binance Futures API Docs](https://binance-docs.github.io/apidocs/futures/en/)
- [Leverage Trading Guide](https://www.binance.com/en/support/faq/360033162192)
