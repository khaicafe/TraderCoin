'use client';

import {useState, useEffect} from 'react';
import {
  BotConfig,
  listBotConfigs,
  createBotConfig,
  deleteBotConfig,
  updateBotConfig,
  setDefaultBotConfig,
  BotConfigCreate,
} from '../../services/botConfigService';
import {getSymbols, Symbol} from '../../services/symbolService';

export default function BotConfigsPage() {
  const [bots, setBots] = useState<BotConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showModal, setShowModal] = useState(false);
  const [editingBot, setEditingBot] = useState<BotConfig | null>(null);

  // Symbol state
  const [symbols, setSymbols] = useState<Symbol[]>([]);
  const [loadingSymbols, setLoadingSymbols] = useState(false);
  const [symbolSearch, setSymbolSearch] = useState('');
  const [showSymbolDropdown, setShowSymbolDropdown] = useState(false);

  // Exchange state
  const [exchangeSearch, setExchangeSearch] = useState('');
  const [showExchangeDropdown, setShowExchangeDropdown] = useState(false);

  // Trading Mode state
  const [tradingModeSearch, setTradingModeSearch] = useState('');
  const [showTradingModeDropdown, setShowTradingModeDropdown] = useState(false);

  const exchanges = [
    {
      value: 'binance',
      label: 'Binance',
      description: 'World largest crypto exchange',
    },
    {value: 'bittrex', label: 'Bittrex', description: 'US-based exchange'},
  ];

  const tradingModes = [
    {value: 'spot', label: 'Spot', description: 'Giao dịch thường'},
    {value: 'futures', label: 'Futures', description: 'Hợp đồng tương lai'},
    {value: 'margin', label: 'Margin', description: 'Giao dịch ký quỹ'},
  ];

  const initialFormData = {
    name: '',
    symbol: '',
    exchange: 'binance',
    trading_mode: 'spot',
    leverage: '1',
    amount: '',
    api_key: '',
    api_secret: '',
    stop_loss_percent: '',
    take_profit_percent: '',
  };

  const [formData, setFormData] = useState(initialFormData);

  const fetchConfigs = async () => {
    try {
      setLoading(true);
      const data = await listBotConfigs();
      console.log('Bot configs fetched:', data.configs); // Debug log
      setBots(data.configs || []);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch bot configs:', err);
      setError('Failed to load bot configurations. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  // Format symbol for display: BTCUSDT -> BTC/USDT
  const formatSymbol = (symbol: string): string => {
    if (symbol.endsWith('USDT')) {
      return symbol.replace('USDT', '/USDT');
    }
    return symbol;
  };

  // Filter symbols based on search
  const filteredSymbols = symbols.filter((sym) =>
    sym.symbol.toLowerCase().includes(symbolSearch.toLowerCase()),
  );

  // Filter exchanges based on search
  const filteredExchanges = exchanges.filter(
    (ex) =>
      ex.label.toLowerCase().includes(exchangeSearch.toLowerCase()) ||
      ex.value.toLowerCase().includes(exchangeSearch.toLowerCase()),
  );

  // Filter trading modes based on search
  const filteredTradingModes = tradingModes.filter(
    (mode) =>
      mode.label.toLowerCase().includes(tradingModeSearch.toLowerCase()) ||
      mode.value.toLowerCase().includes(tradingModeSearch.toLowerCase()),
  );

  useEffect(() => {
    // Check for token before fetching
    const token = localStorage.getItem('token');
    if (token) {
      fetchConfigs();
    } else {
      setError('You must be logged in to view this page.');
      setLoading(false);
    }
  }, []);

  // Load symbols when exchange or trading_mode changes
  useEffect(() => {
    const loadSymbols = async () => {
      if (showModal && formData.exchange) {
        setLoadingSymbols(true);
        try {
          const symbolsList = await getSymbols(
            formData.exchange,
            formData.trading_mode,
          );
          setSymbols(symbolsList);
        } catch (err) {
          console.error('Failed to load symbols:', err);
          setSymbols([]);
        } finally {
          setLoadingSymbols(false);
        }
      }
    };
    loadSymbols();
  }, [showModal, formData.exchange, formData.trading_mode]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      // Don't close if clicking inside dropdown or input
      if (target.closest('.dropdown-container')) {
        return;
      }
      setShowSymbolDropdown(false);
      setShowExchangeDropdown(false);
      setShowTradingModeDropdown(false);
    };
    if (showSymbolDropdown || showExchangeDropdown || showTradingModeDropdown) {
      // Use setTimeout to avoid immediate trigger
      setTimeout(() => {
        document.addEventListener('click', handleClickOutside);
      }, 0);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [showSymbolDropdown, showExchangeDropdown, showTradingModeDropdown]);

  const handleOpenModal = (bot: BotConfig | null = null) => {
    if (bot) {
      setEditingBot(bot);
      setFormData({
        name: bot.name || '',
        symbol: bot.symbol,
        exchange: bot.exchange,
        trading_mode: bot.trading_mode || 'spot',
        leverage: bot.leverage ? String(bot.leverage) : '1',
        amount:
          bot.amount !== undefined && bot.amount !== null
            ? String(bot.amount)
            : '0',
        api_key: '', // Not stored in response for security
        api_secret: '', // Not stored in response for security
        stop_loss_percent: String(bot.stop_loss_percent),
        take_profit_percent: String(bot.take_profit_percent),
      });
    } else {
      setEditingBot(null);
      setFormData(initialFormData);
    }
    setSymbolSearch('');
    setShowSymbolDropdown(false);
    setExchangeSearch('');
    setShowExchangeDropdown(false);
    setTradingModeSearch('');
    setShowTradingModeDropdown(false);
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setEditingBot(null);
    setFormData(initialFormData);
    setSymbolSearch('');
    setShowSymbolDropdown(false);
    setExchangeSearch('');
    setShowExchangeDropdown(false);
    setTradingModeSearch('');
    setShowTradingModeDropdown(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const configData: BotConfigCreate = {
        name: formData.name,
        symbol: formData.symbol,
        exchange: formData.exchange,
        trading_mode: formData.trading_mode,
        leverage: formData.leverage ? parseInt(formData.leverage) : undefined,
        amount: formData.amount ? parseFloat(formData.amount) : undefined,
        api_key: formData.api_key || undefined,
        api_secret: formData.api_secret || undefined,
        stop_loss_percent: parseFloat(formData.stop_loss_percent),
        take_profit_percent: parseFloat(formData.take_profit_percent),
      };

      if (editingBot) {
        await updateBotConfig(editingBot.id, {
          ...configData,
          is_active: editingBot.is_active,
        });
        alert('Configuration updated successfully!');
      } else {
        await createBotConfig(configData);
        alert('Configuration created successfully!');
      }

      fetchConfigs();
      handleCloseModal();
    } catch (err) {
      console.error('Failed to save config:', err);
      alert('Failed to save configuration.');
    }
  };

  const handleDelete = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this configuration?')) {
      try {
        await deleteBotConfig(id);
        alert('Configuration deleted successfully!');
        fetchConfigs();
      } catch (err) {
        console.error('Failed to delete config:', err);
        alert('Failed to delete configuration.');
      }
    }
  };

  const handleToggleStatus = async (bot: BotConfig) => {
    try {
      await updateBotConfig(bot.id, {is_active: !bot.is_active});
      alert('Status updated successfully!');
      fetchConfigs();
    } catch (err) {
      console.error('Failed to update status:', err);
      alert('Failed to update status.');
    }
  };

  const handleSetDefault = async (bot: BotConfig) => {
    try {
      await setDefaultBotConfig(bot.id);
      alert('Default bot set successfully!');
      fetchConfigs();
    } catch (err) {
      console.error('Failed to set default:', err);
      alert('Failed to set default bot.');
    }
  };

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Bot Configurations</h1>
        <button
          onClick={() => handleOpenModal()}
          className="flex items-center gap-2 bg-indigo-600 text-white px-6 py-3 rounded-lg hover:bg-indigo-700 transition-colors shadow-md">
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          Tạo Config Mới
        </button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  ID
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Name
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Symbol
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Sàn
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Mode
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Leverage
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Amount
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  SL %
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  TP %
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Default
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Status
                </th>
                <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">
                  Actions
                </th>
              </tr>
            </thead>
            {!loading && !error && (
              <tbody className="divide-y divide-gray-200">
                {bots.map((bot) => (
                  <tr key={bot.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {bot.id}
                    </td>
                    <td className="px-6 py-4 text-sm font-medium text-gray-900">
                      {bot.name || '-'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {formatSymbol(bot.symbol)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900 capitalize">
                      {bot.exchange}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900 capitalize">
                      {bot.trading_mode || 'spot'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {bot.leverage ? `${bot.leverage}x` : '1x'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {bot.amount !== undefined && bot.amount !== null
                        ? `${bot.amount}`
                        : '-'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {bot.stop_loss_percent}%
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {bot.take_profit_percent}%
                    </td>
                    <td className="px-6 py-4">
                      {bot.is_default ? (
                        <span className="px-3 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                          Default
                        </span>
                      ) : (
                        <button
                          onClick={() => handleSetDefault(bot)}
                          className="text-xs text-gray-500 hover:text-blue-600 hover:underline">
                          Set Default
                        </button>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`px-3 py-1 text-xs font-semibold rounded-full ${
                          bot.is_active
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}>
                        {bot.is_active ? 'Running' : 'Stopped'}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => handleOpenModal(bot)}
                          className="p-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                          title="Edit">
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
                        </button>
                        <button
                          onClick={() => handleDelete(bot.id)}
                          className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                          title="Delete">
                          <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24">
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                            />
                          </svg>
                        </button>
                        <button
                          onClick={() => handleToggleStatus(bot)}
                          className="p-2 text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                          title={bot.is_active ? 'Stop' : 'Start'}>
                          <svg
                            className="w-5 h-5"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24">
                            {bot.is_active ? (
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z"
                              />
                            ) : (
                              <>
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                                />
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                />
                              </>
                            )}
                          </svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            )}
          </table>
        </div>

        {/* Loading State */}
        {loading && (
          <div className="text-center py-12">
            <p className="text-gray-500">Loading configurations...</p>
          </div>
        )}

        {/* Error State */}
        {error && (
          <div className="text-center py-12">
            <p className="text-red-500 font-medium">{error}</p>
          </div>
        )}

        {/* Empty State */}
        {!loading && !error && bots.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">Chưa có bot nào được cấu hình</p>
            <button
              onClick={() => handleOpenModal()}
              className="mt-4 text-indigo-600 hover:text-indigo-700 font-medium">
              Tạo bot đầu tiên
            </button>
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div
          className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4 animate-fadeIn"
          onClick={handleCloseModal}>
          <div
            className="bg-white rounded-lg shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto animate-slideUp"
            onClick={(e) => e.stopPropagation()}>
            {/* Modal Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h2 className="text-2xl font-bold text-gray-900">
                {editingBot ? 'Edit Configuration' : 'Tạo Config Mới'}
              </h2>
              <button
                onClick={handleCloseModal}
                className="text-gray-400 hover:text-gray-600 transition-colors">
                <svg
                  className="w-6 h-6"
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

            {/* Modal Body */}
            <form onSubmit={handleSubmit} className="p-6 space-y-6">
              {/* Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({...formData, name: e.target.value})
                  }
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                  placeholder="My BTC Strategy"
                  required
                />
              </div>

              {/* Exchange & Symbol - 2 columns */}
              <div className="grid grid-cols-2 gap-4">
                {/* Sàn Giao Dịch */}
                <div className="relative dropdown-container">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Sàn Giao Dịch
                  </label>
                  <input
                    type="text"
                    value={
                      formData.exchange
                        ? exchanges.find((ex) => ex.value === formData.exchange)
                            ?.label || formData.exchange
                        : exchangeSearch
                    }
                    onChange={(e) => {
                      setExchangeSearch(e.target.value);
                      setShowExchangeDropdown(true);
                      if (!e.target.value) {
                        setFormData({...formData, exchange: '', symbol: ''});
                      }
                    }}
                    onFocus={() => {
                      setShowExchangeDropdown(true);
                      if (!formData.exchange && !exchangeSearch) {
                        // Show full list when focusing on empty field
                        setExchangeSearch('');
                      }
                    }}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="Search exchange..."
                    required
                    autoComplete="off"
                  />
                  {showExchangeDropdown && filteredExchanges.length > 0 && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                      {filteredExchanges.map((ex) => (
                        <button
                          key={ex.value}
                          type="button"
                          onClick={() => {
                            setFormData({
                              ...formData,
                              exchange: ex.value,
                              symbol: '', // Reset symbol when exchange changes
                            });
                            setExchangeSearch('');
                            setShowExchangeDropdown(false);
                          }}
                          className="w-full px-4 py-3 text-left hover:bg-indigo-50 text-gray-900 transition-colors border-b last:border-b-0">
                          <div className="font-medium">{ex.label}</div>
                          <div className="text-xs text-gray-500">
                            {ex.description}
                          </div>
                        </button>
                      ))}
                    </div>
                  )}
                  {formData.exchange && (
                    <p className="text-xs text-gray-500 mt-1">
                      Selected:{' '}
                      {
                        exchanges.find((ex) => ex.value === formData.exchange)
                          ?.label
                      }
                    </p>
                  )}
                </div>

                {/* Symbol */}
                <div className="relative dropdown-container">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Symbol
                  </label>
                  <input
                    type="text"
                    value={
                      formData.symbol
                        ? formatSymbol(formData.symbol)
                        : symbolSearch
                    }
                    onChange={(e) => {
                      setSymbolSearch(e.target.value);
                      setShowSymbolDropdown(true);
                      if (!e.target.value) {
                        setFormData({...formData, symbol: ''});
                      }
                    }}
                    onFocus={() => {
                      setShowSymbolDropdown(true);
                      if (!formData.symbol && !symbolSearch) {
                        // Show full list when focusing on empty field
                        setSymbolSearch('');
                      }
                    }}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder={
                      loadingSymbols ? 'Loading...' : 'Search symbol...'
                    }
                    required
                    disabled={loadingSymbols}
                    autoComplete="off"
                  />
                  {showSymbolDropdown && filteredSymbols.length > 0 && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                      {filteredSymbols.map((sym) => (
                        <button
                          key={sym.symbol}
                          type="button"
                          onClick={() => {
                            setFormData({...formData, symbol: sym.symbol});
                            setSymbolSearch('');
                            setShowSymbolDropdown(false);
                          }}
                          className="w-full px-4 py-2 text-left hover:bg-indigo-50 text-gray-900 transition-colors">
                          <span className="font-medium">
                            {formatSymbol(sym.symbol)}
                          </span>
                          <span className="text-xs text-gray-500 ml-2">
                            {sym.base_asset}
                          </span>
                        </button>
                      ))}
                    </div>
                  )}
                  {formData.symbol && (
                    <p className="text-xs text-gray-500 mt-1">
                      Selected: {formatSymbol(formData.symbol)}
                    </p>
                  )}
                </div>
              </div>

              {/* Trading Mode & Leverage - 2 columns */}
              <div className="grid grid-cols-2 gap-4">
                {/* Trading Mode */}
                <div className="relative dropdown-container">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Trading Mode
                  </label>
                  <input
                    type="text"
                    value={
                      formData.trading_mode
                        ? tradingModes.find(
                            (mode) => mode.value === formData.trading_mode,
                          )?.label || formData.trading_mode
                        : tradingModeSearch
                    }
                    onChange={(e) => {
                      setTradingModeSearch(e.target.value);
                      setShowTradingModeDropdown(true);
                      if (!e.target.value) {
                        setFormData({
                          ...formData,
                          trading_mode: 'spot',
                          symbol: '',
                        });
                      }
                    }}
                    onFocus={() => {
                      setShowTradingModeDropdown(true);
                      if (!tradingModeSearch) {
                        setTradingModeSearch('');
                      }
                    }}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="Search trading mode..."
                    required
                    autoComplete="off"
                  />
                  {showTradingModeDropdown &&
                    filteredTradingModes.length > 0 && (
                      <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                        {filteredTradingModes.map((mode) => (
                          <button
                            key={mode.value}
                            type="button"
                            onClick={() => {
                              setFormData({
                                ...formData,
                                trading_mode: mode.value,
                                symbol: '', // Reset symbol when trading mode changes
                              });
                              setTradingModeSearch('');
                              setShowTradingModeDropdown(false);
                            }}
                            className="w-full px-4 py-3 text-left hover:bg-indigo-50 text-gray-900 transition-colors border-b last:border-b-0">
                            <div className="font-medium">{mode.label}</div>
                            <div className="text-xs text-gray-500">
                              {mode.description}
                            </div>
                          </button>
                        ))}
                      </div>
                    )}
                  {formData.trading_mode && (
                    <p className="text-xs text-gray-500 mt-1">
                      Selected:{' '}
                      {
                        tradingModes.find(
                          (mode) => mode.value === formData.trading_mode,
                        )?.label
                      }
                    </p>
                  )}
                </div>

                {/* Leverage */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Leverage
                  </label>
                  <input
                    type="number"
                    min="1"
                    max="125"
                    value={formData.leverage}
                    onChange={(e) =>
                      setFormData({...formData, leverage: e.target.value})
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="1"
                  />
                  <p className="text-xs text-gray-500 mt-1">Đòn bẩy 1x-125x</p>
                </div>
              </div>

              {/* Amount */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Amount (USDT)
                </label>
                <input
                  type="number"
                  step="0.01"
                  value={formData.amount}
                  onChange={(e) =>
                    setFormData({...formData, amount: e.target.value})
                  }
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                  placeholder="100"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Số tiền sẽ dùng cho mỗi lệnh trade
                </p>
              </div>

              {/* Stop Loss & Take Profit */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Stop Loss (%)
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.stop_loss_percent}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        stop_loss_percent: e.target.value,
                      })
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="0"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Take Profit (%)
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.take_profit_percent}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        take_profit_percent: e.target.value,
                      })
                    }
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                    placeholder="0"
                    required
                  />
                </div>
              </div>

              {/* API Credentials Section */}
              <div className="border-t border-gray-200 pt-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">
                  API Credentials (Tùy chọn)
                </h3>

                {/* API Key & Secret Key - 2 columns */}
                <div className="grid grid-cols-2 gap-4">
                  {/* API Key */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      API Key
                    </label>
                    <input
                      type="text"
                      value={formData.api_key}
                      onChange={(e) =>
                        setFormData({...formData, api_key: e.target.value})
                      }
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                      placeholder="API Key từ sàn"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Từ{' '}
                      {formData.exchange === 'binance' ? 'Binance' : 'Bittrex'}
                    </p>
                  </div>

                  {/* Secret Key */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Secret Key
                    </label>
                    <input
                      type="password"
                      value={formData.api_secret}
                      onChange={(e) =>
                        setFormData({...formData, api_secret: e.target.value})
                      }
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-gray-900"
                      placeholder="Secret Key từ sàn"
                    />
                    <p className="text-xs text-gray-500 mt-1">Sẽ được mã hóa</p>
                  </div>
                </div>
              </div>

              {/* Modal Footer */}
              <div className="flex items-center justify-end gap-3 pt-6 border-t border-gray-200">
                <button
                  type="button"
                  onClick={handleCloseModal}
                  className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
                  Hủy
                </button>
                <button
                  type="submit"
                  className="px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors">
                  {editingBot ? 'Update Config' : 'Tạo Config'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
