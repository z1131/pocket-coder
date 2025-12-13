import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import 'xterm/css/xterm.css';
import { ArrowLeft, Sparkles, Send, Command, Wifi, WifiOff } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import { api, Session } from '../services/api';
import { ws } from '../services/ws';
import { useAuthStore } from '../store/useStore';
import VirtualKeyboard from '../components/VirtualKeyboard';
import AIAssistant from '../components/AIAssistant';

const TerminalView: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const [session, setSession] = useState<Session | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isAiOpen, setIsAiOpen] = useState(false);
  const [inputValue, setInputValue] = useState('');
  
  const terminalRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const navigate = useNavigate();
  const { token } = useAuthStore();

  // 1. Load Session Info
  useEffect(() => {
    if (!sessionId) return;
    console.log('Fetching session info for ID:', sessionId);
    api.session.get(Number(sessionId))
      .then(res => {
        console.log('Session loaded:', res);
        setSession(res);
      })
      .catch(err => console.error('Session load error:', err));
  }, [sessionId]);

  // 2. Init Xterm & WebSocket
  useEffect(() => {
    console.log('TerminalView Effect Triggered', { token: !!token, sessionId, hasRef: !!terminalRef.current });
    
    if (!token || !sessionId || !terminalRef.current) {
      console.warn('Missing dependencies for terminal init');
      return;
    }

    console.log('Initializing Terminal...');

    // Init Terminal
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#020617', // slate-950
        foreground: '#e2e8f0', // slate-200
      },
      allowProposedApi: true,
    });

    const fitAddon = new FitAddon();
    const webLinksAddon = new WebLinksAddon();
    
    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);
    term.open(terminalRef.current);
    fitAddon.fit();

    xtermRef.current = term;
    fitAddonRef.current = fitAddon;

    // Connect WebSocket
    ws.connect(token);

    // WS Event Handlers
    const offOpen = ws.on('open', () => {
      setIsConnected(true);
      term.write('\x1b[32m\r\n[Connected to Server]\x1b[0m\r\n');
      
      // Request history
      ws.send('terminal:history', { session_id: Number(sessionId) });
      
      // Send resize
      ws.send('terminal:resize', { 
        session_id: Number(sessionId),
        cols: term.cols, 
        rows: term.rows 
      });
    });

    const offClose = ws.on('close', () => {
      setIsConnected(false);
      term.write('\x1b[31m\r\n[Disconnected]\x1b[0m\r\n');
    });

    const offOutput = ws.on('terminal:output', (payload: any) => {
      if (payload.session_id === Number(sessionId) && payload.data) {
        try {
          // Decode Base64 (Standard JS atob)
          const text = atob(payload.data);
          term.write(text);
        } catch (e) {
          console.error('Failed to decode output', e);
        }
      }
    });
    
    const offHistory = ws.on('terminal:history', (payload: any) => {
       if (payload.session_id === Number(sessionId) && payload.data) {
         try {
           const text = atob(payload.data);
           term.write(text);
         } catch (e) {}
       }
    });

    // Terminal Input
    term.onData((data) => {
      if (!ws) return;
      // Encode Base64
      const encoded = btoa(data);
      ws.send('terminal:input', {
        session_id: Number(sessionId),
        data: encoded,
      });
    });

    // Resize Observer
    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit();
      if (ws) {
        ws.send('terminal:resize', { 
          session_id: Number(sessionId),
          cols: term.cols, 
          rows: term.rows 
        });
      }
    });
    resizeObserver.observe(terminalRef.current);

    return () => {
      offOpen();
      offClose();
      offOutput();
      offHistory();
      resizeObserver.disconnect();
      term.dispose();
      ws.disconnect();
    };
  }, [token, sessionId, session]);

  // Handle Input Bar (Quick Command)
  const handleQuickCommand = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue || !xtermRef.current) return;
    
    // Append newline and send
    const data = inputValue + '\r';
    const encoded = btoa(data);
    ws.send('terminal:input', {
      session_id: Number(sessionId),
      data: encoded,
    });
    setInputValue('');
    xtermRef.current.focus();
  };

  // Handle Virtual Keys
  const handleVirtualKey = (key: string) => {
    if (!xtermRef.current) return;
    let sequence = '';
    
    switch (key) {
      case 'UP': sequence = '\x1b[A'; break;
      case 'DOWN': sequence = '\x1b[B'; break;
      case 'LEFT': sequence = '\x1b[D'; break;
      case 'RIGHT': sequence = '\x1b[C'; break;
      case 'ESCAPE': sequence = '\x1b'; break;
      case 'TAB': sequence = '\t'; break;
      case 'CTRL+C': sequence = '\x03'; break;
      default: sequence = key;
    }

    if (sequence) {
      const encoded = btoa(sequence);
      ws.send('terminal:input', {
        session_id: Number(sessionId),
        data: encoded,
      });
    }
    xtermRef.current.focus();
  };

  if (!session) return <div className="bg-slate-950 h-screen text-slate-500 flex items-center justify-center">Loading terminal...</div>;

  return (
    <div className="h-[100dvh] w-full flex flex-col bg-slate-950 overflow-hidden font-sans fixed inset-0 z-50">
      
      {/* Top Bar */}
      <header className="h-12 bg-slate-900 border-b border-slate-800 flex items-center justify-between px-3 shrink-0 z-10 select-none">
        <div className="flex items-center gap-2">
          <button 
            onClick={() => navigate(`/desktops/${session.desktop_id}`)}
            className="p-1.5 hover:bg-slate-800 rounded-md text-slate-400 hover:text-white transition-colors"
          >
            <ArrowLeft size={18} />
          </button>
          <div className="flex flex-col">
            <span className="font-semibold text-slate-200 text-sm leading-tight">Session #{session.id}</span>
            <span className="text-[10px] text-slate-500 leading-tight">
              {session.is_default ? 'Default Terminal' : 'Background Task'}
            </span>
          </div>
        </div>
        
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 px-2 py-1 bg-slate-800 rounded-full border border-slate-700">
            {isConnected ? (
              <Wifi size={14} className="text-emerald-400" />
            ) : (
              <WifiOff size={14} className="text-red-400 animate-pulse" />
            )}
            <span className={`text-xs font-mono ${isConnected ? 'text-emerald-400' : 'text-slate-400'}`}>
              {isConnected ? 'Connected' : 'Offline'}
            </span>
          </div>
        </div>
      </header>

      {/* Terminal Area */}
      <main className="flex-1 relative bg-[#020617] overflow-hidden">
        <div ref={terminalRef} className="absolute inset-0 p-2" />
      </main>

      {/* Footer & Controls */}
      <footer className="bg-slate-950 border-t border-slate-800 shrink-0 z-20 pb-[env(safe-area-inset-bottom)]">
        <VirtualKeyboard onKeyPress={handleVirtualKey} />

        <div className="p-2 flex gap-2 items-center bg-slate-900/50 backdrop-blur">
          <button 
            onClick={() => setIsAiOpen(true)}
            className="flex-shrink-0 w-10 h-10 flex items-center justify-center rounded-lg bg-indigo-600/10 hover:bg-indigo-600/20 border border-indigo-500/30 text-indigo-400 transition-colors"
          >
            <Sparkles size={20} />
          </button>

          <form onSubmit={handleQuickCommand} className="flex-1 relative flex items-center">
            <span className="absolute left-3 text-slate-500 font-mono text-lg">$</span>
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              className="w-full bg-slate-800 text-white font-mono text-sm rounded-lg pl-8 pr-10 py-2.5 border border-slate-700 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all placeholder:text-slate-600"
              placeholder="Quick command..."
              autoComplete="off"
            />
            <button 
              type="submit"
              disabled={!inputValue}
              className={`absolute right-1 p-1.5 rounded-md transition-all ${ 
                inputValue ? 'text-indigo-400 bg-indigo-500/10' : 'text-slate-600'
              }`}
            >
              {inputValue ? <Send size={16} /> : <Command size={16} />}
            </button>
          </form>
        </div>
      </footer>

      <AIAssistant 
        isOpen={isAiOpen} 
        onClose={() => setIsAiOpen(false)} 
        onApplyCommand={(cmd) => {
          setInputValue(cmd);
          // Auto send? or let user confirm. Let's populate input.
        }}
      />
    </div>
  );
};

export default TerminalView;