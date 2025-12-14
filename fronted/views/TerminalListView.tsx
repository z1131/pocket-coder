import React from 'react';
import { ArrowLeft, Terminal as TerminalIcon, Plus, Clock, Cpu } from 'lucide-react';
import { Device, TerminalSession } from '../types';

interface Props {
  device: Device;
  onSelectTerminal: (session: TerminalSession) => void;
  onBack: () => void;
}

const MOCK_SESSIONS: TerminalSession[] = [
  { id: 't1', name: 'zsh', status: 'active', preview: 'user@server:~/projects/web $', uptime: '00:12:44' },
  { id: 't2', name: 'npm run dev', status: 'background', preview: '> ready in 405ms', uptime: '14:32:01' },
  { id: 't3', name: 'docker logs', status: 'background', preview: 'Postgres database system is ready to accept connections', uptime: '2 days' },
];

const TerminalListView: React.FC<Props> = ({ device, onSelectTerminal, onBack }) => {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-200">
      {/* Header */}
      <header className="sticky top-0 z-10 bg-slate-950/80 backdrop-blur-md border-b border-slate-800 px-4 h-16 flex items-center gap-4">
        <button onClick={onBack} className="p-2 -ml-2 text-slate-400 hover:text-white rounded-full hover:bg-slate-800 transition-colors">
          <ArrowLeft size={20} />
        </button>
        <div>
          <h1 className="font-bold text-white">{device.name}</h1>
          <div className="flex items-center gap-2 text-xs text-slate-400">
             <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
             Online
             <span>•</span>
             {device.ip}
          </div>
        </div>
      </header>

      <div className="p-4 pb-24 space-y-6">
        
        {/* Active / Foreground */}
        <section>
          <h2 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3 px-1">Active Session</h2>
          <div 
             onClick={() => onSelectTerminal(MOCK_SESSIONS[0])}
             className="bg-indigo-900/10 border border-indigo-500/30 rounded-xl p-4 active:scale-[0.98] transition-all cursor-pointer hover:bg-indigo-900/20"
          >
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-indigo-500 rounded-lg text-white shadow-lg shadow-indigo-500/20">
                  <TerminalIcon size={20} />
                </div>
                <div>
                  <h3 className="font-bold text-white">{MOCK_SESSIONS[0].name}</h3>
                  <p className="text-xs text-indigo-300">Foreground</p>
                </div>
              </div>
              <div className="flex items-center gap-1 text-xs text-indigo-300 font-mono bg-indigo-500/10 px-2 py-1 rounded">
                <Clock size={12} />
                {MOCK_SESSIONS[0].uptime}
              </div>
            </div>
            <div className="bg-slate-950/50 rounded-lg p-3 font-mono text-xs text-slate-300 border border-slate-800/50 truncate">
              {MOCK_SESSIONS[0].preview}
            </div>
          </div>
        </section>

        {/* Background */}
        <section>
          <h2 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3 px-1">Background Sessions</h2>
          <div className="space-y-3">
            {MOCK_SESSIONS.slice(1).map(session => (
              <div 
                key={session.id}
                onClick={() => onSelectTerminal(session)}
                className="bg-slate-900 border border-slate-800 rounded-xl p-4 active:scale-[0.98] transition-all cursor-pointer hover:border-slate-600 hover:bg-slate-800"
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-slate-800 rounded-lg text-slate-400">
                      <Cpu size={18} />
                    </div>
                    <span className="font-medium text-slate-200">{session.name}</span>
                  </div>
                  <div className="text-xs text-slate-500 font-mono">
                    {session.uptime}
                  </div>
                </div>
                <div className="pl-[3.25rem] text-xs text-slate-500 font-mono truncate opacity-70">
                   {session.preview}
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>

      {/* FAB: Create New Terminal */}
      <div className="fixed bottom-6 right-6">
        <button className="w-14 h-14 bg-indigo-600 rounded-full shadow-[0_4px_14px_0_rgba(79,70,229,0.5)] text-white flex items-center justify-center hover:bg-indigo-500 hover:scale-105 active:scale-95 transition-all">
          <Plus size={28} />
        </button>
      </div>
    </div>
  );
};

export default TerminalListView;
