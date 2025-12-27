'use client';

import {useState, useEffect} from 'react';
import {getOrders, Order} from '@/services/adminService';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [filterExchange, setFilterExchange] = useState('all');
  const [filterMode, setFilterMode] = useState('all');

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const ordersData = await getOrders();

      if (ordersData.success) {
        setOrders(ordersData.orders || []);
      }
    } catch (error) {
      console.error('Error fetching data:', error);
    } finally {
      setLoading(false);
    }
  };

  const getUserInfo = (order: Order) => {
    return {
      email: order.user_email || 'Unknown',
      name: order.user_full_name || 'Unknown',
    };
  };

  const filteredOrders = orders.filter((order) => {
    const userInfo = getUserInfo(order);
    const matchesSearch =
      order.symbol?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      order.order_id?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      userInfo.email.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesStatus =
      filterStatus === 'all' || order.status === filterStatus;
    const matchesExchange =
      filterExchange === 'all' ||
      order.exchange.toLowerCase() === filterExchange.toLowerCase();
    const matchesMode =
      filterMode === 'all' ||
      order.trading_mode?.toLowerCase() === filterMode.toLowerCase();

    return matchesSearch && matchesStatus && matchesExchange && matchesMode;
  });

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'filled':
      case 'closed':
        return 'bg-green-100 text-green-700 border border-green-300';
      case 'pending':
      case 'new':
        return 'bg-yellow-100 text-yellow-700 border border-yellow-300';
      case 'partially_filled':
        return 'bg-blue-100 text-blue-700 border border-blue-300';
      case 'cancelled':
      case 'rejected':
      case 'expired':
        return 'bg-red-100 text-red-700 border border-red-300';
      default:
        return 'bg-gray-100 text-gray-700 border border-gray-300';
    }
  };

  const getSideColor = (side: string) => {
    return side.toLowerCase() === 'buy' || side.toLowerCase() === 'long'
      ? 'text-green-600'
      : 'text-red-600';
  };

  const formatNumber = (num: number, decimals: number = 8) => {
    if (!num) return '0';
    return num.toFixed(decimals).replace(/\.?0+$/, '');
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#EE4D2D] mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading orders...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">
          Orders Management
        </h1>
        <p className="text-gray-500">Monitor and manage all trading orders</p>
      </div>

      {/* Filters */}
      <div className="mb-6 bg-white rounded-lg shadow-md p-6 border-t-4 border-orange-400">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <input
            type="text"
            placeholder="Search by symbol, order ID, or user email..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline-none focus:border-orange-500 transition-colors"
          />

          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-orange-500 transition-colors">
            <option value="all">All Status</option>
            <option value="new">New</option>
            <option value="pending">Pending</option>
            <option value="partially_filled">Partially Filled</option>
            <option value="filled">Filled</option>
            <option value="closed">Closed</option>
            <option value="cancelled">Cancelled</option>
            <option value="rejected">Rejected</option>
          </select>

          <select
            value={filterExchange}
            onChange={(e) => setFilterExchange(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-orange-500 transition-colors">
            <option value="all">All Exchanges</option>
            <option value="binance">Binance</option>
            <option value="bingx">BingX</option>
          </select>

          <select
            value={filterMode}
            onChange={(e) => setFilterMode(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-orange-500 transition-colors">
            <option value="all">All Modes</option>
            <option value="spot">Spot</option>
            <option value="futures">Futures</option>
            <option value="margin">Margin</option>
          </select>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-orange-400">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Total Orders
          </div>
          <div className="text-3xl font-bold text-gray-900">
            {orders.length}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-green-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Filled Orders
          </div>
          <div className="text-3xl font-bold text-green-600">
            {
              orders.filter(
                (o) => o.status === 'filled' || o.status === 'closed',
              ).length
            }
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-yellow-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Pending Orders
          </div>
          <div className="text-3xl font-bold text-yellow-600">
            {
              orders.filter(
                (o) =>
                  o.status === 'pending' ||
                  o.status === 'new' ||
                  o.status === 'partially_filled',
              ).length
            }
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-red-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Failed Orders
          </div>
          <div className="text-3xl font-bold text-red-600">
            {
              orders.filter(
                (o) =>
                  o.status === 'cancelled' ||
                  o.status === 'rejected' ||
                  o.status === 'expired',
              ).length
            }
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-purple-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Total PnL
          </div>
          <div
            className={`text-3xl font-bold ${
              orders.reduce((sum, o) => sum + (o.pnl || 0), 0) >= 0
                ? 'text-green-600'
                : 'text-red-600'
            }`}>
            {orders.reduce((sum, o) => sum + (o.pnl || 0), 0) >= 0 ? '+' : ''}$
            {formatNumber(
              orders.reduce((sum, o) => sum + (o.pnl || 0), 0),
              2,
            )}
          </div>
          <div
            className={`text-sm mt-1 ${
              orders.reduce((sum, o) => sum + (o.pnl_percent || 0), 0) >= 0
                ? 'text-green-500'
                : 'text-red-500'
            }`}>
            {orders.reduce((sum, o) => sum + (o.pnl_percent || 0), 0) >= 0
              ? '+'
              : ''}
            {formatNumber(
              orders.reduce((sum, o) => sum + (o.pnl_percent || 0), 0) /
                orders.filter((o) => o.pnl !== 0).length || 0,
              2,
            )}
            % avg
          </div>
        </div>
      </div>

      {/* Orders Table - Scrollable */}
      <div className="bg-white rounded-lg shadow-md border-t-4 border-orange-400">
        <div className="overflow-x-auto">
          <table className="w-full table-auto">
            <thead className="bg-gradient-to-r from-orange-50 to-orange-100">
              <tr className="border-b-2 border-orange-400">
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  User
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Exchange
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Symbol
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Side
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-4 py-4 text-right text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Quantity
                </th>
                <th className="px-4 py-4 text-right text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Price
                </th>
                <th className="px-4 py-4 text-right text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Filled
                </th>
                <th className="px-4 py-4 text-right text-xs font-medium text-orange-600 uppercase tracking-wider">
                  PnL
                </th>
                <th className="px-4 py-4 text-center text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Created
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {filteredOrders.length === 0 ? (
                <tr>
                  <td
                    colSpan={11}
                    className="px-6 py-8 text-center text-gray-500">
                    No orders found
                  </td>
                </tr>
              ) : (
                filteredOrders.map((order) => {
                  const userInfo = getUserInfo(order);
                  return (
                    <tr
                      key={order.id}
                      className="hover:bg-orange-50 transition-colors">
                      <td className="px-4 py-4">
                        <div className="text-sm">
                          <div className="text-gray-900 font-medium">
                            {userInfo.name}
                          </div>
                          <div className="text-gray-500 text-xs">
                            {userInfo.email}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="text-gray-900 font-medium uppercase">
                            {order.exchange}
                          </span>
                          {order.trading_mode && (
                            <span className="text-xs px-2 py-0.5 bg-purple-100 text-purple-700 rounded border border-purple-300 font-medium">
                              {order.trading_mode.toUpperCase()}
                            </span>
                          )}
                          {order.leverage > 1 && (
                            <span className="text-xs px-2 py-0.5 bg-gradient-to-r from-orange-300 to-orange-400 text-white rounded font-medium shadow-sm">
                              {order.leverage}x
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-4 text-gray-900 font-semibold">
                        {order.symbol}
                      </td>
                      <td className="px-4 py-4">
                        <span
                          className={`font-semibold uppercase ${getSideColor(
                            order.side,
                          )}`}>
                          {order.side}
                        </span>
                      </td>
                      <td className="px-4 py-4 text-gray-700">{order.type}</td>
                      <td className="px-4 py-4 text-right text-gray-700">
                        {formatNumber(order.quantity, 6)}
                      </td>
                      <td className="px-4 py-4 text-right text-gray-700">
                        ${formatNumber(order.price, 4)}
                      </td>
                      <td className="px-4 py-4 text-right">
                        <div className="text-gray-700">
                          {formatNumber(order.filled_quantity || 0, 6)}
                          <span className="text-xs text-gray-500 ml-1">
                            (
                            {formatNumber(
                              (order.filled_quantity / order.quantity) * 100,
                              2,
                            )}
                            %)
                          </span>
                        </div>
                        {order.filled_price > 0 && (
                          <div className="text-xs text-gray-500">
                            @ ${formatNumber(order.filled_price, 4)}
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-4 text-right">
                        {order.pnl !== 0 && (
                          <div>
                            <div
                              className={`font-semibold ${
                                order.pnl > 0
                                  ? 'text-green-600'
                                  : 'text-red-600'
                              }`}>
                              {order.pnl > 0 ? '+' : ''}
                              {formatNumber(order.pnl, 2)}
                            </div>
                            <div
                              className={`text-xs ${
                                order.pnl_percent > 0
                                  ? 'text-green-600'
                                  : 'text-red-600'
                              }`}>
                              ({order.pnl_percent > 0 ? '+' : ''}
                              {formatNumber(order.pnl_percent, 2)}%)
                            </div>
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-4 text-center">
                        <span
                          className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(
                            order.status,
                          )}`}>
                          {order.status.replace('_', ' ').toUpperCase()}
                        </span>
                      </td>
                      <td className="px-4 py-4 text-sm text-gray-600">
                        {formatDate(order.created_at)}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination info - Fixed at bottom */}
        <div className="flex-shrink-0 px-6 py-3 bg-gray-50 border-t border-gray-200 text-sm text-gray-600 text-center">
          Showing {filteredOrders.length} of {orders.length} orders
        </div>
      </div>
    </div>
  );
}
