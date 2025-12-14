import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { useAuthStore } from './store/useStore';
import LoginView from './views/LoginView';
import RegisterView from './views/RegisterView';
import DeviceListView from './views/DeviceListView';
import SessionListView from './views/SessionListView'; // 需要创建
import TerminalView from './views/TerminalView'; // 需要改造

// 路由守卫
const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { token } = useAuthStore();
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
};

// 公开路由守卫（已登录则跳过）
const PublicGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { token } = useAuthStore();
  if (token) {
    return <Navigate to="/" replace />;
  }
  return <>{children}</>;
};

const AppRoutes: React.FC = () => {
  return (
    <Routes>
      <Route
        path="/login"
        element={
          <PublicGuard>
            <LoginView />
          </PublicGuard>
        }
      />
      <Route
        path="/register"
        element={
          <PublicGuard>
            <RegisterView />
          </PublicGuard>
        }
      />
      <Route
        path="/"
        element={
          <AuthGuard>
            <DeviceListView />
          </AuthGuard>
        }
      />
      <Route
        path="/desktops/:desktopId"
        element={
          <AuthGuard>
            <SessionListView />
          </AuthGuard>
        }
      />
      <Route
        path="/sessions/:sessionId"
        element={
          <AuthGuard>
            <TerminalView />
          </AuthGuard>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
};

export default App;
