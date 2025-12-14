import React, { useState } from 'react';
import { Terminal, ArrowLeft } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { useAuthStore } from '../store/useStore';

const RegisterView: React.FC = () => {
  const [username, setUsername] = useState('');
  const [contact, setContact] = useState(''); // Email or Phone
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !contact || !password) return;

    setIsLoading(true);
    setError('');

    // Simple email detection
    const isEmail = contact.includes('@');
    const email = isEmail ? contact : undefined;
    const phone = !isEmail ? contact : undefined;

    try {
      const data = await api.auth.register(username, password, email, phone);
      setAuth(data.user, data.access_token);
      localStorage.setItem('token', data.access_token);
      navigate('/');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Registration failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-black flex flex-col items-center justify-center p-4 text-slate-200 font-sans relative">
      
      {/* Top Bar for Back Navigation */}
      <div className="absolute top-4 left-4">
        <button 
          onClick={() => navigate('/login')}
          className="p-2 text-slate-400 hover:text-white hover:bg-slate-900 rounded-full transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
      </div>

      <div className="w-full max-w-sm space-y-6">
        
        {/* Logo Area */}
        <div className="flex justify-center mb-4">
          <div className="p-3 bg-indigo-600/20 rounded-2xl">
            <Terminal size={48} className="text-indigo-500" />
          </div>
        </div>

        <div className="space-y-2">
          <h1 className="text-3xl font-bold tracking-tight text-white">Create your account</h1>
        </div>

        {/* Register Form */}
        <form onSubmit={handleSubmit} className="space-y-6 mt-8">
          <div className="space-y-5">
            <div className="relative group">
               <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-black border border-slate-700 rounded-md px-4 py-4 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all peer"
              />
            </div>

            <input
              type="text"
              placeholder="Phone or Email"
              value={contact}
              onChange={(e) => setContact(e.target.value)}
              className="w-full bg-black border border-slate-700 rounded-md px-4 py-4 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
            />
            
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-black border border-slate-700 rounded-md px-4 py-4 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
            />
          </div>

          {error && <p className="text-red-500 text-sm">{error}</p>}

          <button
            type="submit"
            disabled={!username || !contact || !password || isLoading}
            className="w-full bg-white text-black font-bold rounded-full py-3 hover:bg-slate-200 transition-colors mt-8 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Creating account...' : 'Sign up'}
          </button>
        </form>
        
        <p className="text-slate-500 text-sm mt-8 text-center">
          Have an account already? <span onClick={() => navigate('/login')} className="text-indigo-400 cursor-pointer hover:underline">Log in</span>
        </p>
      </div>
    </div>
  );
};

export default RegisterView;
