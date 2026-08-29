<template>
  <div class="live-flow-dashboard">
    <div class="live-flow-header">
      <h2>Live Flow Visualization</h2>
      <div class="connection-status">
        <span :class="['status-indicator', connectionStatus]"></span>
        {{ connectionStatusText }}
      </div>
    </div>

    <!-- Debug Info -->
    <div class="live-flow-debug" v-if="debugInfo">
      <span class="debug-item">Flows: {{ safeFlows.length }}</span>
      <span class="debug-item">Messages: {{ debugInfo.messageCount }}</span>
      <span class="debug-item">Reconnects: {{ debugInfo.reconnectAttempts }}</span>
      <span class="debug-item">Last msg: {{ debugInfo.lastMessageTime ? formatTime(debugInfo.lastMessageTime) : 'N/A' }}</span>
    </div>

    <!-- Warnings Banner -->
    <div v-if="liveFlowWarnings.length > 0" class="live-flow-warnings">
      <div class="warnings-header">
        <span class="warnings-icon">⚠</span>
        <span>{{ liveFlowWarnings.length }} warning(s) received from backend</span>
        <button @click="dismissWarnings" class="warnings-dismiss">Dismiss</button>
      </div>
      <div class="warnings-list">
        <div v-for="(warning, idx) in liveFlowWarnings" :key="idx" class="warning-item">
          <span class="warning-message">{{ warning.message }}</span>
          <span v-if="warning.raw" class="warning-raw" @click="toggleRaw(idx)">
            {{ expandedWarnings[idx] ? 'Hide raw' : 'Show raw' }}
          </span>
          <pre v-if="expandedWarnings[idx] && warning.raw" class="warning-raw-data">{{ JSON.stringify(warning.raw, null, 2) }}</pre>
        </div>
      </div>
    </div>

    <div class="live-flow-controls">
      <div class="filter-group">
        <label>Status Filter:</label>
        <select v-model="filters.status" class="filter-select">
          <option value="">All</option>
          <option value="active">Active</option>
          <option value="idle">Idle</option>
          <option value="error">Error</option>
        </select>
      </div>

      <button @click="clearFlows" class="clear-button">
        Clear All
      </button>
    </div>

    <div class="live-flow-canvas" ref="canvas">
      <div v-if="connectionStatus === 'disconnected'" class="no-flows"><!--
    --><span class="no-flows-msg">Disconnected. Attempting to reconnect...</span><!--
  --></div>

      <div v-else-if="filteredFlows.length === 0" class="no-flows">
        Waiting for live traffic data...
      </div>

      <div
        v-for="flow in filteredFlows"
        :key="flow.session_id"
        :class="['flow-card', flow.status]"
      >
        <div class="flow-header">
          <span class="flow-id">{{ flow.session_id ? flow.session_id.substring(0, 12) : '--------' }}</span>
          <span :class="['flow-status', flow.status]">{{ flow.status || 'unknown' }}</span>
        </div>

        <div class="flow-path">
          <div class="endpoint source">
            <div class="endpoint-label">Source</div>
            <div class="endpoint-value">{{ flow.source ? (flow.source.ip + ':' + flow.source.port) : 'N/A' }}</div>
          </div>

          <div class="flow-arrow">
            <span class="arrow-icon">→</span>
            <span class="protocol-badge">{{ flow.protocol || 'N/A' }}</span>
          </div>

          <div class="endpoint destination">
            <div class="endpoint-label">Destination</div>
            <div class="endpoint-value">
              {{ flow.destination ? (flow.destination.domain || flow.destination.ip || 'N/A') : 'N/A' }}<template v-if="flow.destination && flow.destination.port && flow.destination.port !== '-'">:{{ flow.destination.port }}</template>
            </div>
          </div>
        </div>

        <div class="flow-stats">
          <div class="stat">
            <span class="stat-label">Latency:</span>
            <span class="stat-value">{{ flow.speed_bps ? flow.speed_bps + ' ms' : 'N/A' }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">Sent:</span>
            <span class="stat-value">{{ formatBytes(flow.bytes_sent || 0) }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">Recv:</span>
            <span class="stat-value">{{ formatBytes(flow.bytes_recv || 0) }}</span>
          </div>
        </div>

        <div class="flow-meta">
          <span class="meta-item">Started: {{ formatTime(flow.start_time) }}</span>
        </div>
      </div>
    </div>

    <div class="live-flow-legend">
      <div class="legend-item">
        <span class="legend-color active"></span>
        <span>Active</span>
      </div>
      <div class="legend-item">
        <span class="legend-color idle"></span>
        <span>Idle</span>
      </div>
      <div class="legend-item">
        <span class="legend-color error"></span>
        <span>Error</span>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
import liveFlowService from "../plugins/liveflow";

export default {
  name: "LiveFlowDashboard",
  data() {
    return {
      filters: {
        protocol: "",
        status: "",
      },
      expandedWarnings: {},
      debugInfo: null,
      _debugTimer: null,
    };
  },
  computed: {
    ...mapState(["liveFlows", "liveFlowConnected", "liveFlowWarnings"]),
    connectionStatus() {
      return this.liveFlowConnected ? "connected" : "disconnected";
    },
    connectionStatusText() {
      return this.liveFlowConnected ? "Connected" : "Disconnected";
    },
    safeFlows() {
      return (this.liveFlows || []).filter(
        (flow) =>
          flow &&
          typeof flow === "object" &&
          typeof flow.session_id === "string" &&
          typeof flow.protocol === "string"
      );
    },
    filteredFlows() {
      let flows = this.safeFlows;
      if (this.filters.status) {
        flows = flows.filter((flow) => flow.status === this.filters.status);
      }
      return flows;
    },
  },
  mounted() {
    // Auto-connect to live flow WebSocket
    const token = localStorage["token"];
    if (token) {
      console.log("[LiveFlow] Dashboard mounted, connecting...");
      this.$store.dispatch("connectLiveFlow", token);
    } else {
      console.warn("[LiveFlow] No auth token found, cannot connect");
    }

    // Start debug info refresh
    this._debugTimer = setInterval(() => {
      this.refreshDebugInfo();
    }, 3000);
  },
  beforeDestroy() {
    // Disconnect when dashboard is destroyed
    console.log("[LiveFlow] Dashboard destroyed, disconnecting...");
    this.$store.dispatch("disconnectLiveFlow");
    if (this._debugTimer) {
      clearInterval(this._debugTimer);
      this._debugTimer = null;
    }
  },
  methods: {
    refreshDebugInfo() {
      try {
        this.debugInfo = liveFlowService.getDebugInfo();
      } catch (e) {
        // Ignore errors during debug refresh
      }
    },
    formatBytes(bytes) {
      if (!bytes || bytes === 0) return "0 B";
      const k = 1024;
      const sizes = ["B", "KB", "MB", "GB"];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    },
    formatTime(timestamp) {
      if (!timestamp) return "N/A";
      const date = new Date(timestamp);
      if (isNaN(date.getTime())) return "N/A";
      return date.toLocaleTimeString();
    },
    clearFlows() {
      this.$store.commit("SET_LIVE_FLOWS", []);
      this.$store.commit("SET_LIVE_FLOW_WARNINGS", []);
    },
    dismissWarnings() {
      this.$store.commit("SET_LIVE_FLOW_WARNINGS", []);
      this.expandedWarnings = {};
    },
    toggleRaw(idx) {
      this.$set(this.expandedWarnings, idx, !this.expandedWarnings[idx]);
    },
  },
};
</script>

<!-- Use non-scoped style to support body.theme-dark -->
<style>
/* ===== Light Theme (default) ===== */
.live-flow-dashboard {
  padding: 20px;
  background-color: #f5f5f5;
  border-radius: 8px;
}
.live-flow-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.live-flow-header h2 {
  margin: 0;
  color: #333;
}
.connection-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #666;
}
.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background-color: #ccc;
}
.status-indicator.connected {
  background-color: #4caf50;
  animation: lfd-pulse 2s infinite;
}
.status-indicator.disconnected {
  background-color: #f44336;
}
@keyframes lfd-pulse {
  0% { box-shadow: 0 0 0 0 rgba(76, 175, 80, 0.4); }
  70% { box-shadow: 0 0 0 10px rgba(76, 175, 80, 0); }
  100% { box-shadow: 0 0 0 0 rgba(76, 175, 80, 0); }
}
/* Debug info */
.live-flow-debug {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
  padding: 8px 12px;
  background-color: #e8eaf6;
  border-radius: 6px;
  font-size: 12px;
  color: #3949ab;
}
.debug-item {
  font-family: monospace;
}
/* Warnings */
.live-flow-warnings {
  margin-bottom: 20px;
  padding: 12px 16px;
  background-color: #fff3e0;
  border: 1px solid #ffb74d;
  border-radius: 6px;
  animation: lfd-fadeIn 0.3s ease;
}
.warnings-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #e65100;
  font-weight: 500;
}
.warnings-icon { font-size: 18px; }
.warnings-dismiss {
  margin-left: auto;
  padding: 4px 12px;
  background-color: transparent;
  border: 1px solid #e65100;
  border-radius: 4px;
  color: #e65100;
  cursor: pointer;
  font-size: 12px;
}
.warnings-dismiss:hover { background-color: #e65100; color: white; }
.warnings-list { margin-top: 8px; }
.warning-item {
  padding: 6px 0;
  border-bottom: 1px solid rgba(230, 81, 0, 0.1);
  font-size: 13px;
  color: #bf360c;
}
.warning-item:last-child { border-bottom: none; }
.warning-message { display: block; }
.warning-raw {
  display: inline-block;
  margin-top: 4px;
  padding: 2px 8px;
  font-size: 11px;
  color: #e65100;
  border: 1px solid #e65100;
  border-radius: 3px;
  cursor: pointer;
}
.warning-raw:hover { background-color: #e65100; color: white; }
.warning-raw-data {
  margin-top: 6px;
  padding: 8px;
  background-color: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
  font-size: 11px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
/* Controls */
.live-flow-controls {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  padding: 15px;
  background-color: white;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}
.filter-group label { font-size: 14px; color: #666; }
.filter-select {
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  background-color: white;
  color: #333;
}
.clear-button {
  padding: 6px 16px;
  background-color: #f44336;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  margin-left: auto;
}
.clear-button:hover { background-color: #d32f2f; }
/* Canvas */
.live-flow-canvas { min-height: 200px; }
.no-flows { text-align: center; padding: 40px; color: #999; font-size: 16px; }
.no-flows-msg { color: #f44336; }
/* Flow cards */
.flow-card {
  padding: 15px;
  background-color: white;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 15px;
  transition: all 0.3s ease;
  animation: lfd-fadeIn 0.3s ease;
}
@keyframes lfd-fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
.flow-card.active { border-left: 4px solid #4caf50; }
.flow-card.idle { border-left: 4px solid #ff9800; }
.flow-card.error { border-left: 4px solid #f44336; }
.flow-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.flow-id { font-family: monospace; font-size: 12px; color: #666; }
.flow-status { padding: 2px 8px; border-radius: 12px; font-size: 12px; font-weight: 500; }
.flow-status.active { background-color: #e8f5e9; color: #2e7d32; }
.flow-status.idle { background-color: #fff3e0; color: #ef6c00; }
.flow-status.error { background-color: #ffebee; color: #c62828; }
.flow-path {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.endpoint { flex: 1; }
.endpoint-label { font-size: 12px; color: #999; margin-bottom: 4px; }
.endpoint-value { font-family: monospace; font-size: 14px; color: #333; }
.flow-arrow {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 0 15px;
}
.arrow-icon { font-size: 20px; color: #666; animation: lfd-flow 1.5s infinite; }
@keyframes lfd-flow {
  0% { opacity: 0.5; transform: translateX(-2px); }
  50% { opacity: 1; transform: translateX(2px); }
  100% { opacity: 0.5; transform: translateX(-2px); }
}
.protocol-badge {
  font-size: 10px;
  padding: 2px 6px;
  background-color: #e3f2fd;
  color: #1976d2;
  border-radius: 4px;
  margin-top: 4px;
}
.flow-stats { display: flex; gap: 20px; margin-bottom: 12px; }
.stat { display: flex; gap: 6px; }
.stat-label { color: #666; font-size: 12px; }
.stat-value { font-weight: 500; font-size: 12px; color: #333; }
.flow-meta { font-size: 12px; color: #999; }
/* Legend */
.live-flow-legend {
  display: flex;
  gap: 20px;
  margin-top: 20px;
  padding: 15px;
  background-color: white;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #666;
}
.legend-color { width: 12px; height: 12px; border-radius: 3px; }
.legend-color.active { background-color: #4caf50; }
.legend-color.idle { background-color: #ff9800; }
.legend-color.error { background-color: #f44336; }

/* ===== Dark Theme ===== */
body.theme-dark .live-flow-dashboard {
  background-color: #1c1b1f;
}
body.theme-dark .live-flow-header h2 {
  color: #e6e1e5;
}
body.theme-dark .connection-status {
  color: #cac4d0;
}
body.theme-dark .live-flow-debug {
  background-color: #2d2a3e;
  color: #b39ddb;
}
body.theme-dark .live-flow-warnings {
  background-color: #3e2a00;
  border-color: #ff8f00;
}
body.theme-dark .warnings-header {
  color: #ffab40;
}
body.theme-dark .warnings-dismiss {
  border-color: #ffab40;
  color: #ffab40;
}
body.theme-dark .warnings-dismiss:hover {
  background-color: #ffab40;
  color: #000;
}
body.theme-dark .warning-item {
  color: #ffcc80;
  border-bottom-color: rgba(255, 171, 64, 0.1);
}
body.theme-dark .warning-raw {
  color: #ffab40;
  border-color: #ffab40;
}
body.theme-dark .warning-raw:hover {
  background-color: #ffab40;
  color: #000;
}
body.theme-dark .warning-raw-data {
  background-color: rgba(255, 255, 255, 0.05);
}
body.theme-dark .live-flow-controls {
  background-color: #211f26;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}
body.theme-dark .filter-group label {
  color: #cac4d0;
}
body.theme-dark .filter-select {
  background-color: #2d2a3e;
  color: #e6e1e5;
  border-color: #49454f;
}
body.theme-dark .clear-button {
  background-color: #c62828;
}
body.theme-dark .clear-button:hover {
  background-color: #e53935;
}
body.theme-dark .live-flow-canvas .no-flows {
  color: #938f99;
}
body.theme-dark .no-flows-msg {
  color: #ef5350;
}
body.theme-dark .flow-card {
  background-color: #211f26;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}
body.theme-dark .flow-id { color: #cac4d0; }
body.theme-dark .flow-status.active { background-color: #1b3a20; color: #81c784; }
body.theme-dark .flow-status.idle { background-color: #3e2a00; color: #ffb74d; }
body.theme-dark .flow-status.error { background-color: #3e1216; color: #ef5350; }
body.theme-dark .endpoint-label { color: #938f99; }
body.theme-dark .endpoint-value { color: #e6e1e5; }
body.theme-dark .arrow-icon { color: #cac4d0; }
body.theme-dark .protocol-badge {
  background-color: #1a2744;
  color: #64b5f6;
}
body.theme-dark .stat-label { color: #938f99; }
body.theme-dark .stat-value { color: #e6e1e5; }
body.theme-dark .flow-meta { color: #938f99; }
body.theme-dark .live-flow-legend {
  background-color: #211f26;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}
body.theme-dark .legend-item {
  color: #cac4d0;
}
body.theme-dark .flow-card.active {
  border-left-color: #4caf50;
}
body.theme-dark .flow-card.idle {
  border-left-color: #ff9800;
}
body.theme-dark .flow-card.error {
  border-left-color: #f44336;
}
body.theme-dark .legend-color.active { background-color: #4caf50; }
body.theme-dark .legend-color.idle { background-color: #ff9800; }
body.theme-dark .legend-color.error { background-color: #f44336; }
</style>
