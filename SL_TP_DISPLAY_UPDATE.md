# Stop Loss / Take Profit Display Update

## Summary

Updated the orders page to display Stop Loss and Take Profit calculated from bot config percentages instead of reading from the order table.

## Changes Made

### 1. Backend - `Backend/controllers/order.go`

#### Modified `OrderResponse` struct:

```go
type OrderResponse struct {
    models.Order
    BotConfigName      string  `json:"bot_config_name"`
    StopLossPercent    float64 `json:"stop_loss_percent,omitempty"`
    TakeProfitPercent  float64 `json:"take_profit_percent,omitempty"`
}
```

#### Created `getBotConfigInfo()` function:

- Returns: `(botConfigName string, stopLossPercent float64, takeProfitPercent float64)`
- Fetches bot config data using `order.BotConfigID`
- Returns config name and SL/TP percentages

#### Updated all API endpoints:

- `GetOrderHistory()` - Line ~103
- `GetOrders()` - Line ~154
- `GetOrder()` - Line ~192
- `GetCompletedOrders()` - Line ~275

All endpoints now populate `StopLossPercent` and `TakeProfitPercent` in the response.

### 2. Frontend - `frontend/services/orderService.ts`

#### Updated `Order` interface:

```typescript
export interface Order {
  // ... existing fields ...
  stop_loss_percent?: number; // From bot config
  take_profit_percent?: number; // From bot config
  // ... other fields ...
}
```

### 3. Frontend - `frontend/app/orders/page.tsx`

#### Added helper functions:

**`calculateStopLossPrice(entryPrice, slPercent, side)`**

- LONG (BUY): `SL price = entry * (1 - slPercent/100)`
- SHORT (SELL): `SL price = entry * (1 + slPercent/100)`

**`calculateTakeProfitPrice(entryPrice, tpPercent, side)`**

- LONG (BUY): `TP price = entry * (1 + tpPercent/100)`
- SHORT (SELL): `TP price = entry * (1 - tpPercent/100)`

#### Updated SL/TP display column:

- Checks if `order.stop_loss_percent` and `order.take_profit_percent` exist
- Calculates SL/TP prices using entry price and percentages from bot config
- Displays calculated prices with 5 decimal precision
- Shows percentage values in parentheses

## Example Display

For a LONG order with entry price $0.15000:

- If `stop_loss_percent = 2%` → SL: $0.14700 (-2.00%)
- If `take_profit_percent = 5%` → TP: $0.15750 (+5.00%)

For a SHORT order with entry price $0.15000:

- If `stop_loss_percent = 2%` → SL: $0.15300 (-2.00%)
- If `take_profit_percent = 5%` → TP: $0.14250 (+5.00%)

## Testing

1. **Backend**: Successfully compiled with no errors
2. **Frontend**: No TypeScript errors detected
3. **Data Flow**:
   - Bot config percentages → Backend API response
   - Frontend receives percentages in order data
   - Frontend calculates and displays SL/TP prices

## Next Steps

- Test with real orders to verify calculations
- Restart backend service: `cd Backend && ./restart.sh`
- Check frontend: `cd frontend && npm run dev`
- Verify SL/TP values display correctly for both LONG and SHORT positions
