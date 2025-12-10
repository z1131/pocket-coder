import React, { createContext, useContext, useEffect, useMemo, useState, useCallback } from 'react';
import { refreshAccessToken } from '../api/client';

const STORAGE_KEY = 'pc_auth_v1';

// 安全的 base64 解码，支持 UTF-8
function base64UrlDecode(str: string): string {
  try {
    // 替换 URL-safe 字符
    str = str.replace(/-/g, '+').replace(/_/g, '/');
    // 添加 padding
    while (str.length % 4) {
      str += '=';
    }
    // 解码并处理 UTF-8
    const decoded = atob(str);
    // 将二进制字符串转换为 UTF-8
    return decodeURIComponent(
      decoded.split('').map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)).join('')
    );
  } catch (error) {
    console.error('[base64UrlDecode] 解码失败:', error);
    throw error;
  }
}

// 解析 JWT 获取过期时间（毫秒时间戳）
function getTokenExpiry(token: string): number | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      console.error('[getTokenExpiry] token 格式错误，应该有3部分，实际:', parts.length);
      return null;
    }
    const payloadJson = base64UrlDecode(parts[1]);
    const payload = JSON.parse(payloadJson);
    console.log('[getTokenExpiry] 解析成功:', payload);
    return payload.exp ? payload.exp * 1000 : null;
  } catch (error) {
    console.error('[getTokenExpiry] 解析失败:', error);
    return null;
  }
}

// 检查 token 是否过期（预留 30 秒缓冲）
function isTokenExpired(token: string, bufferSeconds = 30): boolean {
  const expiry = getTokenExpiry(token);
  if (!expiry) {
    console.log('[isTokenExpired] 无法解析 token 过期时间');
    return true;
  }
  const now = Date.now();
  const isExpired = now >= expiry - bufferSeconds * 1000;
  console.log('[isTokenExpired]', {
    expiresAt: new Date(expiry).toISOString(),
    now: new Date(now).toISOString(),
    remainingSeconds: Math.floor((expiry - now) / 1000),
    isExpired,
  });
  return isExpired;
}

type AuthContextType = {
  accessToken: string;
  refreshToken: string;
  isLoading: boolean;
  setTokens: (access: string, refresh: string) => void;
  clear: () => void;
  tryRefreshToken: () => Promise<string | null>;
};

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [accessToken, setAccessToken] = useState('');
  const [refreshToken, setRefreshToken] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  // 尝试刷新 token 的函数
  const tryRefreshToken = useCallback(async (): Promise<string | null> => {
    const cached = localStorage.getItem(STORAGE_KEY);
    if (!cached) return null;
    
    try {
      const parsed = JSON.parse(cached);
      const storedRefreshToken = parsed.refreshToken;
      
      if (!storedRefreshToken || isTokenExpired(storedRefreshToken)) {
        return null;
      }
      
      const res = await refreshAccessToken(storedRefreshToken);
      const newAccessToken = res.access_token;
      
      setAccessToken(newAccessToken);
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ 
        accessToken: newAccessToken, 
        refreshToken: storedRefreshToken 
      }));
      
      return newAccessToken;
    } catch {
      return null;
    }
  }, []);

  // 初始化：从 localStorage 恢复并验证 token
  useEffect(() => {
    async function initAuth() {
      console.log('[useAuth] 开始初始化认证...');
      const cached = localStorage.getItem(STORAGE_KEY);
      if (!cached) {
        console.log('[useAuth] localStorage 中没有缓存的 token');
        setIsLoading(false);
        return;
      }

      try {
        const parsed = JSON.parse(cached);
        const storedAccessToken = parsed.accessToken || '';
        const storedRefreshToken = parsed.refreshToken || '';
        
        console.log('[useAuth] 从 localStorage 读取到 token:', {
          hasAccessToken: !!storedAccessToken,
          hasRefreshToken: !!storedRefreshToken,
          accessTokenLength: storedAccessToken.length,
          refreshTokenLength: storedRefreshToken.length,
          accessTokenFull: storedAccessToken,
          refreshTokenFull: storedRefreshToken,
        });

        // 如果 access token 有效，直接使用
        if (storedAccessToken && !isTokenExpired(storedAccessToken)) {
          const expiry = getTokenExpiry(storedAccessToken);
          console.log('[useAuth] Access Token 有效，直接使用', {
            expiresAt: expiry ? new Date(expiry).toISOString() : 'unknown',
            now: new Date().toISOString(),
          });
          setAccessToken(storedAccessToken);
          setRefreshToken(storedRefreshToken);
          setIsLoading(false);
          return;
        }

        console.log('[useAuth] Access Token 已过期或无效，尝试刷新...');

        // 如果 access token 过期，尝试用 refresh token 刷新
        if (storedRefreshToken && !isTokenExpired(storedRefreshToken)) {
          console.log('[useAuth] Refresh Token 有效，开始刷新...');
          try {
            const res = await refreshAccessToken(storedRefreshToken);
            console.log('[useAuth] 刷新成功，获得新的 Access Token');
            setAccessToken(res.access_token);
            setRefreshToken(storedRefreshToken);
            localStorage.setItem(STORAGE_KEY, JSON.stringify({ 
              accessToken: res.access_token, 
              refreshToken: storedRefreshToken 
            }));
            setIsLoading(false);
            return;
          } catch (error) {
            // 刷新失败，清除存储
            console.error('[useAuth] 刷新失败:', error);
            localStorage.removeItem(STORAGE_KEY);
          }
        } else {
          // refresh token 也过期了，清除存储
          console.log('[useAuth] Refresh Token 也已过期，清除存储');
          localStorage.removeItem(STORAGE_KEY);
        }
      } catch (error) {
        console.error('[useAuth] 初始化出错:', error);
        localStorage.removeItem(STORAGE_KEY);
      }
      
      setIsLoading(false);
    }

    initAuth();
  }, []);

  // 保存 token 到 localStorage
  useEffect(() => {
    if (!accessToken && !refreshToken) {
      localStorage.removeItem(STORAGE_KEY);
      return;
    }
    if (accessToken || refreshToken) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ accessToken, refreshToken }));
    }
  }, [accessToken, refreshToken]);

  const value = useMemo<AuthContextType>(() => ({
    accessToken,
    refreshToken,
    isLoading,
    setTokens: (a, r) => {
      setAccessToken(a);
      setRefreshToken(r);
    },
    clear: () => {
      setAccessToken('');
      setRefreshToken('');
    },
    tryRefreshToken,
  }), [accessToken, refreshToken, isLoading, tryRefreshToken]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
