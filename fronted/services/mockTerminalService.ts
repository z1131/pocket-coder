// This simulates the Backend WebSocket/PTY logic
import { TerminalLine } from '../types';

export const generateId = () => Math.random().toString(36).substr(2, 9);

export const simulateCommand = async (cmd: string): Promise<TerminalLine[]> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const command = cmd.trim();
      const timestamp = Date.now();
      
      let lines: TerminalLine[] = [];

      switch (command) {
        case 'help':
          lines = [
            { id: generateId(), type: 'info', text: 'Available commands:', timestamp },
            { id: generateId(), type: 'output', text: '  ls          List directory contents', timestamp },
            { id: generateId(), type: 'output', text: '  ps          Report a snapshot of the current processes', timestamp },
            { id: generateId(), type: 'output', text: '  git status  Show the working tree status', timestamp },
            { id: generateId(), type: 'output', text: '  clear       Clear the terminal screen', timestamp },
            { id: generateId(), type: 'output', text: '  whoami      Print effective userid', timestamp },
          ];
          break;
        case 'ls':
        case 'ls -la':
          lines = [
            { id: generateId(), type: 'output', text: 'total 32', timestamp },
            { id: generateId(), type: 'output', text: 'drwxr-xr-x  5 user  staff  160 Oct 24 10:00 .', timestamp },
            { id: generateId(), type: 'output', text: 'drwxr-xr-x  3 user  staff   96 Oct 24 09:58 ..', timestamp },
            { id: generateId(), type: 'output', text: '-rw-r--r--  1 user  staff  420 Oct 24 10:00 index.ts', timestamp },
            { id: generateId(), type: 'output', text: '-rw-r--r--  1 user  staff  256 Oct 24 09:59 package.json', timestamp },
            { id: generateId(), type: 'output', text: 'drwxr-xr-x  2 user  staff   64 Oct 24 10:01 node_modules', timestamp },
          ];
          break;
        case 'whoami':
          lines = [{ id: generateId(), type: 'output', text: 'root', timestamp }];
          break;
        case 'git status':
          lines = [
            { id: generateId(), type: 'output', text: 'On branch main', timestamp },
            { id: generateId(), type: 'output', text: 'Your branch is up to date with \'origin/main\'.', timestamp },
            { id: generateId(), type: 'output', text: ' ', timestamp },
            { id: generateId(), type: 'output', text: 'nothing to commit, working tree clean', timestamp },
          ];
          break;
        default:
          if (command === '') {
             lines = [];
          } else {
            lines = [{ id: generateId(), type: 'error', text: `zsh: command not found: ${command}`, timestamp }];
          }
      }
      resolve(lines);
    }, 200 + Math.random() * 300); // Simulate network latency
  });
};

export const simulateAIResponse = async (prompt: string): Promise<string> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // Very basic keyword matching to simulate "AI"
      const lower = prompt.toLowerCase();
      if (lower.includes('log') || lower.includes('history')) {
        resolve('git log --oneline -n 10');
      } else if (lower.includes('undo') || lower.includes('revert')) {
        resolve('git reset --soft HEAD~1');
      } else if (lower.includes('install') || lower.includes('add')) {
        resolve('npm install <package_name>');
      } else if (lower.includes('port') || lower.includes('process')) {
        resolve('lsof -i :3000');
      } else {
        resolve('echo "Hello World"');
      }
    }, 1200); // Simulate AI "thinking" time
  });
};
