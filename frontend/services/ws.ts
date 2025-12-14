type MessageHandler = (payload: any) => void;

// 诊断用：给每个连接分配唯一 ID
let wsInstanceCounter = 0;

class WSClient {
  private ws: WebSocket | null = null;
  private currentWsId: number = 0;  // 当前连接的 ID
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectTimer: any = null;
  private heartbeatTimer: any = null;
  private url: string = '';

  connect(token: string) {
    console.log(`[WS DEBUG] connect() called, current wsId=${this.currentWsId}, ws.readyState=${this.ws?.readyState}`);
    console.log(`[WS DEBUG] Call stack:`, new Error().stack);

    // Construct WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host; // Use current host (proxied by Vite)
    this.url = `${protocol}//${host}/ws/mobile?token=${token}`;

    this.initWs();
  }

  private initWs() {
    const newWsId = ++wsInstanceCounter;
    console.log(`[WS DEBUG] initWs() called, creating wsId=${newWsId}, old wsId=${this.currentWsId}, old ws.readyState=${this.ws?.readyState}`);

    if (this.ws) {
      console.log(`[WS DEBUG] Closing old connection wsId=${this.currentWsId}`);
      this.ws.close();
    }

    this.currentWsId = newWsId;
    const capturedWsId = newWsId;  // 闭包捕获，用于 onclose 判断

    this.ws = new WebSocket(this.url);
    console.log(`[WS DEBUG] New WebSocket created, wsId=${capturedWsId}`);

    this.ws.onopen = () => {
      console.log(`[WS DEBUG] onopen fired, wsId=${capturedWsId}, currentWsId=${this.currentWsId}, readyState=${this.ws?.readyState}`);
      if (capturedWsId !== this.currentWsId) {
        console.warn(`[WS DEBUG] ⚠️ onopen: wsId mismatch! This connection (${capturedWsId}) is stale, current is ${this.currentWsId}`);
        return;
      }
      console.log('WS Connected');
      this.startHeartbeat();
      this.emit('open', null);
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        this.emit(msg.type, msg.payload);
      } catch (e) {
        console.error('WS Parse Error', e);
      }
    };

    this.ws.onclose = (event) => {
      console.log(`[WS DEBUG] onclose fired, wsId=${capturedWsId}, currentWsId=${this.currentWsId}, code=${event.code}, reason="${event.reason}"`);

      // 关键检查：如果这是旧连接的 onclose，不要清空 this.ws
      if (capturedWsId !== this.currentWsId) {
        console.warn(`[WS DEBUG] ⚠️ onclose: wsId mismatch! This connection (${capturedWsId}) is stale, current is ${this.currentWsId}. NOT clearing this.ws`);
        return;  // 不执行任何清理操作
      }

      console.log('WS Closed');
      this.stopHeartbeat();
      this.ws = null;
      this.emit('close', null);

      // Auto reconnect
      if (!this.reconnectTimer) {
        console.log(`[WS DEBUG] Setting reconnect timer for wsId=${capturedWsId}`);
        this.reconnectTimer = setTimeout(() => this.initWs(), 3000);
      }
    };

    this.ws.onerror = (e) => {
      console.error(`[WS DEBUG] onerror fired, wsId=${capturedWsId}`, e);
    };
  }

  disconnect() {
    console.log(`[WS DEBUG] disconnect() called, wsId=${this.currentWsId}, ws.readyState=${this.ws?.readyState}`);
    if (this.reconnectTimer) {
      console.log(`[WS DEBUG] Clearing reconnect timer`);
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.onclose = null; // Prevent reconnect
      this.ws.close();
    }
    this.ws = null;
    this.currentWsId = 0;  // 重置
  }

  send(type: string, payload: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload, timestamp: Date.now() }));
    } else {
      console.warn('WS not ready, message dropped', type);
    }
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => this.off(type, handler);
  }

  off(type: string, handler: MessageHandler) {
    const set = this.handlers.get(type);
    if (set) {
      set.delete(handler);
    }
  }

  private emit(type: string, payload: any) {
    const handlers = this.handlers.get(type);
    if (handlers) {
      handlers.forEach(h => h(payload));
    }
  }

  private startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      this.send('heartbeat', {});
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
}

export const ws = new WSClient();
