'use client';

import {useState, useEffect} from 'react';
import {Order, getOrderHistory} from '../../services/orderService';
import {refreshPnL} from '../../services/tradingService';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshingPnL, setRefreshingPnL] = useState<number | null>(null);

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

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      fetchOrders();
    } else {
      setError('You must be logged in to view this page.');
      setLoading(false);
    }
  }, [filters]);

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
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Orders</h1>

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
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {orders.map((order) => (
                  <tr key={order.id} className="hover:bg-gray-50">
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
                          order.side === 'buy'
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
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-red-600">
                      {order.stop_loss_price
                        ? order.stop_loss_price.toFixed(5)
                        : '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-green-600">
                      {order.take_profit_price
                        ? order.take_profit_price.toFixed(5)
                        : '-'}
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
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
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
                    </td>
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
