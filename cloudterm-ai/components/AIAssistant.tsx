import React, { useState, useEffect } from 'react';
import { X, Sparkles, Send, Check } from 'lucide-react';
import { simulateAIResponse } from '../services/mockTerminalService';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onApplyCommand: (cmd: string) => void;
}

const AIAssistant: React.FC<Props> = ({ isOpen, onClose, onApplyCommand }) => {
  const [prompt, setPrompt] = useState('');
  const [isThinking, setIsThinking] = useState(false);
  const [suggestion, setSuggestion] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) {
      // Reset state when closed
      setPrompt('');
      setSuggestion(null);
      setIsThinking(false);
    }
  }, [isOpen]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!prompt.trim()) return;

    setIsThinking(true);
    setSuggestion(null);
    
    // Call mock service
    const result = await simulateAIResponse(prompt);
    
    setSuggestion(result);
    setIsThinking(false);
  };

  const handleSelect = () => {
    if (suggestion) {
      onApplyCommand(suggestion);
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="absolute inset-0 z-50 flex items-end justify-center sm:items-center animate-in fade-in duration-200">
      {/* Invisible backdrop to catch clicks outside, keeping terminal visible */}
      <div 
        className="absolute inset-0 bg-transparent" 
        onClick={onClose}
      />

      <div 
        className="relative w-full sm:w-[500px] bg-slate-900 border-t sm:border border-slate-700 sm:rounded-xl shadow-2xl overflow-hidden flex flex-col max-h-[60vh] sm:max-h-[80vh] mb-0 sm:mb-10"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-800 bg-slate-900">
          <div className="flex items-center gap-2 text-indigo-400">
            <Sparkles size={18} />
            <h3 className="font-semibold text-white">AI Command Assistant</h3>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white p-1">
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4 overflow-y-auto">
          {/* Conversation History / Result */}
          {suggestion && (
            <div className="bg-slate-800 rounded-lg p-3 border border-slate-700 animate-in slide-in-from-bottom-2">
              <div className="text-xs text-slate-400 mb-1">Suggested Command:</div>
              <div className="font-mono text-green-400 text-sm break-all bg-slate-950 p-2 rounded border border-slate-800 mb-3">
                {suggestion}
              </div>
              <div className="flex gap-2 justify-end">
                 <button 
                  onClick={() => setSuggestion(null)} 
                  className="px-3 py-1.5 text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-700 rounded transition-colors"
                >
                  Retry
                </button>
                <button 
                  onClick={handleSelect}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-indigo-600 hover:bg-indigo-500 text-white rounded transition-colors"
                >
                  <Check size={12} />
                  Select
                </button>
              </div>
            </div>
          )}

          {isThinking && (
             <div className="flex items-center gap-3 p-3 text-slate-400 animate-pulse">
                <Sparkles size={16} className="animate-spin" />
                <span className="text-sm">Analyzing intent...</span>
             </div>
          )}

          {/* Empty State / Prompt */}
          {!suggestion && !isThinking && (
             <div className="text-center py-6 text-slate-500">
                <p className="text-sm">Describe what you want to do in natural language.</p>
                <p className="text-xs mt-1">e.g., "Undo the last commit" or "List all node processes"</p>
             </div>
          )}
        </div>

        {/* Input Area */}
        <form onSubmit={handleSubmit} className="p-3 bg-slate-950 border-t border-slate-800 flex gap-2">
          <input
            autoFocus
            type="text"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Ask AI to generate a command..."
            className="flex-1 bg-slate-900 text-slate-200 text-sm rounded-lg border border-slate-700 px-3 py-2 focus:outline-none focus:border-indigo-500 transition-colors"
          />
          <button 
            type="submit" 
            disabled={!prompt.trim() || isThinking}
            className="bg-indigo-600 disabled:bg-slate-800 disabled:text-slate-500 text-white p-2 rounded-lg hover:bg-indigo-500 transition-colors"
          >
            <Send size={18} />
          </button>
        </form>
      </div>
    </div>
  );
};

export default AIAssistant;