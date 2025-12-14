import { VirtualKey } from './types';

export const VIRTUAL_KEYS: VirtualKey[] = [
  { label: 'Esc', value: 'ESCAPE', type: 'control' },
  { label: 'Tab', value: 'TAB', type: 'control' },
  { label: 'Ctrl', value: 'CTRL', type: 'control' },
  { label: 'Alt', value: 'ALT', type: 'control' },
  { label: '/', value: '/', type: 'char' },
  { label: '-', value: '-', type: 'char' },
  { label: '|', value: '|', type: 'char' },
  { label: '▲', value: 'UP', type: 'nav' },
  { label: '▼', value: 'DOWN', type: 'nav' },
  { label: '◀', value: 'LEFT', type: 'nav' },
  { label: '▶', value: 'RIGHT', type: 'nav' },
];

export const MOCK_BOOT_SEQUENCE = [
  "Initializing cloud connection...",
  "Authenticating user (mock_user)...",
  "Establishing secure channel...",
  "Connection established.",
  "Welcome to Pocket Coder v1.0.2",
  "Type 'help' for available commands."
];
