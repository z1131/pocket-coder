import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { registerUser } from '../api/client';

const RegisterPage: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!username || !password) {
      setError('用户名和密码不能为空');
      return;
    }
    if (password.length < 6) {
      setError('密码至少 6 位');
      return;
    }
    if (password !== confirm) {
      setError('两次密码不一致');
      return;
    }

    try {
      setLoading(true);
      await registerUser(username, password, email || undefined);
      setSuccess('注册成功，请登录');
      setTimeout(() => navigate('/login', { replace: true }), 800);
    } catch (err: any) {
      setError(err?.message || '注册失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-black text-white flex items-center justify-center px-4">
      <div className="w-full max-w-md bg-neutral-900/70 border border-neutral-800 rounded-2xl p-8 shadow-xl">
        <h1 className="text-2xl font-semibold mb-2">创建新账号</h1>
        <p className="text-sm text-neutral-400 mb-6">注册后可在手机端管理你的桌面设备。</p>

        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <label className="flex flex-col gap-2 text-sm">
            <span className="text-neutral-300">用户名</span>
            <input
              className="bg-neutral-800 border border-neutral-700 rounded-lg px-3 py-2 focus:outline-none focus:border-neutral-500"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="your-name"
              autoComplete="username"
            />
          </label>

          <label className="flex flex-col gap-2 text-sm">
            <span className="text-neutral-300">邮箱（可选）</span>
            <input
              className="bg-neutral-800 border border-neutral-700 rounded-lg px-3 py-2 focus:outline-none focus:border-neutral-500"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              autoComplete="email"
            />
          </label>

          <label className="flex flex-col gap-2 text-sm">
            <span className="text-neutral-300">密码</span>
            <input
              type="password"
              className="bg-neutral-800 border border-neutral-700 rounded-lg px-3 py-2 focus:outline-none focus:border-neutral-500"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="至少 6 位"
              autoComplete="new-password"
            />
          </label>

          <label className="flex flex-col gap-2 text-sm">
            <span className="text-neutral-300">确认密码</span>
            <input
              type="password"
              className="bg-neutral-800 border border-neutral-700 rounded-lg px-3 py-2 focus:outline-none focus:border-neutral-500"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              placeholder="再次输入密码"
              autoComplete="new-password"
            />
          </label>

          {error && <div className="text-sm text-rose-400">{error}</div>}
          {success && <div className="text-sm text-emerald-400">{success}</div>}

          <button
            type="submit"
            disabled={loading}
            className="h-11 rounded-lg bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 font-semibold"
          >
            {loading ? '注册中...' : '注册并开始使用'}
          </button>

          <div className="text-sm text-neutral-400 text-center">
            已有账号？<Link className="text-emerald-400 hover:text-emerald-300" to="/login">去登录</Link>
          </div>
        </form>
      </div>
    </div>
  );
};

export default RegisterPage;
