<template>
  <section class="live-flow-dashboard">
    <header class="live-flow-header">
      <div>
        <h2>Live Flow</h2>
        <p>Observed routes: source → selected proxy → destination</p>
      </div>
      <div class="connection-status" :title="connectionStatusText">
        <span :class="['status-indicator', connectionStatus]"></span>
        {{ connectionStatusText }}
      </div>
    </header>

    <div v-if="liveFlowWarnings.length" class="live-flow-warning">
      {{
        liveFlowWarnings[0].message || "The server reported a live-flow error."
      }}
      <button type="button" @click="dismissWarnings">Dismiss</button>
    </div>

    <div class="live-flow-controls">
      <label>
        Protocol
        <select v-model="filters.protocol">
          <option value="">All</option>
          <option
            v-for="protocol in protocols"
            :key="protocol"
            :value="protocol"
          >
            {{ protocol.toUpperCase() }}
          </option>
        </select>
      </label>
      <label>
        Status
        <select v-model="filters.status">
          <option value="">All</option>
          <option value="active">Active</option>
          <option value="idle">Idle</option>
          <option value="error">Error</option>
        </select>
      </label>
      <label>
        Proxy
        <select v-model="filters.proxy">
          <option value="">All selected proxies</option>
          <option v-for="proxy in proxies" :key="proxy.id" :value="proxy.id">
            {{ proxy.name }}
          </option>
        </select>
      </label>
      <span class="flow-count">
        {{ filteredFlows.length }} observed route<span
          v-if="filteredFlows.length !== 1"
          >s</span
        >
      </span>
    </div>

    <div
      class="flow-topology"
      role="region"
      aria-label="Live traffic route topology"
    >
      <div v-if="connectionStatus === 'disconnected'" class="flow-placeholder">
        Connecting to live route events…
      </div>
      <div v-else-if="!filteredFlows.length" class="flow-placeholder">
        Waiting for routed traffic. Start a proxied connection to see its path.
      </div>
      <svg
        v-else
        class="flow-svg"
        :viewBox="`0 0 ${graph.width} ${graph.height}`"
        preserveAspectRatio="xMidYMin meet"
      >
        <defs>
          <marker
            id="live-flow-arrow-active"
            markerWidth="8"
            markerHeight="8"
            refX="6"
            refY="3"
            orient="auto"
          >
            <path d="M0,0 L0,6 L7,3 z" class="flow-arrow-head active" />
          </marker>
          <marker
            id="live-flow-arrow-idle"
            markerWidth="8"
            markerHeight="8"
            refX="6"
            refY="3"
            orient="auto"
          >
            <path d="M0,0 L0,6 L7,3 z" class="flow-arrow-head idle" />
          </marker>
          <marker
            id="live-flow-arrow-error"
            markerWidth="8"
            markerHeight="8"
            refX="6"
            refY="3"
            orient="auto"
          >
            <path d="M0,0 L0,6 L7,3 z" class="flow-arrow-head error" />
          </marker>
        </defs>

        <g class="flow-columns">
          <text :x="graph.sourceX" y="34" text-anchor="middle">SOURCE</text>
          <text
            v-for="column in graph.proxyColumns"
            :key="column.index"
            :x="column.x"
            y="34"
            text-anchor="middle"
          >
            {{ column.index === 0 ? "SELECTED PROXY" : "PROXY HOP" }}
          </text>
          <text :x="graph.destinationX" y="34" text-anchor="middle">
            DESTINATION
          </text>
        </g>

        <g class="flow-edges">
          <path
            v-for="edge in graph.edges"
            :key="edge.id"
            :d="edge.path"
            :class="[
              'flow-edge',
              edge.status,
              { selected: edge.sessionId === selectedSessionId },
            ]"
            :style="{ strokeWidth: edge.width }"
            :marker-end="`url(#live-flow-arrow-${edge.status})`"
            tabindex="0"
            @click="selectFlow(edge.sessionId)"
            @keyup.enter="selectFlow(edge.sessionId)"
          >
            <title>{{ edge.label }}</title>
          </path>
        </g>

        <g class="flow-nodes">
          <g
            v-for="node in graph.nodes"
            :key="node.id"
            :class="[
              'flow-node',
              node.kind,
              { selected: node.sessionIds.includes(selectedSessionId) },
            ]"
            tabindex="0"
            @click="selectNode(node)"
            @keyup.enter="selectNode(node)"
          >
            <title>{{ node.detail }}</title>
            <rect
              :x="node.x - 88"
              :y="node.y - 31"
              width="176"
              height="62"
              rx="10"
            />
            <text
              :x="node.x"
              :y="node.y - 7"
              text-anchor="middle"
              class="flow-node-title"
            >
              {{ node.title }}
            </text>
            <text
              :x="node.x"
              :y="node.y + 14"
              text-anchor="middle"
              class="flow-node-detail"
            >
              {{ node.subtitle }}
            </text>
          </g>
        </g>
      </svg>
    </div>

    <aside v-if="selectedFlow" class="flow-details">
      <div>
        <strong>Selected route</strong>
        <span>{{ selectedFlow.session_id }}</span>
      </div>
      <div>
        <strong>Path</strong><span>{{ pathLabel(selectedFlow) }}</span>
      </div>
      <div>
        <strong>Protocol</strong
        ><span>{{ selectedFlow.protocol.toUpperCase() }}</span>
      </div>
      <div>
        <strong>Last observed</strong
        ><span>{{
          formatTime(selectedFlow.last_activity || selectedFlow.start_time)
        }}</span>
      </div>
      <div>
        <strong>Traffic counters</strong
        ><span>{{ trafficSummary(selectedFlow) }}</span>
      </div>
    </aside>

    <footer class="flow-legend">
      <span><i class="legend-line active"></i> Active route</span>
      <span><i class="legend-line idle"></i> Idle route</span>
      <span><i class="legend-line error"></i> Error route</span>
      <span><i class="legend-node proxy"></i> Selected proxy server</span>
    </footer>
  </section>
