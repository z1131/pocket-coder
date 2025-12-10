import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { API_BASE } from '../api/client';
import { ConnectionStatus, PocketEvent } from '../types';

interface UsePocketWSOptions {
  token?: string;
  autoConnect?: boolean;
  onEvent?: (event: PocketEvent) => void;
}

interface SendMessageParams {
  desktopId: number;
  content: string;
  sessionId?: number;
  messageId?: string;
}

const WS_BASE = import.meta.env.VITE_WS_BASE || API_BASE.replace(/^http/, 'ws');

function buildWsUrl(token: string) {
  const base = WS_BASE.endsWith('/') ? WS_BASE.slice(0, -1) : WS_BASE;
  return `${base}/ws/mobile?token=${encodeURIComponent(token)}`;
}

export function usePocketWS(options: UsePocketWSOptions) {
  const { token, autoConnect = true, onEvent } = options;
  const [status, setStatus] = useState(ConnectionStatus.DISCONNECTED);
  const [presence, setPresence] = useState<Record<number, 'online' | 'offline'>>({});
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<number | null>(null);
  const retryCount = useRef(0);
  const connectRef = useRef<() => void>(() => { });

  const clearReconnectTimer = useCallback(() => {
    if (reconnectTimer.current) {
      window.clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
  }, []);

  const disconnect = useCallback(() => {
    clearReconnectTimer();
    setStatus(ConnectionStatus.DISCONNECTED);
    if (socketRef.current) {
      socketRef.current.onopen = null;
      socketRef.current.onclose = null;
      socketRef.current.onmessage = null;
      socketRef.current.onerror = null;
      socketRef.current.close();
    }
    socketRef.current = null;
  }, [clearReconnectTimer]);

  const scheduleReconnect = useCallback(() => {
    if (!autoConnect || !token) return;
    clearReconnectTimer();
    const delay = Math.min(30000, 1000 * Math.pow(2, retryCount.current));
    reconnectTimer.current = window.setTimeout(() => {
      retryCount.current += 1;
      connectRef.current();
    }, delay);
  }, [autoConnect, clearReconnectTimer, token]);

  const handleIncoming = useCallback((msg: any) => {
    const { type, payload } = msg;
    let event: PocketEvent | null = null;

    console.log('[PocketWS] Received message:', type, payload);

    switch (type) {
      case 'terminal:output':
        if (payload?.data) {
          // Decode base64 data
          try {
            const binaryString = atob(payload.data);
            const bytes = new Uint8Array(binaryString.length);
            for (let i = 0; i < binaryString.length; i++) {
              bytes[i] = binaryString.charCodeAt(i);
            }
            const decoded = new TextDecoder().decode(bytes);
            console.log('[PocketWS] Terminal output decoded:', decoded);
            event = { kind: 'terminal:output', data: decoded };
          } catch (e) {
            console.error('Failed to decode terminal output:', e);
            event = { kind: 'terminal:output', data: payload.data };
          }
        }
        break;
      case 'terminal:history':
        if (payload?.data) {
          // Decode base64 history data
          try {
            const binaryString = atob(payload.data);
            const bytes = new Uint8Array(binaryString.length);
            for (let i = 0; i < binaryString.length; i++) {
              bytes[i] = binaryString.charCodeAt(i);
            }
            const decoded = new TextDecoder().decode(bytes);
            console.log('[PocketWS] Terminal history decoded, length:', decoded.length);
            event = { kind: 'terminal:history', data: decoded };
          } catch (e) {
            console.error('Failed to decode terminal history:', e);
            event = { kind: 'terminal:history', data: payload.data };
          }
        }
        break;
      case 'terminal:exit':
        event = { kind: 'terminal:exit', code: payload?.code || 0 };
        break;
      case 'desktop:online':
        if (payload?.desktop_id) {
          setPresence((prev) => ({ ...prev, [payload.desktop_id]: 'online' }));
          event = { kind: 'desktop-status', desktopId: payload.desktop_id, status: 'online' };
        }
        break;
      case 'desktop:offline':
        if (payload?.desktop_id) {
          setPresence((prev) => ({ ...prev, [payload.desktop_id]: 'offline' }));
          event = { kind: 'desktop-status', desktopId: payload.desktop_id, status: 'offline' };
        }
        break;
      case 'session:create':
        if (payload?.session_id) {
          event = { kind: 'session-create', sessionId: payload.session_id, workingDir: payload.working_dir };
        }
        break;
      case 'agent:stream':
        if (payload?.session_id != null && payload?.delta != null) {
          event = { kind: 'agent-stream', sessionId: payload.session_id, delta: payload.delta };
        }
        break;
      case 'agent:response':
        if (payload?.session_id != null && payload?.content != null) {
          event = {
            kind: 'agent-response',
            sessionId: payload.session_id,
            content: payload.content,
            role: payload.role || 'assistant',
          };
        }
        break;
      case 'agent:status':
        if (payload?.status) {
          event = { kind: 'agent-status', status: payload.status, sessionId: payload.session_id };
          if (payload.status === 'running') {
            setStatus(ConnectionStatus.BUSY);
          } else if (payload.status === 'idle') {
            setStatus(ConnectionStatus.CONNECTED);
          }
        }
        break;
      case 'error':
        if (payload?.code) {
          event = { kind: 'error', code: payload.code, message: payload.message || '未知错误' };
        }
        break;
      case 'pong':
        event = { kind: 'pong' };
        break;
      default:
        break;
    }

    if (event && onEvent) {
      onEvent(event);
    }
  }, [onEvent]);

  const connect = useCallback(() => {
    connectRef.current = connect;
    if (!token) return;

    // Prevent duplicate connection if we already have one for this token
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      console.log('[PocketWS] Already connected, skipping connect');
      return;
    }

    disconnect();
    console.log('[PocketWS] Connecting with token:', token.slice(0, 8) + '...');
    setStatus(ConnectionStatus.CONNECTING);
    const wsUrl = buildWsUrl(token);
    const ws = new WebSocket(wsUrl);
    socketRef.current = ws;

    ws.onopen = () => {
      console.log('[PocketWS] Connected');
      retryCount.current = 0;
      setStatus(ConnectionStatus.CONNECTED);
    };

    ws.onclose = (e) => {
      console.log('[PocketWS] Disconnected', e.code, e.reason);
      setStatus(ConnectionStatus.DISCONNECTED);
      if (e.code !== 1000) {
        scheduleReconnect();
      }
    };

    ws.onerror = (e) => {
      console.error('[PocketWS] Error', e);
      ws.close();
    };

    ws.onmessage = (event) => {
      // 服务器可能会把多条消息用换行符合并发送
      const messages = event.data.split('\n').filter((s: string) => s.trim());
      for (const msgStr of messages) {
        try {
          const msg = JSON.parse(msgStr);
          handleIncoming(msg);
        } catch (err) {
          console.error('无法解析 WebSocket 消息', msgStr, err);
        }
      }
    };
  }, [disconnect, scheduleReconnect, token, handleIncoming]);

  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  const sendUserMessage = useCallback((params: SendMessageParams) => {
    const ws = socketRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket 未连接');
    }
    const messageId = params.messageId || (crypto.randomUUID ? crypto.randomUUID() : String(Date.now()));
    const payload = {
      type: 'user:message',
      payload: {
        desktop_id: params.desktopId,
        session_id: params.sessionId,
        content: params.content,
      },
      timestamp: Date.now(),
      message_id: messageId,
    };
    ws.send(JSON.stringify(payload));
    return messageId;
  }, []);

  const sendTerminalInput = useCallback((desktopId: number, data: string) => {
    const ws = socketRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    // 使用 base64 编码数据，与电脑端保持一致
    const encoded = btoa(data);

    const payload = {
      type: 'terminal:input',
      payload: {
        desktop_id: desktopId,
        data: encoded
      },
      timestamp: Date.now()
    };
    console.log('[PocketWS] Sending terminal input:', data, '-> base64:', encoded);
    ws.send(JSON.stringify(payload));
  }, []);

  const sendTerminalResize = useCallback((desktopId: number, cols: number, rows: number) => {
    const ws = socketRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    const payload = {
      type: 'terminal:resize',
      payload: {
        desktop_id: desktopId,
        cols: cols,
        rows: rows
      },
      timestamp: Date.now()
    };
    ws.send(JSON.stringify(payload));
  }, []);

  const requestTerminalHistory = useCallback((desktopId: number) => {
    const ws = socketRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    const payload = {
      type: 'terminal:history',
      payload: {
        desktop_id: desktopId
      },
      timestamp: Date.now()
    };
    console.log('[PocketWS] Requesting terminal history for desktop:', desktopId);
    ws.send(JSON.stringify(payload));
  }, []);



  useEffect(() => {
    if (autoConnect && token) {
      connect();
    }
    return () => {
      disconnect();
    };
  }, [autoConnect, connect, disconnect, token]);

  const presenceList = useMemo(() => presence, [presence]);

  return {
    status,
    presence: presenceList,
    connect,
    disconnect,
    sendUserMessage,
    sendTerminalInput,
    sendTerminalResize,
    requestTerminalHistory,
  };
}
