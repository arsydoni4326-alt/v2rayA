<template>
  <div class="live-flow-dashboard">
    <div class="live-flow-header">
      <h2>Live Flow Visualization</h2>
      <div class="connection-status">
        <span :class="['status-indicator', connectionStatus]"></span>
        {{ connectionStatusText }}
      </div>
    </div>

    <div class="live-flow-controls">
      <div class="filter-group">
        <label>Protocol Filter:</label>
        <select v-model="filters.protocol" class="filter-select">
          <option value="">All</option>
          <option value="tcp">TCP</option>
          <option value="udp">UDP</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Status Filter:</label>
        <select v-model="filters.status" class="filter-select">
          <option value="">All</option>
          <option value="active">Active</option>
          <option value="idle">Idle</option>
        </select>
      </div>

      <button @click="clearFlows" class="clear-button">
        Clear All
      </button>
    </div>

    <div class="live-flow-canvas" ref="canvas">
      <div v-if="filteredFlows.length === 0" class="no-flows">
        No active flows. Connect to a proxy server to see live traffic.
      </div>

      <div
        v-for="flow in filteredFlows"
        :key="flow.session_id"
        :class="['flow-card', flow.status]"
      >
        <div class="flow-header">
          <span class="flow-id">{{ flow.session_id.substring(0, 8) }}</span>
          <span :class="['flow-status', flow.status]">{{ flow.status }}</span>
        </div>

        <div class="flow-path">
          <div class="endpoint source">
            <div class="endpoint-label">Source</div>
            <div class="endpoint-value">{{ flow.source.ip }}:{{ flow.source.port }}</div>
          </div>

          <div class="flow-arrow">
            <span class="arrow-icon">→</span>
            <span class="protocol-badge">{{ flow.protocol }}</span>
          </div>

          <div class="endpoint destination">
            <div class="endpoint-label">Destination</div>
            <div class="endpoint-value">
              {{ flow.destination.domain || flow.destination.ip }}:{{ flow.destination.port }}
            </div>
          </div>
        </div>

        <div class="flow-stats">
          <div class="stat">
            <span class="stat-label">Sent:</span>
            <span class="stat-value">{{ formatBytes(flow.bytes_sent) }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">Recv:</span>
            <span class="stat-value">{{ formatBytes(flow.bytes_recv) }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">Speed:</span>
            <span class="stat-value">{{ formatSpeed(flow.speed_bps) }}</span>
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
<script>
import { mapState, mapGetters } from "vuex";

export default {
  name: "LiveFlowDashboard",
  data() {
    return {
      filters: {
        protocol: "",
        status: "",
      },
    };
  },
  computed: {
    ...mapState(["liveFlows", "liveFlowConnected"]),
    ...mapGetters(["filteredLiveFlows"]),
    connectionStatus() {
      return this.liveFlowConnected ? "connected" : "disconnected";
    },
    connectionStatusText() {
      return this.liveFlowConnected ? "Connected" : "Disconnected";
    },
    filteredFlows() {
      let flows = this.liveFlows;

      if (this.filters.protocol) {
        flows = flows.filter(
          (flow) => flow.protocol.toLowerCase() === this.filters.protocol
        );
      }

      if (this.filters.status) {
        flows = flows.filter((flow) => flow.status === this.filters.status);
      }

      return flows;
    },
  },
  methods: {
    formatBytes(bytes) {
      if (bytes === 0) return "0 B";
      const k = 1024;
      const sizes = ["B", "KB", "MB", "GB"];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    },
    formatSpeed(bps) {
      if (bps === 0) return "0 bps";
      const k = 1000;
      const sizes = ["bps", "Kbps", "Mbps", "Gbps"];
      const i = Math.floor(Math.log(bps) / Math.log(k));
      return parseFloat((bps / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    },
    formatTime(timeString) {
      if (!timeString) return "N/A";
      const date = new Date(timeString);
      return date.toLocaleTimeString();
    },
    clearFlows() {
      this.$store.commit("SET_LIVE_FLOWS", []);
    },
  },
};
</script>
      <div class="legend-item">
<style scoped>
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
  animation: pulse 2s infinite;
}

.status-indicator.disconnected {
  background-color: #f44336;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(76, 175, 80, 0.4);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(76, 175, 80, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(76, 175, 80, 0);
  }
}

.live-flow-controls {
  display: flex;
  gap: 20px;
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

.filter-group label {
  font-weight: 500;
  color: #555;
}

.filter-select {
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: white;
}

.clear-button {
  padding: 6px 16px;
  background-color: #f44336;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.clear-button:hover {
  background-color: #d32f2f;
}

.live-flow-canvas {
  min-height: 400px;
  background-color: white;
  border-radius: 6px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.no-flows {
  text-align: center;
  color: #999;
  padding: 40px;
  font-size: 16px;
}

.flow-card {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 15px;
  transition: all 0.3s ease;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.flow-card.active {
  border-left: 4px solid #4caf50;
}

.flow-card.idle {
  border-left: 4px solid #ff9800;
}

.flow-card.error {
  border-left: 4px solid #f44336;
}

.flow-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.flow-id {
  font-family: monospace;
  font-size: 12px;
  color: #666;
}

.flow-status {
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.flow-status.active {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.flow-status.idle {
  background-color: #fff3e0;
  color: #ef6c00;
}

.flow-status.error {
  background-color: #ffebee;
  color: #c62828;
}

.flow-path {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.endpoint {
  flex: 1;
}

.endpoint-label {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}

.endpoint-value {
  font-family: monospace;
  font-size: 14px;
  color: #333;
}

.flow-arrow {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 0 15px;
}

.arrow-icon {
  font-size: 20px;
  color: #666;
  animation: flow 1.5s infinite;
}

@keyframes flow {
  0% {
    opacity: 0.5;
    transform: translateX(-2px);
  }
  50% {
    opacity: 1;
    transform: translateX(2px);
  }
  100% {
    opacity: 0.5;
    transform: translateX(-2px);
  }
}

.protocol-badge {
  font-size: 10px;
  padding: 2px 6px;
  background-color: #e3f2fd;
  color: #1976d2;
  border-radius: 4px;
  margin-top: 4px;
}

.flow-stats {
  display: flex;
  gap: 20px;
  margin-bottom: 12px;
}

.stat {
  display: flex;
  gap: 6px;
}

.stat-label {
  color: #666;
  font-size: 12px;
}

.stat-value {
  font-weight: 500;
  font-size: 12px;
  color: #333;
}

.flow-meta {
  font-size: 12px;
  color: #999;
}

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

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 3px;
}

.legend-color.active {
  background-color: #4caf50;
}

.legend-color.idle {
  background-color: #ff9800;
}

.legend-color.error {
  background-color: #f44336;
}
</style>
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