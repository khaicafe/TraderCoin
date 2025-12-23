'use client';
import {useState, useEffect} from 'react';
import {Order, getOrderHistory} from '../../services/orderService';
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
      isolated_margin: number;
    };
    timestamp?: number;
  };
}

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [wsStatus, setWsStatus] = useState<string>('DISCONNECTED');

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
  });

  // Filters
  const [filters, setFilters] = useState({
    symbol: '',
    status: '',
    side: '',
  });

  const fetchOrdersBK = async () => {
    try {
      setLoading(true);
      const params: any = {limit: 100, offset: 0};
      if (filters.symbol) params.symbol = filters.symbol;
      if (filters.status) params.status = filters.status;
      if (filters.side) params.side = filters.side;

      const data = await getOrderHistory(params);
      setOrders(data);

      // Calculate stats
      const total = data.length;
      const filled = data.filter((o) =>
        ['filled', 'closed'].includes(o.status?.toLowerCase() ?? ''),
      ).length;
      const New = data.filter((o) =>
        ['open', 'new', 'pending', 'partially_filled'].includes(
          o.status?.toLowerCase() ?? '',
        ),
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

  const refreshOrdersLight = async () => {
    try {
      const params: any = {limit: 100, offset: 0};
      if (filters.symbol) params.symbol = filters.symbol;
      if (filters.status) params.status = filters.status;
      if (filters.side) params.side = filters.side;

      const data = await getOrderHistory(params);
      setOrders(data);

      const total = data.length;
      const filled = data.filter((o) =>
        ['filled', 'closed'].includes(o.status?.toLowerCase() ?? ''),
      ).length;
      const New = data.filter((o) =>
        ['open', 'new', 'pending', 'partially_filled'].includes(
          o.status?.toLowerCase() ?? '',
        ),
      ).length;
      const cancelled = data.filter(
        (o) => o.status?.toLowerCase() === 'cancelled',
      ).length;

      setStats({total, filled, New, cancelled});
    } catch (err) {
      // Silent fail on refresh
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
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
                    isolated_margin: String(
                      data.position.isolated_margin || '0',
                    ),
                    position_side: data.position.position_side || '',
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
    } else {
      setError('You must be logged in to view this page.');
      setLoading(false);
    }
  }, [filters]);

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

    // Spot WebSocket
    if (spotSymbols.length > 0) {
      const streams = spotSymbols.map((s) => `${s}@ticker`).join('/');
      spotWs = new WebSocket(`wss://stream.binance.com:9443/ws/${streams}`);

      spotWs.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.e === '24hrTicker') {
          setRealtimePrices((prev) => ({
            ...prev,
            [data.s.toUpperCase()]: {
              price: parseFloat(data.c),
              change: parseFloat(data.p),
              percent: parseFloat(data.P),
            },
          }));
        }
      };
    }

    // Futures WebSocket - CHÍNH CHO FUTURES
    if (futuresSymbols.length > 0) {
      const streams = futuresSymbols.map((s) => `${s}@ticker`).join('/');
      futuresWs = new WebSocket(`wss://fstream.binance.com/ws/${streams}`);

      futuresWs.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.e === '24hrTicker') {
          setRealtimePrices((prev) => ({
            ...prev,
            [`${data.s.toUpperCase()}_FUTURES`]: {
              price: parseFloat(data.c),
              change: parseFloat(data.p),
              percent: parseFloat(data.P),
            },
          }));
        }
      };

      futuresWs.onerror = (err) => console.warn('Futures WS error:', err);
      futuresWs.onclose = () => console.log('Futures WS closed');
    }

    return () => {
      spotWs?.close();
      futuresWs?.close();
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

  const calculatePnL = (
    order: Order,
    currentPrice: number | null,
  ): number | null => {
    if (!currentPrice || !order.quantity) return null;
    const entryPrice = order.filled_price || order.price;
    if (!entryPrice || entryPrice === 0) return null;
    const quantity = order.quantity;
    const side = order.side?.toLowerCase();

    return side === 'buy'
      ? (currentPrice - entryPrice) * quantity
      : (entryPrice - currentPrice) * quantity;
  };

  const calculateROI = (order: Order, pnl: number | null): number | null => {
    if (pnl === null || !order.quantity) return null;
    const entryPrice = order.filled_price || order.price;
    if (!entryPrice || entryPrice === 0) return null;
    const investment = entryPrice * order.quantity;
    return investment === 0 ? null : (pnl / investment) * 100;
  };

  const getCurrentPriceForPnL = (order: Order): number | null => {
    const status = order.status?.toLowerCase();
    const isFutures =
      order.trading_mode?.toLowerCase() === 'futures' ||
      order.trading_mode?.toLowerCase() === 'future';

    // Futures: Chỉ dừng cập nhật khi status = 'closed'
    // Spot: Dừng cập nhật khi status = 'filled' hoặc 'closed'
    const shouldStopUpdating = isFutures
      ? status === 'closed'
      : ['filled', 'closed'].includes(status ?? '');

    if (shouldStopUpdating) {
      return order.filled_price || null;
    }

    const priceKey = isFutures ? `${order.symbol}_FUTURES` : order.symbol;

    if (realtimePrices[priceKey]) {
      return realtimePrices[priceKey].price;
    }
    if (order.current_price) {
      return order.current_price;
    }
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
                    Order ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Bot
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Symbol
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Mode
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Side
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Type
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Amount
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Entry
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Current
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Position
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Liq Price
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    SL / TP
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    PnL / ROI
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Created
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {orders.map((order) => {
                  const currentPrice = getCurrentPriceForPnL(order);
                  const pnl = calculatePnL(order, currentPrice);
                  const roi = calculateROI(order, pnl);
                  const isOpen = [
                    'new',
                    'open',
                    'pending',
                    'partially_filled',
                  ].includes(order.status?.toLowerCase() ?? '');

                  return (
                    <tr key={order.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 text-sm">{order.id}</td>
                      <td className="px-6 py-4 text-sm font-mono text-blue-600">
                        {order.order_id || '-'}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        {order.bot_config_name || '-'}
                      </td>
                      <td className="px-6 py-4 text-sm font-medium">
                        {formatSymbol(order.symbol)}
                      </td>
                      <td className="px-6 py-4">
                        <span className="px-2 py-1 text-xs font-semibold rounded bg-yellow-100 text-yellow-800 capitalize">
                          {order.trading_mode || 'spot'}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={`px-2 py-1 text-xs font-semibold rounded uppercase ${
                            order.side === 'BUY'
                              ? 'bg-green-100 text-green-800'
                              : 'bg-red-100 text-red-800'
                          }`}>
                          {order.side}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm capitalize">
                        {order.type}
                      </td>
                      <td className="px-6 py-4 text-sm font-bold">
                        {order.quantity}
                      </td>
                      <td className="px-6 py-4 text-sm font-bold">
                        {(order.filled_price || order.price || 0).toFixed(5)}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        {(() => {
                          const status = order.status?.toLowerCase();
                          const isFutures =
                            order.trading_mode?.toLowerCase() === 'futures' ||
                            order.trading_mode?.toLowerCase() === 'future';

                          // Futures: Chỉ dừng cập nhật khi status = 'closed'
                          // Spot: Dừng cập nhật khi status = 'filled' hoặc 'closed'
                          const shouldStopUpdating = isFutures
                            ? status === 'closed'
                            : ['filled', 'closed'].includes(status ?? '');

                          if (shouldStopUpdating && order.filled_price) {
                            return (
                              <span className="font-bold text-green-700">
                                ${order.filled_price.toFixed(5)}
                              </span>
                            );
                          }

                          const priceKey = isFutures
                            ? `${order.symbol}_FUTURES`
                            : order.symbol;

                          if (realtimePrices[priceKey]) {
                            return (
                              <div className="flex flex-col">
                                <span
                                  className={`font-bold text-base animate-pulse ${
                                    realtimePrices[priceKey].percent >= 0
                                      ? 'text-green-600'
                                      : 'text-red-600'
                                  }`}>
                                  ${realtimePrices[priceKey].price.toFixed(5)}
                                </span>
                                {/* <span
                                  className={`text-xs ${
                                    realtimePrices[priceKey].percent >= 0
                                      ? 'text-green-500'
                                      : 'text-red-500'
                                  }`}>
                                  {realtimePrices[priceKey].percent >= 0
                                    ? '+'
                                    : ''}
                                  {realtimePrices[priceKey].percent.toFixed(2)}%
                                </span> */}
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

                          return <span className="text-gray-400">-</span>;
                        })()}
                      </td>
                      {/* Position Info - NEW */}
                      <td className="px-6 py-4 text-sm">
                        {order.position ? (
                          <div className="space-y-1">
                            <div
                              className={`text-xs font-semibold ${
                                parseFloat(order.position.position_amt || '0') >
                                0
                                  ? 'text-green-600'
                                  : parseFloat(
                                      order.position.position_amt || '0',
                                    ) < 0
                                  ? 'text-red-600'
                                  : 'text-gray-500'
                              }`}>
                              {parseFloat(order.position.position_amt || '0') >
                              0
                                ? 'LONG'
                                : parseFloat(
                                    order.position.position_amt || '0',
                                  ) < 0
                                ? 'SHORT'
                                : 'NONE'}{' '}
                              {Math.abs(
                                parseFloat(order.position.position_amt || '0'),
                              )}
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
                      {/* Liquidation Price - NEW */}
                      <td className="px-6 py-4 text-sm">
                        {order.position?.liquidation_price &&
                        parseFloat(order.position.liquidation_price) > 0 ? (
                          <span className="text-red-600 font-semibold text-xs">
                            $
                            {parseFloat(
                              order.position.liquidation_price,
                            ).toFixed(2)}
                          </span>
                        ) : (
                          <span className="text-gray-400">-</span>
                        )}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        {order.stop_loss_price || order.take_profit_price ? (
                          <div className="space-y-1">
                            {order.stop_loss_price && (
                              <div className="text-red-600 font-medium text-xs whitespace-nowrap">
                                SL: ${order.stop_loss_price.toFixed(2)}{' '}
                                {calcTargetPercent(
                                  order.stop_loss_price,
                                  order.filled_price || order.price,
                                  order.side,
                                ) !== null && (
                                  <span>
                                    (
                                    {calcTargetPercent(
                                      order.stop_loss_price,
                                      order.filled_price || order.price,
                                      order.side,
                                    )?.toFixed(2)}
                                    %)
                                  </span>
                                )}
                              </div>
                            )}
                            {order.take_profit_price && (
                              <div className="text-green-600 font-medium text-xs whitespace-nowrap">
                                TP: ${order.take_profit_price.toFixed(2)}{' '}
                                {calcTargetPercent(
                                  order.take_profit_price,
                                  order.filled_price || order.price,
                                  order.side,
                                ) !== null && (
                                  <span>
                                    (+
                                    {calcTargetPercent(
                                      order.take_profit_price,
                                      order.filled_price || order.price,
                                      order.side,
                                    )?.toFixed(2)}
                                    %)
                                  </span>
                                )}
                              </div>
                            )}
                          </div>
                        ) : (
                          '-'
                        )}
                      </td>
                      <td className="px-6 py-4">
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
                      </td>
                      <td className="px-6 py-4 text-sm">
                        {(() => {
                          // Ưu tiên dùng position data nếu có
                          const positionPnl = order.position?.unrealized_profit
                            ? parseFloat(order.position.unrealized_profit)
                            : null;
                          const positionRoi = order.position?.pnl_percent
                            ? parseFloat(order.position.pnl_percent)
                            : null;

                          const displayPnl =
                            positionPnl !== null ? positionPnl : pnl;
                          const displayRoi =
                            positionRoi !== null ? positionRoi : roi;

                          if (displayPnl === null && displayRoi === null) {
                            return <span className="text-gray-400">-</span>;
                          }

                          return (
                            <div className="flex flex-col">
                              {displayPnl !== null && (
                                <span
                                  className={`font-semibold text-base ${
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
                                      ? 'text-green-500'
                                      : 'text-red-500'
                                  }`}>
                                  {displayRoi >= 0 ? '+' : ''}
                                  {displayRoi.toFixed(2)}%
                                </span>
                              )}
                            </div>
                          );
                        })()}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        {formatDate(order.created_at)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
