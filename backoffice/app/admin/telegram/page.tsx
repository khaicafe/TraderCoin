'use client';

import {useEffect, useState} from 'react';
import {useRouter} from 'next/navigation';

interface TelegramConfig {
  id: number;
  user_id: number;
  bot_token: string;
  chat_id: string;
  bot_name: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
  User: {
    id: number;
    email: string;
    full_name: string;
  };
}

export default function TelegramManagement() {
  const router = useRouter();
  const [configs, setConfigs] = useState<TelegramConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [showTestModal, setShowTestModal] = useState(false);
  const [testingConfig, setTestingConfig] = useState<TelegramConfig | null>(
    null,
  );
  const [testLoading, setTestLoading] = useState(false);
  const [testResult, setTestResult] = useState<{
    success: boolean;
    message: string;
  } | null>(null);

  // Edit Config Modal States
  const [showEditModal, setShowEditModal] = useState(false);
  const [editingConfig, setEditingConfig] = useState<TelegramConfig | null>(
    null,
  );
  const [formData, setFormData] = useState({
    bot_token: '',
    chat_id: '',
    bot_name: '',
    is_enabled: true,
  });
  const [saveLoading, setSaveLoading] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('admin_token');
    if (!token) {
      router.push('/');
      return;
    }

    fetchConfigs(token);
  }, [router]);

  const fetchConfigs = async (token: string) => {
    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/admin/telegram`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response.ok) {
        throw new Error('Failed to fetch telegram configs');
      }

      const data = await response.json();
      setConfigs(data.data || []);
    } catch (error) {
      console.error('Error fetching telegram configs:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchUsers = async () => {
    // Không cần nữa vì chỉ có 1 config
  };

  const handleOpenEditModal = () => {
    // Lấy config hiện tại (chỉ có 1 row duy nhất)
    if (configs.length > 0) {
      const config = configs[0];
      setEditingConfig(config);
      setFormData({
        bot_token: config.bot_token,
        chat_id: config.chat_id,
        bot_name: config.bot_name || '',
        is_enabled: config.is_enabled,
      });
      setShowEditModal(true);
    }
  };

  const handleCloseEditModal = () => {
    setShowEditModal(false);
    setEditingConfig(null);
    setFormData({
      bot_token: '',
      chat_id: '',
      bot_name: '',
      is_enabled: true,
    });
  };

  const handleSaveConfig = async () => {
    if (!formData.bot_token || !formData.chat_id) {
      alert('Bot Token and Chat ID are required');
      return;
    }

    if (!editingConfig) {
      alert('No configuration to edit');
      return;
    }

    const token = localStorage.getItem('admin_token');
    if (!token) return;

    setSaveLoading(true);

    try {
      // Sử dụng admin PUT endpoint để update config
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/admin/telegram/${editingConfig.id}`,
        {
          method: 'PUT',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            bot_token: formData.bot_token,
            chat_id: formData.chat_id,
            bot_name: formData.bot_name,
            is_enabled: formData.is_enabled,
          }),
        },
      );

      const data = await response.json();

      if (response.ok) {
        alert('Telegram configuration updated successfully!');
        handleCloseEditModal();
        fetchConfigs(token);
      } else {
        alert(data.error || 'Failed to update configuration');
      }
    } catch (error) {
      console.error('Error updating config:', error);
      alert('Failed to update configuration');
    } finally {
      setSaveLoading(false);
    }
  };

  const handleTestNotification = async () => {
    if (!testingConfig) return;

    const token = localStorage.getItem('admin_token');
    if (!token) return;

    setTestLoading(true);
    setTestResult(null);

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/admin/telegram/test-connection`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            bot_token: testingConfig.bot_token,
            chat_id: testingConfig.chat_id,
          }),
        },
      );

      const data = await response.json();

      if (response.ok) {
        setTestResult({
          success: true,
          message: 'Test message sent successfully! ✅',
        });
      } else {
        setTestResult({
          success: false,
          message: data.error || 'Failed to send test message',
        });
      }
    } catch (error) {
      setTestResult({
        success: false,
        message: 'Network error: Failed to connect to server',
      });
    } finally {
      setTestLoading(false);
    }
  };

  const handleOpenTestModal = (config: TelegramConfig) => {
    setTestingConfig(config);
    setShowTestModal(true);
    setTestResult(null);
  };

  const handleCloseTestModal = () => {
    setShowTestModal(false);
    setTestingConfig(null);
    setTestResult(null);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#EE4D2D] mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading telegram configs...</p>
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
            Telegram Configuration Management
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            View and test user Telegram notification configurations
          </p>
        </div>
        <button
          onClick={handleOpenEditModal}
          disabled={configs.length === 0}
          className="px-6 py-3 bg-gradient-to-r from-orange-400 to-orange-500 text-white rounded-lg hover:from-orange-500 hover:to-orange-600 transition-all shadow-lg hover:shadow-xl flex items-center gap-2 font-medium disabled:opacity-50 disabled:cursor-not-allowed">
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
          Edit Configuration
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow-md p-6 border-t-4 border-blue-500">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Configs</p>
              <p className="text-3xl font-bold text-gray-900">
                {configs.length}
              </p>
            </div>
            <div className="p-3 bg-blue-100 rounded-full">
              <svg
                className="w-8 h-8 text-blue-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                />
              </svg>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6 border-t-4 border-green-500">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Enabled</p>
              <p className="text-3xl font-bold text-green-600">
                {configs.filter((c) => c.is_enabled).length}
              </p>
            </div>
            <div className="p-3 bg-green-100 rounded-full">
              <svg
                className="w-8 h-8 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6 border-t-4 border-gray-400">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Disabled</p>
              <p className="text-3xl font-bold text-gray-600">
                {configs.filter((c) => !c.is_enabled).length}
              </p>
            </div>
            <div className="p-3 bg-gray-100 rounded-full">
              <svg
                className="w-8 h-8 text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"
                />
              </svg>
            </div>
          </div>
        </div>
      </div>

      {/* Configs Table */}
      <div className="bg-white rounded-lg shadow-md overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gradient-to-r from-orange-400 to-orange-500">
              <tr>
                <th className="px-6 py-4 text-left text-xs font-semibold text-white uppercase tracking-wider">
                  Bot Name
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-white uppercase tracking-wider">
                  Chat ID
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-white uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-white uppercase tracking-wider">
                  Created At
                </th>
                <th className="px-6 py-4 text-center text-xs font-semibold text-white uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {configs.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center">
                    <div className="flex flex-col items-center">
                      <svg
                        className="w-16 h-16 text-gray-300 mb-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                        />
                      </svg>
                      <p className="text-gray-500 text-lg font-medium">
                        No Telegram configurations found
                      </p>
                      <p className="text-gray-400 text-sm mt-1">
                        Users haven't set up their Telegram notifications yet
                      </p>
                    </div>
                  </td>
                </tr>
              ) : (
                configs.map((config) => (
                  <tr
                    key={config.id}
                    className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">
                        {config.bot_name || 'Unnamed Bot'}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-600 font-mono">
                        {config.chat_id}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {config.is_enabled ? (
                        <span className="px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                          Enabled
                        </span>
                      ) : (
                        <span className="px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">
                          Disabled
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDate(config.created_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <button
                        onClick={() => handleOpenTestModal(config)}
                        className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-blue-500 to-blue-600 text-white text-sm font-medium rounded-lg hover:from-blue-600 hover:to-blue-700 transition-all shadow-md hover:shadow-lg">
                        <svg
                          className="w-4 h-4 mr-2"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                          />
                        </svg>
                        Send Test
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Test Modal */}
      {showTestModal && testingConfig && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-2xl max-w-md w-full">
            <div className="bg-gradient-to-r from-blue-500 to-blue-600 text-white p-6 rounded-t-lg">
              <h2 className="text-2xl font-bold">Test Telegram Notification</h2>
              <p className="text-blue-100 text-sm mt-1">
                Send a test message to verify connection
              </p>
            </div>

            <div className="p-6">
              {/* User Info */}
              <div className="mb-6 p-4 bg-gray-50 rounded-lg">
                <p className="text-sm text-gray-600 mb-2">
                  <span className="font-semibold">User:</span>{' '}
                  {testingConfig.User?.full_name ||
                    testingConfig.User?.email ||
                    'N/A'}
                </p>
                <p className="text-sm text-gray-600 mb-2">
                  <span className="font-semibold">Bot Name:</span>{' '}
                  {testingConfig.bot_name || 'Unnamed Bot'}
                </p>
                <p className="text-sm text-gray-600">
                  <span className="font-semibold">Chat ID:</span>{' '}
                  <code className="bg-white px-2 py-1 rounded">
                    {testingConfig.chat_id}
                  </code>
                </p>
              </div>

              {/* Test Result */}
              {testResult && (
                <div
                  className={`mb-4 p-4 rounded-lg ${
                    testResult.success
                      ? 'bg-green-50 border border-green-200'
                      : 'bg-red-50 border border-red-200'
                  }`}>
                  <div className="flex items-start">
                    {testResult.success ? (
                      <svg
                        className="w-5 h-5 text-green-600 mt-0.5 mr-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    ) : (
                      <svg
                        className="w-5 h-5 text-red-600 mt-0.5 mr-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    )}
                    <div>
                      <p
                        className={`text-sm font-medium ${
                          testResult.success ? 'text-green-800' : 'text-red-800'
                        }`}>
                        {testResult.success ? 'Success!' : 'Error'}
                      </p>
                      <p
                        className={`text-sm mt-1 ${
                          testResult.success ? 'text-green-700' : 'text-red-700'
                        }`}>
                        {testResult.message}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Message Preview */}
              <div className="mb-6">
                <p className="text-sm font-semibold text-gray-700 mb-2">
                  Test Message Preview:
                </p>
                <div className="bg-gray-100 p-4 rounded-lg border border-gray-200">
                  <p className="text-sm text-gray-700">
                    ✅ <strong>Telegram Bot Connected Successfully!</strong>
                  </p>
                  <p className="text-sm text-gray-600 mt-2">
                    Your bot is ready to send notifications.
                  </p>
                </div>
              </div>
            </div>

            {/* Modal Actions */}
            <div className="bg-gray-50 px-6 py-4 rounded-b-lg flex gap-3 justify-end border-t">
              <button
                onClick={handleCloseTestModal}
                disabled={testLoading}
                className="px-6 py-2 border-2 border-gray-300 text-gray-700 rounded-lg hover:bg-gray-100 transition-colors font-medium disabled:opacity-50">
                Close
              </button>
              <button
                onClick={handleTestNotification}
                disabled={testLoading}
                className="px-6 py-2 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg hover:from-blue-600 hover:to-blue-700 transition-all font-medium shadow-md hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed flex items-center">
                {testLoading ? (
                  <>
                    <svg
                      className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                      fill="none"
                      viewBox="0 0 24 24">
                      <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"></circle>
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Sending...
                  </>
                ) : (
                  <>
                    <svg
                      className="w-4 h-4 mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                      />
                    </svg>
                    Send Test Message
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Config Modal */}
      {showEditModal && editingConfig && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-2xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-gradient-to-r from-orange-400 to-orange-500 text-white p-6 rounded-t-lg">
              <h2 className="text-2xl font-bold">
                Edit Telegram Configuration
              </h2>
              <p className="text-white/90 text-sm mt-1">
                Update Telegram notification settings
              </p>
            </div>

            <div className="p-6 space-y-4">
              {/* Current User Info */}
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
                <p className="text-sm font-semibold text-blue-900 mb-1">
                  Current User
                </p>
                <p className="text-sm text-blue-800">
                  {editingConfig.User?.full_name ||
                    editingConfig.User?.email ||
                    'N/A'}
                </p>
              </div>

              {/* Bot Token */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Bot Token *
                </label>
                <input
                  type="text"
                  placeholder="1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
                  value={formData.bot_token}
                  onChange={(e) =>
                    setFormData({...formData, bot_token: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent font-mono text-sm text-gray-900"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Get from @BotFather on Telegram
                </p>
              </div>

              {/* Chat ID */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Chat ID *
                </label>
                <input
                  type="text"
                  placeholder="123456789"
                  value={formData.chat_id}
                  onChange={(e) =>
                    setFormData({...formData, chat_id: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent font-mono text-sm text-gray-900"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Get from bot API getUpdates
                </p>
              </div>

              {/* Bot Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Bot Name
                </label>
                <input
                  type="text"
                  placeholder="My Trading Bot"
                  value={formData.bot_name}
                  onChange={(e) =>
                    setFormData({...formData, bot_name: e.target.value})
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent text-gray-900"
                />
              </div>

              {/* Is Enabled */}
              <div>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.is_enabled}
                    onChange={(e) =>
                      setFormData({...formData, is_enabled: e.target.checked})
                    }
                    className="rounded text-orange-600 focus:ring-orange-500"
                  />
                  <span className="text-sm font-medium text-gray-700">
                    Enable notifications
                  </span>
                </label>
              </div>

              {/* How to Get Info */}
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <p className="text-sm font-semibold text-blue-900 mb-2">
                  📖 How to get Bot Token and Chat ID:
                </p>
                <ol className="text-xs text-blue-800 space-y-1 list-decimal list-inside">
                  <li>Open Telegram and search for @BotFather</li>
                  <li>Send /newbot command and follow instructions</li>
                  <li>Copy the bot token provided</li>
                  <li>Start a conversation with your bot</li>
                  <li>
                    Visit: https://api.telegram.org/bot&lt;TOKEN&gt;/getUpdates
                  </li>
                  <li>
                    Find "chat":&#123;"id":123456789&#125; in the response
                  </li>
                </ol>
              </div>
            </div>

            {/* Modal Actions */}
            <div className="sticky bottom-0 bg-gray-50 px-6 py-4 rounded-b-lg flex gap-3 justify-end border-t">
              <button
                onClick={handleCloseEditModal}
                disabled={saveLoading}
                className="px-6 py-2 border-2 border-gray-300 text-gray-700 rounded-lg hover:bg-gray-100 transition-colors font-medium disabled:opacity-50">
                Cancel
              </button>
              <button
                onClick={handleSaveConfig}
                disabled={saveLoading}
                className="px-6 py-2 bg-gradient-to-r from-orange-500 to-orange-600 text-white rounded-lg hover:from-orange-600 hover:to-orange-700 transition-all font-medium shadow-md hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed flex items-center">
                {saveLoading ? (
                  <>
                    <svg
                      className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                      fill="none"
                      viewBox="0 0 24 24">
                      <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"></circle>
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Saving...
                  </>
                ) : (
                  'Save Changes'
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
