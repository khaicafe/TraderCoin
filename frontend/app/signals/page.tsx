'use client';

import {useState, useEffect, useCallback, useRef} from 'react';
import {useRouter} from 'next/navigation';
import {
  listSignals,
  executeSignal,
  updateSignalStatus,
  deleteSignal,
  TradingSignal,
} from '@/services/signalService';
import {
  listBotConfigs,
  BotConfig,
  updateBotConfig,
} from '@/services/botConfigService';
import {getAccountInfo, AccountInfo} from '@/services/tradingService';
import websocketService from '@/services/websocketService';
import {getWebhookPrefix, createWebhookPrefix} from '@/services/signalService';
import {getUser} from '@/services/authService';

export default function SignalsPage() {
  const router = useRouter();
  const [signals, setSignals] = useState<TradingSignal[]>([]);
  const [botConfigs, setBotConfigs] = useState<BotConfig[]>([]);
  const [selectedBotConfig, setSelectedBotConfig] = useState<number | null>(
    null,
  );
  const [accountInfo, setAccountInfo] = useState<AccountInfo | null>(null);
  const [loadingAccount, setLoadingAccount] = useState(false);
  const [loading, setLoading] = useState(true);
  const [executing, setExecuting] = useState<number | null>(null);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('');
  const [filterSymbol, setFilterSymbol] = useState<string>('');
  const [filterHours, setFilterHours] = useState<number>(6);
  const [wsStatus, setWsStatus] = useState<string>('DISCONNECTED');
  const [webhookPrefix, setWebhookPrefix] = useState<string>('');
  const [webhookURL, setWebhookURL] = useState<string>('');
  const [showWebhookModal, setShowWebhookModal] = useState(false);
  const [customPrefix, setCustomPrefix] = useState<string>('');
  const [currentUserId, setCurrentUserId] = useState<number | null>(null);
  const [testMode, setTestMode] = useState<boolean>(false); // 🧪 Test mode toggle

  // Amount editing state
  const [editingAmount, setEditingAmount] = useState<{[key: number]: boolean}>(
    {},
  );
  const [tempAmounts, setTempAmounts] = useState<{[key: number]: number}>({});

  // Use ref to store latest fetchData function for WebSocket handler
  const fetchDataRef = useRef<(() => Promise<void>) | null>(null);

  // Define fetchData with useCallback to avoid stale closure in WebSocket handler
  const fetchData = useCallback(async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        router.push('/login');
        return;
      }

      // Get current user ID
      const user = getUser();
      if (user?.id) {
        setCurrentUserId(user.id);
      }

      // Fetch signals
      const nowSec = Math.floor(Date.now() / 1000);
      const sinceTs =
        filterHours && filterHours > 0
          ? Math.floor(nowSec - filterHours * 3600)
          : undefined;

      const signalsData = await listSignals({
        status: filterStatus || undefined,
        symbol: filterSymbol || undefined,
        limit: 100,
        // Ưu tiên since_ts để chuẩn theo 'thời gian hiện tại' của client
        since_ts: sinceTs,
        // Giữ since_hours như fallback
        since_hours: !sinceTs ? filterHours || undefined : undefined,
      });
      setSignals(signalsData.signals || []);

      // Fetch bot configs
      const configsData = await listBotConfigs();
      setBotConfigs(configsData.configs || []);

      // Auto-select default or first active config
      const defaultConfig = configsData.configs.find(
        (c: BotConfig) => c.is_default && c.is_active,
      );
      const firstActive = configsData.configs.find(
        (c: BotConfig) => c.is_active,
      );
      if (defaultConfig) {
        setSelectedBotConfig(defaultConfig.id);
      } else if (firstActive) {
        setSelectedBotConfig(firstActive.id);
      }
    } catch (err: any) {
      console.error('Error fetching data:', err);
      if (err.response?.status === 401) {
        router.push('/login');
      } else {
        setError('Không thể tải dữ liệu');
      }
    } finally {
      setLoading(false);
    }
  }, [filterStatus, filterSymbol, filterHours, router]); // Dependencies: recreate function when filters change

  // Update ref whenever fetchData changes
  useEffect(() => {
    fetchDataRef.current = fetchData;
  }, [fetchData]);

  // Fetch data on mount and when filters change
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Fetch or generate webhook prefix for current user
  useEffect(() => {
    const initPrefix = async () => {
      try {
        const token = localStorage.getItem('token');
        if (!token) return;
        const data = await getWebhookPrefix();
        if (data.prefix) {
          setWebhookPrefix(data.prefix);
          setWebhookURL(data.url);
        }
      } catch (e) {
        console.warn('No existing webhook prefix, you can create one.');
      }
    };
    initPrefix();
  }, []);

  // � Fetch account info when bot config is selected
  useEffect(() => {
    const fetchAccountInfo = async () => {
      if (!selectedBotConfig) {
        setAccountInfo(null);
        return;
      }

      setLoadingAccount(true);
      setAccountInfo(null);
      try {
        const info = await getAccountInfo(selectedBotConfig);
        setAccountInfo(info);
        console.log('✅ Fetched account info:', info);
      } catch (err: any) {
        console.error('❌ Error fetching account info:', err);
        // Don't show error to user, just log it
      } finally {
        setLoadingAccount(false);
      }
    };

    fetchAccountInfo();
  }, [selectedBotConfig]);

  // �🔌 WebSocket Connection for Real-time Signal Notifications
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    // Connect to WebSocket
    websocketService.connect();

    // Update connection status periodically
    const statusInterval = setInterval(() => {
      setWsStatus(websocketService.getConnectionState());
    }, 1000);

    // 📡 Subscribe to signal_new events from TradingView webhook
    // Backend sends: { type: "signal_new", data: { signal_id: 123, symbol: "BTCUSDT", action: "buy", ... } }
    const unsubscribeSignalNew = websocketService.onMessage((message) => {
      console.log('🔍 WebSocket message received:', message);

      if (message.type === 'signal_new') {
        console.log('📥 New signal notification received:', message.data);

        // Show success notification
        setSuccess(
          `🔔 Signal mới từ TradingView!\nSymbol: ${message.data.symbol}\nAction: ${message.data.action}`,
        );

        // Refresh signals list using latest fetchData from ref
        console.log('♻️ Refreshing signals list...');
        if (fetchDataRef.current) {
          fetchDataRef.current();
        }

        // Auto-hide success message after 5 seconds
        setTimeout(() => setSuccess(''), 5000);
      }
    });

    // Cleanup
    return () => {
      console.log('🧹 Cleaning up WebSocket subscription');
      unsubscribeSignalNew();
      clearInterval(statusInterval);
      websocketService.disconnect();
    };
  }, [router]); // Only depend on router - WebSocket setup runs once

  const handleExecuteSignal = async (signalId: number) => {
    if (!selectedBotConfig) {
      setError('Vui lòng chọn Bot Config trước');
      return;
    }

    setError('');
    setSuccess('');
    setExecuting(signalId);

    try {
      const result = await executeSignal(signalId, {
        bot_config_id: selectedBotConfig,
        test_mode: testMode, // 🧪 Pass test mode flag
      });

      const modeLabel = testMode ? ' (TEST MODE - Mock Order)' : '';
      setSuccess(
        `✅ Đặt lệnh thành công!${modeLabel}\nOrder ID: ${result.order.id}\nSymbol: ${result.order.symbol}`,
      );

      // Refresh signals
      fetchData();
    } catch (err: any) {
      console.error('Error executing signal:', err);
      setError(
        err.response?.data?.error ||
          err.response?.data?.details ||
          'Không thể đặt lệnh',
      );
    } finally {
      setExecuting(null);
    }
  };

  const handleIgnoreSignal = async (signalId: number) => {
    try {
      await updateSignalStatus(signalId, 'ignored');
      setSuccess('Signal đã được đánh dấu là ignored');
      fetchData();
    } catch (err: any) {
      console.error('Error ignoring signal:', err);
      setError('Không thể cập nhật signal');
    }
  };

  const handleDeleteSignal = async (signalId: number) => {
    // if (!confirm('Bạn có chắc muốn xóa signal này?')) return;
    // try {
    //   await deleteSignal(signalId);
    //   setSuccess('Signal đã được xóa');
    //   fetchData();
    // } catch (err: any) {
    //   console.error('Error deleting signal:', err);
    //   setError('Không thể xóa signal');
    // }
  };

  const getActionBadge = (action: string) => {
    const badges: any = {
      buy: 'bg-green-100 text-green-800',
      sell: 'bg-red-100 text-red-800',
      close: 'bg-gray-100 text-gray-800',
    };
    return badges[action.toLowerCase()] || 'bg-blue-100 text-blue-800';
  };

  const getStatusBadge = (status: string) => {
    const badges: any = {
      pending: 'bg-yellow-100 text-yellow-800',
      executed: 'bg-green-100 text-green-800',
      failed: 'bg-red-100 text-red-800',
      ignored: 'bg-gray-100 text-gray-800',
    };
    return badges[status] || 'bg-blue-100 text-blue-800';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Đang tải...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">📡 Signals</h1>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setShowWebhookModal(true)}
            className="px-3 py-2 bg-green-600 text-white rounded hover:bg-green-700 text-sm">
            🔗 Tạo Prefix Signal
          </button>

          {/* WebSocket Status Indicator */}
          <div className="flex items-center gap-2">
            <div
              className={`w-3 h-3 rounded-full ${
                wsStatus === 'CONNECTED'
                  ? 'bg-green-500'
                  : wsStatus === 'CONNECTING'
                  ? 'bg-yellow-500 animate-pulse'
                  : 'bg-red-500'
              }`}
            />
            <span className="text-sm text-gray-600">
              {wsStatus === 'CONNECTED'
                ? 'Real-time active'
                : wsStatus === 'CONNECTING'
                ? 'Connecting...'
                : 'Disconnected'}
            </span>
          </div>

          {/* <button
            onClick={fetchData}
            className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
            🔄 Refresh
          </button> */}
        </div>
      </div>

      {/* Bot Config Selection */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Bot Config để đặt lệnh
            </label>
            <select
              value={selectedBotConfig || ''}
              onChange={(e) => setSelectedBotConfig(parseInt(e.target.value))}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 text-gray-900">
              <option value="">-- Chọn Bot Config --</option>
              {botConfigs
                .filter((c) => c.is_active)
                .map((config) => (
                  <option key={config.id} value={config.id}>
                    {config.name ||
                      `${config.exchange.toUpperCase()} - ${config.symbol}`}
                    {config.is_default ? ' (Default)' : ''}
                  </option>
                ))}
            </select>
          </div>

          {/* 🧪 Test Mode Toggle */}
          <div className="flex items-end">
            <div className="flex items-center">
              <input
                type="checkbox"
                id="testMode"
                checked={testMode}
                onChange={(e) => setTestMode(e.target.checked)}
                className="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
              />
              <label
                htmlFor="testMode"
                className="ml-2 text-sm text-gray-700 cursor-pointer">
                🧪 Test Mode{' '}
                <span className="text-xs text-gray-500">
                  (Bypass PlaceOrder)
                </span>
              </label>
            </div>
          </div>
        </div>
      </div>

      {/* Bot Config Info Cards - 3 Cards: Bot Config, Spot Trading, Futures Trading */}
      {selectedBotConfig &&
        botConfigs.length > 0 &&
        (() => {
          const config = botConfigs.find((c) => c.id === selectedBotConfig);
          if (!config) return null;

          const isFutures =
            config.trading_mode?.toLowerCase() === 'futures' ||
            config.trading_mode?.toLowerCase() === 'future';

          return (
            <div className="mb-4">
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-2">
                {/* Card 1: Bot Config Info */}
                <div className="p-2 bg-indigo-50 border border-indigo-200 rounded">
                  <div className="flex items-center justify-between mb-1.5">
                    <h4 className="font-semibold text-indigo-900 text-[11px] flex items-center gap-1">
                      <span>⚙️</span>
                      <span>Config</span>
                    </h4>
                    {config.is_default && (
                      <span className="text-[9px] bg-indigo-100 text-indigo-700 px-1 py-0.5 rounded">
                        Default
                      </span>
                    )}
                  </div>

                  <div className="space-y-1 text-[11px]">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Exchange:</span>
                      <span className="font-medium">
                        {config.exchange.toUpperCase()}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Symbol:</span>
                      <span className="font-medium">{config.symbol}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">Mode:</span>
                      <span
                        className={`px-1 py-0.5 rounded text-[9px] font-semibold ${
                          isFutures
                            ? 'bg-orange-100 text-orange-800'
                            : 'bg-blue-100 text-blue-800'
                        }`}>
                        {config.trading_mode || 'spot'}
                      </span>
                    </div>
                    {config.amount && config.amount > 0 && (
                      <div className="flex justify-between">
                        <span className="text-gray-600">Amount:</span>
                        <span className="font-medium">{config.amount}</span>
                      </div>
                    )}
                    <div className="pt-1 border-t border-indigo-200 space-y-0.5">
                      <div className="flex justify-between">
                        <span className="text-gray-600">SL/TP:</span>
                        <span>
                          <span className="text-red-600 font-semibold">
                            {config.stop_loss_percent}%
                          </span>
                          <span className="text-gray-400 mx-0.5">/</span>
                          <span className="text-green-600 font-semibold">
                            {config.take_profit_percent}%
                          </span>
                        </span>
                      </div>
                      <div className="flex justify-between text-[10px]">
                        <span className="text-gray-500">R:R</span>
                        <span className="text-gray-700">
                          {(
                            config.take_profit_percent /
                            config.stop_loss_percent
                          ).toFixed(1)}
                          :1
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Card 2: SPOT Trading Info */}
                <div className="p-2 bg-blue-50 border border-blue-200 rounded">
                  <div className="flex items-center justify-between mb-1.5">
                    <h4 className="font-semibold text-blue-900 text-[11px] flex items-center gap-1">
                      <span>�</span>
                      <span>SPOT</span>
                    </h4>
                    <span className="text-[9px] bg-blue-100 text-blue-700 px-1 py-0.5 rounded">
                      1x
                    </span>
                  </div>

                  {loadingAccount ? (
                    <div className="flex justify-center py-4">
                      <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600"></div>
                    </div>
                  ) : accountInfo?.spot ? (
                    <div className="space-y-1 text-[11px]">
                      <div className="flex justify-between">
                        <span className="text-gray-600">Tổng:</span>
                        <span className="font-bold text-gray-900">
                          ${accountInfo.spot.total_balance.toFixed(2)}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Khả dụng:</span>
                        <span className="font-bold text-green-600">
                          ${accountInfo.spot.available_balance.toFixed(2)}
                        </span>
                      </div>
                      {accountInfo.spot.balances &&
                        accountInfo.spot.balances.length > 0 && (
                          <div className="pt-1 border-t border-blue-200">
                            <p className="text-gray-500 mb-1 text-[9px]">
                              Top assets:
                            </p>
                            <div className="space-y-0.5 max-h-16 overflow-y-auto">
                              {accountInfo.spot.balances
                                .filter((b) => b.total > 0.00001)
                                .sort((a, b) => b.total - a.total)
                                .slice(0, 3)
                                .map((balance, idx) => (
                                  <div
                                    key={idx}
                                    className="flex justify-between text-[10px]">
                                    <span className="font-medium text-blue-700">
                                      {balance.asset}
                                    </span>
                                    <span className="text-gray-900">
                                      {balance.total.toFixed(4)}
                                    </span>
                                  </div>
                                ))}
                            </div>
                          </div>
                        )}
                    </div>
                  ) : (
                    <div className="space-y-1 text-[11px]">
                      <div className="flex justify-between">
                        <span className="text-gray-600">Tổng:</span>
                        <span className="font-bold text-gray-900">$0.00</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Khả dụng:</span>
                        <span className="font-bold text-green-600">$0.00</span>
                      </div>
                      <div className="pt-1 border-t border-blue-200">
                        <p className="text-gray-400 italic text-center text-[9px] py-1">
                          Chưa có dữ liệu
                        </p>
                      </div>
                    </div>
                  )}
                </div>

                {/* Card 3: FUTURES Trading Info */}
                <div className="p-2 bg-purple-50 border border-purple-200 rounded">
                  <div className="flex items-center justify-between mb-1.5">
                    <h4 className="font-semibold text-purple-900 text-[11px] flex items-center gap-1">
                      <span>🚀</span>
                      <span>FUTURES</span>
                    </h4>
                    <span className="text-[9px] bg-purple-100 text-purple-700 px-1 py-0.5 rounded">
                      125x
                    </span>
                  </div>

                  {loadingAccount ? (
                    <div className="flex justify-center py-4">
                      <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-purple-600"></div>
                    </div>
                  ) : accountInfo?.futures ? (
                    <div className="space-y-1 text-[11px]">
                      <div className="flex justify-between">
                        <span className="text-gray-600">Tổng:</span>
                        <span className="font-bold text-gray-900">
                          ${accountInfo.futures.total_balance.toFixed(2)}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Khả dụng:</span>
                        <span className="font-bold text-green-600">
                          ${accountInfo.futures.available_balance.toFixed(2)}
                        </span>
                      </div>
                      {accountInfo.futures.balances &&
                        accountInfo.futures.balances.length > 0 && (
                          <div className="pt-1 border-t border-purple-200">
                            <p className="text-gray-500 mb-1 text-[9px]">
                              Tài sản chi tiết:
                            </p>
                            <div className="space-y-0.5 max-h-20 overflow-y-auto">
                              {accountInfo.futures.balances
                                .filter((b) => b.total > 0.00001)
                                .sort((a, b) => b.total - a.total)
                                .map((balance, idx) => (
                                  <div
                                    key={idx}
                                    className="flex justify-between text-[10px]">
                                    <span className="font-medium text-purple-700">
                                      {balance.asset}
                                    </span>
                                    <div className="text-right">
                                      <span className="text-gray-900 font-semibold">
                                        {balance.total.toFixed(6)}
                                      </span>
                                      {balance.locked > 0 && (
                                        <span className="text-orange-600 ml-1 text-[9px]">
                                          (🔒{balance.locked.toFixed(6)})
                                        </span>
                                      )}
                                    </div>
                                  </div>
                                ))}
                            </div>
                          </div>
                        )}
                    </div>
                  ) : (
                    <div className="space-y-1 text-[11px]">
                      <div className="flex justify-between">
                        <span className="text-gray-600">Tổng:</span>
                        <span className="font-bold text-gray-900">$0.00</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Khả dụng:</span>
                        <span className="font-bold text-green-600">$0.00</span>
                      </div>
                      <div className="pt-1 border-t border-purple-200">
                        <p className="text-gray-400 italic text-center text-[9px] py-1">
                          Chưa có dữ liệu
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          );
        })()}

      {/* Error/Success Alerts */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {success && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-green-800 whitespace-pre-line">
            {success}
          </p>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Filter by Status */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Filter by Status
            </label>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 text-gray-900">
              <option value="">All Status</option>
              <option value="pending">Pending</option>
              <option value="executed">Executed</option>
              <option value="failed">Failed</option>
              <option value="ignored">Ignored</option>
            </select>
          </div>

          {/* Filter by Symbol */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Filter by Symbol
            </label>
            <input
              type="text"
              value={filterSymbol}
              onChange={(e) => setFilterSymbol(e.target.value)}
              placeholder="BTCUSDT"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 text-gray-900"
            />
          </div>

          {/* Filter by last N hours */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Thời gian (giờ gần đây)
            </label>
            <select
              value={filterHours}
              onChange={(e) => setFilterHours(parseFloat(e.target.value))}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 text-gray-900">
              <option value={0.5}>30 phút</option>
              <option value={1}>1 giờ</option>
              <option value={6}>6 giờ</option>
              <option value={12}>12 giờ</option>
              <option value={24}>24 giờ</option>
              <option value={48}>48 giờ</option>
              <option value={72}>72 giờ</option>
              <option value={168}>7 ngày</option>
              <option value={0}>Tất cả</option>
            </select>
          </div>
        </div>
      </div>

      {/* Signals List */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Time
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Symbol
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Prefix
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Action
                </th>
                {/* <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Price
                </th> */}
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Amount
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  SL / TP
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Strategy
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {signals.length === 0 ? (
                <tr>
                  <td
                    colSpan={10}
                    className="px-6 py-12 text-center text-gray-500">
                    <div className="text-4xl mb-2">📭</div>
                    <p>Chưa có signal nào</p>
                    <p className="text-sm mt-2">
                      Tạo Alert trên TradingView và gửi webhook đến backend
                    </p>
                  </td>
                </tr>
              ) : (
                signals.map((signal) => {
                  const selectedConfig = botConfigs.find(
                    (c) => c.id === selectedBotConfig,
                  );
                  const displayAmount =
                    tempAmounts[signal.id] !== undefined
                      ? tempAmounts[signal.id]
                      : selectedConfig?.amount || 0;

                  return (
                    <tr key={signal.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {new Date(signal.received_at).toLocaleString('vi-VN')}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="font-semibold text-gray-900">
                          {signal.symbol}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                        {signal.webhook_prefix || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 py-1 text-xs font-semibold rounded-full ${getActionBadge(
                            signal.action,
                          )}`}>
                          {signal.action.toUpperCase()}
                        </span>
                      </td>
                      {/* <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {signal.price > 0 ? `$${signal.price.toFixed(2)}` : '-'}
                      </td> */}
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        {editingAmount[signal.id] ? (
                          <div className="flex items-center gap-2">
                            <input
                              type="number"
                              step="0.01"
                              value={displayAmount}
                              onChange={(e) => {
                                const value = parseFloat(e.target.value) || 0;
                                setTempAmounts((prev) => ({
                                  ...prev,
                                  [signal.id]: value,
                                }));
                              }}
                              className="w-24 px-2 py-1 border border-purple-300 rounded focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                              autoFocus
                            />
                            <button
                              onClick={async () => {
                                if (selectedBotConfig) {
                                  try {
                                    await updateBotConfig(selectedBotConfig, {
                                      amount: tempAmounts[signal.id],
                                    });
                                    setEditingAmount((prev) => ({
                                      ...prev,
                                      [signal.id]: false,
                                    }));
                                    setSuccess('Đã cập nhật amount');
                                    fetchData();
                                  } catch (err: any) {
                                    setError('Không thể cập nhật amount');
                                  }
                                }
                              }}
                              className="text-green-600 hover:text-green-800">
                              ✓
                            </button>
                            <button
                              onClick={() => {
                                setEditingAmount((prev) => ({
                                  ...prev,
                                  [signal.id]: false,
                                }));
                                setTempAmounts((prev) => {
                                  const newAmounts = {...prev};
                                  delete newAmounts[signal.id];
                                  return newAmounts;
                                });
                              }}
                              className="text-red-600 hover:text-red-800">
                              ✕
                            </button>
                          </div>
                        ) : (
                          <div
                            className="flex items-center gap-2 cursor-pointer group"
                            onClick={() => {
                              setEditingAmount((prev) => ({
                                ...prev,
                                [signal.id]: true,
                              }));
                              setTempAmounts((prev) => ({
                                ...prev,
                                [signal.id]: selectedConfig?.amount || 0,
                              }));
                            }}>
                            <span className="font-semibold text-gray-900">
                              {displayAmount > 0
                                ? displayAmount.toFixed(4)
                                : '-'}
                            </span>
                            <span className="text-gray-400 group-hover:text-purple-600 text-xs">
                              ✎
                            </span>
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        <div className="space-y-1">
                          {(() => {
                            if (selectedConfig) {
                              return (
                                <>
                                  {selectedConfig.stop_loss_percent > 0 && (
                                    <div className="text-red-600 font-semibold">
                                      SL: {selectedConfig.stop_loss_percent}%
                                    </div>
                                  )}
                                  {selectedConfig.take_profit_percent > 0 && (
                                    <div className="text-green-600 font-semibold">
                                      TP: {selectedConfig.take_profit_percent}%
                                    </div>
                                  )}
                                  {selectedConfig.stop_loss_percent === 0 &&
                                    selectedConfig.take_profit_percent ===
                                      0 && (
                                      <span className="text-gray-400">-</span>
                                    )}
                                </>
                              );
                            }
                            return <span className="text-gray-400">-</span>;
                          })()}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                        {signal.strategy || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 py-1 text-xs font-semibold rounded-full ${getStatusBadge(
                            signal.status,
                          )}`}>
                          {signal.status.toUpperCase()}
                        </span>
                        {signal.order_id && (
                          <div className="text-xs text-gray-500 mt-1">
                            Order #{signal.order_id}
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        <div className="flex items-center gap-2">
                          {signal.status === 'pending' && (
                            // {signal.status === 'failed' && (
                            <>
                              <button
                                onClick={() => handleExecuteSignal(signal.id)}
                                // disabled={
                                //   !selectedBotConfig ||
                                //   executing === signal.id ||
                                //   (signal.executed_by_user_id ===
                                //     currentUserId &&
                                //     currentUserId !== null)
                                // }
                                className="px-3 py-1 bg-green-500 hover:bg-green-600 disabled:bg-gray-300 disabled:cursor-not-allowed text-white rounded text-xs font-semibold"
                                title={
                                  signal.executed_by_user_id ===
                                    currentUserId && currentUserId !== null
                                    ? 'Bạn đã đặt lệnh cho signal này'
                                    : undefined
                                }>
                                {executing === signal.id
                                  ? '⏳'
                                  : signal.executed_by_user_id ===
                                      currentUserId && currentUserId !== null
                                  ? '✅ Đã đặt'
                                  : '✅ Đặt lệnh'}
                              </button>
                            </>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Webhook Info Modal */}
      {showWebhookModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="bg-white w-full max-w-2xl rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                📖 Hướng dẫn cấu hình Signal Alert
              </h3>
              <button
                onClick={() => setShowWebhookModal(false)}
                className="text-gray-500 hover:text-gray-700">
                ✕
              </button>
            </div>
            <div className="space-y-3 text-sm text-gray-800">
              <div>
                <p className="font-medium mb-1">Tạo Webhook Prefix:</p>
                <div className="flex items-center gap-3">
                  <input
                    type="text"
                    value={customPrefix}
                    onChange={(e) => setCustomPrefix(e.target.value)}
                    placeholder="Để trống để tự động tạo"
                    className="flex-1 px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-indigo-500 text-gray-900"
                  />
                  <button
                    onClick={async () => {
                      try {
                        const res = await createWebhookPrefix(
                          customPrefix.trim() || undefined,
                        );
                        console.log('Created new webhook prefix:', res);
                        setWebhookPrefix(res.prefix);
                        setWebhookURL(res.url);
                        setCustomPrefix('');
                        setSuccess(`✅ Prefix tạo thành công: ${res.prefix}`);
                        setTimeout(() => setSuccess(''), 3000);
                      } catch (e: any) {
                        setError(
                          e.response?.data?.error || 'Không thể tạo prefix',
                        );
                      }
                    }}
                    className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm whitespace-nowrap">
                    Tạo prefix
                  </button>
                </div>
              </div>
              <div>
                <p className="font-medium mb-1">
                  Webhook URL (POST để gửi signal từ TradingView):
                </p>
                <code className="block bg-gray-100 p-3 rounded border border-gray-300 text-gray-900 break-all">
                  {typeof window !== 'undefined'
                    ? `${window.location.origin}/api/v1/signals/webhook${
                        webhookPrefix ? `/${webhookPrefix}` : ''
                      }`
                    : `http://localhost:8080/api/v1/signals/webhook${
                        webhookPrefix ? `/${webhookPrefix}` : ''
                      }`}
                </code>
                {!webhookPrefix && (
                  <p className="text-xs text-gray-500 mt-1">
                    Chưa có prefix — tạo prefix ở trên để dùng link này.
                  </p>
                )}
              </div>
              <div>
                {!webhookPrefix && (
                  <p className="text-xs text-gray-500 mt-1">
                    Chưa có prefix — bấm "Tạo prefix" để tạo và dùng link này.
                  </p>
                )}
              </div>
              <div>
                <p className="font-medium mb-1">Message Format (JSON):</p>
                <pre className="block bg-gray-100 p-3 rounded border border-gray-300 text-gray-900 overflow-x-auto">
                  {`{
  "symbol": "{{ticker}}",
  "action": "buy",
  "price": {{close}},
  "stop_loss": {{low}},
  "take_profit": {{high}},
  "strategy": "My Strategy",
  "message": "Signal at {{time}}",
  "timestamp": {{timenow}}
}`}
                </pre>
                <p className="mt-2">
                  <strong>Actions:</strong> <code>buy</code>, <code>sell</code>,{' '}
                  <code>close</code>
                </p>
              </div>
            </div>
            <div className="mt-4 text-right">
              <button
                onClick={() => setShowWebhookModal(false)}
                className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300">
                Đóng
              </button>
            </div>
          </div>
        </div>
      )}

      {/* (Đã bỏ modal Tạo Signal vì yêu cầu: chỉ tạo prefix cho user, không gửi signal thủ công) */}
    </div>
  );
}
