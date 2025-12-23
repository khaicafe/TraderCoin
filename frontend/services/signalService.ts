import api from './api';

export interface TradingSignal {
  id: number;
  symbol: string;
  action: string; // buy, sell, close
  price: number;
  stop_loss: number;
  take_profit: number;
  message: string;
  strategy: string;
  status: string; // pending, executed, failed, ignored
  order_id?: number;
  executed_by_user_id?: number; // User ID who executed this signal
  error_message?: string;
  webhook_prefix?: string;
  received_at: string;
  executed_at?: string;
  created_at: string;
  updated_at: string;
  order?: any; // Order details if executed
}

export interface SignalsResponse {
  signals: TradingSignal[];
  count: number;
}

export interface ExecuteSignalRequest {
  bot_config_id: number;
  test_mode?: boolean; // 🧪 Enable test mode to bypass PlaceOrder
}

export interface ExecuteSignalResponse {
  status: string;
  signal: TradingSignal;
  order: any;
  message: string;
}

export interface CreateSignalPayload {
  symbol: string;
  action: string; // buy, sell, close
  price?: number;
  stopLoss?: number; // camelCase theo backend webhook
  takeProfit?: number; // camelCase theo backend webhook
  message?: string;
  timestamp?: number; // unix seconds
  strategy?: string;
}

// List all trading signals
export const listSignals = async (params?: {
  status?: string;
  symbol?: string;
  prefix?: string;
  limit?: number;
  since_hours?: number;
  since_ts?: number; // unix seconds or milliseconds
}): Promise<SignalsResponse> => {
  const response = await api.get('/signals', {params});
  return response.data;
};

// Get single signal
export const getSignal = async (id: number): Promise<TradingSignal> => {
  const response = await api.get(`/signals/${id}`);
  return response.data;
};

// Execute signal with bot config
export const executeSignal = async (
  signalId: number,
  data: ExecuteSignalRequest,
): Promise<ExecuteSignalResponse> => {
  // Add test_mode as query parameter if enabled
  const params = data.test_mode ? {test_mode: 'true'} : {};
  const response = await api.post(`/signals/${signalId}/execute`, data, {
    params,
  });
  return response.data;
};

// Update signal status (mark as ignored, etc.)
export const updateSignalStatus = async (
  signalId: number,
  status: string,
): Promise<TradingSignal> => {
  const response = await api.put(`/signals/${signalId}/status`, {status});
  return response.data;
};

// Delete signal
export const deleteSignal = async (signalId: number): Promise<void> => {
  await api.delete(`/signals/${signalId}`);
};

// Webhook prefix API
export const getWebhookPrefix = async (): Promise<{
  prefix: string;
  url: string;
}> => {
  const response = await api.get('/signals/webhook/prefix');
  return response.data;
};

export const createWebhookPrefix = async (
  prefix?: string,
): Promise<{prefix: string; url: string}> => {
  const response = await api.post(
    '/signals/webhook/prefix',
    prefix ? {prefix} : {},
  );
  return response.data;
};

// Create a signal via TradingView webhook endpoint (optional prefix)
export const createSignalViaWebhook = async (
  payload: CreateSignalPayload,
  prefix?: string,
): Promise<{
  status: string;
  signal_id: number;
  message: string;
  webhook_prefix?: string;
}> => {
  const url = `/signals/webhook/tradingview${prefix ? `/${prefix}` : ''}`;
  const response = await api.post(url, payload);
  return response.data;
};
