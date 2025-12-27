import api from './api';

export interface SystemLog {
  id: number;
  user_id: number;
  level: string;
  action: string;
  symbol?: string;
  exchange?: string;
  order_id?: number;
  price?: number;
  amount?: number;
  message: string;
  details?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export interface SystemLogStats {
  level: string;
  count: number;
}

export interface GetSystemLogsResponse {
  logs: SystemLog[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export interface GetSystemLogStatsResponse {
  stats: SystemLogStats[];
  total: number;
  period_hours: number;
}

export interface ClearSystemLogsResponse {
  success: boolean;
  message: string;
  deleted_count: number;
}

// Get system logs with filters
export const getSystemLogs = async (params?: {
  level?: string;
  symbol?: string;
  action?: string;
  hours?: number;
  page?: number;
  limit?: number;
}): Promise<GetSystemLogsResponse> => {
  const response = await api.get('/logs', {params});
  return response.data;
};

// Get system log statistics
export const getSystemLogStats = async (
  hours?: number,
): Promise<GetSystemLogStatsResponse> => {
  const response = await api.get('/logs/stats', {params: {hours}});
  return response.data;
};

// Clear old system logs
export const clearSystemLogs = async (
  days?: number,
): Promise<ClearSystemLogsResponse> => {
  const response = await api.delete('/logs/clear', {params: {days}});
  return response.data;
};