</template>

<script>
import { mapState } from "vuex";

const GRAPH_WIDTH = 1200;
const TOP_PADDING = 82;
const NODE_GAP = 92;

function endpointId(endpoint) {
  return `${endpoint.domain || endpoint.ip}:${endpoint.port || 0}`;
}

function endpointLabel(endpoint) {
  const host = endpoint.domain || endpoint.ip;
  return endpoint.port ? `${host}:${endpoint.port}` : host;
}

function layoutNodes(nodes, x) {
  return nodes.map((node, index) => ({
    ...node,
    x,
    y: TOP_PADDING + index * NODE_GAP,
  }));
}

export default {
  name: "LiveFlowDashboard",
  data() {
    return {
      filters: { protocol: "", status: "", proxy: "" },
      selectedSessionId: "",
    };
  },
  computed: {
    ...mapState(["liveFlows", "liveFlowConnected", "liveFlowWarnings"]),
    connectionStatus() {
      return this.liveFlowConnected ? "connected" : "disconnected";
    },
    connectionStatusText() {
      return this.liveFlowConnected
        ? "Live route stream connected"
        : "Live route stream disconnected";
    },
    flows() {
      return (this.liveFlows || []).filter(
        (flow) =>
          flow &&
          typeof flow.session_id === "string" &&
          flow.source &&
          flow.destination &&
          typeof flow.protocol === "string"
      );
    },
    protocols() {
      return [
        ...new Set(this.flows.map((flow) => flow.protocol.toLowerCase())),
      ].sort();
    },
    proxies() {
      const found = new Map();
      this.flows.forEach((flow) => {
        this.flowHops(flow).forEach((proxy) => {
          if (!found.has(proxy.proxy_id)) {
            found.set(proxy.proxy_id, { id: proxy.proxy_id, name: proxy.name });
          }
        });
      });
      return [...found.values()].sort((a, b) => a.name.localeCompare(b.name));
    },
    filteredFlows() {
      return this.flows.filter((flow) => {
        if (
          this.filters.protocol &&
          flow.protocol.toLowerCase() !== this.filters.protocol
        )
          return false;
        if (this.filters.status && flow.status !== this.filters.status)
          return false;
        return (
          !this.filters.proxy ||
          this.flowHops(flow).some((hop) => hop.proxy_id === this.filters.proxy)
        );
      });
    },
    selectedFlow() {
      return (
        this.filteredFlows.find(
          (flow) => flow.session_id === this.selectedSessionId
        ) || null
      );
    },
    graph() {
      const flows = this.filteredFlows;
      const maxHops = Math.max(
        1,
        ...flows.map((flow) => this.flowHops(flow).length)
      );
      const sourceX = 120;
      const destinationX = GRAPH_WIDTH - 120;
      const proxyGap = (destinationX - sourceX) / (maxHops + 1);
      const sourceMap = new Map();
      const destinationMap = new Map();
      const proxyMaps = Array.from({ length: maxHops }, () => new Map());

      flows.forEach((flow) => {
        const sourceKey = `source:${endpointId(flow.source)}`;
        const destinationKey = `destination:${endpointId(flow.destination)}`;
        this.addNode(
          sourceMap,
          sourceKey,
          "source",
          endpointLabel(flow.source),
          flow.source.ip,
          flow.session_id
        );
        this.addNode(
          destinationMap,
          destinationKey,
          "destination",
          endpointLabel(flow.destination),
          flow.destination.ip,
          flow.session_id
        );
        this.flowHops(flow).forEach((proxy, index) => {
          const proxyKey = `proxy:${index}:${proxy.proxy_id}`;
          this.addNode(
            proxyMaps[index],
            proxyKey,
            "proxy",
            proxy.name,
            `${proxy.type} · ${proxy.server}`,
            flow.session_id
          );
        });
      });

      const nodes = [
        ...layoutNodes([...sourceMap.values()], sourceX),
        ...proxyMaps.flatMap((nodeMap, index) =>
          layoutNodes([...nodeMap.values()], sourceX + proxyGap * (index + 1))
        ),
        ...layoutNodes([...destinationMap.values()], destinationX),
      ];
      const positions = new Map(nodes.map((node) => [node.id, node]));
      const edges = [];

      flows.forEach((flow) => {
        const nodeIds = [`source:${endpointId(flow.source)}`];
        this.flowHops(flow).forEach((proxy, index) =>
          nodeIds.push(`proxy:${index}:${proxy.proxy_id}`)
        );
        nodeIds.push(`destination:${endpointId(flow.destination)}`);
        nodeIds.slice(1).forEach((targetId, index) => {
          const from = positions.get(nodeIds[index]);
          const to = positions.get(targetId);
          if (!from || !to) return;
          edges.push({
            id: `${flow.session_id}:${index}`,
            sessionId: flow.session_id,
            status: flow.status || "active",
            width: this.edgeWidth(flow.speed_bps),
            path: this.edgePath(from, to),
            label: `${this.pathLabel(flow)} (${flow.protocol.toUpperCase()})`,
          });
        });
      });

      const highestNodeCount = Math.max(
        1,
        ...[sourceMap, destinationMap, ...proxyMaps].map(
          (nodeMap) => nodeMap.size
        )
      );
      return {
        width: GRAPH_WIDTH,
        height: Math.max(
          300,
          TOP_PADDING + (highestNodeCount - 1) * NODE_GAP + 90
        ),
        sourceX,
        destinationX,
        proxyColumns: proxyMaps.map((_, index) => ({
          index,
          x: sourceX + proxyGap * (index + 1),
        })),
        nodes,
        edges,
      };
    },
  },
  watch: {
    filteredFlows(flows) {
      if (
        this.selectedSessionId &&
        !flows.some((flow) => flow.session_id === this.selectedSessionId)
      ) {
        this.selectedSessionId = "";
      }
    },
  },
  mounted() {
    const token = localStorage["token"];
    if (token) this.$store.dispatch("connectLiveFlow", token);
  },
  beforeDestroy() {
    this.$store.dispatch("disconnectLiveFlow");
  },
  methods: {
    addNode(nodes, id, kind, title, subtitle, sessionId) {
      const existing = nodes.get(id);
      if (existing) {
        existing.sessionIds.push(sessionId);
      } else {
        nodes.set(id, { id, kind, title, subtitle, sessionIds: [sessionId] });
      }
    },
    flowHops(flow) {
      if (Array.isArray(flow.proxy_chain) && flow.proxy_chain.length)
        return flow.proxy_chain;
      return [
        {
          proxy_id: "route:unresolved",
          name: "Route unresolved",
          type: "unknown",
          server: "Selected outbound was not reported",
        },
      ];
    },
    edgePath(from, to) {
      const startX = from.x + 88;
      const endX = to.x - 88;
      const curve = Math.max(48, (endX - startX) * 0.36);
      return `M ${startX} ${from.y} C ${startX + curve} ${from.y}, ${
        endX - curve
      } ${to.y}, ${endX} ${to.y}`;
    },
    edgeWidth(speed) {
      return !speed || speed <= 0 ? 2 : Math.min(8, 2 + Math.log10(speed + 1));
    },
    selectFlow(sessionId) {
      this.selectedSessionId = sessionId;
    },
    selectNode(node) {
      this.selectedSessionId = node.sessionIds[0] || "";
    },
    pathLabel(flow) {
      return [
        endpointLabel(flow.source),
        ...this.flowHops(flow).map((hop) => hop.name),
        endpointLabel(flow.destination),
      ].join(" → ");
    },
    trafficSummary(flow) {
      if (!flow.bytes_sent && !flow.bytes_recv && !flow.speed_bps) {
        return "Unavailable from the current core access event";
      }
      return `${this.formatBytes(flow.bytes_sent)} sent, ${this.formatBytes(
        flow.bytes_recv
      )} received, ${this.formatSpeed(flow.speed_bps)}`;
    },
    formatBytes(bytes) {
      if (!bytes) return "0 B";
      const units = ["B", "KB", "MB", "GB"];
      const exponent = Math.min(
        Math.floor(Math.log(bytes) / Math.log(1024)),
        units.length - 1
      );
      return `${(bytes / Math.pow(1024, exponent)).toFixed(exponent ? 1 : 0)} ${
        units[exponent]
      }`;
    },
    formatSpeed(speed) {
      if (!speed) return "0 bps";
      const units = ["bps", "Kbps", "Mbps", "Gbps"];
      const exponent = Math.min(
        Math.floor(Math.log(speed) / Math.log(1000)),
        units.length - 1
      );
      return `${(speed / Math.pow(1000, exponent)).toFixed(1)} ${
        units[exponent]
      }`;
    },
    formatTime(value) {
      const date = new Date(value);
      return Number.isNaN(date.getTime())
        ? "Unknown"
        : date.toLocaleTimeString();
    },
    dismissWarnings() {
      this.$store.commit("SET_LIVE_FLOW_WARNINGS", []);
    },
  },
};
</script>

