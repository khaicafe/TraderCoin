export default function PortfolioPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Portfolio</h1>
      
      {/* Portfolio Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-sm text-gray-600">Total Balance</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">$12,450.00</p>
          <p className="text-sm text-green-600 mt-1">+5.2% (24h)</p>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-sm text-gray-600">Total Profit</p>
          <p className="text-2xl font-bold text-green-600 mt-2">+$1,250.00</p>
          <p className="text-sm text-gray-600 mt-1">+11.2%</p>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-sm text-gray-600">Active Trades</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">8</p>
          <p className="text-sm text-gray-600 mt-1">3 winning</p>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-sm text-gray-600">Win Rate</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">67%</p>
          <p className="text-sm text-gray-600 mt-1">Last 30 days</p>
        </div>
      </div>

      {/* Holdings */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold">Holdings</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Asset</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Value (USDT)</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">24h Change</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              <tr>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="font-medium text-gray-900">BTC</div>
                  <div className="text-sm text-gray-500">Bitcoin</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">0.25</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">$11,147.00</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-green-600">+3.2%</td>
              </tr>
              <tr>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="font-medium text-gray-900">ETH</div>
                  <div className="text-sm text-gray-500">Ethereum</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">0.5</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">$1,303.00</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-red-600">-1.5%</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
