package v2ray

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	// LiveFlowProduct is the ApiFeed product consumed by the Live Flow WebSocket.
	LiveFlowProduct = "live_flow"

	liveFlowCoreEventPrefix = "V2RAYA_LIVE_FLOW "
	liveFlowSessionTTL      = 30 * time.Second
)

// LiveFlowEndpoint identifies one end of a routed connection.
type LiveFlowEndpoint struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Domain string `json:"domain,omitempty"`
}

// LiveFlowProxyNode identifies a selected proxy server or route hop.
type LiveFlowProxyNode struct {
	ProxyID string `json:"proxy_id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Server  string `json:"server"`
}

// LiveFlowStartData is the payload emitted when a route is first observed.
type LiveFlowStartData struct {
	SessionID   string              `json:"session_id"`
	Source      LiveFlowEndpoint    `json:"source"`
	ProxyChain  []LiveFlowProxyNode `json:"proxy_chain"`
	Destination LiveFlowEndpoint    `json:"destination"`
	Protocol    string              `json:"protocol"`
	StartTime   time.Time           `json:"start_time"`
}

// LiveFlowUpdateData is the payload emitted while a route remains active.
type LiveFlowUpdateData struct {
	SessionID    string `json:"session_id"`
	SpeedBPS     uint64 `json:"speed_bps"`
	BytesSent    uint64 `json:"bytes_sent"`
	BytesRecv    uint64 `json:"bytes_recv"`
	LastActivity string `json:"last_activity"`
	Status       string `json:"status"`
}

// LiveFlowEndData is the payload emitted when an inactive route expires.
type LiveFlowEndData struct {
	SessionID      string  `json:"session_id"`
	EndTime        string  `json:"end_time"`
	TotalBytesSent uint64  `json:"total_bytes_sent"`
	TotalBytesRecv uint64  `json:"total_bytes_recv"`
	DurationSecs   float64 `json:"duration_seconds"`
	EndReason      string  `json:"end_reason"`
}

// LiveFlowEvent is an ApiFeed message body and serializes to the documented
// WebSocket envelope: {"type":"flow_*", "data":{...}}.
type LiveFlowEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// CoreLiveFlowEvent is the private record v2raya-core emits after selecting an
// outbound for an accepted access request.
type CoreLiveFlowEvent struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Detour      string `json:"detour,omitempty"`
	Timestamp   int64  `json:"timestamp"`
}

// LiveFlowSession represents a recently observed route. Standard Xray access
// messages do not expose byte counters or a close event, so sessions are
// refreshed by subsequent observations and expire after bounded inactivity.
type LiveFlowSession struct {
	LiveFlowStartData
	LastActivity time.Time
	BytesSent    uint64
	BytesRecv    uint64
	SpeedBPS     uint64
	Status       string
}

// LiveFlowProducer receives core access events and publishes session events.
type LiveFlowProducer struct {
	feed     *Feed
	sessions map[string]*LiveFlowSession
	mu       sync.Mutex
	start    sync.Once
	done     chan struct{}
}

var (
	liveFlowProducer *LiveFlowProducer
	liveFlowOnce     sync.Once
)

// GetLiveFlowProducer returns the singleton route-event producer.
func GetLiveFlowProducer() *LiveFlowProducer {
	liveFlowOnce.Do(func() {
		liveFlowProducer = &LiveFlowProducer{
			feed:     ApiFeed,
			sessions: make(map[string]*LiveFlowSession),
			done:     make(chan struct{}),
		}
	})
	return liveFlowProducer
}

// Start begins the bounded session expiration loop. It is safe to call more
// than once, including once for each WebSocket client.
func (p *LiveFlowProducer) Start() {
	p.start.Do(func() {
		go p.produceLoop()
	})
}

// Stop is reserved for process shutdown.
func (p *LiveFlowProducer) Stop() {
	select {
	case <-p.done:
	default:
		close(p.done)
	}
}

func (p *LiveFlowProducer) produceLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case now := <-ticker.C:
			p.publishUpdates(now.UTC())
		}
	}
}

func (p *LiveFlowProducer) publishUpdates(now time.Time) {
	var events []LiveFlowEvent

	p.mu.Lock()
	for id, session := range p.sessions {
		if now.Sub(session.LastActivity) >= liveFlowSessionTTL {
			events = append(events, LiveFlowEvent{
				Type: "flow_end",
				Data: LiveFlowEndData{
					SessionID:      id,
					EndTime:        now.Format(time.RFC3339Nano),
					TotalBytesSent: session.BytesSent,
					TotalBytesRecv: session.BytesRecv,
					DurationSecs:   now.Sub(session.StartTime).Seconds(),
					EndReason:      "inactive",
				},
			})
			delete(p.sessions, id)
			continue
		}

		events = append(events, LiveFlowEvent{
			Type: "flow_update",
			Data: LiveFlowUpdateData{
				SessionID:    id,
				SpeedBPS:     session.SpeedBPS,
				BytesSent:    session.BytesSent,
				BytesRecv:    session.BytesRecv,
				LastActivity: session.LastActivity.Format(time.RFC3339Nano),
				Status:       session.Status,
			},
		})
	}
	p.mu.Unlock()

	for _, event := range events {
		p.feed.ProductMessage(LiveFlowProduct, event)
	}
}

// RecordCoreAccess creates or refreshes an event from a real Xray routing
// observation. It never fabricates a source, destination, or selected proxy
// from observatory health checks.
func (p *LiveFlowProducer) RecordCoreAccess(event CoreLiveFlowEvent) {
	source, sourceProtocol, ok := parseLiveFlowEndpoint(event.Source)
	if !ok {
		log.Warn("LiveFlow: discarded access event with invalid source")
		return
	}
	destination, destinationProtocol, ok := parseLiveFlowEndpoint(event.Destination)
	if !ok {
		log.Warn("LiveFlow: discarded access event with invalid destination")
		return
	}

	protocol := destinationProtocol
	if protocol == "" {
		protocol = sourceProtocol
	}
	if protocol == "" {
		protocol = "tcp"
	}

	chain := proxyChainFromDetour(event.Detour)
	sessionID := liveFlowSessionID(source, destination, chain, protocol)
	now := time.Now().UTC()
	if event.Timestamp > 0 {
		now = time.UnixMilli(event.Timestamp).UTC()
	}

	p.Start()
	p.mu.Lock()
	if session, exists := p.sessions[sessionID]; exists {
		session.LastActivity = now
		session.Status = "active"
		p.mu.Unlock()
		return
	}

	session := &LiveFlowSession{
		LiveFlowStartData: LiveFlowStartData{
			SessionID:   sessionID,
			Source:      source,
			ProxyChain:  chain,
			Destination: destination,
			Protocol:    protocol,
			StartTime:   now,
		},
		LastActivity: now,
		Status:       "active",
	}
	p.sessions[sessionID] = session
	p.mu.Unlock()

	p.feed.ProductMessage(LiveFlowProduct, LiveFlowEvent{
		Type: "flow_start",
		Data: session.LiveFlowStartData,
	})
}

// Snapshot returns the active sessions required by a newly connected client.
func (p *LiveFlowProducer) Snapshot() []LiveFlowStartData {
	p.mu.Lock()
	defer p.mu.Unlock()

	sessions := make([]LiveFlowStartData, 0, len(p.sessions))
	for _, session := range p.sessions {
		sessions = append(sessions, session.LiveFlowStartData)
	}
	return sessions
}

// ConsumeCoreLiveFlowEvent parses one core-output line. Returning true means
// the line belongs to the private event channel and must not be logged as
// normal core output, even if it is malformed.
func ConsumeCoreLiveFlowEvent(line string) bool {
	if !strings.HasPrefix(line, liveFlowCoreEventPrefix) {
		return false
	}

	var event CoreLiveFlowEvent
	if err := json.Unmarshal([]byte(strings.TrimPrefix(line, liveFlowCoreEventPrefix)), &event); err != nil {
		log.Warn("LiveFlow: discarded malformed core event")
		return true
	}
	if len(event.Source) == 0 || len(event.Source) > 512 || len(event.Destination) == 0 || len(event.Destination) > 512 || len(event.Detour) > 512 {
		log.Warn("LiveFlow: discarded invalid core event")
		return true
	}

	GetLiveFlowProducer().RecordCoreAccess(event)
	return true
}

func parseLiveFlowEndpoint(raw string) (LiveFlowEndpoint, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 512 {
		return LiveFlowEndpoint{}, "", false
	}

	protocol := ""
	for _, candidate := range []string{"tcp:", "udp:"} {
		if strings.HasPrefix(raw, candidate) {
			protocol = strings.TrimSuffix(candidate, ":")
			raw = strings.TrimPrefix(raw, candidate)
			break
		}
	}

	if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
		raw = parsed.Host
	}

	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		host = strings.Trim(raw, "[]")
		port = ""
	}
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	if host == "" || len(host) > 255 {
		return LiveFlowEndpoint{}, "", false
	}

	endpoint := LiveFlowEndpoint{IP: host}
	if port != "" {
		parsedPort, err := strconv.Atoi(port)
		if err != nil || parsedPort < 0 || parsedPort > 65535 {
			return LiveFlowEndpoint{}, "", false
		}
		endpoint.Port = parsedPort
	}
	if net.ParseIP(host) == nil {
		endpoint.Domain = host
	}
	return endpoint, protocol, true
}

func proxyChainFromDetour(detour string) []LiveFlowProxyNode {
	tag := selectedDetourTag(detour)
	if tag == "" {
		return nil
	}

	lowerTag := strings.ToLower(tag)
	if lowerTag == "direct" || lowerTag == "freedom" {
		return []LiveFlowProxyNode{{
			ProxyID: "route:direct",
			Name:    "Direct",
			Type:    "direct",
			Server:  "direct",
		}}
	}

	var outboundMatches []LiveFlowProxyNode
	if connected := configure.GetConnectedServers(); connected != nil {
		for _, which := range connected.Get() {
			raw, err := which.LocateServerRaw()
			if err != nil || raw == nil || raw.ServerObj == nil {
				continue
			}
			server := raw.ServerObj
			name := server.GetName()
			node := LiveFlowProxyNode{
				ProxyID: "server:" + server.GetProtocol() + ":" + server.GetHostname() + ":" + strconv.Itoa(server.GetPort()) + ":" + name,
				Name:    name,
				Type:    server.GetProtocol(),
				Server:  net.JoinHostPort(server.GetHostname(), strconv.Itoa(server.GetPort())),
			}
			if tag == name || tag == GroupWrapper(name) {
				return []LiveFlowProxyNode{node}
			}
			if which.Outbound == tag {
				outboundMatches = append(outboundMatches, node)
			}
		}
	}

	if len(outboundMatches) == 1 {
		return outboundMatches
	}
	return []LiveFlowProxyNode{{
		ProxyID: "route:" + tag,
		Name:    tag,
		Type:    "outbound",
		Server:  tag,
	}}
}

func selectedDetourTag(detour string) string {
	detour = strings.TrimSpace(detour)
	if detour == "" {
		return ""
	}
	normalized := strings.NewReplacer("==>", "->", ">>", "->").Replace(detour)
	parts := strings.Split(normalized, "->")
	return strings.Trim(strings.TrimSpace(parts[len(parts)-1]), " 『』")
}

func liveFlowSessionID(source, destination LiveFlowEndpoint, chain []LiveFlowProxyNode, protocol string) string {
	parts := []string{protocol, source.IP, strconv.Itoa(source.Port), destination.IP, strconv.Itoa(destination.Port)}
	for _, proxy := range chain {
		parts = append(parts, proxy.ProxyID)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "route-" + hex.EncodeToString(sum[:8])
}
