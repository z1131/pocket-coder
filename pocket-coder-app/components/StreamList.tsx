import React, { useEffect, useRef } from 'react';
import { Message, DiffContent } from '../types';
import DiffCard from './DiffCard';
import { Terminal, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

interface StreamListProps {
  messages: Message[];
  isThinking: boolean;
}

const StreamList: React.FC<StreamListProps> = ({ messages, isThinking }) => {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isThinking]);

  return (
    <div className="flex-1 overflow-y-auto pt-16 pb-32 px-4 space-y-4">
      <div className="text-center py-6">
        <p className="text-xs text-neutral-600 uppercase tracking-widest font-medium">Session Started</p>
        <p className="text-[10px] text-neutral-700 mt-1">Connected via WebSocket Secure</p>
      </div>

      {messages.map((msg) => (
        <div key={msg.id} className="animate-in fade-in slide-in-from-bottom-2 duration-300">
          
          {/* Simple Log */}
          {msg.type === 'log' && (
            <div className="flex gap-3 text-sm font-mono text-neutral-300">
              <Terminal size={16} className="mt-0.5 text-neutral-600 shrink-0" />
              <p className="break-words leading-relaxed">{msg.content as string}</p>
            </div>
          )}

          {/* Info/Success/Error */}
          {msg.type === 'success' && (
             <div className="flex gap-3 text-sm text-emerald-400 bg-emerald-950/10 p-3 rounded-lg border border-emerald-900/30">
               <CheckCircle2 size={18} className="shrink-0" />
               <p>{msg.content as string}</p>
             </div>
          )}
          
          {msg.type === 'error' && (
             <div className="flex gap-3 text-sm text-rose-400 bg-rose-950/10 p-3 rounded-lg border border-rose-900/30">
               <AlertCircle size={18} className="shrink-0" />
               <p>{msg.content as string}</p>
             </div>
          )}

          {/* Diff Visualization */}
          {msg.type === 'diff' && (
            <DiffCard content={msg.content as DiffContent} />
          )}

          {/* Prompt Question (Shown as history) */}
          {msg.type === 'prompt' && (
            <div className="flex justify-end">
              <div className="bg-neutral-800 text-neutral-200 px-4 py-2 rounded-2xl rounded-tr-sm text-sm max-w-[85%]">
                 {msg.content as string}
              </div>
            </div>
          )}
        </div>
      ))}

      {isThinking && (
        <div className="flex items-center gap-2 text-neutral-500 py-2">
           <Loader2 size={14} className="animate-spin" />
           <span className="text-xs font-mono">Agent is thinking...</span>
        </div>
      )}

      <div ref={bottomRef} />
    </div>
  );
};

export default StreamList;