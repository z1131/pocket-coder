import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { usePocketWS } from '../hooks/usePocketWS';
import { ConnectionStatus, PocketEvent } from '../types';
import Terminal from '../components/Terminal';
import { ArrowLeft, Keyboard, MoreVertical } from 'lucide-react';

const SessionPage: React.FC = () => {
  const { desktopId: idParam } = useParams<{ desktopId: string }>();
  const desktopId = idParam ? parseInt(idParam, 10) : 0;

  const navigate = useNavigate();
  const { accessToken } = useAuth();
  const [terminalOutput, setTerminalOutput] = useState<string>('');
  const [showKeyboard, setShowKeyboard] = useState(false);

  const handleEvent = useCallback((evt: PocketEvent) => {
    switch (evt.kind) {
      case 'terminal:output':
        // 累加终端输出，而不是覆盖
        setTerminalOutput(prev => prev + evt.data);
        break;
      case 'terminal:exit':
        console.log('Terminal exited with code:', evt.code);
        navigate('/desktops');
        break;
      case 'error':
        console.error('Error:', evt.message);
        break;
      default:
        break;
    }
  }, [navigate]);

  const { status: wsStatus, sendTerminalInput, sendTerminalResize } = usePocketWS({
    token: accessToken,
    onEvent: handleEvent
  });

  const onTerminalData = (data: string) => {
    if (wsStatus === ConnectionStatus.CONNECTED) {
      sendTerminalInput(desktopId, data);
    }
  };

  const onTerminalResize = (cols: number, rows: number) => {
    if (wsStatus === ConnectionStatus.CONNECTED) {
      sendTerminalResize(desktopId, cols, rows);
    }
  };

  const handleSpecialKey = (key: string) => {
    onTerminalData(key);
  };

  useEffect(() => {
    if (!accessToken) {
      navigate('/desktops', { replace: true });
    }
  }, [accessToken, navigate]);

  return (
    <div className="flex flex-col h-screen bg-black text-white">
      <div className="flex items-center justify-between p-4 border-b border-gray-800">
        <div className="flex items-center gap-4">
          <button
            className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
            onClick={() => navigate('/desktops')}
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex flex-col">
            <h1 className="text-sm font-medium">Remote Session #{desktopId}</h1>
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${wsStatus === ConnectionStatus.CONNECTED ? 'bg-green-500' :
                wsStatus === ConnectionStatus.CONNECTING ? 'bg-yellow-500' : 'bg-red-500'
                }`} />
              <span className="text-xs text-gray-400">
                {wsStatus === ConnectionStatus.CONNECTED ? '已连接' :
                  wsStatus === ConnectionStatus.CONNECTING ? '连接中...' : '未连接'}
              </span>
            </div>
          </div>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowKeyboard(!showKeyboard)}
            className={`p-2 rounded-lg transition-colors ${showKeyboard ? 'bg-blue-600 text-white' : 'hover:bg-gray-800 text-gray-300'}`}
          >
            <Keyboard className="w-5 h-5" />
          </button>
          <button className="p-2 hover:bg-gray-800 rounded-lg transition-colors text-gray-300">
            <MoreVertical className="w-5 h-5" />
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-hidden relative">
        <Terminal
          output={terminalOutput}
          onData={onTerminalData}
          onResize={onTerminalResize}
        />
      </div>

      {/* Virtual Keyboard Toolbar */}
      {showKeyboard && (
        <div className="bg-gray-800 border-t border-gray-700 p-2 grid grid-cols-6 gap-2 shrink-0">
          <KeyButton label="ESC" value="\x1b" onClick={handleSpecialKey} />
          <KeyButton label="TAB" value="\t" onClick={handleSpecialKey} />
          <KeyButton label="CTRL+C" value="\x03" onClick={handleSpecialKey} />
          <KeyButton label="CTRL+Z" value="\x1a" onClick={handleSpecialKey} />
          <KeyButton label="↑" value="\x1b[A" onClick={handleSpecialKey} />
          <KeyButton label="↓" value="\x1b[B" onClick={handleSpecialKey} />
          <KeyButton label="←" value="\x1b[D" onClick={handleSpecialKey} />
          <KeyButton label="→" value="\x1b[C" onClick={handleSpecialKey} />
          <KeyButton label="/" value="/" onClick={handleSpecialKey} />
          <KeyButton label="-" value="-" onClick={handleSpecialKey} />
          <KeyButton label="HOME" value="\x1b[H" onClick={handleSpecialKey} />
          <KeyButton label="END" value="\x1b[F" onClick={handleSpecialKey} />
        </div>
      )}
    </div>
  );
};

const KeyButton: React.FC<{ label: string; value: string; onClick: (v: string) => void }> = ({ label, value, onClick }) => (
  <button
    onClick={() => onClick(value)}
    className="bg-gray-700 hover:bg-gray-600 active:bg-gray-500 text-xs font-mono py-3 rounded transition-colors select-none"
  >
    {label}
  </button>
);

export default SessionPage;
