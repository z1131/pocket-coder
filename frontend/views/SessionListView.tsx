import React, { useEffect, useState } from 'react';
import { ArrowLeft, Terminal as TerminalIcon, Plus, Clock, Cpu, Trash2 } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import { api, Session, Device } from '../services/api';
import { useAppStore } from '../store/useStore';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

dayjs.extend(relativeTime);

// Helper: Decode UTF-8 Base64
function b64_to_utf8(str: string): string {
  try {
    const binString = atob(str);
    const bytes = Uint8Array.from(binString, (m) => m.codePointAt(0) || 0);
    return new TextDecoder().decode(bytes);
  } catch (e) {
    return '';
  }
}

// Helper: Strip ANSI codes
function stripAnsi(str: string): string {
  // eslint-disable-next-line no-control-regex
  return str.replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, '');
}

const SessionListView: React.FC = () => {
  const { desktopId } = useParams<{ desktopId: string }>();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [device, setDevice] = useState<Device | null>(null);
  const [loading, setLoading] = useState(true);
  const setCurrentSession = useAppStore((state) => state.setCurrentSession);
  const navigate = useNavigate();

  useEffect(() => {
    if (!desktopId) return;

    const loadData = async () => {
      try {
        const [deviceData, sessionData] = await Promise.all([
          api.desktop.get(Number(desktopId)),
          api.session.list(Number(desktopId))
        ]);
        setDevice(deviceData);
        setSessions(sessionData.sessions);
      } catch (err) {
        console.error('Failed to load sessions', err);
        navigate('/');
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [desktopId, navigate]);

  const handleCreateSession = async () => {
    if (!desktopId) return;
    try {
      const newSession = await api.session.create(Number(desktopId));
      setSessions([newSession, ...sessions]); // Optimistic update
      handleSelectSession(newSession);
    } catch (err) {
      console.error('Failed to create session', err);
    }
  };

  const handleDeleteSession = async (e: React.MouseEvent, id: number) => {
    e.stopPropagation();
    if (!confirm('Are you sure you want to close this session?')) return;
    try {
      await api.session.delete(id);
      setSessions(sessions.filter(s => s.id !== id));
    } catch (err) {
      console.error('Failed to delete session', err);
    }
  };

  const handleSelectSession = (session: Session) => {
    setCurrentSession(session);
    navigate(`/sessions/${session.id}`);
  };

  const decodePreview = (b64?: string) => {
    if (!b64) return '';
    try {
      const decoded = b64_to_utf8(b64);
      return stripAnsi(decoded);
    } catch (e) {
      return 'Invalid preview data';
    }
  };

  const defaultSession = sessions.find(s => s.is_default);
  const backgroundSessions = sessions.filter(s => !s.is_default);

  if (loading) return <div className="text-center py-10 text-slate-500">Loading sessions...</div>;
  if (!device) return null;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-200 pb-24">
      {/* Header */}
      <header className="sticky top-0 z-10 bg-slate-950/80 backdrop-blur-md border-b border-slate-800 px-4 h-16 flex items-center gap-4">
        <button onClick={() => navigate('/')} className="p-2 -ml-2 text-slate-400 hover:text-white rounded-full hover:bg-slate-800 transition-colors">
          <ArrowLeft size={20} />
        </button>
        <div>
          <h1 className="font-bold text-white">{device.name}</h1>
          <div className="flex items-center gap-2 text-xs text-slate-400">
             <span className={`w-2 h-2 rounded-full ${device.status === 'online' ? 'bg-emerald-500 animate-pulse' : 'bg-slate-500'}`} />
             {device.status === 'online' ? 'Online' : 'Offline'}
             <span>•</span>
             {device.ip || 'Unknown IP'}
          </div>
        </div>
      </header>

      <div className="p-4 space-y-6">
        
        {/* Active / Foreground */}
        {defaultSession && (
          <section>
            <h2 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3 px-1">Default Terminal</h2>
            <div 
               onClick={() => handleSelectSession(defaultSession)}
               className="bg-indigo-900/10 border border-indigo-500/30 rounded-xl p-4 active:scale-[0.98] transition-all cursor-pointer hover:bg-indigo-900/20 group relative"
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-indigo-500 rounded-lg text-white shadow-lg shadow-indigo-500/20">
                    <TerminalIcon size={20} />
                  </div>
                  <div>
                    <h3 className="font-bold text-white">Main Terminal</h3>
                    <p className="text-xs text-indigo-300">Shared Session</p>
                  </div>
                </div>
                <div className="flex items-center gap-1 text-xs text-indigo-300 font-mono bg-indigo-500/10 px-2 py-1 rounded">
                  <Clock size={12} />
                  {dayjs(defaultSession.started_at).fromNow(true)}
                </div>
              </div>
              <div className="bg-slate-950/50 rounded-lg p-3 font-mono text-xs text-slate-300 border border-slate-800/50 truncate h-12 whitespace-pre-wrap overflow-hidden">
                {decodePreview(defaultSession.preview) || 'No output yet...'}
              </div>
            </div>
          </section>
        )}

        {/* Background */}
        <section>
          <h2 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3 px-1">Background Tasks</h2>
          {backgroundSessions.length === 0 ? (
            <div className="text-center py-8 text-slate-600 text-sm italic">No background sessions</div>
          ) : (
            <div className="space-y-3">
              {backgroundSessions.map(session => (
                <div 
                  key={session.id}
                  onClick={() => handleSelectSession(session)}
                  className="bg-slate-900 border border-slate-800 rounded-xl p-4 active:scale-[0.98] transition-all cursor-pointer hover:border-slate-600 hover:bg-slate-800 relative group"
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-slate-800 rounded-lg text-slate-400">
                        <Cpu size={18} />
                      </div>
                      <span className="font-medium text-slate-200">Session #{session.id}</span>
                    </div>
                    <div className="text-xs text-slate-500 font-mono">
                      {dayjs(session.started_at).fromNow(true)}
                    </div>
                  </div>
                  <div className="pl-[3.25rem] text-xs text-slate-500 font-mono truncate opacity-70">
                     {decodePreview(session.preview) || '...'}
                  </div>
                  
                  {/* Delete Button */}
                  <button 
                    onClick={(e) => handleDeleteSession(e, session.id)}
                    className="absolute top-4 right-4 p-2 text-slate-600 hover:text-red-400 hover:bg-red-400/10 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>

      {/* FAB: Create New Terminal */}
      <div className="fixed bottom-6 right-6">
        <button 
          onClick={handleCreateSession}
          className="w-14 h-14 bg-indigo-600 rounded-full shadow-[0_4px_14px_0_rgba(79,70,229,0.5)] text-white flex items-center justify-center hover:bg-indigo-500 hover:scale-105 active:scale-95 transition-all"
        >
          <Plus size={28} />
        </button>
      </div>
    </div>
  );
};

export default SessionListView;
