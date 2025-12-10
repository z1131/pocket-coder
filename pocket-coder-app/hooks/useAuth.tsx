import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';

const STORAGE_KEY = 'pc_auth_v1';

type AuthContextType = {
  accessToken: string;
  refreshToken: string;
  setTokens: (access: string, refresh: string) => void;
  clear: () => void;
};

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [accessToken, setAccessToken] = useState('');
  const [refreshToken, setRefreshToken] = useState('');

  useEffect(() => {
    const cached = localStorage.getItem(STORAGE_KEY);
    if (cached) {
      try {
        const parsed = JSON.parse(cached);
        setAccessToken(parsed.accessToken || '');
        setRefreshToken(parsed.refreshToken || '');
      } catch {
        /* ignore */
      }
    }
  }, []);

  useEffect(() => {
    if (!accessToken && !refreshToken) {
      localStorage.removeItem(STORAGE_KEY);
      return;
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ accessToken, refreshToken }));
  }, [accessToken, refreshToken]);

  const value = useMemo<AuthContextType>(() => ({
    accessToken,
    refreshToken,
    setTokens: (a, r) => {
      setAccessToken(a);
      setRefreshToken(r);
    },
    clear: () => {
      setAccessToken('');
      setRefreshToken('');
    },
  }), [accessToken, refreshToken]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
