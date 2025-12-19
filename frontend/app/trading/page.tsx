'use client';

import {useState, useEffect} from 'react';
import {useRouter} from 'next/navigation';
import {listBotConfigs, BotConfig} from '@/services/botConfigService';
import {
  placeOrder,
  getSymbols,
  PlaceOrderRequest,
  getAccountInfo,
  AccountInfo,
} from '@/services/tradingService';

export default function TradingPage() {
  const router = useRouter();
  const [botConfigs, setBotConfigs] = useState<BotConfig[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<BotConfig | null>(null);
  const [accountInfo, setAccountInfo] = useState<AccountInfo | null>(null);
  const [loadingAccount, setLoadingAccount] = useState(false);
  const [symbols, setSymbols] = useState<string[]>([]);
  const [loadingSymbols, setLoadingSymbols] = useState(false);
  const [orderType, setOrderType] = useState<'market' | 'limit'>('market');
  const [showWarning, setShowWarning] = useState(true);
  const [loading, setLoading] = useState(true);
  const [placing, setPlacing] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Form fields
  const [symbol, setSymbol] = useState('');
  const [amount, setAmount] = useState('');
  const [price, setPrice] = useState('');
  const [symbolSearch, setSymbolSearch] = useState('');
  const [showSymbolDropdown, setShowSymbolDropdown] = useState(false);

  useEffect(() => {
    fetchBotConfigs();
  }, []);

  const fetchBotConfigs = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        router.push('/login');
        return;
      }

      const data = await listBotConfigs();
      setBotConfigs(data.configs);

      // Auto-select default bot or first active bot
      const defaultBot = data.configs.find((c) => c.is_default && c.is_active);
      const firstActive = data.configs.find((c) => c.is_active);

      if (defaultBot) {
        handleSelectConfig(defaultBot);
      } else if (firstActive) {
        handleSelectConfig(firstActive);
      }
    } catch (err: any) {
      console.error('Error fetching bot configs:', err);
      if (err.response?.status === 401) {
        router.push('/login');
      } else {
        setError('Không thể tải danh sách bot config');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSelectConfig = async (config: BotConfig) => {
    setSelectedConfig(config);
    setSymbol(config.symbol || '');
    setSymbolSearch(config.symbol || '');
    setAmount(config.amount?.toString() || '');
    setError('');
    setSuccess('');

    // Fetch symbols from exchange
    setLoadingSymbols(true);
    setSymbols([]);
    try {
      const symbolsData = await getSymbols(config.id);
      setSymbols(symbolsData.symbols || []);
      console.log('Fetched symbols:', symbolsData.symbols);
    } catch (err: any) {
      console.error('Error fetching symbols:', err);
      // Don't show error to user, just log it
    } finally {
      setLoadingSymbols(false);
    }

    // Fetch account info from exchange
    setLoadingAccount(true);
    setAccountInfo(null);
    try {
      const info = await getAccountInfo(config.id);
      setAccountInfo(info);
    } catch (err: any) {
      console.error('Error fetching account info:', err);
      // Don't show error to user, just log it
    } finally {
      setLoadingAccount(false);
    }
  };

  // Filter symbols based on search
  const filteredSymbols = symbols.filter((sym) =>
    sym.toLowerCase().includes(symbolSearch.toLowerCase()),
  );

  const handleSymbolSelect = (sym: string) => {
    setSymbol(sym);
    setSymbolSearch(sym);
    setShowSymbolDropdown(false);
  };

  const handlePlaceOrder = async (side: 'buy' | 'sell') => {
    if (!selectedConfig) {
      setError('Vui lòng chọn Bot Config');
      return;
    }

    setError('');
    setSuccess('');
    setPlacing(true);

    try {
      const orderData: PlaceOrderRequest = {
        bot_config_id: selectedConfig.id,
        side,
        order_type: orderType,
        symbol: symbol || undefined,
        amount: amount ? parseFloat(amount) : undefined,
        price: orderType === 'limit' && price ? parseFloat(price) : undefined,
      };

      // Validate
      if (orderType === 'limit' && !price) {
        setError('Vui lòng nhập giá cho lệnh Limit');
        setPlacing(false);
        return;
      }

      console.log('Placing order with data:', orderData);
      // return;

      const result = await placeOrder(orderData);
      setSuccess(
        `Đặt lệnh ${side.toUpperCase()} thành công!\n` +
          `Order ID: ${result.order_id}\n` +
          `Symbol: ${result.symbol}\n` +
          `Amount: ${result.amount}\n` +
          `Status: ${result.order_status}`,
      );

      // Reset form
      setPrice('');

      // Navigate to orders page after 2 seconds
      setTimeout(() => {
        router.push('/orders');
      }, 2000);
    } catch (err: any) {
      console.error('Error placing order:', err);
      setError(
        err.response?.data?.error || 'Không thể đặt lệnh. Vui lòng thử lại.',
      );
    } finally {
      setPlacing(false);
    }
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
      <h1 className="text-3xl font-bold text-gray-900 mb-6">
        Đặt Lệnh Trading
      </h1>

      {/* Warning Alert */}
      {showWarning && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6 flex items-start gap-3">
          <span className="text-yellow-600 text-sm">⚠️</span>
          <div className="flex-1">
            <p className="text-sm text-yellow-800">
              <strong>Cảnh báo:</strong> Đây là lệnh THẬT trên SÀN GIAO DỊCH!
              {botConfigs.length === 0 && ' Không có bot config nào.'}
            </p>
          </div>
          <button
            onClick={() => setShowWarning(false)}
            className="text-yellow-600 hover:text-yellow-800">
            <svg
              className="w-5 h-5"
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
      )}

      {/* Error Alert */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {/* Success Alert */}
      {success && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-green-800 whitespace-pre-line">
            {success}
          </p>
        </div>
      )}

      {/* No Config Alert */}
      {botConfigs.length === 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-blue-800">
            Bạn chưa có bot config nào.
            <a href="/bot-configs" className="ml-1 underline font-semibold">
              Tạo bot config mới
            </a>
          </p>
        </div>
      )}

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left Column - Bot Config Selection */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900">
            Chọn Bot Config
          </h2>

          <div className="space-y-4">
            <select
              value={selectedConfig?.id || ''}
              onChange={(e) => {
                const config = botConfigs.find(
                  (c) => c.id === parseInt(e.target.value),
                );
                if (config) handleSelectConfig(config);
              }}
              className="w-full px-4 py-3 border-2 border-indigo-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
              disabled={botConfigs.length === 0}>
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

            {/* Selected Config Info */}
            {selectedConfig && (
              <div className="mt-4 p-4 bg-gray-50 rounded-lg space-y-2">
                <h3 className="font-semibold text-gray-900">Thông tin Bot:</h3>
                <div className="text-sm text-gray-600 space-y-1">
                  {selectedConfig.name && (
                    <p>
                      <strong>Name:</strong> {selectedConfig.name}
                    </p>
                  )}
                  <p>
                    <strong>Exchange:</strong>{' '}
                    {selectedConfig.exchange.toUpperCase()}
                  </p>
                  <p>
                    <strong>Symbol:</strong> {selectedConfig.symbol}
                  </p>
                  <p>
                    <strong>Trading Mode:</strong>{' '}
                    {selectedConfig.trading_mode || 'spot'}
                  </p>
                  {selectedConfig.leverage && (
                    <p>
                      <strong>Leverage:</strong> {selectedConfig.leverage}x
                    </p>
                  )}
                  {selectedConfig.amount && (
                    <p>
                      <strong>Default Amount:</strong> {selectedConfig.amount}
                    </p>
                  )}
                  <p>
                    <strong>Stop Loss:</strong>{' '}
                    {selectedConfig.stop_loss_percent}%
                  </p>
                  <p>
                    <strong>Take Profit:</strong>{' '}
                    {selectedConfig.take_profit_percent}%
                  </p>
                </div>
              </div>
            )}

            {/* Account Info Cards - Split Spot and Futures */}
            {selectedConfig && (
              <div className="mt-4 space-y-3">
                <h3 className="font-semibold text-gray-900 flex items-center justify-between">
                  <span>💰 Thông tin tài khoản trên sàn</span>
                  {loadingAccount && (
                    <span className="text-xs text-blue-600">Đang tải...</span>
                  )}
                </h3>

                {loadingAccount ? (
                  <div className="flex justify-center py-8">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                  </div>
                ) : accountInfo ? (
                  <div className="grid grid-cols-1 gap-3">
                    {/* Spot Account Card */}
                    {accountInfo.spot && (
                      <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                        <div className="flex items-center justify-between mb-3">
                          <h4 className="font-semibold text-blue-900 flex items-center gap-2">
                            <span>📊</span>
                            <span>SPOT Trading</span>
                          </h4>
                          <span className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded">
                            No Leverage
                          </span>
                        </div>

                        {/* Spot Summary */}
                        <div className="grid grid-cols-2 gap-2 mb-3">
                          <div className="bg-white p-3 rounded-lg">
                            <p className="text-xs text-gray-500">
                              Tổng tài sản
                            </p>
                            <p className="text-lg font-bold text-gray-900">
                              ${accountInfo.spot.total_balance.toFixed(2)}
                            </p>
                          </div>
                          <div className="bg-white p-3 rounded-lg">
                            <p className="text-xs text-gray-500">Khả dụng</p>
                            <p className="text-lg font-bold text-green-600">
                              ${accountInfo.spot.available_balance.toFixed(2)}
                            </p>
                          </div>
                        </div>

                        {/* Spot Balances */}
                        {accountInfo.spot.balances &&
                          accountInfo.spot.balances.length > 0 && (
                            <div className="bg-white p-3 rounded-lg">
                              <p className="text-xs text-gray-500 mb-2">
                                Tài sản chi tiết:
                              </p>
                              <div className="space-y-2 max-h-32 overflow-y-auto">
                                {accountInfo.spot.balances
                                  .filter((b) => b.total > 0.00001)
                                  .sort((a, b) => b.total - a.total)
                                  .map((balance, idx) => (
                                    <div
                                      key={idx}
                                      className="flex justify-between text-xs">
                                      <span className="font-medium text-blue-700">
                                        {balance.asset}
                                      </span>
                                      <div className="text-right">
                                        <span className="text-gray-900 font-semibold">
                                          {balance.total.toFixed(8)}
                                        </span>
                                        {balance.locked > 0 && (
                                          <span className="text-orange-600 ml-2">
                                            (🔒 {balance.locked.toFixed(8)})
                                          </span>
                                        )}
                                      </div>
                                    </div>
                                  ))}
                              </div>
                            </div>
                          )}
                      </div>
                    )}

                    {/* Futures Account Card */}
                    {accountInfo.futures && (
                      <div className="p-4 bg-purple-50 border border-purple-200 rounded-lg">
                        <div className="flex items-center justify-between mb-3">
                          <h4 className="font-semibold text-purple-900 flex items-center gap-2">
                            <span>🚀</span>
                            <span>FUTURES Trading</span>
                          </h4>
                          <span className="text-xs bg-purple-100 text-purple-700 px-2 py-1 rounded">
                            Up to 125x
                          </span>
                        </div>

                        {/* Futures Summary */}
                        <div className="grid grid-cols-2 gap-2 mb-3">
                          <div className="bg-white p-3 rounded-lg">
                            <p className="text-xs text-gray-500">
                              Tổng tài sản
                            </p>
                            <p className="text-lg font-bold text-gray-900">
                              ${accountInfo.futures.total_balance.toFixed(2)}
                            </p>
                          </div>
                          <div className="bg-white p-3 rounded-lg">
                            <p className="text-xs text-gray-500">Khả dụng</p>
                            <p className="text-lg font-bold text-green-600">
                              $
                              {accountInfo.futures.available_balance.toFixed(2)}
                            </p>
                          </div>
                        </div>

                        {/* Futures Balances */}
                        {accountInfo.futures.balances &&
                          accountInfo.futures.balances.length > 0 && (
                            <div className="bg-white p-3 rounded-lg">
                              <p className="text-xs text-gray-500 mb-2">
                                Tài sản chi tiết:
                              </p>
                              <div className="space-y-2 max-h-32 overflow-y-auto">
                                {accountInfo.futures.balances
                                  .filter((b) => b.total > 0.00001)
                                  .sort((a, b) => b.total - a.total)
                                  .map((balance, idx) => (
                                    <div
                                      key={idx}
                                      className="flex justify-between text-xs">
                                      <span className="font-medium text-purple-700">
                                        {balance.asset}
                                      </span>
                                      <div className="text-right">
                                        <span className="text-gray-900 font-semibold">
                                          {balance.total.toFixed(8)}
                                        </span>
                                        {balance.locked > 0 && (
                                          <span className="text-orange-600 ml-2">
                                            (🔒 {balance.locked.toFixed(8)})
                                          </span>
                                        )}
                                      </div>
                                    </div>
                                  ))}
                              </div>
                            </div>
                          )}
                      </div>
                    )}

                    {/* No Account Data */}
                    {!accountInfo.spot && !accountInfo.futures && (
                      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg text-center">
                        <p className="text-sm text-gray-500 italic">
                          Không có dữ liệu tài khoản
                        </p>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg text-center">
                    <p className="text-sm text-gray-500 italic">
                      Không thể tải thông tin tài khoản
                    </p>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Right Column - Order Form */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900">
            Thông Tin Lệnh
          </h2>

          <form className="space-y-4" onSubmit={(e) => e.preventDefault()}>
            {/* Symbol */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Symbol
              </label>
              {loadingSymbols ? (
                <div className="flex items-center gap-2 text-gray-500 text-sm">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-indigo-600"></div>
                  <span>Đang tải symbols...</span>
                </div>
              ) : symbols.length > 0 ? (
                <div className="relative">
                  <input
                    type="text"
                    value={symbolSearch}
                    onChange={(e) => {
                      setSymbolSearch(e.target.value);
                      setShowSymbolDropdown(true);
                    }}
                    onFocus={() => setShowSymbolDropdown(true)}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="Gõ để tìm symbol..."
                    disabled={!selectedConfig}
                  />
                  {showSymbolDropdown && filteredSymbols.length > 0 && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                      {filteredSymbols.map((sym) => (
                        <div
                          key={sym}
                          onClick={() => handleSymbolSelect(sym)}
                          className={`px-4 py-2 cursor-pointer hover:bg-indigo-50 ${
                            symbol === sym ? 'bg-indigo-100 font-semibold' : ''
                          }`}>
                          {sym}
                        </div>
                      ))}
                    </div>
                  )}
                  {showSymbolDropdown && (
                    <button
                      type="button"
                      onClick={() => setShowSymbolDropdown(false)}
                      className="fixed inset-0 w-full h-full cursor-default z-0"
                      tabIndex={-1}
                    />
                  )}
                </div>
              ) : (
                <input
                  type="text"
                  value={symbol}
                  onChange={(e) => setSymbol(e.target.value)}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                  placeholder="BTC/USDT"
                  disabled={!selectedConfig}
                />
              )}
              <p className="text-xs text-gray-500 mt-1">
                {symbols.length > 0
                  ? `${
                      symbols.length
                    } symbols từ ${selectedConfig?.exchange.toUpperCase()}`
                  : 'Để trống sẽ dùng symbol từ config'}
              </p>
            </div>

            {/* Order Type */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Order Type
              </label>
              <select
                value={orderType}
                onChange={(e) =>
                  setOrderType(e.target.value as 'market' | 'limit')
                }
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                disabled={!selectedConfig}>
                <option value="market">Market (Giá thị trường)</option>
                <option value="limit">Limit (Giá cố định)</option>
              </select>
            </div>

            {/* Price - Show only when Limit is selected */}
            {orderType === 'limit' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Price <span className="text-red-500">*</span>
                </label>
                <input
                  type="number"
                  step="0.00000001"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  className="w-full px-4 py-2 border border-yellow-300 bg-yellow-50 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                  placeholder="Nhập giá"
                  disabled={!selectedConfig}
                />
              </div>
            )}

            {/* Amount */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Amount
              </label>
              <input
                type="number"
                step="0.00000001"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                placeholder="Nhập số lượng"
                disabled={!selectedConfig}
              />
              <p className="text-xs text-gray-500 mt-1">
                Để trống sẽ dùng amount từ config
              </p>
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
                onClick={() => handlePlaceOrder('buy')}
                disabled={!selectedConfig || placing}
                className="w-full bg-green-500 hover:bg-green-600 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-semibold py-3 rounded-lg transition-colors">
                {placing ? 'Đang xử lý...' : 'Đặt lệnh BUY/LONG'}
              </button>
              <button
                type="button"
                onClick={() => handlePlaceOrder('sell')}
                disabled={!selectedConfig || placing}
                className="w-full bg-red-500 hover:bg-red-600 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-semibold py-3 rounded-lg transition-colors">
                {placing ? 'Đang xử lý...' : 'Đặt lệnh SELL/SHORT'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
