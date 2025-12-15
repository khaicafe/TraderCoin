export default function MonitoringPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Monitoring</h1>

      {/* Real-time Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Bot Status</h3>
            <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
          </div>
          <p className="text-sm text-gray-600">Active: 3/5 bots</p>
          <p className="text-2xl font-bold text-green-600 mt-2">Running</p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            CPU Usage
          </h3>
          <div className="relative pt-1">
            <div className="flex mb-2 items-center justify-between">
              <div>
                <span className="text-xs font-semibold inline-block text-purple-600">
                  45%
                </span>
              </div>
            </div>
            <div className="overflow-hidden h-2 mb-4 text-xs flex rounded bg-purple-200">
              <div
                style={{width: '45%'}}
                className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-purple-600"></div>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            Memory Usage
          </h3>
          <div className="relative pt-1">
            <div className="flex mb-2 items-center justify-between">
              <div>
                <span className="text-xs font-semibold inline-block text-blue-600">
                  62%
                </span>
              </div>
            </div>
            <div className="overflow-hidden h-2 mb-4 text-xs flex rounded bg-blue-200">
              <div
                style={{width: '62%'}}
                className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-blue-600"></div>
            </div>
          </div>
        </div>
      </div>

      {/* Active Positions */}
      <div className="bg-white rounded-lg shadow overflow-hidden mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Vị Thế Đang Mở</h2>
        </div>
        <div className="p-6">
          <div className="space-y-4">
            <div className="border border-gray-200 rounded-lg p-4">
              <div className="flex justify-between items-center mb-3">
                <div>
                  <h3 className="font-semibold text-lg text-gray-900">BTC/USDT</h3>
                  <p className="text-sm text-gray-600">Long Position</p>
                </div>
                <div className="text-right">
                  <p className="text-green-600 font-bold text-lg">+$485.50</p>
                  <p className="text-sm text-green-600">+2.15%</p>
                </div>
              </div>
              <div className="grid grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-gray-600">Entry</p>
                  <p className="font-semibold">$43,200</p>
                </div>
                <div>
                  <p className="text-gray-600">Current</p>
                  <p className="font-semibold">$44,129</p>
                </div>
                <div>
                  <p className="text-gray-600">Stop Loss</p>
                  <p className="font-semibold text-red-600">$42,000</p>
                </div>
                <div>
                  <p className="text-gray-600">Take Profit</p>
                  <p className="font-semibold text-green-600">$46,000</p>
                </div>
              </div>
            </div>

            <div className="border border-gray-200 rounded-lg p-4">
              <div className="flex justify-between items-center mb-3">
                <div>
                  <h3 className="font-semibold text-lg text-gray-900">ETH/USDT</h3>
                  <p className="text-sm text-gray-600">Long Position</p>
                </div>
                <div className="text-right">
                  <p className="text-red-600 font-bold text-lg">-$125.30</p>
                  <p className="text-sm text-red-600">-0.85%</p>
                </div>
              </div>
              <div className="grid grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-gray-600">Entry</p>
                  <p className="font-semibold">$2,620</p>
                </div>
                <div>
                  <p className="text-gray-600">Current</p>
                  <p className="font-semibold">$2,598</p>
                </div>
                <div>
                  <p className="text-gray-600">Stop Loss</p>
                  <p className="font-semibold text-red-600">$2,550</p>
                </div>
                <div>
                  <p className="text-gray-600">Take Profit</p>
                  <p className="font-semibold text-green-600">$2,750</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* System Logs Preview */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Recent Activity</h2>
        </div>
        <div className="p-6">
          <div className="space-y-2 font-mono text-sm">
            <div className="text-green-600">
              [10:35:22] ✓ Order BTC/USDT executed at $44,129
            </div>
            <div className="text-blue-600">
              [10:34:15] ℹ Monitoring price changes...
            </div>
            <div className="text-yellow-600">
              [10:33:08] ⚠ High volatility detected on ETH/USDT
            </div>
            <div className="text-green-600">
              [10:32:45] ✓ Take profit triggered for BNB/USDT
            </div>
            <div className="text-blue-600">
              [10:31:30] ℹ Bot #3 started trading
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
