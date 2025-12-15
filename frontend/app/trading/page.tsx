export default function TradingPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Trading</h1>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Active Trades */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Active Trades</h2>
          <div className="space-y-3">
            <div className="p-4 border border-gray-200 rounded-lg">
              <div className="flex justify-between items-center mb-2">
                <span className="font-medium">BTC/USDT</span>
                <span className="text-green-600 text-sm font-medium">+2.5%</span>
              </div>
              <div className="text-sm text-gray-600">
                <p>Entry: $43,500 | Current: $44,588</p>
                <p>Stop Loss: $42,000 | Take Profit: $46,000</p>
              </div>
            </div>
          </div>
        </div>

        {/* Create New Trade */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Create Trade</h2>
          <form className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Trading Pair
              </label>
              <select className="w-full px-3 py-2 border border-gray-300 rounded-md">
                <option>BTC/USDT</option>
                <option>ETH/USDT</option>
                <option>BNB/USDT</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Amount (USDT)
              </label>
              <input
                type="number"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                placeholder="0.00"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Stop Loss (%)
                </label>
                <input
                  type="number"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder="5"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Take Profit (%)
                </label>
                <input
                  type="number"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder="10"
                />
              </div>
            </div>
            <button
              type="submit"
              className="w-full bg-blue-600 text-white py-2 rounded-md hover:bg-blue-700"
            >
              Create Trade
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
