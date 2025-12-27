'use client';

import {useState, useEffect} from 'react';
import {getSignals, Signal} from '@/services/adminService';

export default function SignalsPage() {
  const [signals, setSignals] = useState<Signal[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [filterAction, setFilterAction] = useState('all');

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const data = await getSignals({limit: 100});
      setSignals(data.signals || []);
    } catch (error) {
      console.error('Error fetching signals:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredSignals = signals.filter((signal) => {
    const matchesSearch =
      signal.symbol?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      signal.strategy?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      signal.webhook_prefix?.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesStatus =
      filterStatus === 'all' || signal.status === filterStatus;
    const matchesAction =
      filterAction === 'all' || signal.action === filterAction;

    return matchesSearch && matchesStatus && matchesAction;
  });

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'executed':
        return 'bg-green-100 text-green-700 border border-green-300';
      case 'pending':
        return 'bg-yellow-100 text-yellow-700 border border-yellow-300';
      case 'failed':
        return 'bg-red-100 text-red-700 border border-red-300';
      case 'ignored':
        return 'bg-gray-100 text-gray-700 border border-gray-300';
      default:
        return 'bg-blue-100 text-blue-700 border border-blue-300';
    }
  };

  const getActionColor = (action: string) => {
    switch (action.toLowerCase()) {
      case 'buy':
      case 'long':
        return 'text-green-600';
      case 'sell':
      case 'short':
        return 'text-red-600';
      case 'close':
        return 'text-orange-600';
      default:
        return 'text-gray-600';
    }
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
      second: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-400 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading signals...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">
          Trading Signals
        </h1>
        <p className="text-gray-500">
          Monitor all trading signals from TradingView
        </p>
      </div>

      {/* Filters */}
      <div className="mb-6 bg-white rounded-lg shadow-md p-6 border-t-4 border-orange-400">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <input
            type="text"
            placeholder="Search by symbol, strategy, or prefix..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline-none focus:border-orange-500 transition-colors"
          />

          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-orange-500 transition-colors">
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="executed">Executed</option>
            <option value="failed">Failed</option>
            <option value="ignored">Ignored</option>
          </select>

          <select
            value={filterAction}
            onChange={(e) => setFilterAction(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-orange-500 transition-colors">
            <option value="all">All Actions</option>
            <option value="buy">Buy</option>
            <option value="sell">Sell</option>
            <option value="close">Close</option>
            <option value="long">Long</option>
            <option value="short">Short</option>
          </select>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-orange-400">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Total Signals
          </div>
          <div className="text-3xl font-bold text-gray-900">
            {signals.length}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-green-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">Executed</div>
          <div className="text-3xl font-bold text-green-600">
            {signals.filter((s) => s.status === 'executed').length}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-yellow-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">Pending</div>
          <div className="text-3xl font-bold text-yellow-600">
            {signals.filter((s) => s.status === 'pending').length}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-red-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">Failed</div>
          <div className="text-3xl font-bold text-red-600">
            {signals.filter((s) => s.status === 'failed').length}
          </div>
        </div>
      </div>

      {/* Signals Table */}
      <div className="bg-white rounded-lg shadow-md border-t-4 border-orange-400">
        <div className="overflow-x-auto">
          <table className="w-full table-auto">
            <thead className="bg-gradient-to-r from-orange-50 to-orange-100">
              <tr className="border-b-2 border-orange-400">
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  ID
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Received At
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Prefix
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Symbol
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Action
                </th>
                <th className="px-4 py-4 text-right text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Price
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {filteredSignals.length === 0 ? (
                <tr>
                  <td
                    colSpan={6}
                    className="px-6 py-8 text-center text-gray-500">
                    No signals found
                  </td>
                </tr>
              ) : (
                filteredSignals.map((signal) => (
                  <tr
                    key={signal.id}
                    className="hover:bg-orange-50 transition-colors">
                    <td className="px-4 py-4 text-sm text-gray-900 font-medium">
                      #{signal.id}
                    </td>
                    <td className="px-4 py-4 text-sm text-gray-600">
                      {formatDate(signal.received_at)}
                    </td>
                    <td className="px-4 py-4 text-sm text-gray-600">
                      <span className="px-2 py-1 bg-purple-100 text-purple-700 rounded text-xs font-mono">
                        {signal.webhook_prefix || 'default'}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-gray-900 font-semibold">
                      {signal.symbol}
                    </td>
                    <td className="px-4 py-4">
                      <span
                        className={`font-semibold uppercase ${getActionColor(
                          signal.action,
                        )}`}>
                        {signal.action}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-right text-gray-700">
                      ${formatNumber(signal.price, 4)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination info */}
        <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
          <p className="text-sm text-gray-600">
            Showing {filteredSignals.length} of {signals.length} signals
          </p>
        </div>
      </div>
    </div>
  );
}
