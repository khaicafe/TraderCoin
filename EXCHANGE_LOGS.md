# Exchange Order Logs

## 📋 Overview

Backend hiện đã được cấu hình để hiển thị chi tiết đầy đủ response từ sàn giao dịch khi đặt lệnh.

---

## 🎨 Log Format

### **1. Main Order (Lệnh Chính)**

```
🟡 MAIN ORDER - Exchange Response:
Status Code: 200
Response Body: {"orderId":12345,"symbol":"BTCUSDT","side":"BUY",...}

✅ MAIN ORDER PLACED:
   OrderID: 12345
   Symbol: BTCUSDT
   Type: MARKET
   Side: BUY
   Quantity: 0.01000000
   Filled Price: 40000.50000000
   Status: FILLED
```

### **2. Stop Loss Order**

```
🔵 STOP LOSS ORDER - Exchange Response:
Status Code: 200
Response Body: {"orderId":12346,"symbol":"BTCUSDT","type":"STOP_MARKET",...}

✅ STOP LOSS ORDER PLACED:
   OrderID: 12346
   Symbol: BTCUSDT
   Type: STOP_MARKET
   Side: SELL
   Stop Price: 38000.00000000
   Status: NEW
```

### **3. Take Profit Order**

```
🟢 TAKE PROFIT ORDER - Exchange Response:
Status Code: 200
Response Body: {"orderId":12347,"symbol":"BTCUSDT","type":"TAKE_PROFIT_MARKET",...}

✅ TAKE PROFIT ORDER PLACED:
   OrderID: 12347
   Symbol: BTCUSDT
   Type: TAKE_PROFIT_MARKET
   Side: SELL
   Take Profit Price: 44000.00000000
   Status: NEW
```

---

## ❌ Error Logs

### **Main Order Failed**

```
🟡 MAIN ORDER - Exchange Response:
Status Code: 400
Response Body: {"code":-1111,"msg":"Precision is over the maximum defined"}

❌ MAIN ORDER ERROR: Binance API error (status 400): Precision is over the maximum defined [Code: -1111]
```

### **Stop Loss Failed**

```
🔵 STOP LOSS ORDER - Exchange Response:
Status Code: 400
Response Body: {"code":-2010,"msg":"Account has insufficient balance"}

❌ STOP LOSS ERROR: Stop loss order failed (status 400): Account has insufficient balance
Error Details: map[code:-2010 msg:Account has insufficient balance]
```

### **Take Profit Failed**

```
🟢 TAKE PROFIT ORDER - Exchange Response:
Status Code: 400
Response Body: {"code":-4045,"msg":"Reach max stop order limit"}

❌ TAKE PROFIT ERROR: Take profit order failed (status 400): Reach max stop order limit
Error Details: map[code:-4045 msg:Reach max stop order limit]
```

---

## 🔍 Common Binance Error Codes

| Code  | Error                                                   | Meaning                           |
| ----- | ------------------------------------------------------- | --------------------------------- |
| -1111 | Precision is over the maximum                           | Số thập phân quá nhiều cho symbol |
| -1013 | Invalid quantity                                        | Quantity không đúng (quá nhỏ/lớn) |
| -2010 | Account has insufficient balance                        | Không đủ tiền                     |
| -2011 | Unknown order sent                                      | Order không tồn tại               |
| -4045 | Reach max stop order limit                              | Quá nhiều stop orders             |
| -1021 | Timestamp for this request is outside of the recvWindow | Thời gian máy không sync          |

---

## 📊 Example Full Log Sequence

### **Success Case:**

