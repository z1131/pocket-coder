import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { login } from '../api/client';
import { useAuth } from '../hooks/useAuth';

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const { setTokens } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!username || !password) {
      setError('请输入用户名和密码');
      return;
    }
    try {
      setLoading(true);
      const resp = await login(username, password);
      setTokens(resp.access_token, resp.refresh_token);
      navigate('/desktops', { replace: true });
    } catch (err: any) {
      setError(err?.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-black text-white flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-[400px] flex flex-col">
        {/* Header Section */}
        <div className="mb-12">
          <h1 className="text-5xl font-bold mb-4 tracking-tight">正发生</h1>
          <h2 className="text-3xl font-bold tracking-tight">现在就加入。</h2>
        </div>

        {/* Social Login Buttons */}
        <div className="flex flex-col gap-3 mb-4">
          <button
            disabled
            className="group relative flex items-center justify-center gap-3 w-full bg-white text-black h-12 rounded-full font-medium transition-opacity disabled:opacity-60 disabled:cursor-not-allowed hover:bg-neutral-100"
          >
            <svg className="w-5 h-5" viewBox="0 0 24 24">
              <path
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                fill="#4285F4"
              />
              <path
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                fill="#34A853"
              />
              <path
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                fill="#FBBC05"
              />
              <path
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                fill="#EA4335"
              />
            </svg>
            使用 Google 账号登录
          </button>

          <button
            disabled
            className="group relative flex items-center justify-center gap-3 w-full bg-white text-black h-12 rounded-full font-medium transition-opacity disabled:opacity-60 disabled:cursor-not-allowed hover:bg-neutral-100"
          >
            <svg className="w-6 h-6 text-[#07C160]" viewBox="0 0 24 24" fill="currentColor">
              <path d="M8.696 15.866c0-3.32-3.134-6.012-7-6.012s-7 2.692-7 6.012c0 3.32 3.134 6.012 7 6.012 3.866 0 7-2.692 7-6.012z" transform="matrix(1.137 0 0 1.137 9.075 4.383)" />
              <path d="M19.333 1.833c-4.8 0-8.666 3.233-8.666 7.233 0 4 3.866 7.233 8.666 7.233 4.8 0 8.667-3.233 8.667-7.233 0-4-3.867-7.233-8.667-7.233z" fill="#FFF" />
              <path clipRule="evenodd" d="M18.668 7.792c0 .416-.324.75-.724.75-.4 0-.724-.334-.724-.75s.324-.75.724-.75c.4 0 .724.334.724.75zm5.787 7.232c-.105.006-.21.01-.314.01-4.8 0-8.667-3.233-8.667-7.233 0-2.83 1.942-5.285 4.805-6.425 2.82.936 4.904 3.197 5.584 5.96a7.66 7.66 0 0 0 1.708 3.518c-1.378 2.01-1.786 3.033-3.116 4.17z" fillRule="evenodd" />
              <path d="M8.696 15.866c0-3.32-3.134-6.012-7-6.012s-7 2.692-7 6.012c0 3.32 3.134 6.012 7 6.012 3.866 0 7-2.692 7-6.012z" transform="matrix(1 0 0 1 1.783 1.483)" />
              <path d="M5.963 6.324c0 .324-.253.585-.563.585-.31 0-.563-.26-.563-.585 0-.324.253-.585.563-.585.31 0 .563.26.563.585zM9.544 6.324c0 .324-.253.585-.563.585-.31 0-.563-.26-.563-.585 0-.324.253-.585.563-.585.31 0 .563.26.563.585z" fill="#000" />
            </svg>
            使用微信登录
          </button>

          <button
            disabled
            className="group relative flex items-center justify-center gap-3 w-full bg-transparent border border-neutral-600 text-white h-12 rounded-full font-medium transition-colors disabled:opacity-60 disabled:cursor-not-allowed hover:bg-neutral-900"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect x="5" y="2" width="14" height="20" rx="2" ry="2"></rect>
              <line x1="12" y1="18" x2="12.01" y2="18"></line>
            </svg>
            使用手机号登录
          </button>
        </div>

        {/* Divider */}
        <div className="relative flex items-center justify-center my-4">
          <div className="bg-neutral-800 h-px w-full absolute z-0"></div>
          <span className="bg-black px-2 text-sm text-white z-10 relative">或</span>
        </div>

        <button
          className="w-full bg-[#1d9bf0] hover:bg-[#1a8cd8] text-white h-12 rounded-full font-bold transition-colors mb-2"
          onClick={() => {
            // Toggle visual display of form or just scroll to it if we were doing a single page thing
            // For now we just keep the form visible below but styled differently
            const form = document.querySelector('form');
            form?.scrollIntoView({ behavior: 'smooth' });
            form?.querySelector('input')?.focus();
          }}
        >
          创建账号
        </button>

        <p className="text-[11px] text-neutral-500 mb-8 leading-tight">
          注册即表示同意<a href="#" className="text-[#1d9bf0] hover:underline">服务条款</a>及<a href="#" className="text-[#1d9bf0] hover:underline">隐私政策</a>，其中包括 <a href="#" className="text-[#1d9bf0] hover:underline">Cookie 使用条款</a>。
        </p>

        {/* Login Form Section */}
        <div className="mt-4">
          <h3 className="text-xl font-bold mb-4">已有账号？</h3>

          <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
            <div className="group relative">
              <input
                className="w-full bg-black border border-neutral-600 rounded-md px-3 py-3 text-white focus:outline-none focus:border-[#1d9bf0] focus:ring-1 focus:ring-[#1d9bf0] placeholder-neutral-500 transition-all"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="用户名"
                autoComplete="username"
              />
            </div>

            <div className="group relative">
              <input
                type="password"
                className="w-full bg-black border border-neutral-600 rounded-md px-3 py-3 text-white focus:outline-none focus:border-[#1d9bf0] focus:ring-1 focus:ring-[#1d9bf0] placeholder-neutral-500 transition-all"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="密码"
                autoComplete="current-password"
              />
            </div>

            {error && <div className="text-sm text-rose-500 font-medium px-1">{error}</div>}

            <button
              type="submit"
              disabled={loading}
              className="w-full h-12 rounded-full border border-neutral-600 text-[#1d9bf0] font-bold hover:bg-[#1d9bf0]/10 transition-colors disabled:opacity-50 mt-2"
            >
              {loading ? '登录中...' : '登录'}
            </button>

          </form>
          <div className="mt-4 text-sm text-neutral-500">
            <Link className="hover:underline" to="/register">还没有账号？去注册</Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
