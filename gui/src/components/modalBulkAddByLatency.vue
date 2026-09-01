<template>
  <div class="modal-card" style="max-width: 600px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("bulkAddByLatency.title") }}</p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" size="is-small" :closable="false">
        {{ $t("bulkAddByLatency.messages[0]") }}
      </b-message>
      <b-field :label="$t('bulkAddByLatency.maxLatency')">
        <b-numberinput v-model="maxLatency" :min="1" :max="99999" :step="50"
          :placeholder="$t('bulkAddByLatency.maxLatencyPlaceholder')" controls-position="compact" expanded />
      </b-field>
      <b-field :label="$t('bulkAddByLatency.targetGroup')">
        <b-select v-model="selectedOutbound" expanded>
          <option v-for="outbound in outbounds" :key="outbound" :value="outbound">
            {{ outbound.toUpperCase() }}
          </option>
        </b-select>
      </b-field>
      <div class="bulk-add-preview">
        <div v-if="matchingServers.length > 0" class="has-text-success">
          <b-icon icon="check-circle" size="is-small" />
          {{ $t("bulkAddByLatency.preview", { count: serversToAdd.length, group: selectedOutbound.toUpperCase() }) }}
          <div v-if="alreadyInGroupCount > 0" class="has-text-warning" style="margin-top:4px;font-size:0.85em">
            {{ $t("bulkAddByLatency.alreadyInGroup", { count: alreadyInGroupCount }) }}
          </div>
        </div>
        <div v-else-if="hasLatencyData" class="has-text-danger">
          <b-icon icon="alert-circle" size="is-small" />
          {{ $t("bulkAddByLatency.previewNoMatch", { threshold: maxLatency }) }}
        </div>
        <div v-else class="has-text-warning">
          <b-icon icon="alert" size="is-small" />
          {{ $t("bulkAddByLatency.noLatencyData") }}
        </div>
      </div>
    </section>
    <footer class="modal-card-foot" style="justify-content:flex-end">
      <b-button @click="cancel">{{ $t("operations.cancel") }}</b-button>
      <b-button type="is-success" :disabled="serversToAdd.length===0" :loading="saving" @click="confirm">
        {{ $t("bulkAddByLatency.confirm") }} ({{ serversToAdd.length }})
      </b-button>
    </footer>
  </div>
</template>

