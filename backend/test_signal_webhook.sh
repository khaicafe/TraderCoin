#!/bin/bash

# Test TradingView Signal Webhook with Real-time Notification

echo "==============================================="
echo "🔔 Testing TradingView Signal Webhook"
echo "==============================================="
echo ""
echo "⚠️  IMPORTANT: Before running this test, make sure:"
echo "   1. Backend is running: ./tradercoin"
echo "   2. Frontend is open in browser: http://localhost:3000/signals"
echo "   3. WebSocket status shows GREEN (CONNECTED)"
echo ""
echo "Press Enter to continue or Ctrl+C to cancel..."
read

echo ""
echo "📡 Sending test signal to TradingView webhook..."
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/signals/webhook/tradingview \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "ETHUSDT",
    "action": "BUY",
    "price": 2250.50,
    "stopLoss": 2200.00,
    "takeProfit": 2350.00,
    "strategy": "WebSocket Test",
    "message": "Testing real-time notification from script"
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

echo "HTTP Status Code: $HTTP_CODE"
echo "Response Body: $BODY"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Signal sent successfully!"
    echo ""
    echo "🔍 Check the following:"
    echo "   1. Backend logs should show:"
    echo "      - 📡 TradingView Signal Received: BUY ETHUSDT @ 2250.50"
    echo "      - ✅ Signal saved with ID: X"
    echo "      - 📡 Broadcasted message via connection..."
    echo "      - ✅ Broadcast successful: N messages sent to M users"
    echo ""
    echo "   2. Frontend browser should show:"
    echo "      - Toast notification: 🔔 Signal mới từ TradingView!"
    echo "      - Signals list auto-refreshed"
    echo "      - Console log: 📥 New signal notification received"
    echo ""
    echo "   3. If you DON'T see the notification:"
    echo "      - Check WebSocket status in UI (must be green)"
    echo "      - Check browser console for errors"
    echo "      - Make sure signals page is open (not orders or other pages)"
else
    echo "❌ Failed to send signal!"
    echo ""
    echo "Common issues:"
    echo "   - Backend not running (make sure ./tradercoin is running)"
    echo "   - Port 8080 not accessible"
    echo "   - Invalid JSON payload"
fi

echo ""
echo "==============================================="
