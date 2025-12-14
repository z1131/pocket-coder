import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User, Device, Session } from '../services/api';

interface AuthState {
  user: User | null;
  token: string | null;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
}

interface AppState {
  currentDesktop: Device | null;
  setCurrentDesktop: (desktop: Device | null) => void;
  currentSession: Session | null;
  setCurrentSession: (session: Session | null) => void;
}

// 认证 Store (持久化)
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      setAuth: (user, token) => set({ user, token }),
      logout: () => set({ user: null, token: null }),
    }),
    {
      name: 'auth-storage', // localStorage key
    }
  )
);

// 应用状态 Store (不持久化，或者部分持久化)
export const useAppStore = create<AppState>((set) => ({
  currentDesktop: null,
  setCurrentDesktop: (desktop) => set({ currentDesktop: desktop }),
  currentSession: null,
  setCurrentSession: (session) => set({ currentSession: session }),
}));
