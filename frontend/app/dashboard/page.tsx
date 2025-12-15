export default function DashboardPage() {
  return (
    <div>
      {/* Header */}
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Dashboard</h1>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Total Bots */}
        <div className="bg-gradient-to-br from-indigo-500 to-indigo-600 rounded-xl shadow-lg p-6 text-white">
          <p className="text-sm font-medium opacity-90 mb-2">Total Bots</p>
          <p className="text-5xl font-bold">0</p>
        </div>

        {/* Active Bots */}
        <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl shadow-lg p-6 text-white">
          <p className="text-sm font-medium opacity-90 mb-2">Active Bots</p>
          <p className="text-5xl font-bold">0</p>
        </div>

        {/* Total Orders */}
        <div className="bg-gradient-to-br from-indigo-600 to-purple-600 rounded-xl shadow-lg p-6 text-white">
          <p className="text-sm font-medium opacity-90 mb-2">Total Orders</p>
          <p className="text-5xl font-bold">0</p>
        </div>

        {/* Success Rate */}
        <div className="bg-gradient-to-br from-purple-600 to-pink-500 rounded-xl shadow-lg p-6 text-white">
          <p className="text-sm font-medium opacity-90 mb-2">Success Rate</p>
          <p className="text-5xl font-bold">0%</p>
        </div>
      </div>
    </div>
  );
}
