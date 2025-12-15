export default function SettingsPage() {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Settings</h1>
      
      <div className="bg-white rounded-lg shadow">
        {/* Profile Settings */}
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold">Profile Settings</h2>
        </div>
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Full Name
            </label>
            <input
              type="text"
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              placeholder="John Doe"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Email
            </label>
            <input
              type="email"
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              placeholder="john@example.com"
            />
          </div>
        </div>

        {/* Trading Settings */}
        <div className="px-6 py-4 border-t border-gray-200">
          <h2 className="text-xl font-semibold">Trading Settings</h2>
        </div>
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Default Stop Loss (%)
            </label>
            <input
              type="number"
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              defaultValue="5"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Default Take Profit (%)
            </label>
            <input
              type="number"
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              defaultValue="10"
            />
          </div>
          <div className="flex items-center">
            <input
              type="checkbox"
              id="auto-trading"
              className="h-4 w-4 text-blue-600 border-gray-300 rounded"
            />
            <label htmlFor="auto-trading" className="ml-2 text-sm text-gray-700">
              Enable Auto Trading
            </label>
          </div>
        </div>

        {/* Notification Settings */}
        <div className="px-6 py-4 border-t border-gray-200">
          <h2 className="text-xl font-semibold">Notifications</h2>
        </div>
        <div className="p-6 space-y-3">
          <div className="flex items-center">
            <input
              type="checkbox"
              id="email-notifications"
              className="h-4 w-4 text-blue-600 border-gray-300 rounded"
              defaultChecked
            />
            <label htmlFor="email-notifications" className="ml-2 text-sm text-gray-700">
              Email Notifications
            </label>
          </div>
          <div className="flex items-center">
            <input
              type="checkbox"
              id="trade-alerts"
              className="h-4 w-4 text-blue-600 border-gray-300 rounded"
              defaultChecked
            />
            <label htmlFor="trade-alerts" className="ml-2 text-sm text-gray-700">
              Trade Alerts
            </label>
          </div>
        </div>

        <div className="px-6 py-4 border-t border-gray-200">
          <button className="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700">
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}
