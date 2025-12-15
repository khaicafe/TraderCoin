'use client';

import { useState } from 'react';

export default function TradingPage() {
  const [selectedConfig, setSelectedConfig] = useState('');

  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Đặt Lệnh Trading</h1>

      {/* Warning Alert */}
      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6 flex items-start gap-3">
        <span className="text-yellow-600 text-sm">⚠️</span>
        <p className="text-sm text-yellow-800">
          <strong>Không có config nào</strong>
        </p>
        <button className="ml-auto text-yellow-600 hover:text-yellow-800">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left Column - Bot Config Selection */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Chọn Bot Config</h2>
          
          <div>
            <select 
              value={selectedConfig}
              onChange={(e) => setSelectedConfig(e.target.value)}
              className="w-full px-4 py-3 border-2 border-indigo-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900">
              <option value="">-- Chọn Bot Config --</option>
              <option value="1">Scalping Bot - BTC/USDT</option>
              <option value="2">ETH Trader - ETH/USDT</option>
            </select>
          </div>
        </div>

        {/* Right Column - Order Form */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Thông Tin Lệnh</h2>
          
          <form className="space-y-4">
            {/* Symbol */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Symbol
              </label>
              <input
                type="text"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                placeholder="BTC/USDT"
                disabled
              />
              <p className="text-xs text-gray-500 mt-1">Gõ để tìm kiếm. 40 symbols phổ biến có sẵn</p>
            </div>

            {/* Order Type */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Order Type
              </label>
              <select className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900">
                <option>Market (Giá thị trường)</option>
                <option>Limit (Giá cố định)</option>
              </select>
            </div>

            {/* Amount */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Amount
              </label>
              <input
                type="text"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                placeholder="Nhập số lượng"
              />
              <p className="text-xs text-gray-500 mt-1">Để trống sẽ dùng amount từ config</p>
            </div>

            {/* Warning */}
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 flex items-start gap-2">
              <span className="text-yellow-600">⚠️</span>
              <p className="text-xs text-yellow-800">
                <strong>Cảnh báo:</strong> Đây là lệnh THẬT trên SÀN GIAO DỊCH!
              </p>
            </div>

            {/* Action Buttons */}
            <div className="grid grid-cols-2 gap-4 pt-2">
              <button
                type="button"
                className="w-full bg-green-500 hover:bg-green-600 text-white font-semibold py-3 rounded-lg transition-colors">
                Đặt lệnh BUY
              </button>
              <button
                type="button"
                className="w-full bg-red-400 hover:bg-red-500 text-white font-semibold py-3 rounded-lg transition-colors">
                Đặt lệnh SELL
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