<script>
import axios from "../plugins/axios";
const apiRoot = process.env.VUE_APP_API_ROOT || "/api";
export default {
  name: "ModalBulkAddByLatency",
  props: {
    tabType: { type: String, default: "all" },
    tabIndex: { type: Number, default: -1 },
  },
  data() {
    return {
      maxLatency: 500, selectedOutbound: "proxy", outbounds: ["proxy"],
      touchData: null, currentConnectedServers: [], saving: false,
    };
  },
  computed: {
    hasLatencyData() {
      if (!this.touchData) return false;
      return this.getAllServers().some((s) => s.pingLatency && s.pingLatency.endsWith("ms"));
    },
    matchingServers() {
      if (!this.touchData) return [];
      return this.getAllServers().filter((s) => {
        if (s.disableReason) return false;
        if (!s.pingLatency || !s.pingLatency.endsWith("ms")) return false;
        return !isNaN(parseInt(s.pingLatency, 10)) && parseInt(s.pingLatency, 10) <= this.maxLatency;
      });
    },
    serversToAdd() {
      if (!this.matchingServers.length) return [];
      const existingKeys = new Set(
        this.currentConnectedServers
          .filter((w) => (w.outbound || "proxy") === this.selectedOutbound)
          .map((w) => `${w._type || w.TYPE}-${w.id || w.ID}-${w.sub || 0}`)
      );
      return this.matchingServers.filter((s) => {
        const sub = s._type === "subscriptionServer" ? (s.sub != null ? s.sub : this.findSubIndex(s)) : 0;
        return !existingKeys.has(`${s._type}-${s.id}-${sub}`);
      });
    },
    alreadyInGroupCount() { return this.matchingServers.length - this.serversToAdd.length; },
  },
  async created() { await this.fetchOutbounds(); await this.fetchTouchData(); },
  methods: {
    serverKey(s) {
      return `${s._type}-${s.id}-${s._type === "subscriptionServer" ? (s.sub != null ? s.sub : 0) : 0}`;
    },
    getAllServers() {
      if (!this.touchData) return [];
      if (this.tabType === "server") {
        return (this.touchData.servers || []).map((s) => ({ ...s, _type: "server" }));
      }
      if (this.tabType === "subscription" && this.tabIndex >= 0
          && this.tabIndex < (this.touchData.subscriptions || []).length) {
        const sub = this.touchData.subscriptions[this.tabIndex];
        return (sub.servers || []).map((s) => ({ ...s, _type: "subscriptionServer", sub: this.tabIndex }));
      }
      const servers = (this.touchData.servers || []).map((s) => ({ ...s, _type: "server" }));
      const subServers = [];
      (this.touchData.subscriptions || []).forEach((sub, subIdx) => {
        (sub.servers || []).forEach((s) => subServers.push({ ...s, _type: "subscriptionServer", sub: subIdx }));
      });
      return [...servers, ...subServers];
    },
    findSubIndex(server) {
      if (!this.touchData) return 0;
      for (let i = 0; i < (this.touchData.subscriptions || []).length; i++) {
        if ((this.touchData.subscriptions[i].servers || []).some((s) => s.id === server.id)) return i;
      }
      return 0;
    },
    async fetchOutbounds() {
      try {
        const res = await axios({ url: apiRoot + "/outbounds" });
        if (res.data.code === "SUCCESS" && res.data.data.outbounds) {
          this.outbounds = res.data.data.outbounds;
          if (!this.outbounds.includes(this.selectedOutbound)) this.selectedOutbound = this.outbounds[0] || "proxy";
        }
      } catch (e) { /* ignore */ }
    },
    async fetchTouchData() {
      try {
        const res = await axios({ url: apiRoot + "/touch" });
        if (res.data.code === "SUCCESS") {
          this.touchData = res.data.data.touch;
          this.currentConnectedServers = res.data.data.touch.connectedServer || [];
        }
      } catch (e) { /* ignore */ }
    },
    cancel() { this.$emit("close"); },
    async confirm() {
      if (!this.serversToAdd.length) return;
      this.saving = true;
      try {
        // Preserve ALL currently connected servers for this outbound,
        // not just those from the current tab. The backend API replaces
        // the full list, so filtering by tab here would discard servers
        // added from other subscriptions.
        const existingTouches = this.currentConnectedServers
          .filter((w) => (w.outbound || "proxy") === this.selectedOutbound)
          .map((w) => ({ id: w.id || w.ID, _type: w._type || w.TYPE,
            sub: (w._type || w.TYPE) === "subscriptionServer" ? (w.sub != null ? w.sub : 0) : 0,
            outbound: this.selectedOutbound }));
        const newTouches = this.serversToAdd.map((s) => ({
          id: s.id, _type: s._type,
          sub: s._type === "subscriptionServer" ? (s.sub != null ? s.sub : 0) : 0,
          outbound: this.selectedOutbound }));
        const res = await axios({ url: apiRoot + "/outboundConnections", method: "put",
          data: { outbound: this.selectedOutbound, touches: [...existingTouches, ...newTouches] } });
        if (res.data.code === "SUCCESS") {
          this.$buefy.toast.open({ type: "is-success", position: "is-top", duration: 5000,
            message: this.$t("bulkAddByLatency.success", { count: this.serversToAdd.length, group: this.selectedOutbound.toUpperCase() }) });
          this.$emit("close"); this.$emit("done");
        } else {
          this.$buefy.toast.open({ type: "is-warning", position: "is-top", duration: 5000,
            message: res.data.message || this.$t("common.fail") });
        }
      } catch (err) {
        this.$buefy.toast.open({ type: "is-warning", position: "is-top", duration: 5000,
          message: err?.response?.data?.message || err?.message || this.$t("common.fail") });
      } finally { this.saving = false; }
    },
  },
};
</script>

<style lang="scss" scoped>
.bulk-add-preview {
  padding: 10px 12px;
  border-radius: 6px;
  background: #f5f5f5;
  margin-top: 8px;
  font-size: 0.95em;
}
.bulk-add-server-list {
  max-height: 200px;
  overflow-y: auto;
  margin-top: 8px;
  border: 1px solid #eee;
  border-radius: 4px;
}
.bulk-add-server-item {
  display: flex;
  align-items: center;
  padding: 4px 10px;
  font-size: 0.85em;
  gap: 6px;
  border-bottom: 1px solid #f5f5f5;
  &:last-child { border-bottom: none; }
}
.bulk-add-server-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bulk-add-server-latency { color: #4caf50; font-weight: 500; flex-shrink: 0; }
body.theme-dark {
  .bulk-add-preview { background: #2b2930; }
  .bulk-add-server-list { border-color: #444; }
  .bulk-add-server-item { border-bottom-color: #333; }
}
</style>

