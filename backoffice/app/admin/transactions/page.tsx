'use client';

import {useState, useEffect} from 'react';
import {getTransactions, Transaction} from '@/services/adminService';

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [filterStatus, setFilterStatus] = useState('all');

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const data = await getTransactions();
      setTransactions(data || []);
    } catch (error) {
      console.error('Error fetching transactions:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredTransactions = transactions.filter((tx) => {
    const matchesSearch =
      tx.user_email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      tx.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      tx.id.toString().includes(searchTerm);

    const matchesType = filterType === 'all' || tx.type === filterType;
    const matchesStatus = filterStatus === 'all' || tx.status === filterStatus;

    return matchesSearch && matchesType && matchesStatus;
  });

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'completed':
      case 'success':
        return 'bg-green-100 text-green-700 border border-green-300';
      case 'pending':
        return 'bg-yellow-100 text-yellow-700 border border-yellow-300';
      case 'failed':
      case 'cancelled':
        return 'bg-red-100 text-red-700 border border-red-300';
      default:
        return 'bg-gray-100 text-gray-700 border border-gray-300';
    }
  };

  const getTypeColor = (type: string) => {
    switch (type.toLowerCase()) {
      case 'deposit':
      case 'top-up':
        return 'text-green-600';
      case 'withdrawal':
        return 'text-red-600';
      case 'payment':
        return 'text-orange-600';
      case 'refund':
        return 'text-blue-600';
      default:
        return 'text-gray-600';
    }
  };

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(num);
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

  // Calculate totals
  const totalDeposits = transactions
    .filter((tx) => tx.type === 'deposit' && tx.status === 'completed')
    .reduce((sum, tx) => sum + tx.amount, 0);

  const totalWithdrawals = transactions
    .filter((tx) => tx.type === 'withdrawal' && tx.status === 'completed')
    .reduce((sum, tx) => sum + Math.abs(tx.amount), 0);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-400 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading transactions...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">
          Transactions Management
        </h1>
        <p className="text-gray-500">
          Monitor all user transactions and payments
        </p>
      </div>

      {/* Filters */}
      <div className="mb-6 bg-white rounded-lg shadow-md p-6 border-t-4 border-orange-400">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <input
            type="text"
            placeholder="Search by email, description, or ID..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline-none focus:border-orange-500 transition-colors"
          />

          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-orange-500 transition-colors">
            <option value="all">All Types</option>
            <option value="deposit">Deposit</option>
            <option value="withdrawal">Withdrawal</option>
            <option value="payment">Payment</option>
            <option value="refund">Refund</option>
          </select>

          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-4 py-2 bg-white border-2 border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-orange-500 transition-colors">
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-orange-400">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Total Transactions
          </div>
          <div className="text-3xl font-bold text-gray-900">
            {transactions.length}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-green-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Total Deposits
          </div>
          <div className="text-3xl font-bold text-green-600">
            ${formatNumber(totalDeposits)}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-red-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Total Withdrawals
          </div>
          <div className="text-3xl font-bold text-red-600">
            ${formatNumber(totalWithdrawals)}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border-t-4 border-purple-500">
          <div className="text-gray-500 text-sm mb-2 font-medium">
            Net Balance
          </div>
          <div
            className={`text-3xl font-bold ${
              totalDeposits - totalWithdrawals >= 0
                ? 'text-green-600'
                : 'text-red-600'
            }`}>
            ${formatNumber(totalDeposits - totalWithdrawals)}
          </div>
        </div>
      </div>

      {/* Transactions Table */}
      <div className="bg-white rounded-lg shadow-md border-t-4 border-orange-400">
        <div className="overflow-x-auto">
          <table className="w-full table-auto">
            <thead className="bg-gradient-to-r from-orange-50 to-orange-100">
              <tr className="border-b-2 border-orange-400">
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  ID
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Date
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  User
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-4 py-4 text-right text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Amount
                </th>
                <th className="px-4 py-4 text-center text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-4 text-left text-xs font-medium text-orange-600 uppercase tracking-wider">
                  Description
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {filteredTransactions.length === 0 ? (
                <tr>
                  <td
                    colSpan={7}
                    className="px-6 py-8 text-center text-gray-500">
                    No transactions found
                  </td>
                </tr>
              ) : (
                filteredTransactions.map((tx) => (
                  <tr
                    key={tx.id}
                    className="hover:bg-orange-50 transition-colors">
                    <td className="px-4 py-4 text-sm text-gray-900 font-medium">
                      #{tx.id}
                    </td>
                    <td className="px-4 py-4 text-sm text-gray-600">
                      {formatDate(tx.created_at)}
                    </td>
                    <td className="px-4 py-4">
                      <div className="text-sm">
                        <div className="text-gray-900 font-medium">
                          User #{tx.user_id}
                        </div>
                        <div className="text-gray-500 text-xs">
                          {tx.user_email}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <span
                        className={`font-semibold uppercase ${getTypeColor(
                          tx.type,
                        )}`}>
                        {tx.type}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-right">
                      <span
                        className={`font-bold ${
                          tx.amount >= 0 ? 'text-green-600' : 'text-red-600'
                        }`}>
                        {tx.amount >= 0 ? '+' : ''}${formatNumber(tx.amount)}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span
                        className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(
                          tx.status,
                        )}`}>
                        {tx.status.toUpperCase()}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-sm text-gray-600">
                      {tx.description || '-'}
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
            Showing {filteredTransactions.length} of {transactions.length}{' '}
            transactions
          </p>
        </div>
      </div>
    </div>
  );
}
