package v2ray

import (
	"sync"
	"time"
)

const (
	LiveFlowProduct = "live_flow"
)

// LiveFlowEvent represents a live flow event
type LiveFlowEvent struct {
	Type      string      `json:"type"`
	SessionID string      `json:"session_id"`
	Data      interface{} `json:"data"`
}

// LiveFlowProducer manages live flow event production
type LiveFlowProducer struct {
	feed     *Feed
	sessions map[string]*LiveFlowSession
	mu       sync.RWMutex
	done     chan struct{}
}

// LiveFlowSession represents a tracked session
type LiveFlowSession struct {
	SessionID   string
	StartTime   time.Time
	LastUpdate  time.Time
	BytesSent   uint64
	BytesRecv   uint64
	SpeedBPS    uint64
	Status      string
	SourceIP    string
	SourcePort  int
	DestIP      string
	DestPort    int
	Protocol    string
}

var (
	liveFlowProducer *LiveFlowProducer
	liveFlowOnce     sync.Once
)

// GetLiveFlowProducer returns the singleton LiveFlowProducer
func GetLiveFlowProducer() *LiveFlowProducer {
	liveFlowOnce.Do(func() {
		liveFlowProducer = &LiveFlowProducer{
			feed:     ApiFeed,
			sessions: make(map[string]*LiveFlowSession),
			done:     make(chan struct{}),
		}
		// Register the live_flow product
		liveFlowProducer.feed.RegisterProduct(LiveFlowProduct)
	})
	return liveFlowProducer
}

// Start begins the live flow producer
func (p *LiveFlowProducer) Start() {
	go p.produceLoop()
}

// Stop stops the live flow producer
func (p *LiveFlowProducer) Stop() {
	select {
	case <-p.done:
		// Already closed
	default:
		close(p.done)
	}
}

// produceLoop periodically produces live flow events
func (p *LiveFlowProducer) produceLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.produceUpdates()
		}
	}
}

// produceUpdates produces update events for all active sessions
func (p *LiveFlowProducer) produceUpdates() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for sessionID, session := range p.sessions {
		// Calculate speed (bytes per second)
		now := time.Now()
		elapsed := now.Sub(session.LastUpdate).Seconds()
		if elapsed > 0 {
			session.SpeedBPS = uint64(float64(session.BytesSent+session.BytesRecv) / elapsed)
		}
		session.LastUpdate = now

		// Produce update event
		event := LiveFlowEvent{
			Type:      "flow_update",
			SessionID: sessionID,
			Data: map[string]interface{}{
				"session_id":    sessionID,
				"speed_bps":     session.SpeedBPS,
				"bytes_sent":    session.BytesSent,
				"bytes_recv":    session.BytesRecv,
				"last_activity": session.LastUpdate.Format(time.RFC3339),
				"status":        session.Status,
			},
		}
		p.feed.ProductMessage(LiveFlowProduct, event)
	}
}

// AddSession adds a new session
func (p *LiveFlowProducer) AddSession(session *LiveFlowSession) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sessions[session.SessionID] = session

	// Produce start event
	event := LiveFlowEvent{
		Type:      "flow_start",
		SessionID: session.SessionID,
		Data: map[string]interface{}{
			"session_id": session.SessionID,
			"source": map[string]interface{}{
				"ip":   session.SourceIP,
				"port": session.SourcePort,
			},
			"destination": map[string]interface{}{
				"ip":   session.DestIP,
				"port": session.DestPort,
			},
			"protocol":   session.Protocol,
			"start_time": session.StartTime.Format(time.RFC3339),
		},
	}
	p.feed.ProductMessage(LiveFlowProduct, event)
}

// UpdateSession updates an existing session
func (p *LiveFlowProducer) UpdateSession(sessionID string, bytesSent, bytesRecv uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return
	}

	session.BytesSent = bytesSent
	session.BytesRecv = bytesRecv
	session.LastUpdate = time.Now()
}

// EndSession ends a session
func (p *LiveFlowProducer) EndSession(sessionID string, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return
	}

	endTime := time.Now()
	duration := endTime.Sub(session.StartTime).Seconds()

	// Produce end event
	event := LiveFlowEvent{
		Type:      "flow_end",
		SessionID: sessionID,
		Data: map[string]interface{}{
			"session_id":       sessionID,
			"end_time":         endTime.Format(time.RFC3339),
			"total_bytes_sent": session.BytesSent,
			"total_bytes_recv": session.BytesRecv,
			"duration_seconds": duration,
			"end_reason":       reason,
		},
	}
	p.feed.ProductMessage(LiveFlowProduct, event)

	// Remove session
	delete(p.sessions, sessionID)
}

// GetActiveSessions returns all active sessions
func (p *LiveFlowProducer) GetActiveSessions() []*LiveFlowSession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sessions := make([]*LiveFlowSession, 0, len(p.sessions))
	for _, session := range p.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}