import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './hooks/useAuth';
import LoginPage from './pages/LoginPage';
import DesktopsPage from './pages/DesktopsPage';
import SessionPage from './pages/SessionPage';
import RegisterPage from './pages/RegisterPage';

const Guard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { accessToken } = useAuth();
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
