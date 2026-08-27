package service

import (
	"testing"
	"time"
)

func TestLiveFlowSessionCreation(t *testing.T) {
	// Test session creation
	session := &LiveFlowSession{
		SessionID: "test-session-1",
		Source: FlowEndpoint{
			IP:   "192.168.1.100",
			Port: 12345,
		},
		Destination: FlowEndpoint{
			IP:   "10.0.0.1",
			Port: 443,
		},
		Protocol:  "tcp",
		StartTime: time.Now(),
		Status:    "active",
	}

	if session.SessionID != "test-session-1" {
		t.Errorf("Expected session ID 'test-session-1', got '%s'", session.SessionID)
	}

	if session.Source.IP != "192.168.1.100" {
		t.Errorf("Expected source IP '192.168.1.100', got '%s'", session.Source.IP)
	}

	if session.Destination.IP != "10.0.0.1" {
		t.Errorf("Expected destination IP '10.0.0.1', got '%s'", session.Destination.IP)
	}

	if session.Protocol != "tcp" {
		t.Errorf("Expected protocol 'tcp', got '%s'", session.Protocol)
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := GenerateSessionID()
	id2 := GenerateSessionID()

	if id1 == "" {
		t.Error("Expected non-empty session ID")
	}

	if id2 == "" {
		t.Error("Expected non-empty session ID")
	}

	if id1 == id2 {
		t.Error("Expected unique session IDs")
	}
}