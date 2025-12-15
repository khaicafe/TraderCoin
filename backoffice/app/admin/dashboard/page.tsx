'use client';

import {useEffect, useState} from 'react';
import {useRouter} from 'next/navigation';
import Link from 'next/link';

interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  suspendedUsers: number;
  totalRevenue: number;
  monthlyRevenue: number;
  activeSubscriptions: number;
  totalOrders: number;
  todayOrders: number;
}

export default function AdminDashboard() {
  const router = useRouter();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [admin, setAdmin] = useState<any>(null);

  useEffect(() => {
    const token = localStorage.getItem('admin_token');
    const adminData = localStorage.getItem('admin');

    if (!token) {
      router.push('/');
      return;
    }

    // Parse admin data safely
    if (adminData && adminData !== 'undefined') {
      try {
        setAdmin(JSON.parse(adminData));
      } catch (e) {
        console.error('Failed to parse admin data:', e);
      }
    }

    fetchDashboardStats(token);
  }, [router]);

  const fetchDashboardStats = async (token: string) => {
    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/admin/statistics`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error('Failed to fetch stats');
      }

      const data = await response.json();
      setStats(data);
    } catch (error) {
      console.error('Error fetching dashboard stats:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin');
    router.push('/');
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-slate-900 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                TraderCoin Backoffice
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                Welcome back, {admin?.name || 'Admin'}
              </p>
            </div>
            <button
              onClick={handleLogout}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">
              Logout
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {/* Total Users */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-blue-100 rounded-lg p-3">
                <svg
                  className="h-6 w-6 text-blue-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                  />
                </svg>
              </div>
              <div className="ml-4 flex-1">
                <p className="text-sm font-medium text-gray-500">Total Users</p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats?.totalUsers || 0}
                </p>
              </div>
            </div>
            <div className="mt-3 flex text-sm">
              <span className="text-green-600 font-medium">
                {stats?.activeUsers || 0} active
              </span>
              <span className="text-gray-400 mx-2">•</span>
              <span className="text-red-600 font-medium">
                {stats?.suspendedUsers || 0} suspended
              </span>
            </div>
          </div>

          {/* Revenue */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-green-100 rounded-lg p-3">
                <svg
                  className="h-6 w-6 text-green-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <div className="ml-4 flex-1">
                <p className="text-sm font-medium text-gray-500">
                  Total Revenue
                </p>
                <p className="text-2xl font-bold text-gray-900">
                  ${stats?.totalRevenue?.toLocaleString() || 0}
                </p>
              </div>
            </div>
            <div className="mt-3 text-sm text-gray-600">
              ${stats?.monthlyRevenue?.toLocaleString() || 0} this month
            </div>
          </div>

          {/* Active Subscriptions */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-purple-100 rounded-lg p-3">
                <svg
                  className="h-6 w-6 text-purple-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <div className="ml-4 flex-1">
                <p className="text-sm font-medium text-gray-500">
                  Active Subscriptions
                </p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats?.activeSubscriptions || 0}
                </p>
              </div>
            </div>
            <div className="mt-3 text-sm text-gray-600">
              Monthly recurring revenue
            </div>
          </div>

          {/* Total Orders */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0 bg-orange-100 rounded-lg p-3">
                <svg
                  className="h-6 w-6 text-orange-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                  />
                </svg>
              </div>
              <div className="ml-4 flex-1">
                <p className="text-sm font-medium text-gray-500">
                  Total Orders
                </p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats?.totalOrders || 0}
                </p>
              </div>
            </div>
            <div className="mt-3 text-sm text-gray-600">
              {stats?.todayOrders || 0} orders today
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="bg-white rounded-lg shadow mb-8">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-xl font-bold text-gray-900">Quick Actions</h2>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Link
                href="/admin/users"
                className="flex items-center p-4 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors">
                <svg
                  className="h-8 w-8 text-blue-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                  />
                </svg>
                <div className="ml-4">
                  <h3 className="font-semibold text-gray-900">Manage Users</h3>
                  <p className="text-sm text-gray-600">
                    View, suspend, or activate users
                  </p>
                </div>
              </Link>

              <Link
                href="/admin/subscriptions"
                className="flex items-center p-4 bg-green-50 rounded-lg hover:bg-green-100 transition-colors">
                <svg
                  className="h-8 w-8 text-green-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
                  />
                </svg>
                <div className="ml-4">
                  <h3 className="font-semibold text-gray-900">Subscriptions</h3>
                  <p className="text-sm text-gray-600">
                    Manage user subscriptions
                  </p>
                </div>
              </Link>

              <Link
                href="/admin/transactions"
                className="flex items-center p-4 bg-purple-50 rounded-lg hover:bg-purple-100 transition-colors">
                <svg
                  className="h-8 w-8 text-purple-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                  />
                </svg>
                <div className="ml-4">
                  <h3 className="font-semibold text-gray-900">Transactions</h3>
                  <p className="text-sm text-gray-600">View payment history</p>
                </div>
              </Link>
            </div>
          </div>
        </div>

        {/* Recent Activity (Placeholder) */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-xl font-bold text-gray-900">Recent Activity</h2>
          </div>
          <div className="p-6">
            <p className="text-gray-500 text-center py-8">
              No recent activity to display
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}
