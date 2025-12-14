export type LineType = 'input' | 'output' | 'error' | 'system' | 'info';

export interface TerminalLine {
  id: string;
  text: string;
  type: LineType;
  timestamp: number;
  cwd?: string; // current working directory
}

export interface ConnectionStatus {
  status: 'connecting' | 'connected' | 'disconnected' | 'reconnecting';
  latency: number;
}

export interface VirtualKey {
  label: string;
  value: string; // The actual code to send or action to trigger
  type: 'char' | 'control' | 'nav';
}

// --- New Types for Navigation & Data ---

export type ViewName = 'login' | 'register' | 'device-list' | 'terminal-list' | 'terminal';

export interface Device {
  id: string;
  name: string;
  os: 'linux' | 'macos' | 'windows';
  status: 'online' | 'offline';
  ip: string;
  lastActive: string;
  latency?: number;
}

export interface TerminalSession {
  id: string;
  name: string; // e.g., "zsh", "node server"
  status: 'active' | 'background' | 'idle';
  preview: string; // last line or command
  uptime: string;
}