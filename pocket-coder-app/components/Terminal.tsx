import React, { useEffect, useRef } from 'react';
import { Terminal as XTerm } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';

interface TerminalProps {
  onData: (data: string) => void;
  onResize: (cols: number, rows: number) => void;
  output?: string;
}

const Terminal: React.FC<TerminalProps> = ({ onData, onResize, output }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const terminalRef = useRef<XTerm | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const lastOutputLengthRef = useRef<number>(0);

  useEffect(() => {
    if (!containerRef.current) return;

    const term = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#ffffff',
      },
      allowProposedApi: true,
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    term.open(containerRef.current);
    fitAddon.fit();

    term.onData((data) => {
      onData(data);
    });

    term.onResize((size) => {
      onResize(size.cols, size.rows);
    });

    terminalRef.current = term;
    fitAddonRef.current = fitAddon;

    // Initial resize
    onResize(term.cols, term.rows);

    const handleResize = () => {
      fitAddon.fit();
      onResize(term.cols, term.rows);
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      term.dispose();
    };
  }, []); // Empty dependency array to run only once

  useEffect(() => {
    if (terminalRef.current && output) {
      // 只写入新增的部分，避免重复写入
      const newContent = output.slice(lastOutputLengthRef.current);
      if (newContent) {
        terminalRef.current.write(newContent);
        lastOutputLengthRef.current = output.length;
      }
    }
  }, [output]);

  return (
    <div 
      ref={containerRef} 
      style={{ 
        width: '100%', 
        height: '100%', 
        overflow: 'hidden',
        backgroundColor: '#1e1e1e' 
      }} 
    />
  );
};

export default Terminal;
