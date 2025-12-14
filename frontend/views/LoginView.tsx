import React, { useState } from 'react';
import { Terminal, Smartphone, Chrome, ChevronDown } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { useAuthStore } from '../store/useStore';

const LoginView: React.FC = () => {
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [countryCode, setCountryCode] = useState('+86');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!identifier || !password) return;

    setIsLoading(true);
    setError('');

    // Auto-format phone number for login
    let finalIdentifier = identifier;
    const isEmail = identifier.includes('@');
    // Simple check: if not email and looks like a phone number (digits), add country code
    // Note: Usernames can also be just text, so we need to be careful.
    // However, typical usernames don't start with digits usually, but they can.
    // Strategy: If selected country code is visible (meaning not email), try to prefix if it looks like a phone.
    // A safer bet for Login is: if it's purely digits and length > 6, assume phone.
    const isPhone = !isEmail && /^\d+$/.test(identifier);
    
    if (isPhone) {
       finalIdentifier = `${countryCode}${identifier}`;
    }

    try {
      const data = await api.auth.login(finalIdentifier, password);
      setAuth(data.user, data.access_token);
      localStorage.setItem('token', data.access_token);
      navigate('/');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-black flex flex-col items-center justify-center p-4 text-slate-200 font-sans">
      <div className="w-full max-w-sm space-y-8">
        
        {/* Logo Area */}
        <div className="flex justify-center mb-8">
          <div className="p-3 bg-indigo-600/20 rounded-2xl">
            <Terminal size={48} className="text-indigo-500" />
          </div>
        </div>

        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold tracking-tight text-white">Sign in to Pocket Coder</h1>
          <p className="text-slate-500">Manage your infrastructure from anywhere</p>
        </div>

        {/* Social Login Buttons (Mock) */}
        <div className="space-y-3 mt-8">
          <button className="w-full flex items-center justify-center gap-3 bg-white text-black font-semibold rounded-full py-2.5 hover:bg-slate-200 transition-colors">
            <Chrome size={18} />
            <span>Sign in with Google</span>
          </button>
          
          <button className="w-full flex items-center justify-center gap-3 bg-white text-black font-semibold rounded-full py-2.5 hover:bg-slate-200 transition-colors">
            <Smartphone size={18} />
            <span>Sign in with Phone</span>
          </button>
        </div>

        <div className="relative flex items-center justify-center py-2">
          <div className="border-t border-slate-800 w-full absolute"></div>
          <span className="bg-black px-3 text-slate-500 text-sm relative z-10">or</span>
        </div>

        {/* Login Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-4">
             {/* Identifier Input with Country Code */}
             <div className="flex bg-black border border-slate-700 rounded-md focus-within:border-indigo-500 focus-within:ring-1 focus-within:ring-indigo-500 transition-all overflow-hidden">
              {!identifier.includes('@') && /^\d*$/.test(identifier) && identifier.length > 0 && (
                <div className="flex items-center border-r border-slate-700 bg-slate-900/30">
                  <select 
                    value={countryCode}
                    onChange={(e) => setCountryCode(e.target.value)}
                    className="h-full bg-transparent text-slate-300 text-sm pl-3 pr-8 appearance-none outline-none cursor-pointer hover:text-white transition-colors py-3"
                    style={{ backgroundImage: 'none' }}
                  >
                    <option value="+86">🇨🇳 +86</option>
                    <option value="+1">🇺🇸 +1</option>
                    <option value="+44">🇬🇧 +44</option>
                    <option value="+81">🇯🇵 +81</option>
                  </select>
                  <ChevronDown size={14} className="text-slate-500 absolute left-[4.5rem] pointer-events-none" />
                </div>
              )}
              <input
                type="text"
                placeholder="Username, Email or Phone"
                value={identifier}
                onChange={(e) => setIdentifier(e.target.value)}
                className="flex-1 bg-transparent px-4 py-3 text-white placeholder-slate-500 outline-none border-none min-w-0"
              />
            </div>

            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-black border border-slate-700 rounded-md px-4 py-3 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
            />
          </div>

          {error && <p className="text-red-500 text-sm">{error}</p>}

          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-indigo-600 text-white font-bold rounded-full py-3 hover:bg-indigo-500 transition-colors mt-6 shadow-[0_0_15px_rgba(79,70,229,0.3)] disabled:opacity-50"
          >
            {isLoading ? 'Signing in...' : 'Sign In'}
          </button>
          
          <button type="button" className="w-full bg-black text-white font-bold border border-slate-700 rounded-full py-3 hover:bg-slate-900 transition-colors">
            Forgot password?
          </button>
        </form>

        <p className="text-slate-500 text-sm mt-8 text-center">
          Don't have an account? <span onClick={() => navigate('/register')} className="text-indigo-400 cursor-pointer hover:underline">Sign up</span>
        </p>
      </div>
    </div>
  );
};

export default LoginView;
