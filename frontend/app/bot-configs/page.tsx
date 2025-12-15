export default function BotConfigsPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Bot Configs</h1>
      
      <div className="bg-white rounded-lg shadow p-6">
        <div className="mb-6">
          <h2 className="text-xl font-semibold mb-4">Cấu Hình Bot Trading</h2>
          <p className="text-gray-600 mb-4">Thiết lập các thông số cho bot giao dịch tự động</p>
        </div>

        <form className="space-y-6">
          {/* Bot Name */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Tên Bot
            </label>
            <input
              type="text"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              placeholder="My Trading Bot"
            />
          </div>

          {/* Exchange Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Sàn Giao Dịch
            </label>
            <select className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent">
              <option>Binance</option>
              <option>Bittrex</option>
              <option>Coinbase</option>
            </select>
          </div>

          {/* Trading Pair */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Cặp Giao Dịch
            </label>
            <input
              type="text"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              placeholder="BTC/USDT"
            />
          </div>

          {/* Strategy Settings */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Chiến Lược
              </label>
              <select className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent">
                <option>Scalping</option>
                <option>Day Trading</option>
                <option>Swing Trading</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Khung Thời Gian
              </label>
              <select className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent">
                <option>1m</option>
                <option>5m</option>
                <option>15m</option>
                <option>1h</option>
                <option>4h</option>
              </select>
            </div>
          </div>

          {/* Risk Management */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Stop Loss (%)
              </label>
              <input
                type="number"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                placeholder="3"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Take Profit (%)
              </label>
              <input
                type="number"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                placeholder="5"
              />
            </div>
          </div>

          {/* Max Investment */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Vốn Tối Đa (USDT)
            </label>
            <input
              type="number"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              placeholder="1000"
            />
          </div>

          {/* Bot Status */}
          <div className="flex items-center">
            <input
              type="checkbox"
              id="bot-active"
              className="h-4 w-4 text-purple-600 border-gray-300 rounded focus:ring-purple-500"
            />
            <label htmlFor="bot-active" className="ml-2 text-sm text-gray-700">
              Kích hoạt bot
            </label>
          </div>

          <button
            type="submit"
            className="w-full bg-gradient-to-r from-purple-600 to-indigo-600 text-white py-3 rounded-lg hover:from-purple-700 hover:to-indigo-700 transition-all duration-200 font-medium">
            Lưu Cấu Hình
          </button>
        </form>
      </div>
    </div>
  );
}
