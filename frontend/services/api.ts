import axios, { AxiosError, AxiosResponse } from 'axios';

// 基础配置
const API_URL = '/api/v1';

const apiClient = axios.create({
  baseURL: API_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 拦截器：请求时注入 Token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 拦截器：响应错误处理
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Token 过期，清除并跳转登录（这里先只清除，跳转逻辑由 UI 层或 Store 处理）
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/'; // 强制跳转到根路径（通常是登录页）
    }
    return Promise.reject(error);
  }
);

// 通用响应结构
export interface ApiResponse<T = any> {
  code: number;
  message?: string;
  data: T;
}

// 类型定义
export interface User {
  id: number;
  username: string;
  email?: string;
  phone?: string;
  avatar?: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

export interface Device {
  id: number;
  name: string;
  type: string;
  ip?: string;
  status: 'online' | 'offline';
  os_info?: string;
  last_heartbeat?: string;
}

export interface Session {
  id: number;
  desktop_id: number;
  process_id?: string;
  is_default: boolean;
  preview?: string; // Base64
  status: 'active' | 'ended';
  started_at: string;
}

// API 函数集合
export const api = {
  auth: {
    login: async (identifier: string, password: string) => {
      const res = await apiClient.post<ApiResponse<AuthResponse>>('/auth/login', { identifier, password });
      return res.data.data;
    },
    register: async (username: string, password: string, email?: string, phone?: string) => {
      const res = await apiClient.post<ApiResponse<AuthResponse>>('/auth/register', { username, password, email, phone });
      return res.data.data;
    },
    logout: async () => {
      return apiClient.post('/auth/logout');
    },
  },
  user: {
    getProfile: async () => {
      const res = await apiClient.get<ApiResponse<User>>('/users/me');
      return res.data.data;
    },
  },
  desktop: {
    list: async () => {
      const res = await apiClient.get<ApiResponse<{ desktops: Device[] }>>('/desktops');
      return res.data.data.desktops;
    },
    get: async (id: number) => {
      const res = await apiClient.get<ApiResponse<Device>>(`/desktops/${id}`);
      return res.data.data;
    },
    delete: async (id: number) => {
      return apiClient.delete(`/desktops/${id}`);
    },
  },
  session: {
    list: async (desktopId: number, page = 1, pageSize = 20) => {
      const res = await apiClient.get<ApiResponse<{ sessions: Session[], total: number }>>('/sessions', {
        params: { desktop_id: desktopId, page, page_size: pageSize },
      });
      return res.data.data;
    },
    create: async (desktopId: number, workingDir?: string, isDefault?: boolean) => {
      const res = await apiClient.post<ApiResponse<Session>>('/sessions', {
        desktop_id: desktopId,
        working_dir: workingDir,
        is_default: isDefault, // 注意：通常手机端只能设为 false
      });
      return res.data.data;
    },
    get: async (id: number) => {
      const res = await apiClient.get<ApiResponse<{ session: Session }>>(`/sessions/${id}`);
      return res.data.data.session;
    },
    delete: async (id: number) => {
      return apiClient.delete(`/sessions/${id}`);
    },
    getActive: async (desktopId: number) => {
      const res = await apiClient.get<ApiResponse<{ session: Session | null }>>(`/desktops/${desktopId}/sessions/active`);
      return res.data.data.session;
    },
  },
  ai: {
    generateCommand: async (prompt: string, context?: { os?: string; shell?: string }) => {
      const res = await apiClient.post<ApiResponse<{ command: string; explanation?: string }>>('/ai/generate-command', {
        prompt,
        context: context || {},
      });
      return res.data.data;
    },
  },
};

