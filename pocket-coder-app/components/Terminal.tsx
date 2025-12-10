import React, { useEffect, useRef, useCallback } from 'react';
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
  const inputRef = useRef<HTMLInputElement>(null);
  const composingRef = useRef<boolean>(false);

  // 处理按键（捕获特殊键和普通字符）
  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
    // 正在输入法组合中，不处理
    if (composingRef.current) return;

    const key = e.key;
    
    // 回车键
    if (key === 'Enter') {
      e.preventDefault();
      onData('\r');
      return;
    }
    // 退格键
    if (key === 'Backspace') {
      e.preventDefault();
      onData('\x7f');
      return;
    }
    // Tab 键
    if (key === 'Tab') {
      e.preventDefault();
      onData('\t');
      return;
    }
    // 方向键
    if (key === 'ArrowUp') {
      e.preventDefault();
      onData('\x1b[A');
      return;
    }
    if (key === 'ArrowDown') {
      e.preventDefault();
      onData('\x1b[B');
      return;
    }
    if (key === 'ArrowLeft') {
      e.preventDefault();
      onData('\x1b[D');
      return;
    }
    if (key === 'ArrowRight') {
      e.preventDefault();
      onData('\x1b[C');
      return;
    }
    // Escape
    if (key === 'Escape') {
      e.preventDefault();
      onData('\x1b');
      return;
    }

    // 普通可打印字符（单个字符）
    if (key.length === 1) {
      e.preventDefault();
      onData(key);
    }
  }, [onData]);

  // 处理输入法组合开始
  const handleCompositionStart = useCallback(() => {
    composingRef.current = true;
  }, []);

  // 处理输入法组合结束（中文等）
  const handleCompositionEnd = useCallback((e: React.CompositionEvent<HTMLInputElement>) => {
    composingRef.current = false;
    const data = e.data;
    if (data) {
      onData(data);
    }
    // 清空输入框
    if (inputRef.current) {
      inputRef.current.value = '';
    }
  }, [onData]);

  // 点击终端时聚焦到隐藏输入框
  const handleContainerClick = useCallback(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

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

    // xterm.js 原生键盘事件（桌面端浏览器可能用这个）
    term.onData((data) => {
      console.log('[xterm onData]', JSON.stringify(data));
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

    // 自动聚焦
    setTimeout(() => {
      if (inputRef.current) {
        inputRef.current.focus();
      }
    }, 100);

    return () => {
      window.removeEventListener('resize', handleResize);
      term.dispose();
    };
  }, []);

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
      style={{ 
        width: '100%', 
        height: '100%', 
        overflow: 'hidden',
        backgroundColor: '#1e1e1e',
        position: 'relative',
      }}
      onClick={handleContainerClick}
      onTouchStart={handleContainerClick}
    >
      {/* 隐藏的输入框，用于捕获移动端键盘输入 */}
      <input
        ref={inputRef}
        type="text"
        onKeyDown={handleKeyDown}
        onCompositionStart={handleCompositionStart}
        onCompositionEnd={handleCompositionEnd}
        autoCapitalize="off"
        autoCorrect="off"
        autoComplete="off"
        spellCheck={false}
        enterKeyHint="send"
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '1px',
          height: '1px',
          opacity: 0,
          zIndex: -1,
          border: 'none',
          outline: 'none',
          padding: 0,
          fontSize: '16px', // 防止 iOS 缩放
          background: 'transparent',
        }}
      />
      {/* xterm.js 终端容器 */}
      <div 
        ref={containerRef} 
        style={{ 
          width: '100%', 
          height: '100%', 
          position: 'absolute',
          top: 0,
          left: 0,
          zIndex: 0,
        }} 
      />
    </div>
  );
};

export default Terminal;
