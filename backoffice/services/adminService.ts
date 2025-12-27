import api from './api';

export interface User {
  id: number;
  email: string;
  full_name: string;
  phone?: string;
  status: string;
  subscription_end?: string;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: number;
  user_id: number;
  user_email?: string;
  user_full_name?: string;
  exchange: string;
  symbol: string;
  order_id: string;
  side: string;
  type: string;
  quantity: number;
  price: number;
  filled_price: number;
  filled_quantity: number;
  current_price: number;
  status: string;
  trading_mode: string;
  leverage: number;
  stop_loss_price: number;
  take_profit_price: number;
  pnl: number;
  pnl_percent: number;
  position_side: string;
  liquidation_price: number;
  margin_type: string;
  isolated_margin: number;
  created_at: string;
  updated_at: string;
}

export interface Signal {
  id: number;
  symbol: string;
  action: string;
  price: number;
  stop_loss: number;
  take_profit: number;
  message: string;
  strategy: string;
  status: string;
  order_id?: number;
  executed_by_user_id?: number;
  error_message: string;
  webhook_prefix: string;
  received_at: string;
  executed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: number;
  user_id: number;
  user_email: string;
  amount: number;
  type: string;
  status: string;
  description: string;
  created_at: string;
}

export interface SystemLog {
  id: number;
  user_id: number;
  user_email: string;
  user_full_name: string;
  level: string;
  action: string;
  symbol: string;
  exchange: string;
  order_id?: number;
  price: number;
  amount: number;
  message: string;
  details: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
}

export interface Exchange {
  id: number;
  name: string;
  display_name: string;
  api_url: string;
  is_active: boolean;
  supports_spot: boolean;
  supports_futures: boolean;
  supports_margin: boolean;
  created_at: string;
  updated_at: string;
}

export interface Statistics {
  totalUsers: number;
  activeUsers: number;
  suspendedUsers: number;
  totalRevenue: number;
  activeSubscriptions: number;
  totalOrders: number;
}

// Auth
export const adminLogin = async (email: string, password: string) => {
  const response = await api.post('/api/v1/admin/login', {email, password});
  return response.data;
};

// Users
export const getUsers = async () => {
  const response = await api.get<{success: boolean; users: User[]}>(
    '/api/v1/admin/users',
  );
  return response.data;
};

export const updateUserStatus = async (
  userId: number,
  status: string,
  reason?: string,
) => {
  const response = await api.put(`/api/v1/admin/users/${userId}/status`, {
    status,
    reason,
  });
  return response.data;
};

// Orders
export const getOrders = async () => {
  const response = await api.get<{success: boolean; orders: Order[]}>(
    '/api/v1/admin/orders',
  );
  return response.data;
};

// Exchanges
export const getExchanges = async () => {
  const response = await api.get<{success: boolean; exchanges: Exchange[]}>(
    '/api/v1/admin/exchanges',
  );
  return response.data;
};

export const createExchange = async (data: Partial<Exchange>) => {
  const response = await api.post('/api/v1/admin/exchanges', data);
  return response.data;
};

// Signals
export const getSignals = async (params?: {
  status?: string;
  symbol?: string;
  limit?: number;
  since_hours?: number;
}) => {
  const response = await api.get<{signals: Signal[]; count: number}>(
    '/api/v1/admin/signals',
    {params},
  );
  return response.data;
};

// Transactions
export const getTransactions = async (params?: {
  user_id?: number;
  type?: string;
  status?: string;
}) => {
  const response = await api.get<Transaction[]>('/api/v1/admin/transactions', {
    params,
  });
  return response.data;
};

// System Logs
export const getSystemLogs = async (params?: {
  user_id?: number;
  level?: string;
  symbol?: string;
  action?: string;
  limit?: number;
}) => {
  const response = await api.get<SystemLog[]>('/api/v1/admin/logs', {params});
  return response.data;
};

export const updateExchange = async (id: number, data: Partial<Exchange>) => {
  const response = await api.put(`/api/v1/admin/exchanges/${id}`, data);
  return response.data;
};

export const deleteExchange = async (id: number) => {
  const response = await api.delete(`/api/v1/admin/exchanges/${id}`);
  return response.data;
};

// Statistics
export const getStatistics = async () => {
  const response = await api.get<{success: boolean; stats: Statistics}>(
    '/api/v1/admin/statistics',
  );
  return response.data;
};

// Admin Profile & Settings
export const getAdminProfile = async () => {
  const response = await api.get('/api/v1/admin/profile');
  return response.data;
};

export const updateAdminProfile = async (data: {
  email: string;
  full_name: string;
}) => {
  const response = await api.put('/api/v1/admin/profile', data);
  return response.data;
};

export const changeAdminPassword = async (data: {
  current_password: string;
  new_password: string;
}) => {
  const response = await api.put('/api/v1/admin/password', data);
  return response.data;
};
