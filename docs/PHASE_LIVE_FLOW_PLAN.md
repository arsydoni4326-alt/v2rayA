# Phase Plan: Live Flow Visualization

## Overview

This document outlines the implementation plan for the **Live Flow Visualization** feature, which will provide real-time, animated visualization of proxy traffic flows in the v2rayA frontend. This feature is part of the Medium-term Roadmap and will enhance monitoring and debugging capabilities for users.

**Status:** Planning Phase  
**Priority:** Medium  
**Estimated Timeline:** 4-6 weeks  
**Dependencies:** WebSocket infrastructure, Vue.js frontend, Go backend service

---

## 1. Feature Summary

### 1.1 Purpose
Live Flow Visualization enables users to see real-time traffic flows through their proxy setup, displaying:
- Source → Proxy → Destination paths
- Traffic volume and speed
- Protocol information
- Connection status and duration
- User attribution (in multi-user scenarios)

### 1.2 Key Benefits
- **Enhanced Monitoring:** Visual understanding of traffic patterns
- **Debugging Aid:** Quickly identify connection issues or routing problems
- **Performance Insight:** Real-time bandwidth and latency visualization
- **Educational Value:** Understand how proxy chains work visually

---

## 2. Technical Architecture

### 2.1 System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     User Interface                         │
│              (Vue.js Live Flow Dashboard)                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ WebSocket Connection
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 WebSocket Server Endpoint                  │
│                    (Go/Gin Backend)                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Session Events
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              V2Ray/XRay Core Integration                  │
│           (Session Tracking & Event Emission)             │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Communication Protocol
- **Transport:** WebSocket (wss:// for production)
- **Message Format:** JSON-encoded messages with consistent schema
#### 3.1.2 WebSocket Endpoint
- **Endpoint:** `/api/live-flow`
- **Authentication:** JWT token validation (existing)
- **Connection Management:**
  - Client registration/deregistration
  - Connection lifecycle management
  - Broadcast to all authenticated clients
  - Graceful disconnection handling

#### 3.1.3 Message Emission
- **On Connection:** Send batch state (all active sessions)
- **On Session Start:** Emit `flow_start` message
- **On Periodic Update:** Emit `flow_update` message (throttled)
- **On Session End:** Emit `flow_end` message

### 3.2 Frontend Requirements

#### 3.2.1 WebSocket Client
- **Connection Management:**
  - Initial connection with authentication
  - Reconnection with exponential backoff
  - Error handling and state management
- **Message Processing:**
  - Parse incoming messages
  - Dispatch to state management
  - Handle initial batch state

#### 3.2.2 State Management
- **Store Structure:**
  - Active flows collection
  - Filter states
  - Connection status
- **Operations:**
  - Add new flow on `flow_start`
  - Update flow on `flow_update`
### 3.3 UI/UX Requirements

#### 3.3.1 Filtering Controls
- By protocol (TCP, UDP, etc.)
- By proxy server
- By user (multi-user scenarios)
- By traffic threshold (minimum speed/bytes)

#### 3.3.2 Visualization Features
- **Node Rendering:**
  - Distinct icons/colors for source/proxy/destination
  - Size scaling based on traffic volume
  - Status indicators (active/idle/error)
- **Path Rendering:**
  - Animated lines/arrows showing flow direction
  - Color coding by protocol/status
  - Width/thickness by traffic volume
  - Particle effects for data movement
- **Animations:**
  - Entry/exit transitions for new/closed flows
  - Pulsing effects for active flows
  - Smooth movement indicators

#### 3.3.3 Responsive Design
- Desktop-first with mobile optimization
- Different layouts for various screen sizes
- Touch-friendly controls for mobile

---

## 4. Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Backend: WebSocket endpoint setup with authentication
- [ ] Backend: Basic session event structure
- [ ] Frontend: WebSocket client service
- [ ] Frontend: Basic state management for flows
- [ ] Documentation: API schema finalization

### Phase 2: Core Visualization (Week 2-3)
---

## 5. Testing Strategy

### 5.1 Unit Tests
- Backend: WebSocket message handling
- Frontend: State management logic
- Frontend: Component rendering

### 5.2 Integration Tests
- WebSocket connection lifecycle
- Message flow from backend to frontend
- Authentication and authorization

### 5.3 Performance Tests
- Concurrent connections (100+ clients)
- Message throughput (1000+ messages/second)
- Memory usage under load
- CPU usage optimization

### 5.4 Security Tests
- Authentication bypass attempts
- Input validation
- Memory leak prevention
- Rate limiting effectiveness

---

## 6. Security Considerations

### 6.1 Authentication & Authorization
- Reuse existing JWT authentication mechanism
- Validate tokens on WebSocket connection
- Implement connection limits per user

---

## 7. Performance Considerations

### 7.1 Backend Optimization
- Throttle high-frequency events (max 1 update/second per flow)
- Implement connection pooling
- Use efficient data structures for session tracking
- Consider event batching for multiple updates

### 7.2 Frontend Optimization
- Use requestAnimationFrame for smooth animations
- Implement virtual scrolling for large flow lists
- Optimize re-renders with Vue.js reactivity
- Use Web Workers for heavy computations if needed

### 7.3 Network Optimization
- Compress WebSocket messages in production
- Implement delta updates for state changes
- Consider binary protocol for high-throughput scenarios

---

## 8. Deployment & Configuration

### 8.1 Configuration Options
```yaml
live_flow:
  enabled: true
  max_connections: 100
---

## 9. Documentation Updates Required

### 9.1 User Documentation
- [ ] Update README.md with Live Flow feature description
- [ ] Create user guide for Live Flow Dashboard
- [ ] Document filtering and customization options
- [ ] Add troubleshooting section

### 9.2 Developer Documentation
- [ ] Update ARCHITECTURE.md with WebSocket components
- [ ] Document API changes in SPECIFICATION.md
- [ ] Add WebSocket schema to CONTRIBUTING.md
- [ ] Update development setup instructions

### 9.3 Operational Documentation
- [ ] Deployment guide updates
- [ ] Configuration reference
- [ ] Monitoring setup
- [ ] Performance tuning guide

---

## 10. Risks & Mitigations

### 10.1 Technical Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| WebSocket scalability issues | High | Medium | Implement connection limits, optimize message handling |
---

## 11. Success Criteria

### 11.1 Functional Requirements
- [ ] Real-time flow visualization with < 1 second latency
- [ ] Support for 100+ concurrent flows
- [ ] Filtering by protocol, proxy, user, and traffic threshold
- [ ] Responsive design for desktop and mobile
- [ ] Graceful degradation under high load

### 11.2 Performance Requirements
- [ ] WebSocket connection establishment < 500ms
- [ ] Message processing < 100ms
- [ ] UI update frame rate > 30fps
- [ ] Memory usage < 50MB for 100 flows
- [ ] CPU usage < 10% during normal operation

### 11.3 Quality Requirements
- [ ] 90%+ test coverage for new code
- [ ] No critical security vulnerabilities
- [ ] Documentation complete and accurate
- [ ] Performance baseline established
- [ ] Monitoring and alerting configured

---

## 12. Appendices

### Appendix A: WebSocket Message Schema
See [LIVE_FLOW_IMPLEMENTATION.md](LIVE_FLOW_IMPLEMENTATION.md) Section 2 for detailed message schema.

### Appendix B: Component Architecture
See [LIVE_FLOW_IMPLEMENTATION.md](LIVE_FLOW_IMPLEMENTATION.md) Section 3 for frontend component details.

### Appendix C: Implementation Checklist
See [LIVE_FLOW_IMPLEMENTATION.md](LIVE_FLOW_IMPLEMENTATION.md) Section 4 for detailed task breakdown.

---

**Document Version:** 1.0  
**Last Updated:** 2026-08-27  
**Author:** Buffy (AI Assistant)  
**Review Status:** Ready for Review
| Core integration complexity | High | Medium | Start with minimal integration, iterate |
| Performance degradation | Medium | Medium | Implement throttling, monitor metrics |
| Security vulnerabilities | High | Low | Security audit, input validation |

### 10.2 Schedule Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Core integration delays | High | Medium | Parallel development, mock services |
| UI complexity underestimation | Medium | Medium | Prototype early, iterative development |
| Testing bottlenecks | Medium | Low | Automated testing, CI/CD integration
  update_interval: 1000  # milliseconds
  throttle_threshold: 100  # messages per second per client
  retention_time: 300  # seconds for inactive flow display
```

### 8.2 Monitoring & Logging
- WebSocket connection metrics
- Message throughput statistics
- Error rate monitoring
- Performance baseline tracking

### 8.3 Feature Flags
- Enable/disable live flow feature
- Adjust performance parameters
- Enable debug logging
- Control visibility in UI
### 6.2 Input Validation
- Validate all incoming WebSocket messages
- Sanitize data before processing
- Prevent injection attacks

### 6.3 Rate Limiting
- Implement per-client rate limiting
- Server-side throttling for high-frequency events
- Prevent denial-of-service attacks

### 6.4 Data Protection
- Encrypt WebSocket traffic (wss://)
- Validate session ownership
- Protect sensitive information in messages
- [ ] Backend: Session tracking integration with core
- [ ] Backend: Message emission logic
- [ ] Frontend: Basic canvas rendering
- [ ] Frontend: Node and path components
- [ ] Frontend: Basic animations

### Phase 3: Advanced Features (Week 3-4)
- [ ] Backend: Throttling and performance optimization
- [ ] Frontend: Filtering controls
- [ ] Frontend: Advanced animations and effects
- [ ] Frontend: Responsive design
- [ ] Testing: Integration testing

### Phase 4: Polish & Documentation (Week 4-6)
- [ ] Performance testing and optimization
- [ ] Security audit
- [ ] Documentation updates
- [ ] User guide creation
- [ ] Release preparation
  - Remove flow on `flow_end`
  - Expiration/cleanup for inactive sessions

#### 3.2.3 Visualization Components
- **Main Dashboard:** `LiveFlowDashboard` container
- **Canvas Area:** `LiveFlowCanvas` for rendering flows
- **Node Components:** Source, proxy, destination nodes
- **Path Components:** Animated connection lines
- **Tooltip Components:** Hover/click details
- **Control Panel:** Filtering and legend
- **Authentication:** JWT token-based (existing mechanism)
- **Versioning:** Protocol version 1.0 (initial)

---

## 3. Implementation Requirements

### 3.1 Backend Requirements

#### 3.1.1 Session Tracking
- **Integration Point:** Hook into V2Ray/XRay core for session events
- **Events to Capture:**
  - Session start (connection established)
  - Session update (periodic stats every 1-5 seconds)
  - Session end (connection closed)
- **Metadata Required:**
  - Unique session ID (UUID)
  - Source IP/port
  - Proxy chain (list of proxies used)
  - Destination IP/port/domain
  - Protocol type (TCP, UDP, etc.)
  - Timestamps (start, last update, end)
  - User information (if authenticated)
  - Traffic statistics (bytes sent/received, speed)