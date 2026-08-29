import Vue from "vue";
import Vuex from "vuex";
import i18n from "../plugins/i18n";

Vue.use(Vuex);
export default new Vuex.Store({
  state: {
    nav: "",
    running: i18n.t("common.checkRunning"),
    connectedServer: {},
    liveFlows: [],
    liveFlowConnected: false,
    liveFlowWarnings: [],
  },
  mutations: {
    NAV(state, val) {
      state.nav = val;
    },
    RUNNING(state, val) {
      state.running = val;
    },
    CONNECTED_SERVER(state, val) {
      state.connectedServer = val;
    },
    SET_LIVE_FLOW_CONNECTED(state, val) {
      state.liveFlowConnected = val;
    },
    SET_LIVE_FLOWS(state, flows) {
      state.liveFlows = flows;
    },
    SET_LIVE_FLOW_WARNINGS(state, warnings) {
      state.liveFlowWarnings = warnings;
    },
    ADD_LIVE_FLOW(state, flow) {
      // Check if flow already exists
      const existingIndex = state.liveFlows.findIndex(
        (f) => f.session_id === flow.session_id
      );
      if (existingIndex === -1) {
        state.liveFlows.push(flow);
      }
    },
    UPDATE_LIVE_FLOW(state, update) {
      const index = state.liveFlows.findIndex(
        (f) => f.session_id === update.session_id
      );
      if (index !== -1) {
        // Update the flow with new data
        Vue.set(state.liveFlows, index, {
          ...state.liveFlows[index],
          bytes_sent: update.bytes_sent,
          bytes_recv: update.bytes_recv,
          speed_bps: update.speed_bps,
          last_activity: update.last_activity,
          status: update.status,
        });
      }
    },
    REMOVE_LIVE_FLOW(state, sessionId) {
      const index = state.liveFlows.findIndex(
        (f) => f.session_id === sessionId
      );
      if (index !== -1) {
        state.liveFlows.splice(index, 1);
      }
    },
  },
  actions: {
    connectLiveFlow({ commit }, token) {
      import("../plugins/liveflow").then((module) => {
        const liveFlowService = module.default;
        liveFlowService.on("connected", () => {
          commit("SET_LIVE_FLOW_CONNECTED", true);
        });
        liveFlowService.on("disconnected", () => {
          commit("SET_LIVE_FLOW_CONNECTED", false);
        });
        liveFlowService.connect(token);
      });
    },
    disconnectLiveFlow({ commit }) {
      import("../plugins/liveflow").then((module) => {
        const liveFlowService = module.default;
        liveFlowService.disconnect();
        commit("SET_LIVE_FLOW_CONNECTED", false);
        commit("SET_LIVE_FLOWS", []);
        commit("SET_LIVE_FLOW_WARNINGS", []);
      });
    },
  },
  modules: {},
});
