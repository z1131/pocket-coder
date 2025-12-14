type MessageHandler = (payload: any) => void;

class WSClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectTimer: any = null;
  private heartbeatTimer: any = null;
  private url: string = '';

  connect(token: string) {
    // Construct WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host; // Use current host (proxied by Vite)
    this.url = `${protocol}//${host}/ws/mobile?token=${token}`;

    this.initWs();
  }

  private initWs() {
    if (this.ws) {
      this.ws.close();
    }

    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
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

    this.ws.onclose = () => {
      console.log('WS Closed');
      this.stopHeartbeat();
      this.ws = null;
      this.emit('close', null);
      
      // Auto reconnect
      if (!this.reconnectTimer) {
        this.reconnectTimer = setTimeout(() => this.initWs(), 3000);
      }
    };

    this.ws.onerror = (e) => {
      console.error('WS Error', e);
    };
  }

  disconnect() {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    if (this.ws) {
      this.ws.onclose = null; // Prevent reconnect
      this.ws.close();
    }
    this.ws = null;
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
