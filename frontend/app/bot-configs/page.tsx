'use client';

import {useState} from 'react';

export default function BotConfigsPage() {
  const [bots, setBots] = useState([
    {
      id: 1,
      name: 'Scalping Bot',
      symbol: 'BTC/USDT',
      exchange: 'Binance',
      amount: 1000,
      sl: 3,
      tp: 5,
      mode: 'Auto',
      status: 'Running',
    },
    {
      id: 2,
      name: 'ETH Trader',
      symbol: 'ETH/USDT',
      exchange: 'Binance',
      amount: 500,
      sl: 2,
      tp: 4,
      mode: 'Manual',
      status: 'Stopped',
    },
  ]);

  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    symbol: '',
    exchange: 'Binance',
    amount: '',
    leverage: '1',
    stopLoss: '',
    takeProfit: '',
    tradingMode: 'Spot',
    apiKey: '',
    secretKey: '',
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // Handle form submission
    console.log('Form data:', formData);
    setShowModal(false);
  };

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Bot Configurations</h1>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center gap-2 bg-indigo-600 text-white px-6 py-3 rounded-lg hover:bg-indigo-700 transition-colors shadow-md">
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          Tạo Config Mới
        </button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  ID
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Tên
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Symbol
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Sàn
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Amount
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  SL %
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  TP %
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Mode
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Status
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {bots.map((bot) => (
                <tr key={bot.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm text-gray-900">{bot.id}</td>
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">
                    {bot.name}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-900">
                    {bot.symbol}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-900">
                    {bot.exchange}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-900">
                    ${bot.amount}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-900">{bot.sl}%</td>
                  <td className="px-6 py-4 text-sm text-gray-900">{bot.tp}%</td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-3 py-1 text-xs font-semibold rounded-full ${
                        bot.mode === 'Auto'
                          ? 'bg-blue-100 text-blue-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                      {bot.mode}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-3 py-1 text-xs font-semibold rounded-full ${
                        bot.status === 'Running'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}>
                      {bot.status}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <button
                        className="p-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                        title="Edit">
                        <svg
                          className="w-5 h-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                          />
                        </svg>
                      </button>
                      <button
                        className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                        title="Delete">
                        <svg
                          className="w-5 h-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                      </button>
                      <button
                        className="p-2 text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                        title="Start/Stop">
                        <svg
                          className="w-5 h-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                          />
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                          />
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Empty State */}
        {bots.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">Chưa có bot nào được cấu hình</p>
            <button className="mt-4 text-indigo-600 hover:text-indigo-700 font-medium">
              Tạo bot đầu tiên
            </button>
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div
          className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4 animate-fadeIn"
          onClick={() => setShowModal(false)}>
          <div
            className="bg-white rounded-lg shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto animate-slideUp"
            onClick={(e) => e.stopPropagation()}>
            {/* Modal Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h2 className="text-2xl font-bold text-gray-900">
                Tạo Config Mới
              </h2>
              <button
                onClick={() => setShowModal(false)}
                className="text-gray-400 hover:text-gray-600 transition-colors">
                <svg
                  className="w-6 h-6"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            {/* Modal Body */}
            <form onSubmit={handleSubmit} className="p-6 space-y-6">
              {/* Symbol */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Symbol
                </label>
                <input
                  type="text"
                  value={formData.symbol}
                  onChange={(e) =>
                    setFormData({...formData, symbol: e.target.value})
                  }
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                  placeholder="BTC/USDT"
                  required
                />
              </div>

              {/* Sàn Giao Dịch */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Sàn Giao Dịch
                </label>
                <select
                  value={formData.exchange}
                  onChange={(e) =>
                    setFormData({...formData, exchange: e.target.value})
                  }
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900">
                  <option>Binance</option>
                  <option>Bittrex</option>
                  <option>Coinbase</option>
                </select>
                <p className="text-xs text-gray-500 mt-1">
                  Chọn sàn giao dịch bạn muốn sử dụng
                </p>
              </div>

              {/* Amount & Leverage */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Amount
                  </label>
                  <input
                    type="number"
                    value={formData.amount}
                    onChange={(e) =>
                      setFormData({...formData, amount: e.target.value})
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="0"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Leverage
                  </label>
                  <input
                    type="number"
                    value={formData.leverage}
                    onChange={(e) =>
                      setFormData({...formData, leverage: e.target.value})
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="1"
                    required
                  />
                </div>
              </div>

              {/* Stop Loss & Take Profit */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Stop Loss (%)
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.stopLoss}
                    onChange={(e) =>
                      setFormData({...formData, stopLoss: e.target.value})
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="0"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Take Profit (%)
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.takeProfit}
                    onChange={(e) =>
                      setFormData({...formData, takeProfit: e.target.value})
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="0"
                    required
                  />
                </div>
              </div>

              {/* Trading Mode */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Trading Mode
                </label>
                <select
                  value={formData.tradingMode}
                  onChange={(e) =>
                    setFormData({...formData, tradingMode: e.target.value})
                  }
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900">
                  <option>Spot</option>
                  <option>Futures</option>
                  <option>Margin</option>
                </select>
              </div>

              {/* API Credentials Section */}
              <div className="border-t border-gray-200 pt-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">
                  API Credentials (Tùy chọn)
                </h3>

                {/* API Key */}
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    API Key
                  </label>
                  <input
                    type="text"
                    value={formData.apiKey}
                    onChange={(e) =>
                      setFormData({...formData, apiKey: e.target.value})
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="Nhập API Key từ sàn giao dịch"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    API Key từ Binance (sẽ được mã hóa)
                  </p>
                </div>

                {/* Secret Key */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Secret Key
                  </label>
                  <input
                    type="password"
                    value={formData.secretKey}
                    onChange={(e) =>
                      setFormData({...formData, secretKey: e.target.value})
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="Nhập Secret Key từ sàn giao dịch"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Secret Key từ Binance (sẽ được mã hóa)
                  </p>
                </div>
              </div>

              {/* Modal Footer */}
              <div className="flex items-center justify-end gap-3 pt-6 border-t border-gray-200">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
                  Hủy
                </button>
                <button
                  type="submit"
                  className="px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors">
                  Tạo Config
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