<style>
.live-flow-dashboard {
  padding: 20px;
  border-radius: 8px;
  background: #f5f5f5;
}
.live-flow-header,
.live-flow-controls,
.flow-legend,
.flow-details {
  display: flex;
  align-items: center;
}
.live-flow-header {
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}
.live-flow-header h2 {
  margin: 0;
  color: #333;
}
.live-flow-header p {
  margin: 4px 0 0;
  color: #777;
  font-size: 13px;
}
.connection-status {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 13px;
}
.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #d32f2f;
}
.status-indicator.connected {
  background: #2e7d32;
  animation: live-flow-pulse 2s infinite;
}
@keyframes live-flow-pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(46, 125, 50, 0.45);
  }
  75% {
    box-shadow: 0 0 0 10px rgba(46, 125, 50, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(46, 125, 50, 0);
  }
}
.live-flow-warning {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  padding: 10px 12px;
  border: 1px solid #ffb74d;
  border-radius: 6px;
  background: #fff3e0;
  color: #b85c00;
  font-size: 13px;
}
.live-flow-warning button {
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-weight: 600;
}
.live-flow-controls {
  flex-wrap: wrap;
  gap: 14px;
  margin-bottom: 16px;
  padding: 12px;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}
.live-flow-controls label {
  display: flex;
  align-items: center;
  gap: 7px;
  color: #555;
  font-size: 13px;
}
.live-flow-controls select {
  max-width: 210px;
  padding: 6px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: #fff;
  color: #333;
}
.flow-count {
  margin-left: auto;
  color: #777;
  font-size: 13px;
}
.flow-topology {
  overflow-x: auto;
  min-height: 300px;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}
