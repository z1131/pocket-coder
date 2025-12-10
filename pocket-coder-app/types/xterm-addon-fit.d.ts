declare module 'xterm-addon-fit' {
    import { Terminal, ITerminalAddon } from 'xterm';
    export class FitAddon implements ITerminalAddon {
        constructor();
        activate(terminal: Terminal): void;
        dispose(): void;
        fit(): void;
        proposeDimensions(): { cols: number; rows: number } | undefined;
    }
}