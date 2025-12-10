import React from 'react';
import { DiffContent } from '../types';
import { FileCode2, GitCommitHorizontal } from 'lucide-react';

interface DiffCardProps {
  content: DiffContent;
}

const DiffCard: React.FC<DiffCardProps> = ({ content }) => {
  return (
    <div className="my-3 rounded-lg border border-neutral-800 bg-neutral-950 overflow-hidden shadow-sm">
      {/* File Header */}
      <div className="px-3 py-2 bg-neutral-900 border-b border-neutral-800 flex items-center gap-2">
        <FileCode2 size={14} className="text-blue-400" />
        <span className="text-xs font-mono text-neutral-300 truncate">{content.file}</span>
        <span className="text-[10px] text-neutral-500 ml-auto border border-neutral-800 px-1.5 rounded bg-neutral-950">
          {content.language}
        </span>
      </div>

      {/* Code Content */}
      <div className="overflow-x-auto">
        <div className="min-w-full font-mono text-[11px] leading-5 p-2">
          {content.lines.map((line, idx) => {
            const isAdd = line.startsWith('+');
            const isDel = line.startsWith('-');
            const isMeta = line.startsWith('@@');

            let bgClass = '';
            let textClass = 'text-neutral-400';
            let prefix = ' ';

            if (isAdd) {
              bgClass = 'bg-emerald-950/30';
              textClass = 'text-emerald-400';
              prefix = '+';
            } else if (isDel) {
              bgClass = 'bg-rose-950/30';
              textClass = 'text-rose-400 line-through decoration-rose-900/50';
              prefix = '-';
            } else if (isMeta) {
              textClass = 'text-blue-500 font-bold';
              prefix = ' ';
            }

            // Strip the actual first char if it's the marker to avoid double rendering if data comes raw
            const cleanLine = (isAdd || isDel) ? line.substring(1) : line;

            return (
              <div key={idx} className={`flex ${bgClass} -mx-2 px-2`}>
                <span className={`select-none w-4 inline-block text-center opacity-50 mr-2 ${textClass}`}>
                  {prefix}
                </span>
                <span className={`whitespace-pre ${textClass}`}>
                  {cleanLine}
                </span>
              </div>
            );
          })}
        </div>
      </div>
      
      {/* Footer */}
      <div className="px-3 py-1.5 bg-neutral-900/50 border-t border-neutral-800 flex items-center justify-between">
        <div className="flex items-center gap-1.5">
           <GitCommitHorizontal size={12} className="text-neutral-500" />
           <span className="text-[10px] text-neutral-500">Proposed Change</span>
        </div>
      </div>
    </div>
  );
};

export default DiffCard;