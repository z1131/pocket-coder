import React from 'react';
import { TerminalLine } from '../types';

interface Props {
  line: TerminalLine;
}

const TerminalLineRenderer: React.FC<Props> = ({ line }) => {
  const getStyles = () => {
    switch (line.type) {
      case 'input':
        return 'text-slate-100 font-bold';
      case 'error':
        return 'text-red-400';
      case 'system':
        return 'text-blue-400 italic';
      case 'info':
        return 'text-yellow-400';
      case 'output':
      default:
        return 'text-slate-300';
    }
  };

  if (line.type === 'input') {
    return (
      <div className="flex items-start break-all font-mono text-sm sm:text-base py-0.5">
        <span className="text-green-500 mr-2 shrink-0">➜</span>
        <span className="text-cyan-400 mr-2 shrink-0">~</span>
        <span className={getStyles()}>{line.text}</span>
      </div>
    );
  }

  return (
    <div className={`break-all font-mono text-sm sm:text-base py-0.5 whitespace-pre-wrap ${getStyles()}`}>
      {line.text}
    </div>
  );
};

export default TerminalLineRenderer;
