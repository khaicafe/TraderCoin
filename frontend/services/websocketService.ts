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

/**
 * WebSocketService - Quản lý kết nối WebSocket real-time
 *
 * Service này cung cấp kết nối WebSocket persistent đến backend server,
 * tự động reconnect khi mất kết nối, và quản lý các message handlers.
 */
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

  /**
   * generateSessionId - Tạo ID duy nhất cho session
   *
   * @returns {string} Session ID dạng: "session_1702912345_abc123"
   *
   * Công dụng:
   * - Phân biệt các tab/window khác nhau của cùng 1 user
   * - Backend dùng để track từng connection riêng biệt
   */
  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substring(7)}`;
  }

  /**
   * connect - Kết nối đến WebSocket server
   *
   * Workflow:
   * 1. Kiểm tra nếu đã kết nối → return
   * 2. Lấy token từ localStorage
   * 3. Tạo WebSocket connection với URL có token + session_id
   * 4. Setup các event handlers (onopen, onmessage, onerror, onclose)
   * 5. Gửi auth message sau khi connected
   *
   * Auto-reconnect:
   * - Nếu connection bị đóng (không phải intentional) → tự động reconnect
   * - Sử dụng exponential backoff (1s, 2s, 4s, 8s, ...)
   * - Tối đa 10 lần thử
   *
   * @example
   * ```typescript
   * websocketService.connect();
   * // WebSocket sẽ tự động kết nối đến ws://localhost:8080/api/v1/trading/ws
   * ```
   */
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

  /**
   * scheduleReconnect - Lên lịch reconnect với exponential backoff
   *
   * Chiến lược reconnect:
   * - Lần 1: 1 giây
   * - Lần 2: 2 giây
   * - Lần 3: 4 giây
   * - Lần 4: 8 giây
   * - ...
   * - Tối đa: 30 giây
   * - Giới hạn: 10 lần thử
   *
   * @example
   * ```typescript
   * // Tự động được gọi khi connection bị đóng
   * // Reconnecting in 2000ms (attempt 2/10)
   * // Reconnecting in 4000ms (attempt 3/10)
   * ```
   */
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

  /**
   * disconnect - Ngắt kết nối WebSocket
   *
   * Sử dụng khi:
   * - User logout
   * - Component unmount
   * - Muốn dừng nhận updates
   *
   * Thực hiện:
   * 1. Set flag isIntentionallyClosed = true (để không auto-reconnect)
   * 2. Clear reconnect timeout
   * 3. Đóng WebSocket connection
   * 4. Reset reconnect attempts counter
   *
   * @example
   * ```typescript
   * // Khi user logout
   * websocketService.disconnect();
   * ```
   */
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

  /**
   * send - Gửi message đến WebSocket server
   *
   * @param {WebSocketMessage} message - Message cần gửi với format:
   *   {
   *     type: "message_type",
   *     data: {...}
   *   }
   *
   * Kiểm tra:
   * - Chỉ gửi khi WebSocket đang ở trạng thái OPEN
   * - Nếu không connected → log warning và không gửi
   *
   * @example
   * ```typescript
   * websocketService.send({
   *   type: 'subscribe',
   *   data: { symbol: 'BTCUSDT' }
   * });
   * ```
   */
  send(message: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }

  /**
   * onMessage - Subscribe vào tất cả WebSocket messages
   *
   * @param {MessageHandler} handler - Function xử lý message nhận được
   * @returns {Function} Unsubscribe function để cleanup
   *
   * Cách hoạt động:
   * 1. Thêm handler vào Set messageHandlers
   * 2. Mỗi khi nhận message → gọi tất cả handlers
   * 3. Return function để remove handler khi không cần nữa
   *
   * Sử dụng:
   * - Subscribe vào tất cả message types
   * - Tự filter theo message.type trong handler
   *
   * @example
   * ```typescript
   * const unsubscribe = websocketService.onMessage((message) => {
   *   if (message.type === 'order_update') {
   *     console.log('Order updated:', message.data);
   *   }
   * });
   *
   * // Cleanup khi component unmount
   * return () => {
   *   unsubscribe();
   * };
   * ```
   */
  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler);

    // Return unsubscribe function
    return () => {
      this.messageHandlers.delete(handler);
    };
  }

  /**
   * onOrderUpdate - Subscribe chỉ vào order updates
   *
   * @param {Function} callback - Function nhận OrderUpdate data
   * @returns {Function} Unsubscribe function
   *
   * Tiện ích:
   * - Không cần check message.type
   * - Tự động extract message.data với type safety
   * - Chỉ trigger khi có order_update
   *
   * Backend gửi khi:
   * - Order status thay đổi (new → filled)
   * - Order quantity thay đổi (partially filled)
   * - Order bị cancelled/expired
   *
   * @example
   * ```typescript
   * const unsubscribe = websocketService.onOrderUpdate((update) => {
   *   console.log('Order:', update.order_id);
   *   console.log('Status:', update.status);
   *   console.log('Filled:', update.executed_qty);
   *
   *   // Update UI
   *   refreshOrders();
   * });
   * ```
   */
  onOrderUpdate(callback: (update: OrderUpdate) => void): () => void {
    const handler: MessageHandler = (message) => {
      if (message.type === 'order_update') {
        callback(message.data as OrderUpdate);
      }
    };

    return this.onMessage(handler);
  }

  /**
   * onPriceUpdate - Subscribe chỉ vào price updates
   *
   * @param {Function} callback - Function nhận PriceUpdate data
   * @returns {Function} Unsubscribe function
   *
   * Tiện ích:
   * - Tự động extract price data với type safety
   * - Chỉ trigger khi có price_update
   *
   * Backend gửi khi:
   * - Giá symbol thay đổi
   * - Cập nhật real-time từ exchange
   *
   * @example
   * ```typescript
   * const unsubscribe = websocketService.onPriceUpdate((update) => {
   *   console.log('Symbol:', update.symbol);
   *   console.log('Price:', update.price);
   *   console.log('Change:', update.price_change);
   *   console.log('Percent:', update.price_percent);
   *
   *   // Update price display
   *   updatePriceUI(update);
   * });
   * ```
   */
  onPriceUpdate(callback: (update: PriceUpdate) => void): () => void {
    const handler: MessageHandler = (message) => {
      if (message.type === 'price_update') {
        callback(message.data as PriceUpdate);
      }
    };

    return this.onMessage(handler);
  }

  /**
   * isConnected - Kiểm tra WebSocket có đang connected không
   *
   * @returns {boolean} true nếu connected, false nếu không
   *
   * Sử dụng:
   * - Hiển thị connection status indicator
   * - Disable/enable features based on connection
   * - Debug connection issues
   *
   * @example
   * ```typescript
   * if (websocketService.isConnected()) {
   *   console.log('✅ Connected to server');
   * } else {
   *   console.log('❌ Not connected');
   * }
   *
   * // Trong UI
   * const status = websocketService.isConnected() ? 'online' : 'offline';
   * ```
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  /**
   * getConnectionState - Lấy trạng thái chi tiết của WebSocket
   *
   * @returns {string} Một trong các giá trị:
   *   - 'CONNECTING' - Đang kết nối
   *   - 'CONNECTED' - Đã kết nối
   *   - 'CLOSING' - Đang đóng
   *   - 'DISCONNECTED' - Đã ngắt kết nối
   *   - 'UNKNOWN' - Trạng thái không xác định
   *
   * Sử dụng:
   * - Hiển thị status badge với màu sắc
   * - Debug connection lifecycle
   * - Monitoring và logging
   *
   * @example
   * ```typescript
   * const state = websocketService.getConnectionState();
   * console.log('WebSocket state:', state);
   *
   * // Trong UI
   * const badgeColor = {
   *   'CONNECTING': 'yellow',
   *   'CONNECTED': 'green',
   *   'CLOSING': 'orange',
   *   'DISCONNECTED': 'red'
   * }[state];
   * ```
   */
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
