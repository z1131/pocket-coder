import React, { useRef } from 'react';
import { VIRTUAL_KEYS } from '../constants';

interface Props {
  onKeyPress: (key: string) => void;
  activeModifier?: string | null;
}

const VirtualKeyboard: React.FC<Props> = ({ onKeyPress, activeModifier }) => {
  const lastClickTime = useRef<number>(0);

  const handleKeyPress = (e: React.MouseEvent | React.TouchEvent, key: string) => {
    e.preventDefault();
    const now = Date.now();
    if (now - lastClickTime.current < 150) {
      return; // Debounce fast clicks
    }
    lastClickTime.current = now;
    onKeyPress(key);
  };

  return (
    <div className="w-full bg-slate-900 border-t border-slate-800">
      <div className="flex overflow-x-auto py-2 px-2 gap-2 no-scrollbar scroll-smooth">
        {VIRTUAL_KEYS.map((key) => {
          const isActive = key.value === activeModifier;
          return (
            <button
              key={key.label}
              onClick={(e) => handleKeyPress(e, key.value)}
              className={`
                flex-shrink-0 min-w-[3rem] h-9 px-2 rounded-md font-mono text-sm font-medium
                active:scale-95 transition-transform select-none touch-manipulation
                ${isActive 
                  ? 'bg-indigo-600 text-white shadow-[0_2px_0_0_rgba(67,56,202,1)] translate-y-[1px]' 
                  : ''}
                ${!isActive && key.type === 'control' 
                  ? 'bg-slate-700 text-indigo-300 shadow-[0_2px_0_0_rgba(51,65,85,1)] active:shadow-none active:translate-y-[2px]' 
                  : ''}
                ${key.type === 'nav' 
                  ? 'bg-slate-800 text-slate-300 shadow-[0_2px_0_0_rgba(30,41,59,1)] active:shadow-none active:translate-y-[2px]' 
                  : ''}
                ${key.type === 'char' 
                  ? 'bg-slate-800 text-slate-200 shadow-[0_2px_0_0_rgba(30,41,59,1)] active:shadow-none active:translate-y-[2px]' 
                  : ''}
              `}
            >
              {key.label}
            </button>
          );
        })}
      </div>
    </div>
  );
};

export default VirtualKeyboard;
