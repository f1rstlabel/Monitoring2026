import { ref } from 'vue';

type WSMessageHandler = (data: any) => void;

export class LiveWebSocket {
  private socket: WebSocket | null = null;
  private url: string;
  private handlers: Set<WSMessageHandler> = new Set();
  public isConnected = ref(false);
  public lastUpdatedAgo = ref(0);
  private timer: any = null;

  constructor() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    this.url = `${protocol}//${host}/ws/live`;
  }

  private isIntentionalDisconnect = false;

  public connect() {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }
    this.isIntentionalDisconnect = false;

    try {
      this.socket = new WebSocket(this.url);

      this.socket.onopen = () => {
        this.isConnected.value = true;
        this.startRelativeTimer();
      };

      this.socket.onmessage = (event) => {
        this.lastUpdatedAgo.value = 0;
        try {
          const data = JSON.parse(event.data);
          this.handlers.forEach(handler => handler(data));
        } catch (e) {
          console.error('WS JSON parse error', e);
        }
      };

      this.socket.onclose = () => {
        this.isConnected.value = false;
        if (!this.isIntentionalDisconnect) {
          setTimeout(() => this.connect(), 3000);
        }
      };

      this.socket.onerror = () => {
        this.isConnected.value = false;
      };
    } catch (e) {
      this.isConnected.value = false;
    }
  }

  private startRelativeTimer() {
    if (this.timer) clearInterval(this.timer);
    this.timer = setInterval(() => {
      if (this.isConnected.value) {
        this.lastUpdatedAgo.value++;
      }
    }, 1000);
  }

  public subscribe(handler: WSMessageHandler) {
    this.handlers.add(handler);
    return () => this.handlers.delete(handler);
  }

  public disconnect() {
    this.isIntentionalDisconnect = true;
    if (this.timer) clearInterval(this.timer);
    if (this.socket) {
      this.socket.close();
    }
  }
}

export const wsClient = new LiveWebSocket();
