import store from "../store";

class LiveFlowService {
  constructor() {
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 10;
    this.reconnectDelay = 1000;
    this.isConnected = false;
    this.listeners = new Map();
    this._token = null;
    this._lastMessageTime = 0;
    this._messageCount = 0;
  }

  /**
   * Build the WebSocket URL, matching App.vue's pattern for custom backends.
   */
  _buildUrl(token) {
    // apiRoot is a global defined by Vite: `${localStorage["backendAddress"]}/api`
    let baseUrl = typeof apiRoot !== "undefined" ? apiRoot : "/api";

    if (!baseUrl.trim() || baseUrl.startsWith("/")) {
      baseUrl = location.protocol + "//" + location.host + baseUrl;
    }

    try {
      const u = new URL(baseUrl);
      const protocol = u.protocol === "https:" ? "wss:" : "ws:";
      const basePath = u.pathname.replace(/\/api\/?$/, "");
      return `${protocol}//${u.host}${basePath}/api/live-flow?Authorization=${encodeURIComponent(token)}`;
    } catch (e) {
      // Fallback to simple URL construction
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const host = window.location.host;
      return `${protocol}//${host}/api/live-flow?Authorization=${encodeURIComponent(token)}`;
    }
  }

  /**
   * Connect to WebSocket endpoint
   * @param {string} token - JWT authentication token
   */
  connect(token) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log("[LiveFlow] Already connected, skipping");
      return;
    }

    this._token = token;
    const url = this._buildUrl(token);
    console.log(
      "[LiveFlow] Connecting to:",
      url.replace(/([?&](?:Authorization|token)=)[^&]*/, "$1***")
    );

    try {
      this.ws = new WebSocket(url);
    } catch (err) {
      console.error("[LiveFlow] Failed to create WebSocket:", err);
      this.emit("error", err);
      return;
    }

    this.ws.onopen = () => {
      console.log("[LiveFlow] WebSocket connected successfully");
      this.isConnected = true;
      this.reconnectAttempts = 0;
      this.emit("connected");
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this._lastMessageTime = Date.now();
        this._messageCount++;
        this.handleMessage(message);
      } catch (error) {
        console.error("[LiveFlow] Failed to parse message:", error, event.data);
      }
    };

    this.ws.onclose = (event) => {
      console.log("[LiveFlow] WebSocket closed, code:", event.code, "reason:", event.reason);
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
      console.warn("[LiveFlow] Max reconnect attempts reached");
      this.emit("reconnectFailed");
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    console.log(
      `[LiveFlow] Reconnecting in ${delay}ms (attempt ${
        this.reconnectAttempts + 1
      }/${this.maxReconnectAttempts})`
    );

    setTimeout(() => {
      this.reconnectAttempts++;
      this.connect(token);
    }, delay);
  }

  /**
   * Handle incoming WebSocket messages
   */
  handleMessage(message) {
    const { type, body, data } = message;
    const payload = body || data;

    console.log("[LiveFlow] Received message type:", type, "payload:", payload);

    switch (type) {
      case "flow_start":
        this.emit("flowStart", payload);
        this.updateStoreFlowStart(payload);
        break;
      case "flow_update":
        this.emit("flowUpdate", payload);
        this.updateStoreFlowUpdate(payload);
        break;
      case "flow_end":
        this.emit("flowEnd", payload);
        this.updateStoreFlowEnd(payload);
        break;
      case "batch_state":
        this.emit("batchState", payload);
        this.updateStoreBatchState(payload);
        break;
      case "observatory":
        this.emit("observatory", payload);
        this.updateStoreObservatory(message.produce_time, payload);
        break;
      case "error":
        this.emit("flowError", payload);
        break;
      default:
        console.warn("[LiveFlow] Unknown message type:", type, message);
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
   * Update Vuex store with observatory outbound status data.
   * Transforms each outboundStatus entry into a flow-compatible card
   * so the LiveFlowDashboard can display them with animations.
   * Collects warnings for any malformed entries.
   */
  updateStoreObservatory(produceTime, body) {
    const warnings = [];

    if (!body || typeof body !== "object") {
      warnings.push({
        time: Date.now(),
        message: "Observatory message body is missing or not an object",
        raw: body,
      });
      console.warn("[LiveFlow] Observatory body invalid:", body);
      store.commit("SET_LIVE_FLOW_WARNINGS", warnings);
      return;
    }

    if (!Array.isArray(body.outboundStatus)) {
      warnings.push({
        time: Date.now(),
        message: `outboundStatus is not an array (got ${typeof body.outboundStatus})`,
        raw: body.outboundStatus,
      });
      console.warn(
        "[LiveFlow] outboundStatus is not an array:",
        body.outboundStatus
      );
      store.commit("SET_LIVE_FLOW_WARNINGS", warnings);
      return;
    }

    const outboundName = body.outboundName || "unknown";

    const flows = body.outboundStatus
      .map((status, index) => {
        if (!status || typeof status !== "object") {
          warnings.push({
            time: Date.now(),
            message: `outboundStatus[${index}] is not an object`,
            raw: status,
          });
          console.warn(
            `[LiveFlow] outboundStatus[${index}] invalid:`,
            status
          );
          return null;
        }

        if (typeof status.outbound_tag !== "string" || !status.outbound_tag) {
          warnings.push({
            time: Date.now(),
            message: `outboundStatus[${index}] missing or invalid outbound_tag`,
            raw: status,
          });
          console.warn(
            `[LiveFlow] outboundStatus[${index}] missing outbound_tag:`,
            status
          );
        }

        if (typeof status.alive !== "boolean") {
          warnings.push({
            time: Date.now(),
            message: `outboundStatus[${index}] missing or invalid "alive" field (got ${typeof status.alive})`,
            raw: status,
          });
          console.warn(
            `[LiveFlow] outboundStatus[${index}] invalid alive:`,
            status.alive
          );
        }

        if (typeof status.delay !== "number") {
          warnings.push({
            time: Date.now(),
            message: `outboundStatus[${index}] missing or invalid "delay" field (got ${typeof status.delay})`,
            raw: status,
          });
          console.warn(
            `[LiveFlow] outboundStatus[${index}] invalid delay:`,
            status.delay
          );
        }

        const sessionId = `obs-${outboundName}-${index}`;

        return {
          session_id: sessionId,
          status: status.alive === true ? "active" : "error",
          source: {
            ip: outboundName,
            port: "-",
          },
          destination: {
            ip: status.outbound_tag || "unknown",
            port: "-",
            domain: status.outbound_tag || "unknown",
          },
          protocol: "proxy",
          bytes_sent: 0,
          bytes_recv: 0,
          speed_bps: typeof status.delay === "number" ? status.delay : 0,
          last_activity: produceTime ? produceTime * 1000 : Date.now(),
          start_time: produceTime ? produceTime * 1000 : Date.now(),
        };
      })
      .filter(Boolean);

    console.log(
      `[LiveFlow] Observatory update: ${flows.length} flows committed, ${warnings.length} warnings`
    );

    store.commit("SET_LIVE_FLOWS", flows);
    store.commit("SET_LIVE_FLOW_WARNINGS", warnings);

    if (warnings.length > 0) {
      console.warn(
        `[LiveFlow] Observatory processing completed with ${warnings.length} warning(s)`,
        warnings
      );
    }
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
    console.log("[LiveFlow] Disconnecting...");
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

  /**
   * Get debug info
   */
  getDebugInfo() {
    return {
      isConnected: this.isConnected,
      reconnectAttempts: this.reconnectAttempts,
      lastMessageTime: this._lastMessageTime,
      messageCount: this._messageCount,
    };
  }
}

// Export singleton instance
export default new LiveFlowService();
