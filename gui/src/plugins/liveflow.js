import store from "../store";

class LiveFlowService {
  constructor() {
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.isConnected = false;
    this.listeners = new Map();
  }

  /**
   * Connect to WebSocket endpoint
   * @param {string} token - JWT authentication token
   */
  connect(token) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return;
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const url = `${protocol}//${host}/api/auth/live-flow?token=${token}`;

    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      console.log("[LiveFlow] WebSocket connected");
      this.isConnected = true;
      this.reconnectAttempts = 0;
      this.emit("connected");
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error("[LiveFlow] Failed to parse message:", error);
      }
    };

    this.ws.onclose = (event) => {
      console.log("[LiveFlow] WebSocket closed:", event.code, event.reason);
      this.isConnected = false;
      this.emit("disconnected");
      this.attemptReconnect(token);
    };

    this.ws.onerror = (error) => {
      console.error("[LiveFlow] WebSocket error:", error);
      this.emit("error", error);
    };
  }

  /**
   * Attempt to reconnect with exponential backoff
   */
  attemptReconnect(token) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log("[LiveFlow] Max reconnect attempts reached");
      this.emit("reconnectFailed");
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    console.log(`[LiveFlow] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1})`);

    setTimeout(() => {
      this.reconnectAttempts++;
      this.connect(token);
    }, delay);
  }

  /**
   * Handle incoming WebSocket messages
   */
  handleMessage(message) {
    const { type, data } = message;

    switch (type) {
      case "flow_start":
        this.emit("flowStart", data);
        this.updateStoreFlowStart(data);
        break;
      case "flow_update":
        this.emit("flowUpdate", data);
        this.updateStoreFlowUpdate(data);
        break;
      case "flow_end":
        this.emit("flowEnd", data);
        this.updateStoreFlowEnd(data);
        break;
      case "batch_state":
        this.emit("batchState", data);
        this.updateStoreBatchState(data);
        break;
      case "error":
        this.emit("flowError", data);
        break;
      default:
        console.log("[LiveFlow] Unknown message type:", type);
    }
  }

  /**
   * Update Vuex store with flow start data
   */
  updateStoreFlowStart(data) {
    store.commit("ADD_LIVE_FLOW", data);
  }

  /**
   * Update Vuex store with flow update data
   */
  updateStoreFlowUpdate(data) {
    store.commit("UPDATE_LIVE_FLOW", data);
  }

  /**
   * Update Vuex store with flow end data
   */
  updateStoreFlowEnd(data) {
    store.commit("REMOVE_LIVE_FLOW", data.session_id);
  }

  /**
   * Update Vuex store with batch state data
   */
  updateStoreBatchState(data) {
    store.commit("SET_LIVE_FLOWS", data.sessions);
  }

  /**
   * Add event listener
   */
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  /**
   * Remove event listener
   */
  off(event, callback) {
    if (!this.listeners.has(event)) {
      return;
    }
    const callbacks = this.listeners.get(event);
    const index = callbacks.indexOf(callback);
    if (index > -1) {
      callbacks.splice(index, 1);
    }
  }

  /**
   * Emit event to listeners
   */
  emit(event, data) {
    if (!this.listeners.has(event)) {
      return;
    }
    this.listeners.get(event).forEach((callback) => {
      try {
        callback(data);
      } catch (error) {
        console.error(`[LiveFlow] Error in ${event} listener:`, error);
      }
    });
  }

  /**
   * Disconnect from WebSocket
   */
  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
    this.listeners.clear();
  }

  /**
   * Get connection status
   */
  getConnectionStatus() {
    return this.isConnected;
  }
}

// Export singleton instance
export default new LiveFlowService();