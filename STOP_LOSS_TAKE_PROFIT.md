# Stop Loss & Take Profit Implementation

## 🎯 Overview

TraderCoin automatically places **Stop Loss** and **Take Profit** orders on Binance exchange after your main order is filled. This ensures automatic risk management 24/7 without manual monitoring.

---

## 🔧 How It Works

### **Step 1: User Places Order**

User creates an order via Frontend:

```
Symbol: BTCUSDT
Side: BUY
Amount: 0.01 BTC
Bot Config: (với Stop Loss 5%, Take Profit 10%)
```

### **Step 2: Main Order Executed**

Backend places the main order on Binance:

```
POST /fapi/v1/order
{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "MARKET",
  "quantity": 0.01
}

Response:
{
  "orderId": 12345,
  "status": "FILLED",
  "avgPrice": 40000.00
}
```

### **Step 3: Calculate SL/TP Prices**

Based on filled price and config percentages:

```
Filled Price: $40,000
Stop Loss %: 5%
Take Profit %: 10%

For BUY orders:
├─ Stop Loss Price = $40,000 × (1 - 5%) = $38,000
└─ Take Profit Price = $40,000 × (1 + 10%) = $44,000

For SELL orders (SHORT):
├─ Stop Loss Price = $40,000 × (1 + 5%) = $42,000
└─ Take Profit Price = $40,000 × (1 - 10%) = $36,000
```

### **Step 4: Place Stop Loss Order on Binance**

```
POST /fapi/v1/order
{
  "symbol": "BTCUSDT",
  "side": "SELL",              // Opposite side
  "type": "STOP_MARKET",       // Market order when price hits stop
  "stopPrice": 38000.00,       // Trigger price
  "closePosition": true        // Close entire position
}

Response:
{
  "orderId": 12346,
  "status": "NEW"               // Waiting to be triggered
}
```

### **Step 5: Place Take Profit Order on Binance**

```
POST /fapi/v1/order
{
  "symbol": "BTCUSDT",
  "side": "SELL",              // Opposite side
  "type": "TAKE_PROFIT_MARKET",// Market order when price hits TP
  "stopPrice": 44000.00,       // Trigger price
  "closePosition": true        // Close entire position
}

Response:
{
  "orderId": 12347,
  "status": "NEW"               // Waiting to be triggered
}
```

### **Step 6: Binance Monitors 24/7**

Binance server automatically:

- ✅ Monitors price continuously
- ✅ Triggers Stop Loss when price ≤ $38,000
- ✅ Triggers Take Profit when price ≥ $44,000
- ✅ Executes market order immediately
- ✅ Cancels the other order (if SL triggers, TP is auto-cancelled)

---

## 📊 Example Scenarios

### **Scenario 1: Price Goes Up - Take Profit Triggered ✅**

```
T = 0:   BUY 0.01 BTC @ $40,000
         SL: $38,000 (NEW)
         TP: $44,000 (NEW)

T = 1h:  Price rises to $44,000
         TP triggered! → SELL 0.01 BTC @ $44,000
         SL automatically cancelled

Profit: ($44,000 - $40,000) × 0.01 = $40 (+10%)
```

### **Scenario 2: Price Goes Down - Stop Loss Triggered ❌**

```
T = 0:   BUY 0.01 BTC @ $40,000
         SL: $38,000 (NEW)
         TP: $44,000 (NEW)

T = 30m: Price drops to $38,000
         SL triggered! → SELL 0.01 BTC @ $38,000
         TP automatically cancelled

Loss: ($38,000 - $40,000) × 0.01 = -$20 (-5%)
```

### **Scenario 3: SHORT Position**

```
T = 0:   SELL 0.01 BTC @ $40,000
         SL: $42,000 (price goes UP - bad for SHORT)
         TP: $36,000 (price goes DOWN - good for SHORT)

T = 2h:  Price drops to $36,000
         TP triggered! → BUY 0.01 BTC @ $36,000

Profit: ($40,000 - $36,000) × 0.01 = $40 (+10%)
```

---

## 🔑 Key Features

### **1. Automatic Execution**

