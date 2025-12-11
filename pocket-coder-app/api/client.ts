export const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080';

export type LoginResponse = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
};

export type RegisterResponse = {
  user_id: number;
  username: string;
};

export type DesktopItem = {
  id: number;
  name: string;
  agent_type: string;
  status: string;
  os_info?: string | null;
  working_dir?: string | null;
  last_heartbeat?: string | null;
};

export type DesktopsResponse = {
  desktops: DesktopItem[];
};

async function request<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> | undefined),
  };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const resp = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const data = await resp.json();

  if (data.code !== 0) {
    throw new Error(data.message || '请求失败');
  }
  return data.data as T;
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  return request<LoginResponse>("/api/v1/auth/login", {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function registerUser(username: string, password: string, email?: string): Promise<RegisterResponse> {
  return request<RegisterResponse>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, password, email }),
  });
}

export async function fetchDesktops(token: string): Promise<DesktopsResponse> {
  return request<DesktopsResponse>("/api/v1/desktops", { method: 'GET' }, token);
}

export async function renameDesktop(id: number, name: string, token: string): Promise<void> {
  await request(`/api/v1/desktops/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  }, token);
}

// 会话类型
export type SessionItem = {
  id: number;
  desktop_id: number;
  agent_type: string;
  working_dir?: string | null;
  title?: string | null;
  summary?: string | null;
  status: string;
  started_at: string;
  ended_at?: string | null;
};

export type SessionsResponse = {
  sessions: SessionItem[];
  total: number;
};

// 获取设备的会话列表
export async function fetchSessions(desktopId: number, token: string, page: number = 1, pageSize: number = 20): Promise<SessionsResponse> {
  return request<SessionsResponse>(`/api/v1/sessions?desktop_id=${desktopId}&page=${page}&page_size=${pageSize}`, { method: 'GET' }, token);
}

// 创建新会话
export async function createSession(desktopId: number, token: string, workingDir?: string): Promise<SessionItem> {
  // 注意：后端现在改为从 Body 读取 desktop_id，且 path 为 /api/v1/sessions
  // 返回的是 SessionResponse (即 SessionItem)
  return request<SessionItem>('/api/v1/sessions', {
    method: 'POST',
    body: JSON.stringify({ desktop_id: desktopId, working_dir: workingDir }),
  }, token);
}

// 刷新 Token 响应类型
export type RefreshTokenResponse = {
  access_token: string;
  expires_in: number;
};

// 刷新 Access Token
export async function refreshAccessToken(refreshToken: string): Promise<RefreshTokenResponse> {
  console.log('[API] 开始刷新 Access Token...');
  try {
    const result = await request<RefreshTokenResponse>('/api/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    console.log('[API] Access Token 刷新成功');
    return result;
  } catch (error) {
    console.error('[API] Access Token 刷新失败:', error);
    throw error;
  }
}
