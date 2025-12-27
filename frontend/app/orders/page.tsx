'use client';
import {useState, useEffect} from 'react';
import {useRouter} from 'next/navigation';
import {Order, getOrderHistory, closeOrder} from '../../services/orderService';
import websocketService from '../../services/websocketService';

// WebSocket message interface for order updates with position data
interface OrderUpdateMessage {
  type: string;
  data: {
    order_id: number;
    user_id?: number;
    symbol: string;
    side: string;
    status: string;
    trading_mode?: string;
    position?: {
      symbol: string;
      position_amt: number;
      position_side: string;
      entry_price: number;
      mark_price: number;
      liquidation_price: number;
      unrealized_profit: number;
      pnl_percent: number;
      leverage: number;
      margin_type: string;
      isolated: boolean;
      isolated_margin: number;
    };
    timestamp?: number;
  };
}

export default function OrdersPage() {
  const router = useRouter();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [wsStatus, setWsStatus] = useState<string>('DISCONNECTED');

  // Modal state
  const [showCloseModal, setShowCloseModal] = useState(false);
  const [orderToClose, setOrderToClose] = useState<Order | null>(null);
  const [closing, setClosing] = useState(false);

  // Toast notification state
  const [toast, setToast] = useState<{
    show: boolean;
    message: string;
    type: 'success' | 'error';
  }>({show: false, message: '', type: 'success'});

  // Realtime prices: key = "ETHUSDT_FUTURES" hoặc "BTCUSDT" (Spot)
  const [realtimePrices, setRealtimePrices] = useState<{
    [key: string]: {price: number; change: number; percent: number};
  }>({});

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    filled: 0,
    New: 0,
    cancelled: 0,
    totalPnl: 0,
    filledPnl: 0,
    newPnl: 0,
    cancelledPnl: 0,
  });

  // Filters
  const [filters, setFilters] = useState({
    symbol: '',
    status: '',
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
      console.log(
        '⏱️  Note: Position data will be populated via WebSocket updates (every 5s)',
      );

      // Debug: Check which orders have position data
      const ordersWithPosition = data.filter((o) => o.position);
      console.log(
        `📊 Orders with position data: ${ordersWithPosition.length}/${data.length}`,
      );
      ordersWithPosition.forEach((o) => {
        console.log(
          `  Order ${o.id}: position_amt=${o.position?.position_amt}`,
        );
      });

      // Calculate stats with PnL
      const total = data.length;
      const filledOrders = data.filter(
        (o) =>
          o.status?.toLowerCase() === 'filled' ||
          o.status?.toLowerCase() === 'closed',
      );
      const filled = filledOrders.length;

      const newOrders = data.filter(
        (o) =>
          o.status?.toLowerCase() === 'open' ||
          o.status?.toLowerCase() === 'new',
      );
      const New = newOrders.length;

      // For PnL calculation, only count orders with status = 'new' (not 'open', 'pending', etc.)
      const newOrdersForPnl = data.filter(
        (o) => o.status?.toLowerCase() === 'new',
      );

      const cancelledOrders = data.filter(
        (o) => o.status?.toLowerCase() === 'cancelled',
      );
      const cancelled = cancelledOrders.length;

      // Calculate PnL for each category
      const calculateOrderPnl = (order: Order): number => {
        const status = order.status?.toLowerCase();
        const isFutures =
          order.trading_mode?.toLowerCase() === 'futures' ||
          order.trading_mode?.toLowerCase() === 'future';

        // Futures closed: use order.pnl
        if (isFutures && status === 'closed' && order.pnl) {
          return order.pnl;
        }

        // Futures open: use position.unrealized_profit
        if (isFutures && order.position?.unrealized_profit) {
          return parseFloat(order.position.unrealized_profit);
        }

        return 0;
      };

      const totalPnl = data.reduce((sum, o) => sum + calculateOrderPnl(o), 0);
      const filledPnl = filledOrders.reduce(
        (sum, o) => sum + calculateOrderPnl(o),
        0,
      );
      const newPnl = newOrdersForPnl.reduce(
        (sum, o) => sum + calculateOrderPnl(o),
        0,
      );
      const cancelledPnl = cancelledOrders.reduce(
        (sum, o) => sum + calculateOrderPnl(o),
        0,
      );

      setStats({
        total,
        filled,
        New,
        cancelled,
        totalPnl,
        filledPnl,
        newPnl,
        cancelledPnl,
      });
      setError(null);
    } catch (err) {
      console.error('Failed to fetch orders:', err);
      setError('Failed to load orders. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const refreshOrdersLight = async () => {
    try {
      const params: any = {limit: 100, offset: 0};
      if (filters.symbol) params.symbol = filters.symbol;
      if (filters.status) params.status = filters.status;
      if (filters.side) params.side = filters.side;

      const data = await getOrderHistory(params);
      setOrders(data);

      const total = data.length;
      const filledOrders = data.filter((o) =>
        ['filled', 'closed'].includes(o.status?.toLowerCase() ?? ''),
      );
      const filled = filledOrders.length;

      const newOrders = data.filter((o) =>
        ['open', 'new', 'pending', 'partially_filled'].includes(
          o.status?.toLowerCase() ?? '',
        ),
      );
      const New = newOrders.length;

      // For PnL calculation, only count orders with status = 'new' (not 'open', 'pending', etc.)
      const newOrdersForPnl = data.filter(
        (o) => o.status?.toLowerCase() === 'new',
      );

      const cancelledOrders = data.filter(
        (o) => o.status?.toLowerCase() === 'cancelled',
      );
      const cancelled = cancelledOrders.length;

      // Calculate PnL for each category
      const calculateOrderPnl = (order: Order): number => {
        const status = order.status?.toLowerCase();
        const isFutures =
          order.trading_mode?.toLowerCase() === 'futures' ||
          order.trading_mode?.toLowerCase() === 'future';

        if (isFutures && status === 'closed' && order.pnl) {
          return order.pnl;
        }

        if (isFutures && order.position?.unrealized_profit) {
          return parseFloat(order.position.unrealized_profit);
        }

        return 0;
      };

      const totalPnl = data.reduce((sum, o) => sum + calculateOrderPnl(o), 0);
      const filledPnl = filledOrders.reduce(
        (sum, o) => sum + calculateOrderPnl(o),
        0,
      );
      const newPnl = newOrdersForPnl.reduce(
        (sum, o) => sum + calculateOrderPnl(o),
        0,
      );
      const cancelledPnl = cancelledOrders.reduce(
        (sum, o) => sum + calculateOrderPnl(o),
        0,
      );

      setStats({
        total,
        filled,
        New,
        cancelled,
        totalPnl,
        filledPnl,
        newPnl,
        cancelledPnl,
      });
    } catch (err) {
      // Silent fail on refresh
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    // Validate token format (JWT has 3 parts separated by dots)
    const tokenParts = token.split('.');
    if (tokenParts.length !== 3) {
      console.error('Invalid token format');
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      router.push('/login');
      return;
    }

    // Check token expiration
    try {
      const payload = JSON.parse(atob(tokenParts[1]));
      const expirationTime = payload.exp * 1000; // Convert to milliseconds
      const currentTime = Date.now();

      if (currentTime >= expirationTime) {
        console.error('Token expired');
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        router.push('/login');
        return;
      }

      // Optional: Show warning if token expires in less than 5 minutes
      const timeUntilExpiry = expirationTime - currentTime;
      if (timeUntilExpiry < 5 * 60 * 1000) {
        console.warn(
          `Token expires in ${Math.floor(timeUntilExpiry / 1000 / 60)} minutes`,
        );
      }
    } catch (err) {
      console.error('Failed to decode token:', err);
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      router.push('/login');
      return;
    }

    fetchOrders();

    websocketService.connect();

    const statusInterval = setInterval(() => {
      setWsStatus(websocketService.getConnectionState());
    }, 1000);

    const unsubscribeOrderUpdates = websocketService.onMessage((message) => {
      if (message.type === 'order_update') {
        const updateMsg = message as OrderUpdateMessage;
        console.log('📡 Received order update:', updateMsg);

        // Extract data from message.data
        const data = updateMsg.data;
        if (!data) {
          console.warn('❌ Invalid message format: missing data field');
          return;
        }

        console.log('📦 Order update data:', {
          order_id: data.order_id,
          symbol: data.symbol,
          status: data.status,
          has_position: !!data.position,
          position_data: data.position,
          isolated: data.position?.isolated,
        });

        // Update specific order in state with position data
        setOrders((prevOrders) => {
          console.log(
            `🔍 Looking for order ${data.order_id} in ${prevOrders.length} orders`,
          );
          console.log(
            '  Available order IDs:',
            prevOrders.map((o) => o.id),
          );

          const orderIndex = prevOrders.findIndex(
            (o) => o.id === data.order_id,
          );

          if (orderIndex === -1) {
            // Order not found, do full refresh
            console.log(
              `⚠️  Order ${data.order_id} not found in current list, refreshing...`,
            );
            refreshOrdersLight();
            return prevOrders;
          }

          // Update order with new data including position info
          const updatedOrders = [...prevOrders];
          const existingOrder = updatedOrders[orderIndex];

          updatedOrders[orderIndex] = {
            ...existingOrder,
            status: data.status,
            position: data.position
              ? {
                  position_amt: String(data.position.position_amt || '0'),
                  entry_price: String(data.position.entry_price || '0'),
                  mark_price: String(data.position.mark_price || '0'),
                  liquidation_price: String(
                    data.position.liquidation_price || '0',
                  ),
                  unrealized_profit: String(
                    data.position.unrealized_profit || '0',
                  ),
                  pnl_percent: String(data.position.pnl_percent || '0'),
                  leverage: String(data.position.leverage || '0'),
                  margin_type: data.position.margin_type || '',
                  isolated_margin: String(data.position.isolated_margin || '0'),
                  position_side: data.position.position_side || '',
                  isolated: data.position.isolated || false,
                }
              : undefined,
            // Update PnL from position if available
            ...(data.position && {
              pnl: Number(data.position.unrealized_profit) || 0,
              pnl_percent: Number(data.position.pnl_percent) || 0,
            }),
          };

          console.log(
            `✅ Updated order ${data.order_id} with status=${data.status}${
              data.position ? ', position data included' : ''
            }`,
          );
          console.log('Updated order object:', updatedOrders[orderIndex]);
          return updatedOrders;
        });
      }
    });

    return () => {
      unsubscribeOrderUpdates();
      clearInterval(statusInterval);
      websocketService.disconnect();
    };
  }, [filters]);

  // Recalculate stats (especially PnL) when orders change (e.g., WebSocket updates)
  useEffect(() => {
    if (orders.length === 0) return;

    const total = orders.length;
    const filledOrders = orders.filter((o) =>
      ['filled', 'closed'].includes(o.status?.toLowerCase() ?? ''),
    );
    const filled = filledOrders.length;

    const newOrders = orders.filter((o) =>
      ['open', 'new', 'pending', 'partially_filled'].includes(
        o.status?.toLowerCase() ?? '',
      ),
    );
    const New = newOrders.length;

    // For PnL calculation, only count orders with status = 'new'
    const newOrdersForPnl = orders.filter(
      (o) => o.status?.toLowerCase() === 'new',
    );

    const cancelledOrders = orders.filter(
      (o) => o.status?.toLowerCase() === 'cancelled',
    );
    const cancelled = cancelledOrders.length;

    // Calculate PnL for each category
    const calculateOrderPnl = (order: Order): number => {
      const status = order.status?.toLowerCase();
      const isFutures =
        order.trading_mode?.toLowerCase() === 'futures' ||
        order.trading_mode?.toLowerCase() === 'future';

      if (isFutures && status === 'closed' && order.pnl) {
        return order.pnl;
      }

      if (isFutures && order.position?.unrealized_profit) {
        return parseFloat(order.position.unrealized_profit);
      }

      return 0;
    };

    const totalPnl = orders.reduce((sum, o) => sum + calculateOrderPnl(o), 0);
    const filledPnl = filledOrders.reduce(
      (sum, o) => sum + calculateOrderPnl(o),
      0,
    );
    const newPnl = newOrdersForPnl.reduce(
      (sum, o) => sum + calculateOrderPnl(o),
      0,
    );
    const cancelledPnl = cancelledOrders.reduce(
      (sum, o) => sum + calculateOrderPnl(o),
      0,
    );

    setStats({
      total,
      filled,
      New,
      cancelled,
      totalPnl,
      filledPnl,
      newPnl,
      cancelledPnl,
    });
  }, [orders]); // Re-run when orders change

  // 🔥 REAL-TIME PRICE VIA BINANCE FUTURES + SPOT WEBSOCKET
  useEffect(() => {
    if (orders.length === 0) return;
    console.log(
      `📊 Setting up price subscriptions for ${orders.length} orders`,
    );

    // Filter orders that need realtime price updates
    const ordersNeedingPrice = orders.filter((order) => {
      const status = order.status?.toLowerCase();
      const isFutures =
        order.trading_mode?.toLowerCase() === 'futures' ||
        order.trading_mode?.toLowerCase() === 'future';

      // Futures: Subscribe cho tất cả trừ 'closed'
      // Spot: Chỉ subscribe cho 'new', 'pending', 'partially_filled', 'open'
      if (isFutures) {
        return status !== 'closed';
      } else {
        return ['new', 'pending', 'partially_filled', 'open'].includes(
          status ?? '',
        );
      }
    });

    if (ordersNeedingPrice.length === 0) {
      console.log('📊 No orders need price subscription');
      return;
    }

    const spotSymbols = Array.from(
      new Set(
        ordersNeedingPrice
          .filter(
            (o) => !o.trading_mode || o.trading_mode.toLowerCase() === 'spot',
          )
          .map((o) => o.symbol.toLowerCase()),
      ),
    );

    const futuresSymbols = Array.from(
      new Set(
        ordersNeedingPrice
          .filter(
            (o) =>
              o.trading_mode?.toLowerCase() === 'futures' ||
              o.trading_mode?.toLowerCase() === 'future',
          )
          .map((o) => o.symbol.toLowerCase()),
      ),
    );

    console.log(
      `📊 Subscribing to ${spotSymbols.length} spot + ${futuresSymbols.length} futures symbols`,
    );

    let spotWs: WebSocket | null = null;
    let futuresWs: WebSocket | null = null;

    // ===== SPOT WEBSOCKET =====
    if (spotSymbols.length > 0) {
      const streams = spotSymbols.map((s) => `${s}@ticker`).join('/');

      console.log(`🔧 Spot symbols:`, spotSymbols);
      console.log(`🔧 Spot streams:`, streams);

      // Always use combined streams format (more reliable)
      const url = `wss://stream.binance.com:9443/stream?streams=${streams}`;
      console.log(`🔌 Connecting to Spot WS:`, url);

      spotWs = new WebSocket(url);

      spotWs.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          // Combined stream wraps data in "data" field
          if (message.data && message.data.e === '24hrTicker') {
            const data = message.data;
            setRealtimePrices((prev) => ({
              ...prev,
              [data.s.toUpperCase()]: {
                price: parseFloat(data.c),
                change: parseFloat(data.p),
                percent: parseFloat(data.P),
              },
            }));
          }
        } catch (err) {
          console.error('Error parsing Spot WS message:', err);
        }
      };

      spotWs.onopen = () => {
        console.log(`✅ Spot WS connected (${spotSymbols.length} symbols)`);
      };

      spotWs.onerror = (err) => {
        console.error('❌ Spot WS error:', err);
        console.error('   Symbols:', spotSymbols);
        console.error('   URL:', url);
      };

      spotWs.onclose = (event) => {
        console.log(
          `🔌 Spot WS closed. Code: ${event.code}, Reason: ${
            event.reason || 'None'
          }`,
        );
      };
    }

    // ===== FUTURES WEBSOCKET =====
    if (futuresSymbols.length > 0) {
      const streams = futuresSymbols.map((s) => `${s}@ticker`).join('/');

      console.log(`🔧 Futures symbols:`, futuresSymbols);
      console.log(`🔧 Futures streams:`, streams);

      // Always use combined streams format (more reliable)
      const url = `wss://fstream.binance.com/stream?streams=${streams}`;
      console.log(`🔌 Connecting to Futures WS:`, url);

      futuresWs = new WebSocket(url);

      futuresWs.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          // Combined stream wraps data in "data" field
          if (message.data && message.data.e === '24hrTicker') {
            const data = message.data;
            setRealtimePrices((prev) => ({
              ...prev,
              [`${data.s.toUpperCase()}_FUTURES`]: {
                price: parseFloat(data.c),
                change: parseFloat(data.p),
                percent: parseFloat(data.P),
              },
            }));
          }
        } catch (err) {
          console.error('Error parsing Futures WS message:', err);
        }
      };

      futuresWs.onopen = () => {
        console.log(
          `✅ Futures WS connected (${futuresSymbols.length} symbols)`,
        );
      };

      futuresWs.onerror = (err) => {
        console.error('❌ Futures WS error:', err);
        console.error('   Symbols:', futuresSymbols);
        console.error('   URL:', url);
      };

      futuresWs.onclose = (event) => {
        console.log(
          `🔌 Futures WS closed. Code: ${event.code}, Reason: ${
            event.reason || 'None'
          }`,
        );
      };
    }

    return () => {
      if (spotWs) {
        console.log('🔌 Closing Spot WS...');
        spotWs.close();
      }
      if (futuresWs) {
        console.log('🔌 Closing Futures WS...');
        futuresWs.close();
      }
    };
  }, [orders]);

  const formatSymbol = (symbol: string): string => {
    return symbol.endsWith('USDT') ? symbol.replace('USDT', '/USDT') : symbol;
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString('vi-VN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const calcTargetPercent = (
    target: number | undefined | null,
    entry: number | undefined | null,
    side: string,
  ): number | null => {
    if (!target || !entry || entry === 0) return null;
    const base = ((target - entry) / entry) * 100;
    return side?.toLowerCase() === 'sell' ? -base : base;
  };

  // Calculate Stop Loss price from percentage
  const calculateStopLossPrice = (
    entryPrice: number,
    slPercent: number,
    side: string,
  ): number => {
    // LONG: SL price = entry * (1 - slPercent/100)
    // SHORT: SL price = entry * (1 + slPercent/100)
    if (side?.toLowerCase() === 'buy') {
      return entryPrice * (1 - slPercent / 100);
    } else {
      return entryPrice * (1 + slPercent / 100);
    }
  };

  // Calculate Take Profit price from percentage
  const calculateTakeProfitPrice = (
    entryPrice: number,
    tpPercent: number,
    side: string,
  ): number => {
    // LONG: TP price = entry * (1 + tpPercent/100)
    // SHORT: TP price = entry * (1 - tpPercent/100)
    if (side?.toLowerCase() === 'buy') {
      return entryPrice * (1 + tpPercent / 100);
    } else {
      return entryPrice * (1 - tpPercent / 100);
    }
  };

  // ⭐ Lấy PnL từ data có sẵn (API hoặc WebSocket), KHÔNG tính toán
  const getPnL = (order: Order): number | null => {
    const status = order.status?.toLowerCase();
    const isFutures =
      order.trading_mode?.toLowerCase() === 'futures' ||
      order.trading_mode?.toLowerCase() === 'future';

    // Futures đã đóng: Lấy từ database (order.pnl)
    if (isFutures && status === 'closed') {
      return order.pnl ?? null;
    }

    // Futures đang mở: Lấy từ WebSocket position data
    if (isFutures && order.position?.unrealized_profit) {
      return parseFloat(order.position.unrealized_profit);
    }

    // Spot hoặc không có data: null
    return null;
  };

  const getPnLPercent = (order: Order): number | null => {
    const status = order.status?.toLowerCase();
    const isFutures =
      order.trading_mode?.toLowerCase() === 'futures' ||
      order.trading_mode?.toLowerCase() === 'future';

    // Futures đã đóng: Lấy từ database (order.pnl_percent)
    if (isFutures && status === 'closed') {
      return order.pnl_percent ?? null;
    }

    // Futures đang mở: Lấy từ WebSocket position data
    if (isFutures && order.position?.pnl_percent) {
      return parseFloat(order.position.pnl_percent);
    }

    // Spot hoặc không có data: null
    return null;
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Orders</h1>
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
          <div className="mt-3 pt-3 border-t border-purple-400">
            <p className="text-xs opacity-75">PnL:</p>
            <p
              className={`text-lg font-semibold ${
                (stats.totalPnl || 0) >= 0 ? 'text-green-200' : 'text-red-200'
              }`}>
              {(stats.totalPnl || 0) >= 0 ? '+' : ''}$
              {(stats.totalPnl || 0).toFixed(2)}
            </p>
          </div>
        </div>
        <div className="bg-gradient-to-br from-green-500 to-green-600 text-white rounded-lg shadow p-6">
          <p className="text-sm opacity-90">Đã Khớp</p>
          <p className="text-3xl font-bold mt-2">{stats.filled}</p>
          <div className="mt-3 pt-3 border-t border-green-400">
            <p className="text-xs opacity-75">PnL:</p>
            <p
              className={`text-lg font-semibold ${
                (stats.filledPnl || 0) >= 0 ? 'text-green-200' : 'text-red-200'
              }`}>
              {(stats.filledPnl || 0) >= 0 ? '+' : ''}$
              {(stats.filledPnl || 0).toFixed(2)}
            </p>
          </div>
        </div>
        <div className="bg-gradient-to-br from-yellow-500 to-yellow-600 text-white rounded-lg shadow p-6">
          <p className="text-sm opacity-90">Đang Chờ</p>
          <p className="text-3xl font-bold mt-2">{stats.New}</p>
          <div className="mt-3 pt-3 border-t border-yellow-400">
            <p className="text-xs opacity-75">PnL (Realtime):</p>
            <p
              className={`text-lg font-semibold animate-pulse ${
                (stats.newPnl || 0) >= 0 ? 'text-green-200' : 'text-red-200'
              }`}>
              {(stats.newPnl || 0) >= 0 ? '+' : ''}$
              {(stats.newPnl || 0).toFixed(2)}
            </p>
          </div>
        </div>
        <div className="bg-gradient-to-br from-red-500 to-red-600 text-white rounded-lg shadow p-6">
          <p className="text-sm opacity-90">Đã Hủy</p>
          <p className="text-3xl font-bold mt-2">{stats.cancelled}</p>
          <div className="mt-3 pt-3 border-t border-red-400">
            <p className="text-xs opacity-75">PnL:</p>
            <p
              className={`text-lg font-semibold ${
                (stats.cancelledPnl || 0) >= 0
                  ? 'text-green-200'
                  : 'text-red-200'
              }`}>
              {(stats.cancelledPnl || 0) >= 0 ? '+' : ''}$
              {(stats.cancelledPnl || 0).toFixed(2)}
            </p>
          </div>
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
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
              placeholder="ETHUSDT"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Status
            </label>
            <select
              value={filters.status}
              onChange={(e) => setFilters({...filters, status: e.target.value})}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500">
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
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500">
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
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Order Details
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Status / Exchange
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Trading Info
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Position / Liq
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Price / PnL
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    SL / TP
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Action
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {orders.map((order) => {
                  const pnl = getPnL(order);
                  const pnlPercent = getPnLPercent(order);
                  const isOpen = [
                    'new',
                    'open',
                    'pending',
                    'partially_filled',
                  ].includes(order.status?.toLowerCase() ?? '');

                  return (
                    <tr key={order.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 text-sm">{order.id}</td>
                      <td className="px-6 py-4">
                        <div className="flex flex-col gap-1">
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Created:
                            </span>
                            <span className="text-sm text-gray-600">
                              {formatDate(order.created_at)}
                            </span>
                          </div>
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Exchange:
                            </span>
                            <span className="px-2 py-1 text-xs font-semibold rounded bg-blue-100 text-blue-800 uppercase">
                              {order.exchange || 'Unknown'}
                            </span>
                          </div>

                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Order:
                            </span>
                            <span className="text-sm font-mono text-blue-600">
                              {order.order_id || '-'}
                            </span>
                          </div>
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Bot:
                            </span>
                            <span className="text-sm text-gray-900">
                              {order.bot_config_name || '-'}
                            </span>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex flex-col items-center gap-1">
                          <span
                            className={`px-2 py-1 text-xs font-semibold rounded capitalize ${
                              ['filled', 'closed'].includes(
                                order.status?.toLowerCase() ?? '',
                              )
                                ? 'bg-green-100 text-green-800'
                                : order.status?.toLowerCase() === 'new'
                                ? 'bg-yellow-100 text-yellow-800'
                                : order.status?.toLowerCase() === 'cancelled'
                                ? 'bg-red-100 text-red-800'
                                : 'bg-gray-100 text-gray-800'
                            }`}>
                            {order.status}
                          </span>
                          <span className="uppercase px-2 py-0.5 text-xs font-semibold rounded bg-yellow-100 text-yellow-800 capitalize">
                            {order.trading_mode || 'spot'}
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex flex-col gap-1.5">
                          {/* Symbol */}
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Symbol:
                            </span>
                            <span className="text-sm font-bold">
                              {formatSymbol(order.symbol)}
                            </span>
                          </div>
                          {/* Mode */}
                          {/* <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Mode:
                            </span>
                            <span className="px-2 py-0.5 text-xs font-semibold rounded bg-yellow-100 text-yellow-800 capitalize">
                              {order.trading_mode || 'spot'}
                            </span>
                          </div> */}
                          {/* Side */}
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Side:
                            </span>
                            <span
                              className={`px-2 py-0.5 text-xs font-semibold rounded uppercase ${
                                order.side === 'BUY'
                                  ? 'bg-green-100 text-green-800'
                                  : 'bg-red-100 text-red-800'
                              }`}>
                              {order.side}
                            </span>
                          </div>
                          {/* Type */}
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Type:
                            </span>
                            <span className="px-2 py-0.5 text-xs font-semibold rounded bg-blue-100 text-blue-800 capitalize">
                              {order.type}
                            </span>
                          </div>
                          {/* Amount */}
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Amount:
                            </span>
                            <span className="text-sm font-bold text-gray-900">
                              {order.quantity}
                            </span>
                          </div>
                        </div>
                      </td>
                      {/* Position & Liquidation Price - Combined */}
                      <td className="px-6 py-4 text-sm">
                        {(() => {
                          // Lấy position side từ DB (LONG/SHORT)
                          const positionSide =
                            order.position_side ||
                            order.position?.position_side;
                          const leverage =
                            order.leverage || order.position?.leverage;
                          const marginType =
                            order.margin_type || order.position?.margin_type;
                          const isIsolated =
                            marginType === 'isolated' ||
                            order.position?.isolated;

                          // Liq price: Ưu tiên lấy từ WebSocket, fallback về DB
                          const liqPrice = order.position?.liquidation_price
                            ? parseFloat(order.position.liquidation_price)
                            : order.liquidation_price;

                          // Nếu không có thông tin gì
                          if (
                            !positionSide &&
                            !leverage &&
                            !marginType &&
                            !liqPrice
                          ) {
                            return <span className="text-gray-400">-</span>;
                          }

                          return (
                            <div className="space-y-1.5">
                              {/* Position Direction (LONG/SHORT only, không show BOTH) */}
                              {positionSide && positionSide !== 'BOTH' && (
                                <div className="flex items-center gap-1">
                                  <span className="text-xs text-gray-500">
                                    Position:
                                  </span>
                                  <span
                                    className={`text-xs font-semibold ${
                                      positionSide === 'LONG'
                                        ? 'text-green-600'
                                        : positionSide === 'SHORT'
                                        ? 'text-red-600'
                                        : 'text-gray-500'
                                    }`}>
                                    {positionSide}
                                  </span>
                                </div>
                              )}
                              {/* Leverage */}
                              {leverage && (
                                <div className="flex items-center gap-1">
                                  <span className="text-xs text-gray-500">
                                    Leverage:
                                  </span>
                                  <span className="text-xs text-purple-600 font-medium">
                                    {leverage}x
                                  </span>
                                </div>
                              )}
                              {/* Margin Type */}
                              {marginType && (
                                <div className="flex items-center gap-1">
                                  <span className="text-xs text-gray-500">
                                    Margin:
                                  </span>
                                  <span
                                    className={`text-xs px-2 py-0.5 rounded font-medium ${
                                      isIsolated
                                        ? 'bg-orange-100 text-orange-700'
                                        : 'bg-blue-100 text-blue-700'
                                    }`}>
                                    {isIsolated ? 'Isolated' : 'Cross'}
                                  </span>
                                </div>
                              )}
                              {/* Liquidation Price */}
                              {liqPrice && liqPrice > 0 && (
                                <div className="flex items-center gap-1">
                                  <span className="text-xs text-gray-500">
                                    Liq:
                                  </span>
                                  <span className="text-red-600 font-semibold text-xs">
                                    ${liqPrice.toFixed(2)}
                                  </span>
                                </div>
                              )}
                            </div>
                          );
                        })()}
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex flex-col gap-1">
                          {/* Entry Price */}
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Entry:
                            </span>
                            <span className="text-sm font-bold text-gray-900">
                              $
                              {(order.filled_price || order.price || 0).toFixed(
                                5,
                              )}
                            </span>
                          </div>
                          {/* Current Price */}
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 font-medium">
                              Current:
                            </span>
                            {(() => {
                              const status = order.status?.toLowerCase();
                              const isFutures =
                                order.trading_mode?.toLowerCase() ===
                                  'futures' ||
                                order.trading_mode?.toLowerCase() === 'future';

                              // Futures: Chỉ dừng cập nhật khi status = 'closed'
                              // Spot: Dừng cập nhật khi status = 'filled' hoặc 'closed'
                              const shouldStopUpdating = isFutures
                                ? status === 'closed'
                                : ['filled', 'closed'].includes(status ?? '');

                              if (shouldStopUpdating && order.filled_price) {
                                return (
                                  <span className="text-sm font-bold text-green-700">
                                    ${order.filled_price.toFixed(5)}
                                  </span>
                                );
                              }

                              const priceKey = isFutures
                                ? `${order.symbol}_FUTURES`
                                : order.symbol;

                              if (realtimePrices[priceKey]) {
                                return (
                                  <span
                                    className={`text-sm font-bold animate-pulse ${
                                      realtimePrices[priceKey].percent >= 0
                                        ? 'text-green-600'
                                        : 'text-red-600'
                                    }`}>
                                    ${realtimePrices[priceKey].price.toFixed(5)}
                                  </span>
                                );
                              }

                              if (order.current_price) {
                                return (
                                  <span className="text-sm font-semibold text-blue-600">
                                    ${order.current_price.toFixed(5)}
                                  </span>
                                );
                              }

                              return (
                                <span className="text-sm text-gray-400">-</span>
                              );
                            })()}
                          </div>
                          {/* PnL / ROI */}
                          <div className="flex items-center gap-1 mt-1 pt-1 border-t border-gray-200">
                            <span className="text-xs text-gray-500 font-medium">
                              PnL:
                            </span>
                            {(() => {
                              // Lấy PnL từ data có sẵn (không tính toán)
                              const displayPnl = pnl;
                              const displayRoi = pnlPercent;

                              if (displayPnl === null && displayRoi === null) {
                                return (
                                  <span className="text-sm text-gray-400">
                                    -
                                  </span>
                                );
                              }

                              return (
                                <div className="flex items-center gap-1.5">
                                  {displayPnl !== null && (
                                    <span
                                      className={`text-sm font-semibold ${
                                        displayPnl >= 0
                                          ? 'text-green-600'
                                          : 'text-red-600'
                                      } ${isOpen ? 'animate-pulse' : ''}`}>
                                      {displayPnl >= 0 ? '+' : ''}$
                                      {displayPnl.toFixed(2)}
                                    </span>
                                  )}
                                  {displayRoi !== null && (
                                    <span
                                      className={`text-xs font-medium ${
                                        displayRoi >= 0
                                          ? 'text-green-700'
                                          : 'text-red-500'
                                      }`}>
                                      ({displayRoi >= 0 ? '+' : ''}
                                      {displayRoi.toFixed(2)}%)
                                    </span>
                                  )}
                                </div>
                              );
                            })()}
                          </div>
                        </div>
                      </td>

                      {/* SL/TP Column */}
                      <td className="px-6 py-4 text-sm">
                        {(() => {
                          const entryPrice = order.filled_price || order.price;
                          const hasSlPercent =
                            order.stop_loss_percent &&
                            order.stop_loss_percent > 0;
                          const hasTpPercent =
                            order.take_profit_percent &&
                            order.take_profit_percent > 0;

                          if (!hasSlPercent && !hasTpPercent) {
                            return <span className="text-gray-400">-</span>;
                          }

                          return (
                            <div className="space-y-1">
                              {hasSlPercent && (
                                <div className="text-red-600 font-medium text-xs whitespace-nowrap">
                                  SL: $
                                  {calculateStopLossPrice(
                                    entryPrice,
                                    order.stop_loss_percent!,
                                    order.side,
                                  ).toFixed(5)}{' '}
                                  <span className="text-red-500">
                                    (-{order.stop_loss_percent!.toFixed(2)}%)
                                  </span>
                                </div>
                              )}
                              {hasTpPercent && (
                                <div className="text-green-600 font-medium text-xs whitespace-nowrap">
                                  TP: $
                                  {calculateTakeProfitPrice(
                                    entryPrice,
                                    order.take_profit_percent!,
                                    order.side,
                                  ).toFixed(5)}{' '}
                                  <span className="text-green-700">
                                    (+{order.take_profit_percent!.toFixed(2)}%)
                                  </span>
                                </div>
                              )}
                            </div>
                          );
                        })()}
                      </td>

                      {/* Action Column */}
                      <td className="px-6 py-4 text-sm text-center">
                        {(() => {
                          const isSpot =
                            order.trading_mode?.toLowerCase() === 'spot';
                          const isFutures =
                            order.trading_mode?.toLowerCase() === 'futures';
                          const status = order.status?.toLowerCase();

                          // Enable button logic:
                          // - Spot: status = 'filled'
                          // - Futures: status = 'new'
                          const canClose =
                            (isSpot && status === 'filled') ||
                            (isFutures && status === 'new');

                          return (
                            <button
                              onClick={() => {
                                if (canClose) {
                                  setOrderToClose(order);
                                  setShowCloseModal(true);
                                }
                              }}
                              disabled={!canClose}
                              className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                                canClose
                                  ? 'bg-red-500 hover:bg-red-600 text-white cursor-pointer'
                                  : 'bg-gray-200 text-gray-400 cursor-not-allowed'
                              }`}>
                              Close
                            </button>
                          );
                        })()}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Close Order Confirmation Modal */}
      {showCloseModal && orderToClose && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
            <h3 className="text-lg font-bold text-gray-900 mb-4">
              Confirm Close Order
            </h3>

            <div className="mb-6 space-y-2 bg-gray-50 p-4 rounded">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Symbol:</span>
                <span className="font-medium">{orderToClose.symbol}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Mode:</span>
                <span className="font-medium">{orderToClose.trading_mode}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Side:</span>
                <span
                  className={`font-medium ${
                    orderToClose.side?.toLowerCase() === 'buy'
                      ? 'text-green-600'
                      : 'text-red-600'
                  }`}>
                  {orderToClose.side}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Amount:</span>
                <span className="font-medium">{orderToClose.quantity}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Entry Price:</span>
                <span className="font-medium">
                  $
                  {(
                    orderToClose.filled_price ||
                    orderToClose.price ||
                    0
                  ).toFixed(5)}
                </span>
              </div>
            </div>

            <div className="bg-yellow-50 border border-yellow-200 rounded p-3 mb-6">
              <p className="text-sm text-yellow-800">
                <strong>Warning:</strong> This will close your position
                immediately at market price. This action cannot be undone.
              </p>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowCloseModal(false);
                  setOrderToClose(null);
                }}
                disabled={closing}
                className="flex-1 px-4 py-2 border border-gray-300 rounded text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
                Cancel
              </button>
              <button
                onClick={async () => {
                  if (!orderToClose) return;

                  setClosing(true);
                  try {
                    // Call API to close all orders and position for this symbol
                    const result = await closeOrder(orderToClose.id);

                    // Refresh orders after successful close
                    await fetchOrders();

                    // Close modal first
                    setShowCloseModal(false);
                    setOrderToClose(null);

                    // Show success toast
                    setToast({
                      show: true,
                      message: result.message || 'Order closed successfully!',
                      type: 'success',
                    });

                    // Auto hide toast after 3 seconds
                    setTimeout(() => {
                      setToast({show: false, message: '', type: 'success'});
                    }, 3000);
                  } catch (err: any) {
                    console.error('Failed to close order:', err);

                    // Close modal
                    setShowCloseModal(false);
                    setOrderToClose(null);

                    // Show error toast with API error message if available
                    const errorMessage =
                      err.response?.data?.error ||
                      err.response?.data?.details ||
                      'Failed to close order. Please try again.';
                    setToast({
                      show: true,
                      message: errorMessage,
                      type: 'error',
                    });

                    // Auto hide toast after 3 seconds
                    setTimeout(() => {
                      setToast({show: false, message: '', type: 'error'});
                    }, 3000);
                  } finally {
                    setClosing(false);
                  }
                }}
                disabled={closing}
                className="flex-1 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium">
                {closing ? 'Closing...' : 'Close Order'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Toast Notification */}
      {toast.show && (
        <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 animate-slideUp">
          <div
            className={`px-6 py-4 rounded-lg shadow-lg flex items-center gap-3 ${
              toast.type === 'success'
                ? 'bg-green-500 text-white'
                : 'bg-red-500 text-white'
            }`}>
            {toast.type === 'success' ? (
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            ) : (
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            )}
            <span className="font-medium">{toast.message}</span>
          </div>
        </div>
      )}
    </div>
  );
}
