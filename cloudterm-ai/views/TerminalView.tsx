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

// Helper functions for UTF-8 Base64 encoding/decoding
function utf8_to_b64(str: string): string {
  const bytes = new TextEncoder().encode(str);
  const binString = Array.from(bytes, (byte) =>
    String.fromCodePoint(byte),
  ).join("");
  return btoa(binString);
}

function b64_to_utf8(str: string): string {
  const binString = atob(str);
  const bytes = Uint8Array.from(binString, (m) => m.codePointAt(0) || 0);
  return new TextDecoder().decode(bytes);
}

const TerminalView: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const [session, setSession] = useState<Session | null>(null);
  
  // Connection States
  const [isWsConnected, setIsWsConnected] = useState(false); // Mobile <-> Server
  const [isDesktopOnline, setIsDesktopOnline] = useState(false); // Desktop <-> Server
  const [desktopName, setDesktopName] = useState('');
  
  // Computed final status for UI
  const isConnected = isWsConnected && isDesktopOnline;

  const [isAiOpen, setIsAiOpen] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [activeModifier, setActiveModifier] = useState<string | null>(null);
  
  const terminalRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const navigate = useNavigate();
  const { token } = useAuthStore();

  // 1. Load Session & Desktop Info
  useEffect(() => {
    if (!sessionId) return;
    
    console.log('[TerminalView] Fetching session details for ID:', sessionId);

    api.session.get(Number(sessionId))
      .then(res => {
        console.log('[TerminalView] Session loaded successfully:', res);
        setSession(res);
        // Fetch desktop info to get name and initial status
        if (res.desktop_id) {
            console.log('[TerminalView] Fetching desktop info for ID:', res.desktop_id);
            api.desktop.get(res.desktop_id).then(d => {
                console.log('[TerminalView] Desktop loaded:', d);
                setDesktopName(d.name);
                setIsDesktopOnline(d.status === 'online');
            }).catch((err) => {
              console.error('[TerminalView] Failed to load desktop info:', err);
              setDesktopName('Unknown Host');
              setIsDesktopOnline(false);
            });
        }
      })
      .catch(err => {
        console.error('[TerminalView] Failed to load session:', err);
      });
  }, [sessionId]);

  // 2. Init Xterm & WebSocket
  useEffect(() => {
    if (!token || !sessionId || !terminalRef.current) return;

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
      setIsWsConnected(true);
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
      setIsWsConnected(false);
      term.write('\x1b[31m\r\n[Disconnected from Server]\x1b[0m\r\n');
    });

    const offOutput = ws.on('terminal:output', (payload: any) => {
      if (payload.session_id === Number(sessionId) && payload.data) {
        try {
          const text = b64_to_utf8(payload.data);
          term.write(text);
        } catch (e) {
          console.error('Failed to decode output', e);
        }
      }
    });
    
    const offHistory = ws.on('terminal:history', (payload: any) => {
       if (payload.session_id === Number(sessionId) && payload.data) {
         try {
           const text = b64_to_utf8(payload.data);
           term.write(text);
         } catch (e) {}
       }
    });

    const offOffline = ws.on('desktop:offline', (payload: any) => {
      if (session && payload.desktop_id === session.desktop_id) {
          term.write('\x1b[31m\r\n[Desktop Disconnected]\x1b[0m\r\n');
          setIsDesktopOnline(false);
      }
    });

    const offOnline = ws.on('desktop:online', (payload: any) => {
      if (session && payload.desktop_id === session.desktop_id) {
           term.write('\x1b[32m\r\n[Desktop Reconnected]\x1b[0m\r\n');
           setIsDesktopOnline(true);
      }
    });

    // Terminal Input
    term.onData((data) => {
      if (!ws) return;
      // Encode Base64 with UTF-8 support
      const encoded = utf8_to_b64(data);
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
      offOffline();
      offOnline();
      resizeObserver.disconnect();
      term.dispose();
      ws.disconnect();
    };
  }, [token, sessionId, session]);

  // Handle Input Change (for Sticky Keys)
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    
    // If a modifier is active and we have input
    if (activeModifier && newValue.length > 0) {
      // Get the last character typed (assuming append)
      const char = newValue.slice(-1);
      
      if (activeModifier === 'CTRL') {
        const code = char.toUpperCase().charCodeAt(0);
        if (code >= 64 && code <= 95) {
          // A-Z, [, \, ], ^, _
          const ctrlChar = String.fromCharCode(code - 64);
          const encoded = utf8_to_b64(ctrlChar);
          ws.send('terminal:input', {
            session_id: Number(sessionId),
            data: encoded,
          });
        } else if (code >= 97 && code <= 122) {
          // a-z
          const ctrlChar = String.fromCharCode(code - 96);
          const encoded = utf8_to_b64(ctrlChar);
          ws.send('terminal:input', {
            session_id: Number(sessionId),
            data: encoded,
          });
        }
      } else if (activeModifier === 'ALT') {
         // ESC + char
         const encoded = utf8_to_b64('\x1b' + char);
         ws.send('terminal:input', {
            session_id: Number(sessionId),
            data: encoded,
         });
      }

      // Clear input and reset modifier
      setInputValue('');
      setActiveModifier(null);
      // Try to keep focus on input for continuous typing
      // xtermRef.current?.focus(); 
    } else {
      setInputValue(newValue);
    }
  };

  // Handle Input Bar (Quick Command)
  const handleQuickCommand = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue || !xtermRef.current) return;
    
    // Append newline and send
    const data = inputValue + '\r';
    const encoded = utf8_to_b64(data);
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
    
    // Toggle Modifiers
    if (key === 'CTRL' || key === 'ALT') {
      setActiveModifier(current => current === key ? null : key);
      return;
    }

    // If modifier is active, apply simple reset for now to avoid confusion
    if (activeModifier) {
       setActiveModifier(null);
    }

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
      const encoded = utf8_to_b64(sequence);
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
            <span className="font-semibold text-slate-200 text-sm leading-tight">
                {session.is_default ? (desktopName || 'Terminal') : (session.agent_type || 'Session')}
            </span>
            <span className="text-[10px] text-slate-500 leading-tight">
              {session.is_default ? 'Local Shell' : `Session #${session.id}`}
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
              {isConnected ? 'Connected' : (isWsConnected ? 'Desktop Offline' : 'Offline')}
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
        <VirtualKeyboard onKeyPress={handleVirtualKey} activeModifier={activeModifier} />

        <div className="p-2 flex gap-2 items-center bg-slate-900/50 backdrop-blur">
          <button 
            onClick={() => setIsAiOpen(true)}
            className="flex-shrink-0 w-10 h-10 flex items-center justify-center rounded-lg bg-indigo-600/10 hover:bg-indigo-600/20 border border-indigo-500/30 text-indigo-400 transition-colors"
          >
            <Sparkles size={20} />
          </button>

          <form onSubmit={handleQuickCommand} className="flex-1 relative flex items-center">
            <span className="absolute left-3 text-slate-500 font-mono text-lg select-none">{activeModifier ? `${activeModifier} +` : '$'}</span>
            <input
              type="text"
              value={inputValue}
              onChange={handleInputChange}
              className={`w-full bg-slate-800 text-white font-mono text-sm rounded-lg pr-10 py-2.5 border border-slate-700 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all placeholder:text-slate-600 ${
                activeModifier ? 'pl-24 ring-1 ring-indigo-500 border-indigo-500 bg-indigo-900/20' : 'pl-8'
              }`}
              placeholder={isConnected ? (activeModifier ? "Type a key..." : "Quick command...") : "Disconnected"}
              disabled={!isConnected}
              autoComplete="off"
            />
            <button 
              type="submit"
              disabled={!inputValue || !isConnected}
              className={`absolute right-1 p-1.5 rounded-md transition-all ${ 
                inputValue && isConnected ? 'text-indigo-400 bg-indigo-500/10' : 'text-slate-600'
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
        }}
      />
    </div>
  );
};

export default TerminalView;
