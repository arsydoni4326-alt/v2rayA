# Live Flow Topology Phase Plan

## Status

**Experimental implementation complete; end-to-end deployment validation pending.**

The roadmap feature is implemented as a Vue-native SVG topology rather than React Flow. v2rayA is a Vue 2 application and React Flow requires a parallel React runtime; the SVG graph supplies the needed nodes, directed edges, animation, interaction, filtering, and responsive layout without introducing that dependency.

## Product Definition

The feature shows an observed proxy route as:

```text
source → selected proxy server / reported proxy chain → destination
```

It is not an observatory health dashboard. Observatory `alive` and latency results remain server-health data and must not be displayed as user traffic.

## Architecture

```text
Xray AccessMessage (after route selection)
        │
        │ V2RAYA_LIVE_FLOW JSON line on core stdout
        ▼
service/kernel/v2ray/process.go
        │ validates and consumes private core events
        ▼
LiveFlowProducer → ApiFeed("live_flow")
        │
        ▼
/api/live-flow WebSocket (JWT, same-origin browser upgrades)
        │
        ▼
Vuex → LiveFlowDashboard SVG topology
```

## Delivered Scope

### Route events

- [x] Structured `flow_start`, `flow_update`, `flow_end`, and `batch_state` messages.
- [x] Source, destination, protocol, timestamp, and selected detour captured from the bundled core.
- [x] Selected server metadata resolved from the connected-server configuration where available.
- [x] Route sessions expire after 30 seconds without a refreshed observation.

### Visualization

- [x] SVG nodes for source, selected proxy/hops, and destination.
- [x] Directed animated paths with active, idle, and error states.
- [x] Shared-node topology layout, proxy/protocol/status filters, and route inspection.
- [x] Dark theme and `prefers-reduced-motion` support.

### Security and reliability

- [x] JWT-protected WebSocket endpoint.
- [x] Request-host browser upgrade policy.
- [x] Hashed-token rate-limit keys; five connection attempts per minute per client.
- [x] Bounded WebSocket reads, pings, and outbound queues.
- [x] Validated, bounded private core-event parser.

## Deliberately Deferred

| Capability | Reason | Follow-up |
|---|---|---|
| Exact per-session byte counters and speed | Standard Xray access events do not carry per-session counters. | Add a core stats/connection wrapper event source. |
| Exact close time/reason | Access events are emitted after routing, without close callbacks. | Add core connection lifecycle instrumentation. |
| Concrete member for every outbound group | Some selected group routes expose only a group tag. | Preserve selected member in core routing event if available. |
| Pan/zoom/drag graph editing | The topology is an observation surface, not a route editor. | Evaluate only if large live graphs require it. |

## Acceptance Criteria

1. A new proxied route appears as `source → proxy → destination` after the core selects its outbound.
2. The proxy node shows the concrete configured server when the detour identifies it; otherwise it shows the actual route/group tag.
3. No observatory health result is represented as a user traffic route.
4. Browser-origin WebSocket requests from another host are rejected.
5. Slow clients cannot cause unbounded event queues or block the shared feed.
6. The feature is documented as route topology; unavailable byte counters are not fabricated.

## Validation

See [LIVE_FLOW_SESSION_RESUME.md](LIVE_FLOW_SESSION_RESUME.md) for commands and deployment checks, and [LIVE_FLOW_SECURITY_CHECKLIST.md](LIVE_FLOW_SECURITY_CHECKLIST.md) for the security review.

**Last updated:** 2026-08-30