- No manual monitoring required
- Orders execute even if you're offline
- Binance server handles everything

### **2. Risk Management**

- Automatic stop loss protects capital
- Automatic take profit locks in gains
- One triggers → other cancels

### **3. Support for Spot & Futures**

**Futures:**

```go
type: "STOP_MARKET"
type: "TAKE_PROFIT_MARKET"
closePosition: true
```

**Spot:**

```go
type: "STOP_LOSS_LIMIT"
type: "TAKE_PROFIT_LIMIT"
quantity: specified
price: slightly offset for execution
```

---

## 🛠️ Implementation Details

### **Code Location**

**Backend:**

- `/backend/services/trading.go` - PlaceStopLossOrder(), PlaceTakeProfitOrder()
- `/backend/controllers/trading.go` - PlaceOrderDirect() calls SL/TP functions

### **Functions**

```go
// Place Stop Loss order on Binance
func (ts *TradingService) PlaceStopLossOrder(
    config *models.TradingConfig,
    symbol string,
    stopPrice float64,
    quantity float64,
    side string
) OrderResult

// Place Take Profit order on Binance
func (ts *TradingService) PlaceTakeProfitOrder(
    config *models.TradingConfig,
    symbol string,
    takeProfitPrice float64,
    quantity float64,
    side string
) OrderResult
```

### **Error Handling**

If SL/TP orders fail to place:

- ✅ Main order still succeeds
- ⚠️ Warning logged
- 📊 SL/TP prices still saved in database
- 🔔 User should manually set SL/TP on exchange

---

## 🚨 Important Notes

### **1. Exchange Support**

- ✅ **Binance**: Fully supported (Spot & Futures)
- ❌ **Bittrex**: Not supported (different order types)

### **2. Order Types**

**Futures (Recommended):**

- Uses `STOP_MARKET` and `TAKE_PROFIT_MARKET`
- `closePosition: true` automatically closes entire position
- No need to specify exact quantity

**Spot:**

- Uses `STOP_LOSS_LIMIT` and `TAKE_PROFIT_LIMIT`
- Requires exact quantity
- Slightly offset price to ensure execution

### **3. Limitations**

- **One position per symbol**: closePosition affects entire position
- **Binance only**: Other exchanges need different implementation
- **Network dependent**: Requires stable internet for order placement
- **Max orders**: Binance has limits on open stop orders

---

## 📈 Testing

### **Test Case 1: Long Position with SL/TP**

```
1. Place BUY order: BTCUSDT, 0.01, Market
2. Check Binance: Should see 3 orders
   - FILLED: BUY order
   - NEW: STOP_MARKET (Stop Loss)
   - NEW: TAKE_PROFIT_MARKET (Take Profit)
3. Wait for price to hit one
4. Check: One triggered (FILLED), other cancelled
```

### **Test Case 2: Short Position with SL/TP**

```
1. Place SELL order: BTCUSDT, 0.01, Market (Futures)
2. Check Binance: Should see 3 orders
   - FILLED: SELL order
   - NEW: STOP_MARKET (Stop Loss - BUY)
   - NEW: TAKE_PROFIT_MARKET (Take Profit - BUY)
3. Wait for price to hit one
4. Check: Position closed automatically
```

---

## 🎯 Benefits

1. **24/7 Protection**: Orders execute even when you sleep
2. **No Monitoring**: Set and forget
3. **Fast Execution**: Binance servers are faster than our backend
4. **Reliable**: No network issues between backend and exchange
5. **Professional**: Industry-standard approach

---

## 🔄 Future Enhancements

- [ ] Add Trailing Stop Loss
- [ ] Support partial take profit (multiple TP levels)
- [ ] Add Bittrex support
- [ ] Implement OCO (One-Cancels-Other) orders
- [ ] Add notification when SL/TP triggers

---

## 📞 Support

If you encounter issues:

1. Check Binance order history
2. Verify API permissions (enable Futures trading if needed)
3. Check backend logs for error messages
4. Ensure sufficient balance for orders

---

**Last Updated**: December 18, 2025
**Version**: 1.0
