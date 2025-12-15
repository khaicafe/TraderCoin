'use client';

import {useEffect, useState} from 'react';
import {useRouter} from 'next/navigation';
import Link from 'next/link';

interface ExchangeKey {
  id: number;
  exchange: string;
  api_key: string;
  is_active: boolean;
  created_at: string;
}

export default function ExchangeKeysPage() {
  const router = useRouter();
  const [keys, setKeys] = useState<ExchangeKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [formData, setFormData] = useState({
    exchange: 'binance',
    api_key: '',
    api_secret: '',
  });

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    fetchKeys(token);
  }, [router]);

  const fetchKeys = async (token: string) => {
    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/keys`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error('Failed to fetch keys');
      }

      const data = await response.json();
      // Backend returns array directly, not wrapped in object
      setKeys(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error fetching keys:', error);
      setKeys([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const token = localStorage.getItem('token');
    if (!token) return;

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/keys`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(formData),
        },
      );

      if (!response.ok) {
        throw new Error('Failed to add key');
      }

      const data = await response.json();
      // After adding, refetch the keys list to get updated data
      fetchKeys(token);
      setShowAddModal(false);
      setFormData({exchange: 'binance', api_key: '', api_secret: ''});
      alert('Exchange key added successfully!');
    } catch (error) {
      console.error('Error adding key:', error);
      alert('Failed to add exchange key');
    }
  };

  const handleDelete = async (keyId: number) => {
    if (!confirm('Are you sure you want to delete this API key?')) return;

    const token = localStorage.getItem('token');
    if (!token) return;

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/keys/${keyId}`,
        {
          method: 'DELETE',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error('Failed to delete key');
      }

      setKeys(keys.filter((key) => key.id !== keyId));
      alert('Exchange key deleted successfully!');
    } catch (error) {
      console.error('Error deleting key:', error);
      alert('Failed to delete exchange key');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading exchange keys...</p>
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
              <Link
                href="/dashboard"
                className="text-sm text-blue-600 hover:text-blue-800 mb-2 inline-block">
                ← Back to Dashboard
              </Link>
              <h1 className="text-3xl font-bold text-gray-900">
                Exchange API Keys
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                Manage your exchange API keys for automated trading
              </p>
            </div>
            <button
              onClick={() => setShowAddModal(true)}
              className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold">
              + Add New Key
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Info Alert */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <div className="flex">
            <svg
              className="h-5 w-5 text-blue-400 mt-0.5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">
                Security Notice
              </h3>
              <p className="mt-1 text-sm text-blue-700">
                Your API keys are encrypted and stored securely. Make sure to
                only use API keys with trading permissions and never share them.
              </p>
            </div>
          </div>
        </div>

        {/* Keys List */}
        {keys.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
              />
            </svg>
            <h3 className="mt-4 text-lg font-medium text-gray-900">
              No API Keys Yet
            </h3>
            <p className="mt-2 text-sm text-gray-500">
              Get started by adding your first exchange API key.
            </p>
            <button
              onClick={() => setShowAddModal(true)}
              className="mt-6 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold">
              Add Your First Key
            </button>
          </div>
        ) : (
          <div className="grid gap-6">
            {keys.map((key) => (
              <div
                key={key.id}
                className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow">
                <div className="flex items-start justify-between">
                  <div className="flex items-center">
                    <div className="flex-shrink-0 h-12 w-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                      <span className="text-white font-bold text-lg">
                        {key.exchange.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div className="ml-4">
                      <h3 className="text-lg font-semibold text-gray-900 capitalize">
                        {key.exchange}
                      </h3>
                      <p className="text-sm text-gray-500 mt-1">
                        API Key: {key.api_key.substring(0, 8)}••••••••
                      </p>
                      <p className="text-xs text-gray-400 mt-1">
                        Added on {new Date(key.created_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>

                  <button
                    onClick={() => handleDelete(key.id)}
                    className="px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors font-medium">
                    Delete
                  </button>
                </div>

                <div className="mt-4 flex items-center gap-2 text-sm">
                  <span className="px-2 py-1 bg-green-100 text-green-800 rounded">
                    🔒 Encrypted
                  </span>
                  <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded">
                    ✓ Active
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {/* Add Key Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full p-8">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-bold text-gray-900">
                Add Exchange API Key
              </h2>
              <button
                onClick={() => setShowAddModal(false)}
                className="text-gray-400 hover:text-gray-600">
                <svg
                  className="h-6 w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  Exchange
                </label>
                <select
                  value={formData.exchange}
                  onChange={(e) =>
                    setFormData({...formData, exchange: e.target.value})
                  }
                  className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-blue-500 focus:outline-none text-gray-900 bg-white">
                  <option value="binance">Binance</option>
                  <option value="bittrex">Bittrex</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  API Key
                </label>
                <input
                  type="text"
                  required
                  value={formData.api_key}
                  onChange={(e) =>
                    setFormData({...formData, api_key: e.target.value})
                  }
                  className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-blue-500 focus:outline-none text-gray-900 bg-white"
                  placeholder="Enter your API key"
                />
              </div>

              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  API Secret
                </label>
                <input
                  type="password"
                  required
                  value={formData.api_secret}
                  onChange={(e) =>
                    setFormData({...formData, api_secret: e.target.value})
                  }
                  className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-blue-500 focus:outline-none text-gray-900 bg-white"
                  placeholder="Enter your API secret"
                />
              </div>

              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                <p className="text-xs text-yellow-800">
                  ⚠️ Make sure your API key has trading permissions enabled.
                </p>
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="flex-1 px-4 py-3 border-2 border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-semibold">
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold">
                  Add Key
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
