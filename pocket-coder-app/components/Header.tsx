import React from 'react';
import { Wifi, WifiOff, Cpu, Activity } from 'lucide-react';
import { ConnectionStatus } from '../types';

interface HeaderProps {
  status: ConnectionStatus;
  machineName: string;
}

const Header: React.FC<HeaderProps> = ({ status, machineName }) => {
  const getStatusColor = () => {
    switch (status) {
      case ConnectionStatus.CONNECTED: return 'text-emerald-500';
      case ConnectionStatus.BUSY: return 'text-amber-500';
      case ConnectionStatus.DISCONNECTED: return 'text-rose-500';
      default: return 'text-neutral-500';
    }
  };

  return (
    <header className="fixed top-0 left-0 right-0 h-14 bg-black/80 backdrop-blur-md border-b border-neutral-800 z-50 flex items-center justify-between px-4 select-none">
      <div className="flex items-center gap-3">
        <div className={`p-1.5 rounded-md bg-neutral-900 border border-neutral-800 ${getStatusColor()}`}>
          {status === ConnectionStatus.CONNECTED ? <Wifi size={16} /> : 
           status === ConnectionStatus.BUSY ? <Activity size={16} className="animate-pulse" /> :
           <WifiOff size={16} />}
        </div>
        <div>
          <h1 className="text-sm font-semibold text-neutral-200 tracking-tight flex items-center gap-2">
            {machineName}
            <span className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-neutral-900 border border-neutral-800 text-neutral-400">
              SSH
            </span>
          </h1>
          <p className="text-[10px] text-neutral-500 font-medium tracking-wide uppercase">
            {status}
          </p>
        </div>
      </div>
      
      <div className="flex items-center">
        <Cpu size={18} className="text-neutral-600" />
      </div>
    </header>
  );
};

export default Header;