import axios from 'axios';

// Get token from localStorage
const getToken = () => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('token');
  }
  return null;
};

// API base URL - sử dụng /api/account thay vì /api/v1
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '') ||
  'http://localhost:8080';

const userApi = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor for auth token
userApi.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

export interface User {
  id: number;
  username: string;
  email: string;
  full_name: string;
  phone: string;
  chat_id: string;
  is_active: boolean;
  created_at: string;
}

export interface UpdateProfileRequest {
  username?: string;
  email?: string;
  full_name?: string;
  phone?: string;
  chat_id?: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

const userService = {
  // Lấy thông tin profile user
  getProfile: async (): Promise<User> => {
    const response = await userApi.get<User>('/api/account/profile');
    return response.data;
  },

  // Cập nhật thông tin profile
  updateProfile: async (data: UpdateProfileRequest): Promise<User> => {
    const response = await userApi.put<User>('/api/account/profile', data);
    return response.data;
  },

  // Đổi mật khẩu
  changePassword: async (
    data: ChangePasswordRequest,
  ): Promise<{message: string}> => {
    const response = await userApi.put<{message: string}>(
      '/api/account/change-password',
      data,
    );
    return response.data;
  },
};

export default userService;