```bash
$ go run .

[GIN-debug] POST /api/v1/orders/place --> handler

🟡 MAIN ORDER - Exchange Response:
Status Code: 200
Response Body: {
  "orderId": 123456789,
  "symbol": "BTCUSDT",
  "status": "FILLED",
  "side": "BUY",
  "type": "MARKET",
  "origQty": "0.01000000",
  "executedQty": "0.01000000",
  "cummulativeQuoteQty": "400.50000000",
  "avgPrice": "40050.00000000",
  "fills": [
    {
      "price": "40050.00",
      "qty": "0.01000000"
    }
  ]
}

✅ MAIN ORDER PLACED:
   OrderID: 123456789
   Symbol: BTCUSDT
   Type: MARKET
   Side: BUY
   Quantity: 0.01000000
   Filled Price: 40050.00000000
   Status: FILLED

🔵 STOP LOSS ORDER - Exchange Response:
Status Code: 200
Response Body: {
  "orderId": 123456790,
  "symbol": "BTCUSDT",
  "status": "NEW",
  "side": "SELL",
  "type": "STOP_MARKET",
  "stopPrice": "38047.50"
}

✅ STOP LOSS ORDER PLACED:
   OrderID: 123456790
   Symbol: BTCUSDT
   Type: STOP_MARKET
   Side: SELL
   Stop Price: 38047.50000000
   Status: NEW

🟢 TAKE PROFIT ORDER - Exchange Response:
Status Code: 200
Response Body: {
  "orderId": 123456791,
  "symbol": "BTCUSDT",
  "status": "NEW",
  "side": "SELL",
  "type": "TAKE_PROFIT_MARKET",
  "stopPrice": "44055.00"
}

✅ TAKE PROFIT ORDER PLACED:
   OrderID: 123456791
   Symbol: BTCUSDT
   Type: TAKE_PROFIT_MARKET
   Side: SELL
   Take Profit Price: 44055.00000000
   Status: NEW

Order created successfully: ID=42, Exchange OrderID=123456789
[GIN] 200 | POST /api/v1/orders/place
```

---

## 🎯 Benefits

### **1. Debugging**

- See exact response from Binance
- Identify error codes immediately
- Track order IDs for verification

### **2. Monitoring**

- Verify orders are placed correctly
- Confirm stop prices
- Check order status

### **3. Troubleshooting**

- Quick error identification
- See raw JSON response
- Match with Binance API docs

---

## 🔧 Customization

### **Turn Off Verbose Logging**

Edit `/backend/services/trading.go`:

```go
// Comment out these lines to disable verbose logging:
// fmt.Printf("\n🟡 MAIN ORDER - Exchange Response:\n")
// fmt.Printf("Response Body: %s\n\n", string(body))
```

### **Log to File Instead of Console**

Replace `fmt.Printf` with logger:

```go
import "log"

// Instead of:
fmt.Printf("✅ MAIN ORDER PLACED:\n")

// Use:
log.Printf("✅ MAIN ORDER PLACED: OrderID=%d, Symbol=%s", orderId, symbol)
```

---

## 📝 Response Fields Explained

### **Main Order Response:**

```json
{
  "orderId": 123456789, // Binance order ID
  "symbol": "BTCUSDT", // Trading pair
  "status": "FILLED", // Order status (NEW/FILLED/CANCELED)
  "side": "BUY", // BUY or SELL
  "type": "MARKET", // Order type
  "origQty": "0.01000000", // Original quantity
  "executedQty": "0.01000000", // Filled quantity
  "avgPrice": "40050.00", // Average fill price
  "fills": [
    // Fill details
    {
      "price": "40050.00", // Fill price
      "qty": "0.01000000" // Fill quantity
    }
  ]
}
```

### **Stop Loss/Take Profit Response:**

```json
{
  "orderId": 123456790, // Binance order ID
  "symbol": "BTCUSDT", // Trading pair
  "status": "NEW", // Always NEW (waiting to trigger)
  "side": "SELL", // Closing side
  "type": "STOP_MARKET", // Order type
  "stopPrice": "38047.50" // Trigger price
}
```

---

## 🚨 Important Notes

1. **Logs contain sensitive data** - Don't share publicly
2. **API keys are NOT logged** - Only responses are shown
3. **Status Code 200 = Success** - Any other code is an error
4. **Stop orders show status NEW** - They're waiting to trigger
5. **Check Binance directly** - Use logs to find order IDs

---

## 🔗 Useful Commands

### **Filter Logs:**

```bash
# Only show order placements
go run . | grep "ORDER PLACED"

# Only show errors
go run . | grep "ERROR"

# Show exchange responses only
go run . | grep "Exchange Response"
```

### **Save Logs to File:**

```bash
go run . 2>&1 | tee server.log
```

---

**Last Updated**: December 18, 2025
