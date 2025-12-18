'use client';

import {useState, useEffect} from 'react';
import {Order, getOrderHistory} from '../../services/orderService';
import {refreshPnL} from '../../services/tradingService';
import websocketService, {
  OrderUpdate,
  PriceUpdate,
} from '../../services/websocketService';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshingPnL, setRefreshingPnL] = useState<number | null>(null);
  const [wsStatus, setWsStatus] = useState<string>('DISCONNECTED');

  // Real-time prices state - keyed by symbol
  const [realtimePrices, setRealtimePrices] = useState<{
    [key: string]: {price: number; change: number; percent: number};
  }>({});

  // Stats state
  const [stats, setStats] = useState({
    total: 0,
    filled: 0,
    New: 0,
    cancelled: 0,
  });

  // Filters state - Default status = "new" để hiển thị orders đang chờ
  const [filters, setFilters] = useState({
    symbol: '',
    status: 'new',
    side: '',
  });

  const fetchOrders = async () => {
    try {
      setLoading(true);
      const params: any = {
        limit: 100,
        offset: 0,
      };

      if (filters.symbol) params.symbol = filters.symbol;
      if (filters.status) params.status = filters.status;
      if (filters.side) params.side = filters.side;

      const data = await getOrderHistory(params);
      setOrders(data);
      console.log('Fetched orders:', data);

      // Calculate stats
      const total = data.length;
      const filled = data.filter(
        (o) =>
          o.status?.toLowerCase() === 'filled' ||
          o.status?.toLowerCase() === 'closed',
      ).length;
      const New = data.filter(
        (o) =>
          o.status?.toLowerCase() === 'open' ||
          o.status?.toLowerCase() === 'new',
      ).length;
      const cancelled = data.filter(
        (o) => o.status?.toLowerCase() === 'cancelled',
      ).length;

      setStats({total, filled, New, cancelled});
      setError(null);
    } catch (err) {
      console.error('Failed to fetch orders:', err);
      setError('Failed to load orders. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  // Lightweight refresh that doesn't toggle loading state
  const refreshOrdersLight = async () => {
    try {
      const params: any = {
        limit: 100,
        offset: 0,
      };
      if (filters.symbol) params.symbol = filters.symbol;
      if (filters.status) params.status = filters.status;
      if (filters.side) params.side = filters.side;

      const data = await getOrderHistory(params);
      setOrders(data);

      // Update stats without touching loading
      const total = data.length;
      const filled = data.filter(
        (o) =>
          o.status?.toLowerCase() === 'filled' ||
          o.status?.toLowerCase() === 'closed',
      ).length;
      const New = data.filter(
        (o) =>
          o.status?.toLowerCase() === 'open' ||
          o.status?.toLowerCase() === 'new',
      ).length;
      const cancelled = data.filter(
        (o) => o.status?.toLowerCase() === 'cancelled',
      ).length;
      setStats({total, filled, New, cancelled});
    } catch (err) {
      // ignore transient errors
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      // Initial fetch
      fetchOrders();

      // Connect to WebSocket
      websocketService.connect();

      // Update connection status periodically
      const statusInterval = setInterval(() => {
        setWsStatus(websocketService.getConnectionState());
      }, 1000);

      // ✅ NEW: Subscribe to order_update events from backend worker
      // Backend sends: { type: "order_update", data: { order_id: 123, timestamp: ... } }
      const unsubscribeOrderUpdates = websocketService.onMessage((message) => {
        if (message.type === 'order_update') {
          console.log('📥 Order update notification received:', message.data);

          // Refresh orders from API (lightweight, < 100ms)
          refreshOrdersLight();
        }
      });

      // ❌ REMOVED: Polling interval (backend worker handles it)
      // No need for 5s polling - backend worker checks and pushes updates

      // Subscribe to legacy order updates (keep for compatibility)
      const unsubscribeOrders = websocketService.onOrderUpdate(
        (update: OrderUpdate) => {
          console.log('📥 Legacy order update received:', update);

          // Update the order in the list
          setOrders((prevOrders) => {
            const existingOrder = prevOrders.find(
              (o) => o.order_id === update.order_id,
            );

            if (existingOrder) {
              // Update existing order
              return prevOrders.map((order) =>
                order.order_id === update.order_id
                  ? {
                      ...order,
                      status: update.status.toLowerCase(),
                      filled_quantity: update.executed_qty,
                      filled_price: update.executed_price,
                      current_price: update.current_price,
                    }
                  : order,
              );
            } else {
              // New order - fetch full data
              refreshOrdersLight();
              return prevOrders;
            }
          });
        },
      );

      // Cleanup
      return () => {
        unsubscribeOrderUpdates();
        unsubscribeOrders();
        clearInterval(statusInterval);
        websocketService.disconnect();
      };
    } else {
      setError('You must be logged in to view this page.');
      setLoading(false);
    }
  }, [filters]);

  /**
   * 📊 Fetch Real-time Prices từ Binance
   *
   * API Endpoints:
   * - SPOT: /api/v3/ticker/24hr
   * - FUTURES: /fapi/v1/ticker/24hr
   *
   * Logic:
   * 1. Chỉ fetch giá cho orders ĐANG MỞ (new/pending/partially_filled/open)
   * 2. Bỏ qua orders đã FILLED/CLOSED (không cần real-time)
   * 3. Phân biệt Spot vs Futures để dùng đúng endpoint
   * 4. Fetch mỗi 2 giây để cập nhật giá
   *
   * Tại sao filter?
   * - Tiết kiệm bandwidth (không fetch giá cho orders đã hoàn thành)
   * - Giảm API calls đến Binance
   * - Orders đã filled có giá cố định (filled_price), không cần real-time
   *
   * Example:
   * - 10 orders total: 5 Spot + 5 Futures
   * - 3 Spot orders đang mở → fetch từ Spot API
   * - 2 Futures orders đang mở → fetch từ Futures API
   * - 5 orders đã filled → bỏ qua
   */
  useEffect(() => {
    if (orders.length === 0) return;

    let cancelled = false;

    // 🎯 Filter: CHỈ lấy orders ĐANG MỞ
    const openOrders = orders.filter((order) => {
      const status = order.status?.toLowerCase();
      return (
        status === 'new' ||
        status === 'pending' ||
        status === 'partially_filled' ||
        status === 'open'
      );
    });

    // Nếu không có order đang mở → không cần fetch
    if (openOrders.length === 0) {
      console.log('📊 No open orders - skipping real-time price fetch');
      return;
    }

    // Group orders by trading mode (spot vs futures)
    const spotOrders = openOrders.filter(
      (o) => !o.trading_mode || o.trading_mode.toLowerCase() === 'spot',
    );
    const futuresOrders = openOrders.filter(
      (o) =>
        o.trading_mode?.toLowerCase() === 'futures' ||
        o.trading_mode?.toLowerCase() === 'future',
    );

    const spotSymbols = Array.from(new Set(spotOrders.map((o) => o.symbol)));
    const futuresSymbols = Array.from(
      new Set(futuresOrders.map((o) => o.symbol)),
    );

    console.log(
      `📊 Fetching prices: ${spotSymbols.length} spot symbols, ${futuresSymbols.length} futures symbols`,
    );

    const fetchRealtimePrices = async () => {
      // ✅ Fetch SPOT prices
      for (const symbol of spotSymbols) {
        try {
          // ✅ FIX: Sử dụng endpoint đúng /api/v3/ticker/24hr
          // Testnet: https://testnet.binance.vision
          // Production: https://api.binance.com
          const baseURL = 'https://api.binance.com';
          const response = await fetch(
            `${baseURL}/api/v3/ticker/24hr?symbol=${symbol}`,
          );

          if (!response.ok) {
            console.warn(
              `Failed to fetch SPOT price for ${symbol}: ${response.status}`,
            );
            continue;
          }

          const data = await response.json();
          if (cancelled) return;

          setRealtimePrices((prev) => ({
            ...prev,
            [symbol]: {
              price: parseFloat(data.lastPrice),
              change: parseFloat(data.priceChange),
              percent: parseFloat(data.priceChangePercent),
            },
          }));
        } catch (e) {
          console.warn(`Error fetching SPOT price for ${symbol}:`, e);
        }
      }

      // ✅ Fetch FUTURES prices
      for (const symbol of futuresSymbols) {
        try {
          // ✅ FIX: Sử dụng endpoint đúng /api/v3/ticker/24hr
          // Testnet: https://testnet.binance.vision
          // Production: https://fapi.binance.com/
          const baseURL = 'https://fapi.binance.com/';
          const response = await fetch(
            `${baseURL}/fapi/v1/ticker/24hr?symbol=${symbol}`,
          );

          if (!response.ok) {
            console.warn(
              `Failed to fetch FUTURES price for ${symbol}: ${response.status}`,
            );
            continue;
          }

          const data = await response.json();
          if (cancelled) return;

          setRealtimePrices((prev) => ({
            ...prev,
            [`${symbol}_FUTURES`]: {
              // Add suffix to differentiate
              price: parseFloat(data.lastPrice),
              change: parseFloat(data.priceChange),
              percent: parseFloat(data.priceChangePercent),
            },
          }));
        } catch (e) {
          console.warn(`Error fetching FUTURES price for ${symbol}:`, e);
        }
      }
    };

    // initial + interval
    fetchRealtimePrices();
    const interval = setInterval(fetchRealtimePrices, 2000);

    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [orders]);

  const formatSymbol = (symbol: string): string => {
    if (symbol.endsWith('USDT')) {
      return symbol.replace('USDT', '/USDT');
    }
    return symbol;
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleString('vi-VN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  // Calculate percent distance from entry to a target (SL/TP)
  const calcTargetPercent = (
    target: number | undefined | null,
    entry: number | undefined | null,
    side: string,
  ): number | null => {
    if (!target || !entry || entry === 0) return null;
    const base = ((target - entry) / entry) * 100;
    // For SELL orders, invert so that favorable TP is positive and SL is negative
    return side?.toLowerCase() === 'sell' ? -base : base;
  };

  /**
   * 💰 Calculate Real-time PnL (Profit and Loss)
   *
   * ============================================================================
   * 📖 CÔNG THỨC TÍNH PnL
   * ============================================================================
   *
   * PnL (Profit and Loss) = Lợi nhuận hoặc lỗ của giao dịch
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  LỆNH MUA (BUY ORDER)                                               │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │  PnL = (Giá Hiện Tại - Giá Vào) × Số Lượng                        │
   * │  PnL = (Current Price - Entry Price) × Quantity                     │
   * │                                                                      │
   * │  Logic:                                                             │
   * │  - Mua BTC ở giá thấp                                              │
   * │  - Giá tăng → Profit (PnL > 0) ✅                                   │
   * │  - Giá giảm → Loss (PnL < 0) ❌                                     │
   * │                                                                      │
   * │  Ví dụ:                                                             │
   * │  - Mua 0.1 BTC ở $40,000 (Entry)                                   │
   * │  - Giá hiện tại: $42,000                                           │
   * │  - PnL = (42000 - 40000) × 0.1 = 2000 × 0.1 = $200 (Lãi) ✅       │
   * │                                                                      │
   * │  - Mua 0.1 BTC ở $40,000 (Entry)                                   │
   * │  - Giá hiện tại: $38,000                                           │
   * │  - PnL = (38000 - 40000) × 0.1 = -2000 × 0.1 = -$200 (Lỗ) ❌      │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  LỆNH BÁN (SELL ORDER)                                              │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │  PnL = (Giá Vào - Giá Hiện Tại) × Số Lượng                        │
   * │  PnL = (Entry Price - Current Price) × Quantity                     │
   * │                                                                      │
   * │  Logic:                                                             │
   * │  - Bán BTC ở giá cao (short)                                       │
   * │  - Giá giảm → Profit (PnL > 0) ✅                                   │
   * │  - Giá tăng → Loss (PnL < 0) ❌                                     │
   * │                                                                      │
   * │  Ví dụ:                                                             │
   * │  - Bán (Short) 0.1 BTC ở $42,000 (Entry)                          │
   * │  - Giá hiện tại: $40,000 (giảm)                                    │
   * │  - PnL = (42000 - 40000) × 0.1 = 2000 × 0.1 = $200 (Lãi) ✅       │
   * │                                                                      │
   * │  - Bán (Short) 0.1 BTC ở $42,000 (Entry)                          │
   * │  - Giá hiện tại: $44,000 (tăng)                                    │
   * │  - PnL = (42000 - 44000) × 0.1 = -2000 × 0.1 = -$200 (Lỗ) ❌      │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ============================================================================
   * 📊 CASE STUDIES
   * ============================================================================
   *
   * Case 1: BUY BTC - Thị trường tăng giá (Bull Market)
   * ────────────────────────────────────────────────────
   * Entry Price:    $40,000
   * Current Price:  $45,000
   * Quantity:       0.5 BTC
   * Side:           BUY
   *
   * Calculation:
   * PnL = (45000 - 40000) × 0.5
   *     = 5000 × 0.5
   *     = $2,500 ✅ (Profit)
   *
   * Investment = 40000 × 0.5 = $20,000
   * Return: +$2,500 trên vốn $20,000
   *
   * ────────────────────────────────────────────────────
   * Case 2: BUY BTC - Thị trường giảm giá (Bear Market)
   * ────────────────────────────────────────────────────
   * Entry Price:    $40,000
   * Current Price:  $35,000
   * Quantity:       0.5 BTC
   * Side:           BUY
   *
   * Calculation:
   * PnL = (35000 - 40000) × 0.5
   *     = -5000 × 0.5
   *     = -$2,500 ❌ (Loss)
   *
   * Investment = 40000 × 0.5 = $20,000
   * Loss: -$2,500 trên vốn $20,000
   *
   * ────────────────────────────────────────────────────
   * Case 3: SELL (SHORT) BTC - Giá giảm (Profitable Short)
   * ────────────────────────────────────────────────────
   * Entry Price:    $45,000
   * Current Price:  $40,000
   * Quantity:       0.5 BTC
   * Side:           SELL
   *
   * Calculation:
   * PnL = (45000 - 40000) × 0.5
   *     = 5000 × 0.5
   *     = $2,500 ✅ (Profit - giá giảm như dự đoán)
   *
   * Logic: Short ở $45k, giá giảm xuống $40k
   * → Lãi $5k/BTC × 0.5 BTC = $2,500
   *
   * ────────────────────────────────────────────────────
   * Case 4: SELL (SHORT) BTC - Giá tăng (Loss)
   * ────────────────────────────────────────────────────
   * Entry Price:    $40,000
   * Current Price:  $45,000
   * Quantity:       0.5 BTC
   * Side:           SELL
   *
   * Calculation:
   * PnL = (40000 - 45000) × 0.5
   *     = -5000 × 0.5
   *     = -$2,500 ❌ (Loss - giá tăng ngược dự đoán)
   *
   * Logic: Short ở $40k, giá tăng lên $45k
   * → Lỗ $5k/BTC × 0.5 BTC = -$2,500
   *
   * ============================================================================
   * 🔑 KEY POINTS
   * ============================================================================
   *
   * 1. Entry Price (Giá Vào):
   *    - Ưu tiên: filled_price (giá khớp thực tế)
   *    - Fallback: price (giá đặt lệnh)
   *
   * 2. Current Price (Giá Hiện Tại):
   *    - Real-time từ Binance API (cập nhật mỗi 2s)
   *    - Fallback: DB price (cập nhật mỗi 5s)
   *
   * 3. Quantity (Số Lượng):
   *    - Số lượng BTC/crypto đã mua/bán
   *
   * 4. PnL = 0 khi:
   *    - Current Price = Entry Price (giá không đổi)
   *
   * 5. PnL > 0 (Profit):
   *    - BUY: Current > Entry (giá tăng)
   *    - SELL: Entry > Current (giá giảm)
   *
   * 6. PnL < 0 (Loss):
   *    - BUY: Current < Entry (giá giảm)
   *    - SELL: Entry < Current (giá tăng)
   *
   * ============================================================================
   *
   * @param order - Order object chứa thông tin giao dịch
   * @param currentPrice - Giá hiện tại của crypto
   * @returns PnL value in USDT (hoặc null nếu không tính được)
   */
  const calculatePnL = (
    order: Order,
    currentPrice: number | null,
  ): number | null => {
    if (!currentPrice || !order.quantity) return null;

    // Get entry price (filled_price > price > null)
    const entryPrice = order.filled_price || order.price;
    if (!entryPrice || entryPrice === 0) return null;

    const quantity = order.quantity;
    const side = order.side?.toLowerCase();

    if (side === 'buy') {
      // BUY: Profit when price increases
      return (currentPrice - entryPrice) * quantity;
    } else if (side === 'sell') {
      // SELL: Profit when price decreases
      return (entryPrice - currentPrice) * quantity;
    }

    return null;
  };

  /**
   * 📊 Calculate ROI (Return on Investment) in percentage
   *
   * ============================================================================
   * 📖 CÔNG THỨC TÍNH ROI
   * ============================================================================
   *
   * ROI (Return on Investment) = Tỷ suất lợi nhuận trên vốn đầu tư
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  CÔNG THỨC                                                          │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  ROI% = (PnL / Investment) × 100                                    │
   * │                                                                      │
   * │  Trong đó:                                                          │
   * │  - PnL (Profit and Loss) = Lợi nhuận hoặc lỗ (tính từ hàm trên)  │
   * │  - Investment = Vốn đầu tư ban đầu                                 │
   * │  - Investment = Entry Price × Quantity                              │
   * │                                                                      │
   * │  ROI > 0 → Lãi (màu xanh) ✅                                        │
   * │  ROI < 0 → Lỗ (màu đỏ) ❌                                           │
   * │  ROI = 0 → Hòa vốn (không lãi, không lỗ) ⚪                          │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ============================================================================
   * 📊 CASE STUDIES
   * ============================================================================
   *
   * Case 1: BUY BTC - Lãi 5%
   * ────────────────────────────────────────────────────
   * Entry Price:    $40,000
   * Current Price:  $42,000
   * Quantity:       0.1 BTC
   * Side:           BUY
   *
   * Bước 1 - Tính Investment:
   * Investment = Entry Price × Quantity
   *            = 40,000 × 0.1
   *            = $4,000 (Vốn bỏ ra)
   *
   * Bước 2 - Tính PnL:
   * PnL = (Current - Entry) × Quantity
   *     = (42,000 - 40,000) × 0.1
   *     = 2,000 × 0.1
   *     = $200 (Lãi)
   *
   * Bước 3 - Tính ROI:
   * ROI% = (PnL / Investment) × 100
   *      = (200 / 4,000) × 100
   *      = 0.05 × 100
   *      = 5% ✅
   *
   * Ý nghĩa: Đầu tư $4,000, lãi $200 → Lãi 5%
   *
   * ────────────────────────────────────────────────────
   * Case 2: BUY BTC - Lỗ 12.5%
   * ────────────────────────────────────────────────────
   * Entry Price:    $40,000
   * Current Price:  $35,000
   * Quantity:       0.5 BTC
   * Side:           BUY
   *
   * Bước 1 - Tính Investment:
   * Investment = 40,000 × 0.5 = $20,000
   *
   * Bước 2 - Tính PnL:
   * PnL = (35,000 - 40,000) × 0.5
   *     = -5,000 × 0.5
   *     = -$2,500 (Lỗ)
   *
   * Bước 3 - Tính ROI:
   * ROI% = (-2,500 / 20,000) × 100
   *      = -0.125 × 100
   *      = -12.5% ❌
   *
   * Ý nghĩa: Đầu tư $20,000, lỗ $2,500 → Lỗ 12.5%
   *
   * ────────────────────────────────────────────────────
   * Case 3: SELL (SHORT) BTC - Lãi 25%
   * ────────────────────────────────────────────────────
   * Entry Price:    $50,000
   * Current Price:  $40,000
   * Quantity:       0.2 BTC
   * Side:           SELL
   *
   * Bước 1 - Tính Investment:
   * Investment = 50,000 × 0.2 = $10,000
   *
   * Bước 2 - Tính PnL:
   * PnL = (50,000 - 40,000) × 0.2
   *     = 10,000 × 0.2
   *     = $2,000 (Lãi - giá giảm như dự đoán)
   *
   * Bước 3 - Tính ROI:
   * ROI% = (2,000 / 10,000) × 100
   *      = 0.2 × 100
   *      = 20% ✅
   *
   * Ý nghĩa: Short $10,000, giá giảm 20% → Lãi 20%
   *
   * ────────────────────────────────────────────────────
   * Case 4: Multiple Small Profits (Scalping)
   * ────────────────────────────────────────────────────
   * Entry Price:    $40,000
   * Current Price:  $40,100
   * Quantity:       1 BTC
   * Side:           BUY
   *
   * Bước 1 - Investment:
   * Investment = 40,000 × 1 = $40,000
   *
   * Bước 2 - PnL:
   * PnL = (40,100 - 40,000) × 1 = $100
   *
   * Bước 3 - ROI:
   * ROI% = (100 / 40,000) × 100 = 0.25% ✅
   *
   * Ý nghĩa: Scalping với lãi nhỏ 0.25%
   * Nếu trade 10 lần/ngày → 2.5% profit/day
   *
   * ============================================================================
   * 📈 ROI BENCHMARKS (Tham Khảo)
   * ============================================================================
   *
   * │ ROI Range        │ Đánh Giá                    │ Màu Sắc │
   * ├──────────────────┼─────────────────────────────┼─────────┤
   * │ > +50%           │ Excellent (Rất tốt)        │ 🟢      │
   * │ +20% to +50%     │ Very Good (Tốt)            │ 🟢      │
   * │ +10% to +20%     │ Good (Khá tốt)             │ 🟢      │
   * │ +5% to +10%      │ Moderate (Trung bình)      │ 🟢      │
   * │ +0% to +5%       │ Small Profit (Lãi nhỏ)     │ 🟢      │
   * │ 0%               │ Break Even (Hòa vốn)       │ ⚪      │
   * │ -5% to 0%        │ Small Loss (Lỗ nhỏ)       │ 🔴      │
   * │ -10% to -5%      │ Moderate Loss (Lỗ TB)     │ 🔴      │
   * │ -20% to -10%     │ Significant Loss (Lỗ lớn) │ 🔴      │
   * │ < -20%           │ Heavy Loss (Lỗ nặng)       │ 🔴      │
   *
   * ============================================================================
   * 🎯 SO SÁNH PnL vs ROI
   * ============================================================================
   *
   * Scenario A:
   * ───────────
   * Investment: $1,000
   * PnL: $100
   * ROI: 10%
   *
   * Scenario B:
   * ───────────
   * Investment: $10,000
   * PnL: $100
   * ROI: 1%
   *
   * Nhận xét:
   * - Cùng PnL = $100
   * - Nhưng ROI khác nhau (10% vs 1%)
   * - ROI đo lường hiệu quả sử dụng vốn
   * - Scenario A hiệu quả hơn (10% > 1%)
   *
   * ============================================================================
   * 🔑 KEY POINTS
   * ============================================================================
   *
   * 1. ROI phụ thuộc vào:
   *    - PnL (Lợi nhuận/Lỗ)
   *    - Investment (Vốn đầu tư)
   *
   * 2. ROI giúp:
   *    - So sánh hiệu quả giữa các giao dịch
   *    - Đánh giá performance của chiến lược
   *    - Quyết định stop loss / take profit
   *
   * 3. ROI càng cao càng tốt:
   *    - ROI > 0: Đang lãi ✅
   *    - ROI = 0: Hòa vốn ⚪
   *    - ROI < 0: Đang lỗ ❌
   *
   * 4. Risk Management:
   *    - Set Stop Loss khi ROI < -5% (ví dụ)
   *    - Take Profit khi ROI > +10% (ví dụ)
   *    - Điều chỉnh theo risk tolerance
   *
   * ============================================================================
   * 🎯 DỰA VÀO PnL & ROI - TA BIẾT ĐƯỢC GÌ?
   * ============================================================================
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  1. HIỆU SUẤT GIAO DỊCH (Trading Performance)                      │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  ✅ Biết giao dịch đang lãi hay lỗ bao nhiêu                        │
   * │  ✅ Đánh giá hiệu quả sử dụng vốn                                   │
   * │  ✅ So sánh performance giữa các orders                             │
   * │                                                                      │
   * │  Example:                                                            │
   * │  Order A: PnL = $100, ROI = 10%  (Hiệu quả cao)                    │
   * │  Order B: PnL = $100, ROI = 1%   (Hiệu quả thấp)                   │
   * │  → Order A tốt hơn dù cùng PnL                                      │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  2. QUYẾT ĐỊNH STOP LOSS / TAKE PROFIT                              │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  📊 Stop Loss Trigger:                                              │
   * │  - ROI < -5%  → Cảnh báo (Warning)                                 │
   * │  - ROI < -10% → Cân nhắc stop loss                                 │
   * │  - ROI < -20% → Nên stop loss ngay                                 │
   * │                                                                      │
   * │  📈 Take Profit Trigger:                                            │
   * │  - ROI > +10%  → Có thể take profit một phần                       │
   * │  - ROI > +20%  → Nên take profit                                    │
   * │  - ROI > +50%  → Chốt lời ngay (quá tốt)                           │
   * │                                                                      │
   * │  Example:                                                            │
   * │  Order đang có ROI = -8%                                            │
   * │  → Gần stop loss threshold                                          │
   * │  → Cân nhắc: Giữ tiếp hay cắt lỗ?                                  │
   * │  → Xem market trend để quyết định                                   │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  3. ĐÁNH GIÁ CHIẾN LƯỢC TRADING                                     │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  📊 Win Rate (Tỷ lệ thắng):                                        │
   * │  - Bao nhiêu % orders có ROI > 0?                                   │
   * │  - Win Rate = (Orders Lãi / Tổng Orders) × 100                     │
   * │                                                                      │
   * │  💰 Average ROI:                                                    │
   * │  - ROI trung bình của tất cả orders                                │
   * │  - Đánh giá chiến lược có profitable không                         │
   * │                                                                      │
   * │  📈 Profit Factor:                                                  │
   * │  - Tổng Lãi / Tổng Lỗ                                              │
   * │  - Profit Factor > 1 → Chiến lược tốt                              │
   * │                                                                      │
   * │  Example:                                                            │
   * │  10 orders:                                                         │
   * │  - 7 orders lãi (avg ROI: +8%)                                     │
   * │  - 3 orders lỗ (avg ROI: -5%)                                      │
   * │  → Win Rate = 70%                                                   │
   * │  → Average ROI = +4.1%                                              │
   * │  → Chiến lược khá tốt ✅                                            │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  4. QUẢN LÝ RỦI RO (Risk Management)                                │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  🎯 Position Sizing:                                                │
   * │  - Nếu order đang lỗ (ROI < 0)                                     │
   * │  → Không mở thêm position tương tự                                  │
   * │  - Nếu order đang lãi tốt (ROI > +10%)                             │
   * │  → Có thể thêm position (scale in)                                  │
   * │                                                                      │
   * │  💸 Capital Allocation:                                             │
   * │  - Xem tổng PnL của tất cả orders                                  │
   * │  - Đảm bảo không vượt quá risk limit                               │
   * │  - Ví dụ: Max drawdown = 20% portfolio                             │
   * │                                                                      │
   * │  🔄 Portfolio Rebalancing:                                          │
   * │  - Orders lãi quá nhiều → Take profit một phần                     │
   * │  - Orders lỗ nhiều → Stop loss                                     │
   * │  - Giữ portfolio balance và risk control                            │
   * │                                                                      │
   * │  Example:                                                            │
   * │  Portfolio: $10,000                                                 │
   * │  Order A: ROI = -15% (lỗ $1,500)                                   │
   * │  Order B: ROI = -8% (lỗ $800)                                      │
   * │  → Tổng lỗ = $2,300 (23% portfolio)                                │
   * │  → Vượt risk limit (20%)                                            │
   * │  → Cần stop loss ngay! ⚠️                                           │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  5. TÂM LÝ TRADING (Trading Psychology)                             │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  😊 PnL/ROI dương (Profit):                                        │
   * │  - Tâm lý thoải mái, tự tin                                        │
   * │  - ⚠️ Cẩn thận: Overconfidence → Sai lầm                           │
   * │  - Giữ discipline, không FOMO                                       │
   * │                                                                      │
   * │  😰 PnL/ROI âm (Loss):                                             │
   * │  - Tâm lý stress, muốn "gỡ vốn"                                    │
   * │  - ⚠️ Nguy hiểm: Revenge trading                                    │
   * │  - Cần bình tĩnh, stop loss đúng lúc                               │
   * │                                                                      │
   * │  🎯 Quy tắc vàng:                                                   │
   * │  - Không để emotion chi phối                                        │
   * │  - Follow plan, không trade cảm tính                                │
   * │  - PnL/ROI là số liệu, không phải cảm xúc                          │
   * │                                                                      │
   * │  Example:                                                            │
   * │  Trader A: ROI = -10%                                               │
   * │  Emotion: "Phải gỡ vốn ngay!"                                      │
   * │  Action: Mở thêm 5 orders liều (Revenge trading)                   │
   * │  Result: ROI = -30% (Tệ hơn) ❌                                     │
   * │                                                                      │
   * │  Trader B: ROI = -10%                                               │
   * │  Action: Stop loss, nghỉ ngơi, review strategy                     │
   * │  Result: Giữ được vốn, comeback sau ✅                              │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  6. TIMING THỊ TRƯỜNG (Market Timing)                               │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  📊 Entry Timing:                                                   │
   * │  - Xem ROI của orders trước                                         │
   * │  - Nếu nhiều orders lỗ → Market không thuận lợi                    │
   * │  → Chờ đợi, không vào lệnh mới                                      │
   * │                                                                      │
   * │  📈 Exit Timing:                                                    │
   * │  - ROI đạt target → Take profit                                     │
   * │  - ROI xuống stop loss → Cut loss                                   │
   * │  - Market đảo chiều → Chốt lời/cắt lỗ                              │
   * │                                                                      │
   * │  🔄 Market Condition:                                               │
   * │  - Nhiều orders ROI > 0 → Bull market, trend tốt                   │
   * │  - Nhiều orders ROI < 0 → Bear market, trend xấu                   │
   * │  - Điều chỉnh strategy theo market                                  │
   * │                                                                      │
   * │  Example:                                                            │
   * │  5 orders gần đây đều có ROI < -5%                                 │
   * │  → Market đang sideways/downtrend                                   │
   * │  → Không nên open thêm LONG positions                               │
   * │  → Cân nhắc SHORT hoặc chờ đợi                                      │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  7. PHÂN TÍCH SYMBOL/COIN                                           │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  🪙 Performance theo Symbol:                                        │
   * │  - BTCUSDT orders: Average ROI = +5%                                │
   * │  - ETHUSDT orders: Average ROI = -3%                                │
   * │  → BTC trade tốt hơn ETH                                            │
   * │  → Focus vào BTC, giảm ETH                                          │
   * │                                                                      │
   * │  📊 Best/Worst Performers:                                          │
   * │  - Symbol nào cho ROI tốt nhất?                                    │
   * │  - Symbol nào hay lỗ?                                               │
   * │  - Điều chỉnh portfolio allocation                                  │
   * │                                                                      │
   * │  Example:                                                            │
   * │  BTCUSDT: 10 orders, avg ROI = +8%                                 │
   * │  ETHUSDT: 10 orders, avg ROI = +2%                                 │
   * │  BNBUSDT: 10 orders, avg ROI = -5%                                 │
   * │  → Tăng tỷ trọng BTC                                                │
   * │  → Giảm/Dừng trade BNB                                              │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ┌─────────────────────────────────────────────────────────────────────┐
   * │  8. BÁO CÁO & THUẾ (Reporting & Tax)                                │
   * ├─────────────────────────────────────────────────────────────────────┤
   * │                                                                      │
   * │  📊 Tính toán lợi nhuận thực tế:                                   │
   * │  - Tổng PnL của tất cả orders                                      │
   * │  - Profit/Loss cho kỳ (ngày/tuần/tháng)                            │
   * │  - Report cho thuế (Capital Gains Tax)                              │
   * │                                                                      │
   * │  💰 Realized vs Unrealized PnL:                                     │
   * │  - Realized: Orders đã đóng (filled/closed)                         │
   * │  - Unrealized: Orders đang mở (new/pending)                         │
   * │  - Chỉ realized PnL mới tính thuế                                   │
   * │                                                                      │
   * │  📈 Performance Tracking:                                           │
   * │  - Track PnL/ROI theo thời gian                                    │
   * │  - Xem trend: Đang improve hay decline?                            │
   * │  - Adjust strategy accordingly                                      │
   * │                                                                      │
   * │  Example:                                                            │
   * │  Tháng 1: Total PnL = +$5,000, ROI = +15%                          │
   * │  Tháng 2: Total PnL = -$2,000, ROI = -6%                           │
   * │  Tháng 3: Total PnL = +$8,000, ROI = +24%                          │
   * │  → Q1 profit: $11,000 (cần report thuế)                            │
   * │  → Trend improving ✅                                               │
   * └─────────────────────────────────────────────────────────────────────┘
   *
   * ============================================================================
   * 🎯 TÓM TẮT - HÀNH ĐỘNG DỰA VÀO PnL/ROI
   * ============================================================================
   *
   * ┌──────────────────┬─────────────────────────────────────────────────┐
   * │  Tình Huống      │  Hành Động                                      │
   * ├──────────────────┼─────────────────────────────────────────────────┤
   * │  ROI > +20%      │  ✅ Take profit (chốt lời một phần/toàn bộ)    │
   * │  ROI = +10~+20%  │  ✅ Set trailing stop, protect profit          │
   * │  ROI = 0~+10%    │  ✅ Theo dõi, chờ tăng thêm                    │
   * │  ROI = 0%        │  ⚪ Hòa vốn, cân nhắc exit                      │
   * │  ROI = -5~0%     │  ⚠️ Cảnh báo, monitor chặt                     │
   * │  ROI = -10~-5%   │  ⚠️ Cân nhắc stop loss                         │
   * │  ROI < -10%      │  🔴 Stop loss ngay (protect capital)            │
   * │  ROI < -20%      │  🔴🔴 Emergency exit!                           │
   * ├──────────────────┼─────────────────────────────────────────────────┤
   * │  Nhiều orders    │  📊 Review chiến lược trading                   │
   * │  cùng lỗ         │  🔍 Check market condition                      │
   * │                  │  ⏸️ Tạm dừng trading, rest                      │
   * ├──────────────────┼─────────────────────────────────────────────────┤
   * │  Nhiều orders    │  📈 Chiến lược đang work                        │
   * │  cùng lãi        │  ✅ Tiếp tục follow plan                        │
   * │                  │  ⚠️ Cẩn thận overconfidence                     │
   * └──────────────────┴─────────────────────────────────────────────────┘
   *
   * ============================================================================
   *
   * @param order - Order object chứa thông tin giao dịch
   * @param pnl - Calculated PnL value (từ hàm calculatePnL)
   * @returns ROI percentage (hoặc null nếu không tính được)
   */
  const calculateROI = (order: Order, pnl: number | null): number | null => {
    if (pnl === null || !order.quantity) return null;

    const entryPrice = order.filled_price || order.price;
    if (!entryPrice || entryPrice === 0) return null;

    const investment = entryPrice * order.quantity;
    if (investment === 0) return null;

    return (pnl / investment) * 100;
  };

  /**
   * 🎯 Get current price for PnL calculation
   *
   * Priority:
   * 1. realtimePrices[symbol] - Real-time from Binance API (most accurate)
   * 2. order.current_price - From DB (5s delay)
   * 3. order.filled_price - For filled orders (static)
   * 4. null - Cannot calculate
   *
   * Note: Futures orders have "_FUTURES" suffix in realtimePrices object
   */
  const getCurrentPriceForPnL = (order: Order): number | null => {
    const status = order.status?.toLowerCase();

    // For filled/closed orders, use filled_price (no PnL change)
    if (status === 'filled' || status === 'closed') {
      return order.filled_price || null;
    }

    // For open orders, use real-time price
    // Check if this is a Futures order
    const isFutures =
      order.trading_mode?.toLowerCase() === 'futures' ||
      order.trading_mode?.toLowerCase() === 'future';

    const priceKey = isFutures ? `${order.symbol}_FUTURES` : order.symbol;

    console.log(
      `🔍 Getting price for ${order.symbol} (${
        isFutures ? 'Futures' : 'Spot'
      }) - Key: ${priceKey}`,
    );

    if (realtimePrices[priceKey]) {
      console.log(
        `✅ Found real-time price: $${realtimePrices[priceKey].price}`,
      );
      return realtimePrices[priceKey].price;
    }

    if (order.current_price) {
      console.log(`⚠️ Using DB price: $${order.current_price}`);
      return order.current_price;
    }

    console.log(`❌ No price available for ${order.symbol}`);
    return null;
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Orders</h1>

        {/* WebSocket Status Indicator */}
        <div className="flex items-center gap-2">
          <div
            className={`w-3 h-3 rounded-full ${
              wsStatus === 'CONNECTED'
                ? 'bg-green-500'
                : wsStatus === 'CONNECTING'
                ? 'bg-yellow-500 animate-pulse'
                : 'bg-red-500'
            }`}
          />
          <span className="text-sm text-gray-600">
            {wsStatus === 'CONNECTED'
              ? 'Real-time updates active'
              : wsStatus === 'CONNECTING'
              ? 'Connecting...'
              : 'Disconnected'}
          </span>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gradient-to-br from-purple-500 to-purple-600 text-white rounded-lg shadow p-6">
          <p className="text-sm opacity-90">Tổng Lệnh</p>
          <p className="text-3xl font-bold mt-2">{stats.total}</p>
        </div>
        <div className="bg-gradient-to-br from-green-500 to-green-600 text-white rounded-lg shadow p-6">
          <p className="text-sm opacity-90">Đã Khớp</p>
          <p className="text-3xl font-bold mt-2">{stats.filled}</p>
        </div>
        <div className="bg-gradient-to-br from-yellow-500 to-yellow-600 text-white rounded-lg shadow p-6">
          <p className="text-sm opacity-90">Đang Chờ</p>
          <p className="text-3xl font-bold mt-2">{stats.New}</p>
        </div>
        <div className="bg-gradient-to-br from-red-500 to-red-600 text-white rounded-lg shadow p-6">
          <p className="text-sm opacity-90">Đã Hủy</p>
          <p className="text-3xl font-bold mt-2">{stats.cancelled}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Symbol
            </label>
            <input
              type="text"
              value={filters.symbol}
              onChange={(e) =>
                setFilters({...filters, symbol: e.target.value.toUpperCase()})
              }
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
              placeholder="BTCUSDT"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Status
            </label>
            <select
              value={filters.status}
              onChange={(e) => setFilters({...filters, status: e.target.value})}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900">
              <option value="">All</option>
              <option value="new">New</option>
              <option value="filled">Filled</option>
              <option value="closed">Closed</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Side
            </label>
            <select
              value={filters.side}
              onChange={(e) => setFilters({...filters, side: e.target.value})}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900">
              <option value="">All</option>
              <option value="buy">Buy</option>
              <option value="sell">Sell</option>
            </select>
          </div>
        </div>
      </div>

      {/* Orders Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">
            Danh Sách Lệnh
          </h2>
        </div>

        {loading && (
          <div className="text-center py-12">
            <p className="text-gray-500">Loading orders...</p>
          </div>
        )}

        {error && (
          <div className="text-center py-12">
            <p className="text-red-500 font-medium">{error}</p>
          </div>
        )}

        {!loading && !error && orders.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">Chưa có lệnh nào</p>
          </div>
        )}

        {!loading && !error && orders.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Order ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Bot Config
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Symbol
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Mode
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Side
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Type
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Amount
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Entry Price
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Current Price
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Stop Loss
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Take Profit
                  </th>
                  <th className="px-6 py-3 text-lef`t text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    PnL
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ROI
                  </th>
                  {/* <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th> */}
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {orders.map((order) => (
                  <tr key={order.id} className={`hover:bg-gray-50 border-l-4 `}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {order.id}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-blue-600">
                      {order.order_id || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {order.bot_config_name || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="font-medium text-gray-900">
                        {formatSymbol(order.symbol)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="px-2 py-1 text-xs font-semibold rounded capitalize bg-yellow-100 text-yellow-800">
                        {order.trading_mode || 'spot'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs font-semibold rounded uppercase ${
                          order.side === 'BUY'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}>
                        {order.side}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 capitalize">
                      {order.type}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-bold text-gray-900">
                      {order.quantity}
                    </td>

                    <td className="px-6 py-4 whitespace-nowrap text-sm font-bold text-gray-900">
                      {order.price && order.price !== 0
                        ? order.price.toFixed(5)
                        : order.filled_price
                        ? order.filled_price.toFixed(5)
                        : '-'}
                    </td>

                    {/* 
                      💰 CURRENT PRICE COLUMN - Logic hiển thị theo trạng thái order
                      
                      📋 RULE 1: Order ĐÃ FILLED/CLOSED → Hiển thị giá khớp (không real-time)
                         ✅ Status: filled, closed
                         ✅ Hiển thị: order.filled_price
                         🎨 Style: Bold, màu xanh lá (text-green-700)
                         📊 Example: $42,150.50 (cố định, không thay đổi)
                         💡 Lý do: Lệnh đã hoàn thành, giá không còn thay đổi
                      
                      📋 RULE 2: Order ĐANG MỞ → Hiển thị giá real-time (new/pending/partially_filled)
                         ✅ Status: new, pending, partially_filled, open
                         
                         Priority hiển thị (từ cao xuống thấp):
                         
                         1️⃣ realtimePrices[order.symbol] - GIÁ REAL-TIME TỪ BINANCE API
                            ✅ Nguồn: Fetch trực tiếp từ Binance testnet mỗi 2s
                            ✅ Hiển thị: 
                               - Giá lớn, bold, màu xanh/đỏ theo % thay đổi
                               - Có animate-pulse effect (nổi bật)
                               - Kèm % thay đổi 24h bên dưới
                            📊 Example: $42,150.50 (màu xanh, pulse) với +2.35%
                            
                         2️⃣ order.current_price - GIÁ TỪ DATABASE
                            📦 Nguồn: Backend worker update mỗi 5s
                            🎨 Hiển thị: Bold, màu xanh dương (font-semibold text-blue-600)
                            📊 Example: $42,150.50
                            
                         3️⃣ order.price - GIÁ ĐẶT LỆNH
                            📝 Nguồn: Giá ban đầu user đặt
                            🎨 Hiển thị: Medium, màu xám (font-medium text-gray-600)
                            📊 Example: 42,150.50
                            
                         4️⃣ "-" - KHÔNG CÓ GIÁ
                            ⚠️ Fallback cuối cùng
                      
                      🎯 Flow Logic:
                         1. Check order.status
                            ├─ filled/closed → Show filled_price (RULE 1)
                            └─ other → Show real-time price (RULE 2)
                         
                         2. Nếu RULE 2, check theo priority:
                            realtimePrices → current_price → price → "-"
                      
                      🎨 Màu sắc:
                         - 🟢 Xanh đậm (green-700): Giá đã khớp (filled order)
                         - 🟢 Xanh lá (green-600): Giá tăng (real-time, positive %)
                         - 🔴 Đỏ (red-600): Giá giảm (real-time, negative %)
                         - 💙 Xanh dương (blue-600): Giá từ DB
                         - ⚪ Xám (gray-600): Giá đặt lệnh
                    */}
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {(() => {
                        const status = order.status?.toLowerCase();
                        const isFilled =
                          status === 'filled' || status === 'closed';

                        // RULE 1: Order đã filled → show filled price (cố định)
                        if (isFilled && order.filled_price) {
                          return (
                            <div className="flex flex-col">
                              <span className="font-bold text-base text-green-700">
                                ${order.filled_price.toFixed(5)}
                              </span>
                              <span className="text-xs text-green-600">
                                Filled ✓
                              </span>
                            </div>
                          );
                        }

                        // RULE 2: Order đang mở → show real-time price
                        // Priority: realtimePrices → current_price → price → "-"

                        // Check if this is a Futures order (need to use _FUTURES suffix)
                        const isFutures =
                          order.trading_mode?.toLowerCase() === 'futures' ||
                          order.trading_mode?.toLowerCase() === 'future';
                        const priceKey = isFutures
                          ? `${order.symbol}_FUTURES`
                          : order.symbol;

                        if (realtimePrices[priceKey]) {
                          return (
                            <div className="flex flex-col">
                              <span
                                className={`font-bold text-base ${
                                  realtimePrices[priceKey].percent >= 0
                                    ? 'text-green-600'
                                    : 'text-red-600'
                                } animate-pulse`}>
                                ${realtimePrices[priceKey].price.toFixed(5)}
                              </span>
                              <span
                                className={`text-xs ${
                                  realtimePrices[priceKey].percent >= 0
                                    ? 'text-green-500'
                                    : 'text-red-500'
                                }`}>
                                {realtimePrices[priceKey].percent >= 0
                                  ? '+'
                                  : ''}
                                {realtimePrices[priceKey].percent.toFixed(2)}%
                                {isFutures && (
                                  <span className="ml-1 text-purple-600">
                                    📊
                                  </span>
                                )}
                              </span>
                            </div>
                          );
                        }

                        if (order.current_price) {
                          return (
                            <span className="font-semibold text-blue-600">
                              ${order.current_price.toFixed(5)}
                            </span>
                          );
                        }

                        if (order.price) {
                          return (
                            <span className="font-medium text-gray-600">
                              {order.price.toFixed(5)}
                            </span>
                          );
                        }

                        return <span className="text-gray-400">-</span>;
                      })()}
                    </td>

                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {order.stop_loss_price ? (
                        (() => {
                          const entry = order.filled_price ?? order.price;
                          const pct = calcTargetPercent(
                            order.stop_loss_price,
                            entry,
                            order.side,
                          );
                          return (
                            <div className="flex flex-col">
                              <span className="text-red-600 font-medium">
                                {order.stop_loss_price.toFixed(5)}
                              </span>
                              {pct !== null && (
                                <span
                                  className={`text-xs ${
                                    pct >= 0 ? 'text-green-600' : 'text-red-600'
                                  }`}>
                                  {pct >= 0 ? '+' : ''}
                                  {pct.toFixed(2)}%
                                </span>
                              )}
                            </div>
                          );
                        })()
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {order.take_profit_price ? (
                        (() => {
                          const entry = order.filled_price ?? order.price;
                          const pct = calcTargetPercent(
                            order.take_profit_price,
                            entry,
                            order.side,
                          );
                          return (
                            <div className="flex flex-col">
                              <span className="text-green-600 font-medium">
                                {order.take_profit_price.toFixed(5)}
                              </span>
                              {pct !== null && (
                                <span
                                  className={`text-xs ${
                                    pct >= 0 ? 'text-green-600' : 'text-red-600'
                                  }`}>
                                  {pct >= 0 ? '+' : ''}
                                  {pct.toFixed(2)}%
                                </span>
                              )}
                            </div>
                          );
                        })()
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs font-semibold rounded capitalize ${
                          order.status?.toLowerCase() === 'filled' ||
                          order.status?.toLowerCase() === 'closed'
                            ? 'bg-green-100 text-green-800'
                            : order.status?.toLowerCase() === 'new'
                            ? 'bg-yellow-100 text-yellow-800'
                            : order.status?.toLowerCase() === 'cancelled'
                            ? 'bg-red-100 text-red-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}>
                        {order.status}
                      </span>
                    </td>
                    {/* 
                      💰 PnL COLUMN - Real-time Profit/Loss calculation
                      
                      Logic:
                      1. Tính PnL real-time dựa trên current price
                      2. BUY: (Current - Entry) × Quantity
                      3. SELL: (Entry - Current) × Quantity
                      4. Hiển thị màu xanh (profit) / đỏ (loss)
                      5. Animate pulse cho orders đang mở
                    */}
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {(() => {
                        const currentPrice = getCurrentPriceForPnL(order);
                        const pnl = calculatePnL(order, currentPrice);
                        const status = order.status?.toLowerCase();
                        const isOpen =
                          status === 'new' ||
                          status === 'open' ||
                          status === 'pending';

                        if (pnl !== null) {
                          return (
                            <span
                              className={`font-semibold ${
                                pnl >= 0 ? 'text-green-600' : 'text-red-600'
                              } ${isOpen ? 'animate-pulse' : ''}`}>
                              {pnl >= 0 ? '+' : ''}${pnl.toFixed(2)}
                            </span>
                          );
                        }

                        return <span className="text-gray-400">-</span>;
                      })()}
                    </td>
                    {/* 
                      📊 ROI COLUMN - Return on Investment percentage
                      
                      Formula: (PnL / Investment) × 100
                      Investment = Entry Price × Quantity
                      
                      Example:
                      - Entry: $40,000 × 0.1 BTC = $4,000
                      - PnL: $200
                      - ROI: (200 / 4000) × 100 = 5%
                    */}
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {(() => {
                        const currentPrice = getCurrentPriceForPnL(order);
                        const pnl = calculatePnL(order, currentPrice);
                        const roi = calculateROI(order, pnl);
                        const status = order.status?.toLowerCase();
                        const isOpen =
                          status === 'new' ||
                          status === 'open' ||
                          status === 'pending';

                        if (roi !== null) {
                          return (
                            <span
                              className={`font-semibold ${
                                roi >= 0 ? 'text-green-600' : 'text-red-600'
                              } ${isOpen ? 'animate-pulse' : ''}`}>
                              {roi >= 0 ? '+' : ''}
                              {roi.toFixed(2)}%
                            </span>
                          );
                        }

                        return <span className="text-gray-400">-</span>;
                      })()}
                    </td>
                    {/* <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <button
                        onClick={() => handleRefreshPnL(order.id)}
                        disabled={refreshingPnL === order.id}
                        className={`text-blue-600 hover:text-blue-800 font-medium ${
                          refreshingPnL === order.id
                            ? 'opacity-50 cursor-not-allowed'
                            : ''
                        }`}
                        title="Refresh PnL">
                        {refreshingPnL === order.id ? '⏳' : '🔄'} Refresh PnL
                      </button>
                    </td> */}
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDate(order.created_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
