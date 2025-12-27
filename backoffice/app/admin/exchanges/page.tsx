'use client';

import {useEffect, useState} from 'react';
import {useRouter} from 'next/navigation';

interface ExchangeConfig {
  id: number;
  exchange: string;
  display_name: string;
  spot_api_url: string;
  spot_api_testnet_url: string;
  spot_ws_url: string;
  spot_ws_testnet_url: string;
  futures_api_url: string;
  futures_api_testnet_url: string;
  futures_ws_url: string;
  futures_ws_testnet_url: string;
  is_active: boolean;
  support_spot: boolean;
  support_futures: boolean;
  support_margin: boolean;
  max_leverage: number;
  min_order_size: number;
  maker_fee: number;
  taker_fee: number;
  notes: string;
}

export default function ExchangesManagement() {
  const router = useRouter();
  const [exchanges, setExchanges] = useState<ExchangeConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editingExchange, setEditingExchange] = useState<ExchangeConfig | null>(
    null,
  );
  const [formData, setFormData] = useState<Partial<ExchangeConfig>>({
    exchange: '',
    display_name: '',
    spot_api_url: '',
    spot_api_testnet_url: '',
    spot_ws_url: '',
    spot_ws_testnet_url: '',
    futures_api_url: '',
    futures_api_testnet_url: '',
    futures_ws_url: '',
    futures_ws_testnet_url: '',
    is_active: true,
    support_spot: true,
    support_futures: false,
    support_margin: false,
    max_leverage: 1,
    min_order_size: 0.0001,
    maker_fee: 0.001,
    taker_fee: 0.001,
    notes: '',
  });

  useEffect(() => {
    const token = localStorage.getItem('admin_token');
    if (!token) {
      router.push('/');
      return;
    }

    fetchExchanges(token);
  }, [router]);

  const fetchExchanges = async (token: string) => {
    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/exchanges`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error('Failed to fetch exchanges');
      }

      const data = await response.json();
      setExchanges(data || []);
    } catch (error) {
      console.error('Error fetching exchanges:', error);
    } finally {
      setLoading(false);
    }
  };

  const toggleExchangeStatus = async (exchangeId: number) => {
    const token = localStorage.getItem('admin_token');
    if (!token) return;

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/exchanges/${exchangeId}/toggle`,
        {
          method: 'PATCH',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error('Failed to toggle exchange status');
      }

      // Refresh exchanges list
      fetchExchanges(token);
    } catch (error) {
      console.error('Error toggling exchange status:', error);
      alert('Failed to update exchange status');
    }
  };

  const handleAddExchange = async () => {
    const token = localStorage.getItem('admin_token');
    if (!token) return;

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/exchanges`,
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
        throw new Error('Failed to create exchange');
      }

      // Reset form and close modal
      setShowAddModal(false);
      setFormData({
        exchange: '',
        display_name: '',
        rest_api_url: '',
        rest_api_testnet_url: '',
        websocket_url: '',
        futures_api_url: '',
        is_active: true,
        support_spot: true,
        support_futures: false,
        support_margin: false,
        max_leverage: 1,
        min_order_size: 0.0001,
        maker_fee: 0.001,
        taker_fee: 0.001,
        notes: '',
      });

      // Refresh exchanges list
      fetchExchanges(token);
      alert('Exchange added successfully!');
    } catch (error) {
      console.error('Error adding exchange:', error);
      alert('Failed to add exchange');
    }
  };

  const handleEditClick = (exchange: ExchangeConfig) => {
    setEditingExchange(exchange);
    setFormData({
      exchange: exchange.exchange,
      display_name: exchange.display_name,
      rest_api_url: exchange.rest_api_url,
      rest_api_testnet_url: exchange.rest_api_testnet_url,
      websocket_url: exchange.websocket_url,
      futures_api_url: exchange.futures_api_url,
      is_active: exchange.is_active,
      support_spot: exchange.support_spot,
      support_futures: exchange.support_futures,
      support_margin: exchange.support_margin,
      max_leverage: exchange.max_leverage,
      min_order_size: exchange.min_order_size,
      maker_fee: exchange.maker_fee,
      taker_fee: exchange.taker_fee,
      notes: exchange.notes,
    });
    setShowEditModal(true);
  };

  const handleUpdateExchange = async () => {
    const token = localStorage.getItem('admin_token');
    if (!token || !editingExchange) return;

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/exchanges/${editingExchange.id}`,
        {
          method: 'PUT',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(formData),
        },
      );

      if (!response.ok) {
        throw new Error('Failed to update exchange');
      }

      // Reset form and close modal
      setShowEditModal(false);
      setEditingExchange(null);
      setFormData({
        exchange: '',
        display_name: '',
        rest_api_url: '',
        rest_api_testnet_url: '',
        websocket_url: '',
        futures_api_url: '',
        is_active: true,
        support_spot: true,
        support_futures: false,
        support_margin: false,
        max_leverage: 1,
        min_order_size: 0.0001,
        maker_fee: 0.001,
        taker_fee: 0.001,
        notes: '',
      });

      // Refresh exchanges list
      fetchExchanges(token);
      alert('Exchange updated successfully!');
    } catch (error) {
      console.error('Error updating exchange:', error);
      alert('Failed to update exchange');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#EE4D2D] mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading exchanges...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      {/* Page Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            Exchange Management
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage cryptocurrency exchange configurations
          </p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="px-6 py-3 bg-gradient-to-r from-orange-400 to-orange-500 text-white rounded-lg hover:from-orange-500 hover:to-orange-600 transition-all shadow-lg hover:shadow-xl flex items-center gap-2 font-medium">
          <span className="text-xl">+</span>
          Add New Exchange
        </button>
      </div>

      {/* Exchanges Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {exchanges.map((exchange) => (
          <div
            key={exchange.id}
            className={`bg-white rounded-lg shadow-md overflow-hidden border-t-4 transition-all hover:shadow-lg ${
              exchange.is_active
                ? 'border-[#EE4D2D]'
                : 'border-gray-300 opacity-60'
            }`}>
            {/* Card Header */}
            <div
              className={`p-4 ${
                exchange.is_active
                  ? 'bg-gradient-to-r from-orange-400 to-orange-500'
                  : 'bg-gray-400'
              }`}>
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-white">
                  {exchange.display_name}
                </h3>
                <span
                  className={`px-3 py-1 rounded-full text-xs font-semibold ${
                    exchange.is_active
                      ? 'bg-white text-orange-600'
                      : 'bg-gray-200 text-gray-600'
                  }`}>
                  {exchange.is_active ? 'Active' : 'Inactive'}
                </span>
              </div>
            </div>

            {/* Card Body */}
            <div className="p-6">
              {/* Trading Modes */}
              <div className="mb-4">
                <p className="text-sm font-semibold text-gray-700 mb-2">
                  Supported Modes:
                </p>
                <div className="flex flex-wrap gap-2">
                  {exchange.support_spot && (
                    <span className="px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full font-medium">
                      Spot
                    </span>
                  )}
                  {exchange.support_futures && (
                    <span className="px-2 py-1 bg-purple-100 text-purple-700 text-xs rounded-full font-medium">
                      Futures
                    </span>
                  )}
                  {exchange.support_margin && (
                    <span className="px-2 py-1 bg-orange-100 text-orange-700 text-xs rounded-full font-medium">
                      Margin
                    </span>
                  )}
                </div>
              </div>

              {/* Details */}
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Max Leverage:</span>
                  <span className="font-semibold text-gray-900">
                    {exchange.max_leverage}x
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Maker Fee:</span>
                  <span className="font-semibold text-gray-900">
                    {(exchange.maker_fee * 100).toFixed(2)}%
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Taker Fee:</span>
                  <span className="font-semibold text-gray-900">
                    {(exchange.taker_fee * 100).toFixed(2)}%
                  </span>
                </div>
              </div>

              {/* APIs */}
              <div className="mt-4 pt-4 border-t border-gray-200 space-y-3">
                {exchange.support_spot && exchange.rest_api_url && (
                  <div>
                    <p className="text-xs font-semibold text-gray-600 mb-1">
                      Spot API:
                    </p>
                    <p className="text-xs text-gray-700 font-mono truncate bg-gray-50 px-2 py-1 rounded">
                      {exchange.rest_api_url}
                    </p>
                  </div>
                )}
                {exchange.support_futures && exchange.futures_api_url && (
                  <div>
                    <p className="text-xs font-semibold text-gray-600 mb-1">
                      Futures API:
                    </p>
                    <p className="text-xs text-gray-700 font-mono truncate bg-gray-50 px-2 py-1 rounded">
                      {exchange.futures_api_url}
                    </p>
                  </div>
                )}
                {exchange.websocket_url && (
                  <div>
                    <p className="text-xs font-semibold text-gray-600 mb-1">
                      WebSocket:
                    </p>
                    <p className="text-xs text-gray-700 font-mono truncate bg-gray-50 px-2 py-1 rounded">
                      {exchange.websocket_url}
                    </p>
                  </div>
                )}
              </div>

              {/* Notes */}
              {exchange.notes && (
                <div className="mt-4">
                  <p className="text-xs text-gray-500 italic">
                    {exchange.notes}
                  </p>
                </div>
              )}

              {/* Actions */}
              <div className="mt-6 pt-4 border-t border-gray-200 flex gap-2">
                <button
                  onClick={() => handleEditClick(exchange)}
                  className="flex-1 px-4 py-2 bg-gradient-to-r from-blue-500 to-blue-600 text-white hover:from-blue-600 hover:to-blue-700 rounded-lg font-medium transition-all shadow-md hover:shadow-lg">
                  Edit
                </button>
                <button
                  onClick={() => toggleExchangeStatus(exchange.id)}
                  className={`flex-1 px-4 py-2 rounded-lg font-medium transition-all shadow-md hover:shadow-lg ${
                    exchange.is_active
                      ? 'bg-gradient-to-r from-red-500 to-red-600 text-white hover:from-red-600 hover:to-red-700'
                      : 'bg-gradient-to-r from-green-500 to-green-600 text-white hover:from-green-600 hover:to-green-700'
                  }`}>
                  {exchange.is_active ? 'Disable' : 'Enable'}
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Summary */}
      <div className="mt-8 bg-white rounded-lg shadow-md p-6 border-t-4 border-[#EE4D2D]">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-center">
          <div>
            <p className="text-3xl font-bold text-gray-900">
              {exchanges.length}
            </p>
            <p className="text-sm text-gray-600 font-medium">Total Exchanges</p>
          </div>
          <div>
            <p className="text-3xl font-bold text-green-600">
              {exchanges.filter((e) => e.is_active).length}
            </p>
            <p className="text-sm text-gray-600 font-medium">Active</p>
          </div>
          <div>
            <p className="text-3xl font-bold text-gray-400">
              {exchanges.filter((e) => !e.is_active).length}
            </p>
            <p className="text-sm text-gray-600 font-medium">Inactive</p>
          </div>
        </div>
      </div>

      {/* Add Exchange Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-gradient-to-r from-orange-400 to-orange-500 text-white p-6 rounded-t-lg">
              <h2 className="text-2xl font-bold">Add New Exchange</h2>
              <p className="text-white/90 text-sm mt-1">
                Configure a new cryptocurrency exchange
              </p>
            </div>

            <div className="p-6 space-y-4">
              {/* Basic Info */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Exchange Code *
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. binance"
                    value={formData.exchange || ''}
                    onChange={(e) =>
                      setFormData({...formData, exchange: e.target.value})
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Display Name *
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. Binance"
                    value={formData.display_name || ''}
                    onChange={(e) =>
                      setFormData({...formData, display_name: e.target.value})
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* API URLs */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Spot REST API URL *
                </label>
                <input
                  type="text"
                  placeholder="https://api.exchange.com"
                  value={formData.rest_api_url || ''}
                  onChange={(e) =>
                    setFormData({...formData, rest_api_url: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Futures REST API URL
                </label>
                <input
                  type="text"
                  placeholder="https://fapi.exchange.com"
                  value={formData.futures_api_url || ''}
                  onChange={(e) =>
                    setFormData({...formData, futures_api_url: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  WebSocket URL
                </label>
                <input
                  type="text"
                  placeholder="wss://stream.exchange.com"
                  value={formData.websocket_url || ''}
                  onChange={(e) =>
                    setFormData({...formData, websocket_url: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Testnet API URL
                </label>
                <input
                  type="text"
                  placeholder="https://testnet.exchange.com"
                  value={formData.rest_api_testnet_url || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      rest_api_testnet_url: e.target.value,
                    })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                />
              </div>

              {/* Support Checkboxes */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Trading Modes
                </label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={formData.support_spot || false}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          support_spot: e.target.checked,
                        })
                      }
                      className="rounded text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="text-sm text-gray-700">Spot Trading</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={formData.support_futures || false}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          support_futures: e.target.checked,
                        })
                      }
                      className="rounded text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="text-sm text-gray-700">Futures</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={formData.support_margin || false}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          support_margin: e.target.checked,
                        })
                      }
                      className="rounded text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="text-sm text-gray-700">Margin</span>
                  </label>
                </div>
              </div>

              {/* Numbers */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Max Leverage
                  </label>
                  <input
                    type="number"
                    value={formData.max_leverage || 1}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        max_leverage: parseInt(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Min Order Size
                  </label>
                  <input
                    type="number"
                    step="0.0001"
                    value={formData.min_order_size || 0.0001}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        min_order_size: parseFloat(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Maker Fee (%)
                  </label>
                  <input
                    type="number"
                    step="0.001"
                    value={(formData.maker_fee || 0) * 100}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        maker_fee: parseFloat(e.target.value) / 100,
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Taker Fee (%)
                  </label>
                  <input
                    type="number"
                    step="0.001"
                    value={(formData.taker_fee || 0) * 100}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        taker_fee: parseFloat(e.target.value) / 100,
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* Notes */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Notes
                </label>
                <textarea
                  rows={3}
                  placeholder="Additional notes or configuration details..."
                  value={formData.notes || ''}
                  onChange={(e) =>
                    setFormData({...formData, notes: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                />
              </div>

              {/* Active Status */}
              <div>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.is_active || false}
                    onChange={(e) =>
                      setFormData({...formData, is_active: e.target.checked})
                    }
                    className="rounded text-indigo-600 focus:ring-indigo-500"
                  />
                  <span className="text-sm font-medium text-gray-700">
                    Set as Active
                  </span>
                </label>
              </div>
            </div>

            {/* Modal Actions */}
            <div className="sticky bottom-0 bg-gray-50 px-6 py-4 rounded-b-lg flex gap-3 justify-end border-t">
              <button
                onClick={() => setShowAddModal(false)}
                className="px-6 py-2 border-2 border-gray-300 text-gray-700 rounded-lg hover:bg-gray-100 transition-colors font-medium">
                Cancel
              </button>
              <button
                onClick={handleAddExchange}
                className="px-6 py-2 bg-gradient-to-r from-orange-500 to-orange-600 text-white rounded-lg hover:from-orange-600 hover:to-orange-700 transition-all font-medium shadow-md hover:shadow-lg">
                Add Exchange
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Exchange Modal */}
      {showEditModal && editingExchange && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-gradient-to-r from-orange-400 to-orange-500 text-white p-6 rounded-t-lg">
              <h2 className="text-2xl font-bold">Edit Exchange</h2>
              <p className="text-white/90 text-sm mt-1">
                Update {editingExchange.display_name} configuration
              </p>
            </div>

            <div className="p-6 space-y-4">
              {/* Basic Info */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Exchange Code
                  </label>
                  <input
                    type="text"
                    value={formData.exchange || ''}
                    disabled
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-100 text-gray-500 cursor-not-allowed"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Display Name *
                  </label>
                  <input
                    type="text"
                    value={formData.display_name || ''}
                    onChange={(e) =>
                      setFormData({...formData, display_name: e.target.value})
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* API URLs */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Spot REST API URL *
                </label>
                <input
                  type="text"
                  value={formData.rest_api_url || ''}
                  onChange={(e) =>
                    setFormData({...formData, rest_api_url: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Futures REST API URL
                </label>
                <input
                  type="text"
                  value={formData.futures_api_url || ''}
                  onChange={(e) =>
                    setFormData({...formData, futures_api_url: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  WebSocket URL
                </label>
                <input
                  type="text"
                  value={formData.websocket_url || ''}
                  onChange={(e) =>
                    setFormData({...formData, websocket_url: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Testnet API URL
                </label>
                <input
                  type="text"
                  value={formData.rest_api_testnet_url || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      rest_api_testnet_url: e.target.value,
                    })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* Support Checkboxes */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Trading Modes
                </label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={formData.support_spot || false}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          support_spot: e.target.checked,
                        })
                      }
                      className="rounded text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-sm text-gray-700">Spot Trading</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={formData.support_futures || false}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          support_futures: e.target.checked,
                        })
                      }
                      className="rounded text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-sm text-gray-700">Futures</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={formData.support_margin || false}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          support_margin: e.target.checked,
                        })
                      }
                      className="rounded text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-sm text-gray-700">Margin</span>
                  </label>
                </div>
              </div>

              {/* Numbers */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Max Leverage
                  </label>
                  <input
                    type="number"
                    value={formData.max_leverage || 1}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        max_leverage: parseInt(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Min Order Size
                  </label>
                  <input
                    type="number"
                    step="0.0001"
                    value={formData.min_order_size || 0.0001}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        min_order_size: parseFloat(e.target.value),
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Maker Fee (%)
                  </label>
                  <input
                    type="number"
                    step="0.001"
                    value={(formData.maker_fee || 0) * 100}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        maker_fee: parseFloat(e.target.value) / 100,
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Taker Fee (%)
                  </label>
                  <input
                    type="number"
                    step="0.001"
                    value={(formData.taker_fee || 0) * 100}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        taker_fee: parseFloat(e.target.value) / 100,
                      })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* Notes */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Notes
                </label>
                <textarea
                  rows={3}
                  value={formData.notes || ''}
                  onChange={(e) =>
                    setFormData({...formData, notes: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* Active Status */}
              <div>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.is_active || false}
                    onChange={(e) =>
                      setFormData({...formData, is_active: e.target.checked})
                    }
                    className="rounded text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm font-medium text-gray-700">
                    Active Exchange
                  </span>
                </label>
              </div>
            </div>

            {/* Modal Actions */}
            <div className="sticky bottom-0 bg-gray-50 px-6 py-4 rounded-b-lg flex gap-3 justify-end border-t">
              <button
                onClick={() => {
                  setShowEditModal(false);
                  setEditingExchange(null);
                }}
                className="px-6 py-2 border-2 border-gray-300 text-gray-700 rounded-lg hover:bg-gray-100 transition-colors font-medium">
                Cancel
              </button>
              <button
                onClick={handleUpdateExchange}
                className="px-6 py-2 bg-gradient-to-r from-orange-500 to-orange-600 text-white rounded-lg hover:from-orange-600 hover:to-orange-700 transition-all font-medium shadow-md hover:shadow-lg">
                Update Exchange
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
