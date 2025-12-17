import api from './api';

export interface Order {
  id: number;
  user_id: number;
  exchange: string;
  symbol: string;
  order_id?: string;
  side: string; // 'buy' or 'sell'
  type: string; // 'limit', 'market', etc.
  quantity: number;
  price: number;
  filled_price?: number;
  filled_quantity?: number;
  current_price?: number; // Current market price from exchange
  status: string; // 'pending', 'filled', 'closed', 'cancelled'
  trading_mode?: string; // 'spot', 'futures', 'margin'
  leverage?: number;
  stop_loss_price?: number;
  take_profit_price?: number;
  pnl?: number;
  pnl_percent?: number;
  bot_config_name?: string;
  created_at: string;
  updated_at: string;
}

export interface OrderHistoryParams {
  bot_config_id?: number;
  symbol?: string;
  status?: string;
  side?: string;
  start_date?: string;
  end_date?: string;
  limit?: number;
  offset?: number;
}

// Get order history with filters
export const getOrderHistory = async (
  params?: OrderHistoryParams,
): Promise<Order[]> => {
  const response = await api.get('/orders/history', {params});
  return response.data;
};

// List all orders
export const listOrders = async (skip = 0, limit = 100): Promise<Order[]> => {
  const response = await api.get('/orders', {
    params: {skip, limit},
  });
  return response.data;
};

// Get single order
export const getOrder = async (id: number): Promise<Order> => {
  const response = await api.get(`/orders/${id}`);
  return response.data;
};

// Get completed orders (filled or closed)
export const getCompletedOrders = async (
  params?: OrderHistoryParams,
): Promise<Order[]> => {
  const response = await api.get('/orders/completed', {params});
  return response.data;
};
