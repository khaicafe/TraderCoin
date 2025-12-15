'use client';

import {useState, useEffect} from 'react';
import {useRouter} from 'next/navigation';
import Link from 'next/link';
import {
  TrendingUp,
  Settings,
  History,
  Key,
  LogOut,
  RefreshCw,
  AlertCircle,
} from 'lucide-react';

interface User {
  id: number;
  email: string;
  full_name: string;
  status: string;
  subscription_end: string;
}

interface TradingConfig {
  id: number;
  exchange: string;
  symbol: string;
  stop_loss_percent: number;
  take_profit_percent: number;
  is_active: boolean;
}

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [configs, setConfigs] = useState<TradingConfig[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    fetchUserData();
    fetchTradingConfigs();
  }, []);

  const fetchUserData = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/user/profile`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error('Failed to fetch user data');
      }

      const data = await response.json();
      setUser(data);
    } catch (error) {
      console.error('Error fetching user data:', error);
      router.push('/login');
    } finally {
      setLoading(false);
    }
  };

  const fetchTradingConfigs = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/trading/configs`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (response.ok) {
        const data = await response.json();
        // Ensure configs is always an array
        setConfigs(Array.isArray(data) ? data : data.configs || []);
      }
    } catch (error) {
      console.error('Error fetching configs:', error);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    router.push('/login');
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  const subscriptionDate = user?.subscription_end
    ? new Date(user.subscription_end)
    : null;
  const isExpired = subscriptionDate ? subscriptionDate < new Date() : true;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold text-indigo-600">
                💰 TraderCoin
              </h1>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-gray-700">
                Welcome, <strong>{user?.full_name}</strong>
              </span>
              <button
                onClick={handleLogout}
                className="flex items-center gap-2 px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg transition-colors">
                <LogOut size={18} />
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
        {/* Subscription Alert */}
        {isExpired && (
          <div className="mb-6 p-4 bg-red-50 border-2 border-red-200 rounded-lg flex items-center gap-3">
            <AlertCircle className="text-red-600" size={24} />
            <div className="flex-1">
              <h3 className="font-bold text-red-900">Subscription Expired</h3>
              <p className="text-sm text-red-700">
                Your subscription has expired. Please renew to continue trading.
              </p>
            </div>
            <Link
              href="/subscription"
              className="px-6 py-2 bg-red-600 hover:bg-red-700 text-white font-bold rounded-lg transition-colors">
              Renew Now
            </Link>
          </div>
        )}

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow-md p-6 border-l-4 border-blue-500">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Active Configs</p>
                <p className="text-3xl font-bold text-gray-900">
                  {Array.isArray(configs)
                    ? configs.filter((c) => c.is_active).length
                    : 0}
                </p>
              </div>
              <TrendingUp className="text-blue-500" size={40} />
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-md p-6 border-l-4 border-green-500">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Total Configs</p>
                <p className="text-3xl font-bold text-gray-900">
                  {Array.isArray(configs) ? configs.length : 0}
                </p>
              </div>
              <Settings className="text-green-500" size={40} />
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-md p-6 border-l-4 border-purple-500">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Status</p>
                <p
                  className={`text-xl font-bold ${
                    user?.status === 'active'
                      ? 'text-green-600'
                      : 'text-red-600'
                  }`}>
                  {user?.status?.toUpperCase()}
                </p>
              </div>
              <RefreshCw className="text-purple-500" size={40} />
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Link
            href="/exchange-keys"
            className="bg-white rounded-xl shadow-md p-6 hover:shadow-xl transition-shadow cursor-pointer border-2 border-transparent hover:border-indigo-500">
            <div className="flex flex-col items-center text-center gap-4">
              <div className="p-4 bg-indigo-100 rounded-full">
                <Key className="text-indigo-600" size={32} />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">
                  Exchange Keys
                </h3>
                <p className="text-sm text-gray-600">Manage your API keys</p>
              </div>
            </div>
          </Link>

          <Link
            href="/trading-configs"
            className="bg-white rounded-xl shadow-md p-6 hover:shadow-xl transition-shadow cursor-pointer border-2 border-transparent hover:border-green-500">
            <div className="flex flex-col items-center text-center gap-4">
              <div className="p-4 bg-green-100 rounded-full">
                <Settings className="text-green-600" size={32} />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">
                  Trading Config
                </h3>
                <p className="text-sm text-gray-600">
                  Set stop loss & take profit
                </p>
              </div>
            </div>
          </Link>

          <Link
            href="/orders"
            className="bg-white rounded-xl shadow-md p-6 hover:shadow-xl transition-shadow cursor-pointer border-2 border-transparent hover:border-purple-500">
            <div className="flex flex-col items-center text-center gap-4">
              <div className="p-4 bg-purple-100 rounded-full">
                <History className="text-purple-600" size={32} />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">
                  Order History
                </h3>
                <p className="text-sm text-gray-600">View your trades</p>
              </div>
            </div>
          </Link>

          <Link
            href="/profile"
            className="bg-white rounded-xl shadow-md p-6 hover:shadow-xl transition-shadow cursor-pointer border-2 border-transparent hover:border-orange-500">
            <div className="flex flex-col items-center text-center gap-4">
              <div className="p-4 bg-orange-100 rounded-full">
                <TrendingUp className="text-orange-600" size={32} />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">Profile</h3>
                <p className="text-sm text-gray-600">Manage your account</p>
              </div>
            </div>
          </Link>
        </div>

        {/* Recent Configs */}
        <div className="mt-8 bg-white rounded-xl shadow-md p-6">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">
            Recent Trading Configs
          </h2>
          {!Array.isArray(configs) || configs.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">
                No trading configurations yet
              </p>
              <Link
                href="/trading-configs"
                className="inline-block px-6 py-3 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-lg transition-colors">
                Create Your First Config
              </Link>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">
                      Exchange
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">
                      Symbol
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">
                      Stop Loss
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">
                      Take Profit
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {Array.isArray(configs) &&
                    configs.slice(0, 5).map((config) => (
                      <tr key={config.id} className="hover:bg-gray-50">
                        <td className="px-6 py-4 whitespace-nowrap font-semibold">
                          {config.exchange}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {config.symbol}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-red-600">
                          -{config.stop_loss_percent}%
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-green-600">
                          +{config.take_profit_percent}%
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-3 py-1 rounded-full text-xs font-semibold ${
                              config.is_active
                                ? 'bg-green-100 text-green-800'
                                : 'bg-gray-100 text-gray-800'
                            }`}>
                            {config.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
