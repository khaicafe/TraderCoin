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
    pending: 0,
    cancelled: 0,
  });

  // Filters state
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

      // Calculate stats
      const total = data.length;
      const filled = data.filter((o) => o.status === 'filled').length;
      const pending = data.filter((o) => o.status === 'pending').length;
      const cancelled = data.filter((o) => o.status === 'cancelled').length;

      setStats({total, filled, pending, cancelled});
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
      const filled = data.filter((o) => o.status === 'filled').length;
      const pending = data.filter((o) => o.status === 'pending').length;
      const cancelled = data.filter((o) => o.status === 'cancelled').length;
      setStats({total, filled, pending, cancelled});
    } catch (err) {
      // ignore transient errors
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      fetchOrders();

      // Connect to WebSocket
      websocketService.connect();

      // Update connection status periodically
      const statusInterval = setInterval(() => {
        setWsStatus(websocketService.getConnectionState());
      }, 1000);

      // Fallback: poll order statuses every 5s if no realtime update comes
      const ordersRefreshInterval = setInterval(() => {
        refreshOrdersLight();
      }, 5000);

      // Subscribe to order updates
      const unsubscribeOrders = websocketService.onOrderUpdate(
        (update: OrderUpdate) => {
          console.log('Order update received:', update);

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
              fetchOrders();
              return prevOrders;
            }
          });

          // Update stats
          setStats((prevStats) => {
            const updatedOrders = orders.map((o) =>
              o.order_id === update.order_id
                ? {...o, status: update.status.toLowerCase()}
                : o,
            );

            return {
              total: updatedOrders.length,
              filled: updatedOrders.filter((o) => o.status === 'filled').length,
              pending: updatedOrders.filter((o) => o.status === 'pending')
                .length,
              cancelled: updatedOrders.filter((o) => o.status === 'cancelled')
                .length,
            };
          });
        },
      );

      // Cleanup
      return () => {
        unsubscribeOrders();
        clearInterval(statusInterval);
        clearInterval(ordersRefreshInterval);
        websocketService.disconnect();
      };
    } else {
      setError('You must be logged in to view this page.');
      setLoading(false);
    }
  }, [filters]);

  // Fetch realtime prices directly from Binance when orders change
  useEffect(() => {
    if (orders.length === 0) return;

    let cancelled = false;

    const symbols = Array.from(new Set(orders.map((o) => o.symbol)));

    const fetchRealtimePrices = async () => {
      for (const symbol of symbols) {
        try {
          const response = await fetch(
            `https://testnet.binance.vision/api/v3/ticker/24hr?symbol=${symbol}`,
          );
          if (!response.ok) continue;
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
          // silent - network errors are okay
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

  const handleRefreshPnL = async (orderId: number) => {
    try {
      setRefreshingPnL(orderId);
      const result = await refreshPnL(orderId);

      // Update the order in the list with new PnL values
      setOrders((prevOrders) =>
        prevOrders.map((order) =>
          order.id === orderId
            ? {
                ...order,
                pnl: result.pnl,
                pnl_percent: result.pnl_percent,
              }
            : order,
        ),
      );

      console.log('PnL refreshed:', result);
    } catch (err: any) {
      console.error('Error refreshing PnL:', err);
      alert(err.response?.data?.error || 'Failed to refresh PnL');
    } finally {
      setRefreshingPnL(null);
    }
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
          <p className="text-3xl font-bold mt-2">{stats.pending}</p>
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
              <option value="pending">Pending</option>
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
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
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
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {order.quantity}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {order.filled_price
                        ? order.filled_price.toFixed(5)
                        : order.price.toFixed(5)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {realtimePrices[order.symbol] ? (
                        <div className="flex flex-col">
                          <span
                            className={`font-bold text-base ${
                              realtimePrices[order.symbol].percent >= 0
                                ? 'text-green-600'
                                : 'text-red-600'
                            } animate-pulse`}>
                            ${realtimePrices[order.symbol].price.toFixed(5)}
                          </span>
                          <span
                            className={`text-xs ${
                              realtimePrices[order.symbol].percent >= 0
                                ? 'text-green-500'
                                : 'text-red-500'
                            }`}>
                            {realtimePrices[order.symbol].percent >= 0
                              ? '+'
                              : ''}
                            {realtimePrices[order.symbol].percent.toFixed(2)}%
                          </span>
                        </div>
                      ) : order.current_price ? (
                        <span className="font-semibold text-blue-600">
                          ${order.current_price.toFixed(5)}
                        </span>
                      ) : order.filled_price ? (
                        <span className="font-medium text-gray-700">
                          {order.filled_price.toFixed(5)}
                        </span>
                      ) : order.price ? (
                        <span className="font-medium text-gray-600">
                          {order.price.toFixed(5)}
                        </span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
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
                          order.status === 'filled' || order.status === 'closed'
                            ? 'bg-green-100 text-green-800'
                            : order.status === 'pending'
                            ? 'bg-yellow-100 text-yellow-800'
                            : order.status === 'cancelled'
                            ? 'bg-red-100 text-red-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}>
                        {order.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {order.pnl !== undefined && order.pnl !== null ? (
                        <span
                          className={`font-semibold ${
                            order.pnl >= 0 ? 'text-green-600' : 'text-red-600'
                          }`}>
                          {order.pnl >= 0 ? '+' : ''}
                          {order.pnl.toFixed(2)}
                        </span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {order.pnl_percent !== undefined &&
                      order.pnl_percent !== null ? (
                        <span
                          className={`font-semibold ${
                            order.pnl_percent >= 0
                              ? 'text-green-600'
                              : 'text-red-600'
                          }`}>
                          {order.pnl_percent >= 0 ? '+' : ''}
                          {order.pnl_percent.toFixed(2)}%
                        </span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
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
