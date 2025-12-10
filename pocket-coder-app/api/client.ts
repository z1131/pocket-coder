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
