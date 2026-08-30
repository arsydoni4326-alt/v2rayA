import store from "../store";

function isObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function normalizeEndpoint(value) {
  if (!isObject(value) || typeof value.ip !== "string" || !value.ip) {
    return null;
  }

  return {
    ip: value.ip,
    port: Number.isInteger(value.port) ? value.port : 0,
    domain: typeof value.domain === "string" ? value.domain : "",
  };
}

function normalizeProxyChain(value) {
  if (!Array.isArray(value)) {
    return [];
  }

  return value
    .filter(
      (proxy) =>
        isObject(proxy) &&
        typeof proxy.proxy_id === "string" &&
        typeof proxy.name === "string"
    )
    .map((proxy) => ({
      proxy_id: proxy.proxy_id,
      name: proxy.name,
      type: typeof proxy.type === "string" ? proxy.type : "unknown",
      server: typeof proxy.server === "string" ? proxy.server : "",
    }));
}

function normalizeFlowStart(data) {
  if (
    !isObject(data) ||
    typeof data.session_id !== "string" ||
    !data.session_id
  ) {
    return null;
  }

  const source = normalizeEndpoint(data.source);
  const destination = normalizeEndpoint(data.destination);
  if (!source || !destination || typeof data.protocol !== "string") {
    return null;
  }

  return {
    session_id: data.session_id,
    source,
    proxy_chain: normalizeProxyChain(data.proxy_chain),
    destination,
    protocol: data.protocol || "tcp",
    start_time: data.start_time || new Date().toISOString(),
    bytes_sent: 0,
    bytes_recv: 0,
    speed_bps: 0,
    last_activity: data.start_time || new Date().toISOString(),
    status: "active",
  };
}

class LiveFlowService {
  constructor() {
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 10;
    this.reconnectDelay = 1000;
    this.isConnected = false;
    this.shouldReconnect = false;
    this.reconnectTimer = null;
    this.lastMessageTime = 0;
    this.messageCount = 0;
  }

  buildUrl(token) {
    let baseUrl = typeof apiRoot !== "undefined" ? apiRoot : "/api";

    if (!baseUrl.trim() || baseUrl.startsWith("/")) {
      baseUrl = `${window.location.protocol}//${window.location.host}${baseUrl}`;
    }

    try {
      const url = new URL(baseUrl);
      const protocol = url.protocol === "https:" ? "wss:" : "ws:";
      const basePath = url.pathname.replace(/\/api\/?$/, "");
      return `${protocol}//${
        url.host
      }${basePath}/api/live-flow?token=${encodeURIComponent(token)}`;
    } catch (error) {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      return `${protocol}//${
        window.location.host
      }/api/live-flow?token=${encodeURIComponent(token)}`;
    }
  }

  connect(token) {
    if (!token) {
      return;
    }
    if (
      this.ws &&
      (this.ws.readyState === WebSocket.OPEN ||
        this.ws.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    this.shouldReconnect = true;
    this.ws = new WebSocket(this.buildUrl(token));

    this.ws.onopen = () => {
      this.isConnected = true;
      this.reconnectAttempts = 0;
      store.commit("SET_LIVE_FLOW_CONNECTED", true);
    };

    this.ws.onmessage = (event) => {
      try {
        this.lastMessageTime = Date.now();
        this.messageCount += 1;
        this.handleMessage(JSON.parse(event.data));
      } catch (error) {
        console.warn("[LiveFlow] Ignored malformed WebSocket message", error);
      }
    };

    this.ws.onclose = () => {
      this.ws = null;
      this.isConnected = false;
      store.commit("SET_LIVE_FLOW_CONNECTED", false);
      this.scheduleReconnect(token);
    };
  }

  scheduleReconnect(token) {
    if (
      !this.shouldReconnect ||
      this.reconnectAttempts >= this.maxReconnectAttempts
    ) {
      return;
    }

    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts),
      30000
    );
    this.reconnectAttempts += 1;
    clearTimeout(this.reconnectTimer);
    this.reconnectTimer = setTimeout(() => this.connect(token), delay);
  }

  handleMessage(message) {
    if (!isObject(message) || typeof message.type !== "string") {
      return;
    }

    switch (message.type) {
      case "flow_start": {
        const flow = normalizeFlowStart(message.data);
        if (flow) {
          store.commit("ADD_LIVE_FLOW", flow);
        }
        break;
      }
      case "flow_update":
        if (
          isObject(message.data) &&
          typeof message.data.session_id === "string"
        ) {
          store.commit("UPDATE_LIVE_FLOW", message.data);
        }
        break;
      case "flow_end":
        if (
          isObject(message.data) &&
          typeof message.data.session_id === "string"
        ) {
          store.commit("REMOVE_LIVE_FLOW", message.data.session_id);
        }
        break;
      case "batch_state": {
        const sessions =
          isObject(message.data) && Array.isArray(message.data.sessions)
            ? message.data.sessions.map(normalizeFlowStart).filter(Boolean)
            : [];
        store.commit("SET_LIVE_FLOWS", sessions);
        break;
      }
      case "error":
        if (
          isObject(message.data) &&
          typeof message.data.message === "string"
        ) {
          store.commit("SET_LIVE_FLOW_WARNINGS", [message.data]);
        }
        break;
      default:
        // Observatory messages intentionally do not represent traffic sessions.
        break;
    }
  }

  disconnect() {
    this.shouldReconnect = false;
    clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
    store.commit("SET_LIVE_FLOW_CONNECTED", false);
  }

  getDebugInfo() {
    return {
      isConnected: this.isConnected,
      reconnectAttempts: this.reconnectAttempts,
      lastMessageTime: this.lastMessageTime,
      messageCount: this.messageCount,
    };
  }
}

export default new LiveFlowService();
