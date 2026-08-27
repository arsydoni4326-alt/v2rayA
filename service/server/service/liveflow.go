package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// LiveFlowSession represents a single proxy flow session
type LiveFlowSession struct {
	SessionID   string            `json:"session_id"`
	Source      FlowEndpoint      `json:"source"`
	ProxyChain  []ProxyNode       `json:"proxy_chain"`
	Destination FlowEndpoint      `json:"destination"`
	Protocol    string            `json:"protocol"`
	StartTime   time.Time         `json:"start_time"`
	LastUpdate  time.Time         `json:"last_update"`
	User        *User             `json:"user,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	BytesSent   uint64            `json:"bytes_sent"`
	BytesRecv   uint64            `json:"bytes_recv"`
	SpeedBPS    uint64            `json:"speed_bps"`
	Status      string            `json:"status"` // active, idle, error
}

// FlowEndpoint represents source or destination endpoint
type FlowEndpoint struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Domain string `json:"domain,omitempty"`
}

// ProxyNode represents a proxy in the chain
type ProxyNode struct {
	ProxyID string `json:"proxy_id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Server  string `json:"server"`
}

// User represents user information
type User struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// LiveFlowMessage represents WebSocket message for live flow
type LiveFlowMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// FlowStartData represents flow start message data
type FlowStartData struct {
	SessionID   string            `json:"session_id"`
	Source      FlowEndpoint      `json:"source"`
	ProxyChain  []ProxyNode       `json:"proxy_chain"`
	Destination FlowEndpoint      `json:"destination"`
	Protocol    string            `json:"protocol"`
	StartTime   time.Time         `json:"start_time"`
	User        *User             `json:"user,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FlowUpdateData represents flow update message data
type FlowUpdateData struct {
	SessionID    string `json:"session_id"`
	SpeedBPS     uint64 `json:"speed_bps"`
	BytesSent    uint64 `json:"bytes_sent"`
	BytesRecv    uint64 `json:"bytes_recv"`
	LastActivity string `json:"last_activity"`
	Status       string `json:"status"`
}

// FlowEndData represents flow end message data
type FlowEndData struct {
	SessionID      string  `json:"session_id"`
	EndTime        string  `json:"end_time"`
	TotalBytesSent uint64  `json:"total_bytes_sent"`
	TotalBytesRecv uint64  `json:"total_bytes_recv"`
	DurationSecs   float64 `json:"duration_seconds"`
	EndReason      string  `json:"end_reason"`
}

// BatchStateData represents batch state message data
type BatchStateData struct {
	Sessions    []FlowStartData `json:"sessions"`
	Timestamp   string          `json:"timestamp"`
	TotalActive int             `json:"total_active"`
}

// LiveFlowHandler manages WebSocket connections for live flow visualization
type LiveFlowHandler struct {
	conn     *websocket.Conn
	sessions map[string]*LiveFlowSession
	mu       sync.RWMutex
	send     chan LiveFlowMessage
	Done     chan struct{}
	// Throttling
	lastUpdate time.Time
	updateInterval time.Duration
}

// NewLiveFlowHandler creates a new live flow handler
func NewLiveFlowHandler(conn *websocket.Conn) *LiveFlowHandler {
	h := &LiveFlowHandler{
		conn:           conn,
		sessions:       make(map[string]*LiveFlowSession),
		send:           make(chan LiveFlowMessage, 256),
		Done:           make(chan struct{}),
		lastUpdate:     time.Now(),
		updateInterval: 100 * time.Millisecond, // 10 updates per second max
	}
	// Subscribe to live flow events from the kernel
	go h.subscribeToLiveFlowEvents()
	return h
}

// Start begins the live flow handler
func (h *LiveFlowHandler) Start() {
	go h.writePump()
	go h.readPump()
}

// AddSession adds a new flow session
func (h *LiveFlowHandler) AddSession(session *LiveFlowSession) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[session.SessionID] = session

	// Send flow start message
	msg := LiveFlowMessage{
		Type: "flow_start",
		Data: FlowStartData{
			SessionID:   session.SessionID,
			Source:      session.Source,
			ProxyChain:  session.ProxyChain,
			Destination: session.Destination,
			Protocol:    session.Protocol,
			StartTime:   session.StartTime,
			User:        session.User,
			Metadata:    session.Metadata,
		},
	}
	h.send <- msg
}

// UpdateSession updates an existing flow session
func (h *LiveFlowHandler) UpdateSession(sessionID string, bytesSent, bytesRecv, speedBPS uint64, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, exists := h.sessions[sessionID]
	if !exists {
		return
	}

	session.BytesSent = bytesSent
	session.BytesRecv = bytesRecv
	session.SpeedBPS = speedBPS
	session.Status = status
	session.LastUpdate = time.Now()

	// Throttle updates
	now := time.Now()
	if now.Sub(h.lastUpdate) < h.updateInterval {
		return
	}
	h.lastUpdate = now

	// Send flow update message
	msg := LiveFlowMessage{
		Type: "flow_update",
		Data: FlowUpdateData{
			SessionID:    sessionID,
			SpeedBPS:     speedBPS,
			BytesSent:    bytesSent,
			BytesRecv:    bytesRecv,
			LastActivity: session.LastUpdate.Format(time.RFC3339),
			Status:       status,
		},
	}
	h.send <- msg
}

// EndSession ends a flow session
// GetActiveSessions returns all active sessions
func (h *LiveFlowHandler) GetActiveSessions() []FlowStartData {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessions := make([]FlowStartData, 0, len(h.sessions))
	for _, session := range h.sessions {
		sessions = append(sessions, FlowStartData{
			SessionID:   session.SessionID,
			Source:      session.Source,
			ProxyChain:  session.ProxyChain,
			Destination: session.Destination,
			Protocol:    session.Protocol,
			StartTime:   session.StartTime,
			User:        session.User,
			Metadata:    session.Metadata,
		})
	}
	return sessions
}

// SendBatchState sends current state to newly connected client
func (h *LiveFlowHandler) SendBatchState() {
	sessions := h.GetActiveSessions()
	msg := LiveFlowMessage{
		Type: "batch_state",
		Data: BatchStateData{
			Sessions:    sessions,
			Timestamp:   time.Now().Format(time.RFC3339),
			TotalActive: len(sessions),
		},
	}
	h.send <- msg
}

// readPump reads messages from the WebSocket connection
func (h *LiveFlowHandler) readPump() {
	defer func() {
		h.Stop()
	}()
	for {
		_, _, err := h.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump writes messages to the WebSocket connection
func (h *LiveFlowHandler) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		h.conn.Close()
	}()

	for {
		select {
		case <-h.Done:
			return
		case message, ok := <-h.send:
			if !ok {
				h.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			h.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := h.conn.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			h.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := h.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GenerateSessionID generates a unique session ID
func GenerateSessionID() string {
	return uuid.New().String()
}

// subscribeToLiveFlowEvents subscribes to live flow events from the kernel
func (h *LiveFlowHandler) subscribeToLiveFlowEvents() {
	// This is a placeholder for future integration with the kernel's LiveFlowProducer
	// For now, we'll just log that we're ready to receive events
	log.Info("LiveFlow handler subscribed to live flow events")
}
func (h *LiveFlowHandler) EndSession(sessionID string, reason string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, exists := h.sessions[sessionID]
	if !exists {
		return
	}

	endTime := time.Now()
	duration := endTime.Sub(session.StartTime).Seconds()

	// Send flow end message
	msg := LiveFlowMessage{
		Type: "flow_end",
		Data: FlowEndData{
			SessionID:      sessionID,
			EndTime:        endTime.Format(time.RFC3339),
			TotalBytesSent: session.BytesSent,
			TotalBytesRecv: session.BytesRecv,
			DurationSecs:   duration,
			EndReason:      reason,
		},
	}
	h.send <- msg

	// Remove session
	delete(h.sessions, sessionID)
}
// Stop stops the live flow handler
func (h *LiveFlowHandler) Stop() {
	select {
	case <-h.Done:
		// Already closed
	default:
		close(h.Done)
	}
	h.conn.Close()
}