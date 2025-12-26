import api from './api';

export interface BotConfig {
  id: number;
  user_id: number;
  name: string;
  symbol: string;
  exchange: string;
  amount?: number;
  trading_mode?: string;
  leverage?: number;
  margin_mode?: string;
  stop_loss_percent: number;
  take_profit_percent: number;
  trailing_stop_percent?: number;
  enable_trailing_stop?: boolean;
  activation_price?: number;
  callback_rate?: number;
  is_default?: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface BotConfigCreate {
  name: string;
  symbol: string;
  exchange: string;
  amount?: number;
  trading_mode?: string;
  leverage?: number;
  margin_mode?: string;
  api_key?: string;
  api_secret?: string;
  stop_loss_percent: number;
  take_profit_percent: number;
  trailing_stop_percent?: number;
  enable_trailing_stop?: boolean;
  activation_price?: number;
  callback_rate?: number;
}

export interface BotConfigUpdate {
  symbol?: string;
  exchange?: string;
  amount?: number;
  trading_mode?: string;
  leverage?: number;
  margin_mode?: string;
  api_key?: string;
  api_secret?: string;
  stop_loss_percent?: number;
  take_profit_percent?: number;
  trailing_stop_percent?: number;
  enable_trailing_stop?: boolean;
  activation_price?: number;
  callback_rate?: number;
  is_active?: boolean;
}

export const createBotConfig = async (
  configData: BotConfigCreate,
): Promise<BotConfig> => {
  const response = await api.post('/config', configData);
  return response.data.config;
};

export const listBotConfigs = async (
  skip = 0,
  limit = 100,
): Promise<{configs: BotConfig[]; total: number}> => {
  const response = await api.get('/config/list', {
    params: {skip, limit},
  });
  return response.data;
};

export const getBotConfig = async (id: number): Promise<BotConfig> => {
  const response = await api.get(`/config/${id}`);
  return response.data;
};

export const updateBotConfig = async (
  id: number,
  configData: BotConfigUpdate,
): Promise<BotConfig> => {
  const response = await api.put(`/config/${id}`, configData);
  return response.data.config;
};

export const deleteBotConfig = async (id: number): Promise<void> => {
  await api.delete(`/config/${id}`);
};

export const setDefaultBotConfig = async (id: number): Promise<BotConfig> => {
  const response = await api.put(`/config/${id}/set-default`);
  return response.data.config;
};
