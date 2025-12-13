import React, { useEffect, useState } from 'react';
import { Monitor, Server, MoreVertical, Wifi, WifiOff, Plus } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { api, Device } from '../services/api';
import { ws } from '../services/ws';
import { useAuthStore, useAppStore } from '../store/useStore';

const DeviceListView: React.FC = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const { user, logout } = useAuthStore();
  const setCurrentDesktop = useAppStore((state) => state.setCurrentDesktop);
  const navigate = useNavigate();

  useEffect(() => {
    // Connect WebSocket globally when in device list to receive updates
    const token = localStorage.getItem('token');
    if (token) {
      ws.connect();
    }

    const loadDevices = async () => {
      try {
        const data = await api.desktop.list();
        setDevices(data);
      } catch (err) {
        console.error('Failed to load devices', err);
      } finally {
        setLoading(false);
      }
    };
    loadDevices();

    // Listen for status updates
    const offOnline = ws.on('desktop:online', (payload: any) => {
      setDevices(prev => prev.map(d => 
        d.id === payload.desktop_id ? { ...d, status: 'online' } : d
      ));
    });

    const offOffline = ws.on('desktop:offline', (payload: any) => {
      setDevices(prev => prev.map(d => 
        d.id === payload.desktop_id ? { ...d, status: 'offline' } : d
      ));
    });

    return () => {
      offOnline();
      offOffline();
      // We don't disconnect here because we want to keep the connection alive for other views, 
      // or we can disconnect if we want strict resource management, but for SPA it's usually fine.
      // Ideally App.tsx handles the global connection.
    };
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const handleSelectDevice = (device: Device) => {
    if (device.status === 'online') {
      setCurrentDesktop(device);
      navigate(`/desktops/${device.id}`);
    }
  };

  const handleAddDevice = () => {
    alert("To add a device, please run 'pocket-coder login' on your computer CLI.");
  };

  const getIcon = (osInfo?: string) => {
    const os = osInfo?.toLowerCase() || '';
    if (os.includes('mac') || os.includes('darwin')) return <Monitor className="text-slate-300" size={24} />;
    if (os.includes('win')) return <Monitor className="text-blue-400" size={24} />;
    return <Server className="text-orange-400" size={24} />;
  };

  // Get initials for avatar placeholder
  const initials = user?.username ? user.username.substring(0, 2).toUpperCase() : 'U';

  return (
    <div className="min-h-screen bg-slate-950 text-slate-200 pb-20">
      {/* Header */}
      <header className="sticky top-0 z-10 bg-slate-950/80 backdrop-blur-md border-b border-slate-800 px-4 h-16 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-indigo-600 flex items-center justify-center text-sm font-bold">
            {initials}
          </div>
          <h1 className="font-bold text-lg">Devices</h1>
        </div>
        <button onClick={handleLogout} className="text-slate-400 hover:text-white text-sm font-medium">
          Log out
        </button>
      </header>

      {/* Device List */}
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">
          <span>Your Devices ({devices.length})</span>
        </div>

        {loading ? (
          <div className="text-center py-10 text-slate-500">Loading devices...</div>
        ) : (
          devices.map(device => (
            <div 
              key={device.id}
              onClick={() => handleSelectDevice(device)}
              className={`group relative bg-slate-900 border border-slate-800 rounded-xl p-4 transition-all ${
                device.status === 'online' 
                  ? 'hover:border-indigo-500/50 hover:bg-slate-800 active:scale-[0.98] cursor-pointer' 
                  : 'opacity-60 cursor-not-allowed'
              }`}
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-4">
                  <div className={`p-3 rounded-lg ${device.status === 'online' ? 'bg-slate-800' : 'bg-slate-900'}`}>
                    {getIcon(device.os_info)}
                  </div>
                  <div>
                    <h3 className="font-semibold text-white text-lg">{device.name}</h3>
                    <div className="flex items-center gap-2 text-sm text-slate-400 mt-0.5">
                      <span className="font-mono text-xs opacity-70">{device.ip || 'Unknown IP'}</span>
                      <span>•</span>
                      <span className="text-xs truncate max-w-[150px]">{device.os_info || 'Unknown OS'}</span>
                    </div>
                  </div>
                </div>
                
                <div className="flex flex-col items-end gap-2">
                  <button className="text-slate-500 hover:text-white p-1">
                    <MoreVertical size={18} />
                  </button>
                  {device.status === 'online' ? (
                    <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
                      <Wifi size={12} />
                      Online
                    </div>
                  ) : (
                     <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-slate-800 border border-slate-700 text-slate-500 text-xs font-medium">
                      <WifiOff size={12} />
                      Offline
                    </div>
                  )}
                </div>
              </div>
              
              {/* Status Indicator Bar */}
              <div className={`absolute left-0 top-4 bottom-4 w-1 rounded-r-full ${
                device.status === 'online' ? 'bg-emerald-500' : 'bg-slate-700'
              }`} />
            </div>
          ))
        )}
        
        {/* Add Device Button */}
        <button 
          onClick={handleAddDevice}
          className="w-full border-2 border-dashed border-slate-800 rounded-xl p-4 flex flex-col items-center justify-center gap-2 text-slate-500 hover:text-indigo-400 hover:border-indigo-500/30 hover:bg-slate-900/50 transition-all"
        >
          <div className="p-2 bg-slate-900 rounded-full">
            <Plus size={24} />
          </div>
          <span className="font-medium text-sm">Connect New Device</span>
        </button>
      </div>
    </div>
  );
};

export default DeviceListView;