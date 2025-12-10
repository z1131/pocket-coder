import React, { useState } from 'react';
import { PromptAction, ConnectionStatus } from '../types';
import { Send, XCircle, ArrowUp, ArrowDown, Keyboard } from 'lucide-react';

interface ActionBarProps {
  activePrompt: { id: string; actions?: PromptAction[] } | null;
  onRespond: (actionValue: string) => void;
  status: ConnectionStatus;
}

const ActionBar: React.FC<ActionBarProps> = ({ activePrompt, onRespond, status }) => {
  const [inputText, setInputText] = useState('');
  const [showInput, setShowInput] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputText.trim()) return;
    onRespond(inputText);
    setInputText('');
    setShowInput(false);
  };

  // 1. Prompt Mode: We have an active question from the AI
  if (activePrompt && activePrompt.actions && activePrompt.actions.length > 0) {
    return (
      <div className="fixed bottom-0 left-0 right-0 p-4 bg-black/90 backdrop-blur-xl border-t border-neutral-800 pb-8 z-40 animate-in slide-in-from-bottom-5 duration-300">
        <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between px-1">
                <span className="text-[10px] text-neutral-500 uppercase tracking-widest font-semibold">Action Required</span>
            </div>
            <div className="grid grid-cols-2 gap-3">
            {activePrompt.actions.map((action) => (
                <button
                key={action.value}
                onClick={() => onRespond(action.value)}
                className={`
                    h-14 rounded-xl font-medium text-sm transition-all active:scale-95 flex items-center justify-center
                    ${action.type === 'primary' 
                    ? 'bg-emerald-600 text-white shadow-[0_0_20px_rgba(16,185,129,0.2)] hover:bg-emerald-500 border border-emerald-500' 
                    : action.type === 'danger'
                    ? 'bg-rose-950/40 text-rose-400 border border-rose-900/50 hover:bg-rose-900/50 hover:text-rose-200'
                    : 'bg-neutral-800 text-neutral-300 border border-neutral-700 hover:bg-neutral-700'}
                `}
                >
                {action.label}
                </button>
            ))}
            </div>
            {/* Fallback to keyboard if options aren't enough */}
            <button 
                onClick={() => setShowInput(true)} 
                className="mx-auto mt-1 text-xs text-neutral-500 flex items-center gap-1 hover:text-neutral-300 p-2"
            >
                <Keyboard size={12} />
                Type custom response
            </button>
        </div>
      </div>
    );
  }

  // 2. Custom Input Mode (Overlay)
  if (showInput) {
      return (
        <div className="fixed bottom-0 left-0 right-0 p-4 bg-black/95 backdrop-blur-xl border-t border-neutral-800 pb-8 z-40">
             <form onSubmit={handleSubmit} className="flex gap-2 items-center">
                <button 
                    type="button"
                    onClick={() => setShowInput(false)}
                    className="p-3 text-neutral-500 hover:text-neutral-300"
                >
                    <XCircle size={24} />
                </button>
                <input 
                    autoFocus
                    type="text" 
                    value={inputText}
                    onChange={(e) => setInputText(e.target.value)}
                    placeholder="Type a command..."
                    className="flex-1 bg-neutral-900 border border-neutral-700 text-neutral-100 rounded-full h-12 px-5 focus:outline-none focus:border-neutral-500 font-mono text-sm"
                />
                <button 
                    type="submit"
                    className="p-3 bg-blue-600 text-white rounded-full hover:bg-blue-500 disabled:opacity-50"
                    disabled={!inputText.trim()}
                >
                    <Send size={20} />
                </button>
            </form>
        </div>
      )
  }

  // 3. Idle / Navigation Mode
  return (
    <div className="fixed bottom-0 left-0 right-0 p-4 bg-black/80 backdrop-blur-md border-t border-neutral-800 pb-8 z-40">
      <div className="flex items-center gap-3">
        <button 
            onClick={() => setShowInput(true)}
            disabled={status !== ConnectionStatus.CONNECTED}
            className="flex-1 h-12 bg-neutral-900 border border-neutral-800 rounded-full flex items-center px-4 text-neutral-500 hover:text-neutral-300 hover:border-neutral-700 transition-colors"
        >
            <span className="font-mono text-sm">Send command...</span>
        </button>
        
        {/* Quick Nav Keys (For terminal history, etc) */}
        <div className="flex gap-1">
             <button className="h-12 w-12 flex items-center justify-center rounded-full bg-neutral-900 border border-neutral-800 text-neutral-400 active:bg-neutral-800">
                <ArrowUp size={20} />
             </button>
             <button className="h-12 w-12 flex items-center justify-center rounded-full bg-neutral-900 border border-neutral-800 text-neutral-400 active:bg-neutral-800">
                <ArrowDown size={20} />
             </button>
        </div>
      </div>
    </div>
  );
};

export default ActionBar;