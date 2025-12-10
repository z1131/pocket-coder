import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './hooks/useAuth';
import LoginPage from './pages/LoginPage';
import DesktopsPage from './pages/DesktopsPage';
import SessionPage from './pages/SessionPage';
import RegisterPage from './pages/RegisterPage';

// 加载中组件
const LoadingScreen: React.FC = () => (
  <div style={{
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100vh',
    backgroundColor: '#1a1a2e',
    color: '#fff',
    fontSize: '18px',
  }}>
    <div style={{ textAlign: 'center' }}>
      <div style={{ marginBottom: '16px' }}>🔐</div>
      <div>正在验证登录状态...</div>
    </div>
  </div>
);

const Guard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { accessToken, isLoading } = useAuth();
  
  // 正在加载时显示 loading 状态
  if (isLoading) {
    return <LoadingScreen />;
  }
  
  if (!accessToken) return <Navigate to="/login" replace />;
  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="/desktops"
          element={
            <Guard>
              <DesktopsPage />
            </Guard>
          }
        />
        <Route
          path="/desktops/:desktopId"
          element={
            <Guard>
              <SessionPage />
            </Guard>
          }
        />
        <Route path="*" element={<Navigate to="/desktops" replace />} />
      </Routes>
    </AuthProvider>
  );
};

export default App;