.flow-svg {
  display: block;
  min-width: 760px;
  width: 100%;
}
.flow-placeholder {
  display: grid;
  min-height: 300px;
  place-items: center;
  padding: 24px;
  color: #888;
  text-align: center;
}
.flow-columns text {
  fill: #8a8a8a;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}
.flow-edge {
  fill: none;
  stroke: #2e7d32;
  stroke-linecap: round;
  stroke-dasharray: 8 12;
  animation: live-flow-dash 1s linear infinite;
  cursor: pointer;
  opacity: 0.72;
  transition: opacity 0.2s, stroke-width 0.2s;
}
.flow-edge.idle {
  stroke: #ed8b00;
  animation-duration: 2s;
}
.flow-edge.error {
  stroke: #d32f2f;
  animation: none;
  stroke-dasharray: 3 9;
}
.flow-edge.selected,
.flow-edge:hover {
  opacity: 1;
}
.flow-arrow-head.active {
  fill: #2e7d32;
}
.flow-arrow-head.idle {
  fill: #ed8b00;
}
.flow-arrow-head.error {
  fill: #d32f2f;
}
@keyframes live-flow-dash {
  to {
    stroke-dashoffset: -20;
  }
}
.flow-node {
  cursor: pointer;
}
.flow-node rect {
  stroke-width: 2;
  transition: stroke-width 0.2s, filter 0.2s;
}
.flow-node.source rect {
  fill: #e3f2fd;
  stroke: #1976d2;
}
.flow-node.proxy rect {
  fill: #f3e5f5;
  stroke: #7b1fa2;
}
.flow-node.destination rect {
  fill: #e8f5e9;
  stroke: #2e7d32;
}
.flow-node.selected rect,
.flow-node:hover rect {
  stroke-width: 4;
  filter: drop-shadow(0 2px 3px rgba(0, 0, 0, 0.2));
}
.flow-node-title {
  fill: #202124;
  font-size: 14px;
  font-weight: 700;
}
.flow-node-detail {
  fill: #555;
  font-size: 11px;
}
.flow-details {
  align-items: stretch;
  flex-wrap: wrap;
  gap: 12px 24px;
  margin-top: 16px;
  padding: 12px;
  border-left: 4px solid #7b1fa2;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  font-size: 13px;
}
.flow-details div {
  display: grid;
  gap: 2px;
  max-width: 100%;
}
.flow-details strong {
  color: #666;
  font-size: 11px;
  text-transform: uppercase;
}
.flow-details span {
  overflow-wrap: anywhere;
  color: #333;
}
.flow-legend {
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 16px;
  color: #666;
  font-size: 12px;
}
.flow-legend span {
  display: flex;
  align-items: center;
  gap: 6px;
}
.legend-line {
  width: 22px;
  border-top: 3px dashed #2e7d32;
}
.legend-line.idle {
  border-color: #ed8b00;
}
.legend-line.error {
  border-color: #d32f2f;
}
.legend-node {
  width: 13px;
  height: 13px;
  border: 2px solid #7b1fa2;
  border-radius: 3px;
  background: #f3e5f5;
}
body.theme-dark .live-flow-dashboard {
  background: #1c1b1f;
}
body.theme-dark .live-flow-header h2,
body.theme-dark .flow-details span {
  color: #e6e1e5;
}
body.theme-dark .live-flow-header p,
body.theme-dark .connection-status,
body.theme-dark .flow-count,
body.theme-dark .flow-legend {
  color: #cac4d0;
}
body.theme-dark .live-flow-controls,
body.theme-dark .flow-topology,
body.theme-dark .flow-details {
  background: #211f26;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}
body.theme-dark .live-flow-controls label {
  color: #cac4d0;
}
body.theme-dark .live-flow-controls select {
  border-color: #49454f;
  background: #2d2a3e;
  color: #e6e1e5;
}
body.theme-dark .flow-placeholder {
  color: #938f99;
}
body.theme-dark .flow-columns text {
  fill: #938f99;
}
body.theme-dark .flow-node.source rect {
  fill: #102a43;
  stroke: #64b5f6;
}
body.theme-dark .flow-node.proxy rect {
  fill: #33213b;
  stroke: #ce93d8;
}
body.theme-dark .flow-node.destination rect {
  fill: #1b3a20;
  stroke: #81c784;
}
body.theme-dark .flow-node-title {
  fill: #e6e1e5;
}
body.theme-dark .flow-node-detail {
  fill: #cac4d0;
}
body.theme-dark .flow-details strong {
  color: #938f99;
}
body.theme-dark .live-flow-warning {
  border-color: #ffb74d;
  background: #3e2a00;
  color: #ffcc80;
}
@media (max-width: 700px) {
  .live-flow-dashboard {
    padding: 12px;
  }
  .live-flow-header {
    align-items: flex-start;
    flex-direction: column;
  }
  .flow-count {
    width: 100%;
    margin-left: 0;
  }
}
@media (prefers-reduced-motion: reduce) {
  .status-indicator.connected,
  .flow-edge {
    animation: none;
  }
}
</style>
