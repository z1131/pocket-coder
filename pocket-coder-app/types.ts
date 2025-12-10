export type MessageType = 'log' | 'diff' | 'prompt' | 'success' | 'error' | 'info';

export interface DiffContent {
  file: string;
  language: string;
  lines: string[];
}

export interface PromptAction {
  label: string;
  value: string;
  type: 'primary' | 'danger' | 'neutral';
}

export interface Message {
  id: string;
  type: MessageType;
  content: string | DiffContent;
  timestamp: number;
  // If type is prompt, these populate the UI buttons
  actions?: PromptAction[]; 
}

export enum ConnectionStatus {
  CONNECTING = 'CONNECTING',
  CONNECTED = 'CONNECTED',
  DISCONNECTED = 'DISCONNECTED',
  BUSY = 'BUSY'
}

export type PocketEvent =
  | { kind: 'desktop-status'; desktopId: number; status: 'online' | 'offline' }
  | { kind: 'session-create'; sessionId: number; workingDir?: string }
  | { kind: 'agent-stream'; sessionId: number; delta: string }
  | { kind: 'agent-response'; sessionId: number; content: string; role: string }
  | { kind: 'agent-status'; status: string; sessionId?: number }
  | { kind: 'terminal:output'; data: string }
  | { kind: 'terminal:history'; data: string }
  | { kind: 'terminal:exit'; code: number }
  | { kind: 'error'; code: number; message: string }
  | { kind: 'pong' };