import api from './api';

export interface PlaceOrderRequest {
  bot_config_id: number;
  symbol?: string;
  side: 'buy' | 'sell';
  order_type: 'market' | 'limit';
  amount?: number;
  price?: number;
}

export interface PlaceOrderResponse {
  status: string;
  order_id: number;
  exchange_order_id: string;
  symbol: string;
  side: string;
  order_type: string;
  amount: number;
  price: number;
  filled_price: number;
  stop_loss: number;
  take_profit: number;
  order_status: string;
}

export interface SymbolsResponse {
  symbols: string[];
  count: number;
  exchange: string;
  trading_mode: string;
}

export interface OrderStatusResponse {
  order_id: number;
  exchange_order_id: string;
  status: string;
  filled: number;
  remaining: number;
}

export interface AccountInfo {
  exchange: string;
  total_balance: number;
  available_balance: number;
  in_order: number;
  balances: Balance[];
}

export interface Balance {
  asset: string;
  free: number;
  locked: number;
  total: number;
}

// Đặt lệnh trực tiếp
export const placeOrder = async (
  orderData: PlaceOrderRequest,
): Promise<PlaceOrderResponse> => {
  const response = await api.post('/trading/place-order', orderData);
  return response.data;
};

// Đóng lệnh
export const closeOrder = async (orderId: number): Promise<any> => {
  const response = await api.post(`/trading/close-order/${orderId}`);
  return response.data;
};

// Lấy danh sách symbols từ exchange
export const getSymbols = async (
  configId: number,
): Promise<SymbolsResponse> => {
  const response = await api.get(`/trading/symbols/${configId}`);
  return response.data;
};

// Kiểm tra trạng thái lệnh
export const checkOrderStatus = async (
  orderId: number,
): Promise<OrderStatusResponse> => {
  const response = await api.get(`/trading/check-order/${orderId}`);
  return response.data;
};

// Refresh PnL từ sàn
export const refreshPnL = async (orderId: number): Promise<any> => {
  const response = await api.post(`/trading/refresh-pnl/${orderId}`);
  return response.data;
};

// Lấy thông tin tài khoản từ sàn
export const getAccountInfo = async (
  configId: number,
): Promise<AccountInfo> => {
  const response = await api.get(`/trading/account-info/${configId}`);
  return response.data;
};
