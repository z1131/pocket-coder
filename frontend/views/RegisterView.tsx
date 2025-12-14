import React, { useState } from 'react';
import { Terminal, ArrowLeft, ChevronDown } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { useAuthStore } from '../store/useStore';

const RegisterView: React.FC = () => {
  const [username, setUsername] = useState('');
  const [contact, setContact] = useState(''); // Email or Phone
  const [password, setPassword] = useState('');
  const [countryCode, setCountryCode] = useState('+86');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);

  // Validation Patterns
  const usernamePattern = /^[a-zA-Z0-9_]{3,20}$/;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !contact || !password) return;

    // Frontend Validation
    if (!usernamePattern.test(username)) {
      setError('Username must be 3-20 chars (letters, numbers, underscore only)');
      return;
    }
    if (password.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }

    setIsLoading(true);
    setError('');

    // Simple email detection
    const isEmail = contact.includes('@');
    const email = isEmail ? contact : undefined;
    
    // For phone, combine country code and contact (remove any user-entered spaces/dashes)
    const cleanContact = contact.replace(/[\s-]/g, '');
    const phone = !isEmail ? `${countryCode}${cleanContact}` : undefined;

    try {
      const data = await api.auth.register(username, password, email, phone);
      setAuth(data.user, data.access_token);
      localStorage.setItem('token', data.access_token);
      navigate('/');
    } catch (err: any) {
      console.error(err);
      // Improve error message display
      let msg = err.response?.data?.message || 'Registration failed';
      if (msg.includes("'Phone' failed")) {
        msg = 'Invalid phone number format';
      } else if (msg.includes("Username")) {
        msg = 'Username already exists or invalid';
      }
      setError(msg);
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
                placeholder="Username (3-20 chars)"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-black border border-slate-700 rounded-md px-4 py-4 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all peer"
              />
            </div>

            {/* Phone/Email Input with Country Code Selector */}
            <div className="flex bg-black border border-slate-700 rounded-md focus-within:border-indigo-500 focus-within:ring-1 focus-within:ring-indigo-500 transition-all overflow-hidden">
              {!contact.includes('@') && (
                <div className="flex items-center border-r border-slate-700 bg-slate-900/30">
                  <select 
                    value={countryCode}
                    onChange={(e) => setCountryCode(e.target.value)}
                    className="h-full bg-transparent text-slate-300 text-sm pl-3 pr-8 appearance-none outline-none cursor-pointer hover:text-white transition-colors py-4"
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
                placeholder="Phone number or Email"
                value={contact}
                onChange={(e) => setContact(e.target.value)}
                className="flex-1 bg-transparent px-4 py-4 text-white placeholder-slate-500 outline-none border-none min-w-0"
              />
            </div>
            
            <input
              type="password"
              placeholder="Password (min 6 chars)"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-black border border-slate-700 rounded-md px-4 py-4 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
            />
          </div>

          {error && <p className="text-red-500 text-sm px-1">{error}</p>}

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
