// WebSocket Service for real-time order updates
type WebSocketMessage = {
  type: string;
  data: any;
};

type OrderUpdate = {
  user_id: number;
  exchange_key_id: number;
  exchange: string;
  trading_mode: string;
  order_id: string;
  client_order_id: string;
  symbol: string;
  side: string;
  type: string;
  status: string;
  price: number;
  quantity: number;
  executed_qty: number;
  executed_price: number;
  current_price: number;
  update_time: number;
};

type PriceUpdate = {
  symbol: string;
  price: number;
  price_change: number;
  price_percent: number;
  update_time: number;
};

type MessageHandler = (message: WebSocketMessage) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private messageHandlers: Set<MessageHandler> = new Set();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000; // Start with 1 second
  private isIntentionallyClosed = false;
  private sessionId: string;

  constructor() {
    this.sessionId = this.generateSessionId();
  }

  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substring(7)}`;
  }

  connect(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected');
      return;
    }

    this.isIntentionallyClosed = false;
    const token = localStorage.getItem('token'); // Changed from 'access_token' to 'token'

    if (!token) {
      console.error('No access token found');
      return;
    }

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost =
      process.env.NEXT_PUBLIC_API_URL?.replace('http://', '')
        .replace('https://', '')
        .replace('/api/v1', '') || 'localhost:8080';
    // Add token as query parameter for authentication
    const wsUrl = `${wsProtocol}//${wsHost}/api/v1/trading/ws?token=${encodeURIComponent(
      token,
    )}&session_id=${this.sessionId}`;

    console.log('Connecting to WebSocket:', wsUrl.replace(token, 'HIDDEN'));

    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;

        // Send auth token
        this.send({
          type: 'auth',
          data: {token},
        });
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          console.log('WebSocket message received:', message);

          // Notify all handlers
          this.messageHandlers.forEach((handler) => {
            try {
              handler(message);
            } catch (error) {
              console.error('Error in message handler:', error);
            }
          });
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.ws = null;

        if (!this.isIntentionallyClosed) {
          this.scheduleReconnect();
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached');
      return;
    }

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      30000,
    );

    console.log(
      `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`,
    );

    this.reconnectTimeout = setTimeout(() => {
      this.connect();
    }, delay);
  }

  disconnect(): void {
    this.isIntentionallyClosed = true;

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.reconnectAttempts = 0;
  }

  send(message: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }

  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler);

    // Return unsubscribe function
    return () => {
      this.messageHandlers.delete(handler);
    };
  }

  onOrderUpdate(callback: (update: OrderUpdate) => void): () => void {
    const handler: MessageHandler = (message) => {
      if (message.type === 'order_update') {
        callback(message.data as OrderUpdate);
      }
    };

    return this.onMessage(handler);
  }

  onPriceUpdate(callback: (update: PriceUpdate) => void): () => void {
    const handler: MessageHandler = (message) => {
      if (message.type === 'price_update') {
        callback(message.data as PriceUpdate);
      }
    };

    return this.onMessage(handler);
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  getConnectionState(): string {
    if (!this.ws) return 'DISCONNECTED';

    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'CONNECTING';
      case WebSocket.OPEN:
        return 'CONNECTED';
      case WebSocket.CLOSING:
        return 'CLOSING';
      case WebSocket.CLOSED:
        return 'DISCONNECTED';
      default:
        return 'UNKNOWN';
    }
  }
}

// Singleton instance
const websocketService = new WebSocketService();

export default websocketService;
export type {OrderUpdate, PriceUpdate, WebSocketMessage};
