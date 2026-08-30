package service

import (
	"encoding/json"
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

func TestLiveFlowMessageJSONEnvelope(t *testing.T) {
	encoded, err := json.Marshal(LiveFlowMessage{
		Type: "batch_state",
		Data: BatchStateData{TotalActive: 2},
	})
	if err != nil {
		t.Fatalf("marshal live flow message: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal live flow message: %v", err)
	}
	if decoded["type"] != "batch_state" {
		t.Errorf("expected type batch_state, got %#v", decoded["type"])
	}
	if _, ok := decoded["data"]; !ok {
		t.Error("expected data field in live flow message")
	}
}

func TestObservatoryMessageJSONEnvelope(t *testing.T) {
	encoded, err := json.Marshal(ObservatoryMessage{
		ProduceTime: 1788075070,
		Type:        "observatory",
		Body: map[string]interface{}{
			"outboundName": "proxy",
		},
	})
	if err != nil {
		t.Fatalf("marshal observatory message: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal observatory message: %v", err)
	}
	if decoded["type"] != "observatory" {
		t.Errorf("expected type observatory, got %#v", decoded["type"])
	}
	if _, ok := decoded["body"]; !ok {
		t.Error("expected body field in observatory message")
	}
	if _, ok := decoded["data"]; ok {
		t.Error("observatory message must not have a data wrapper")
	}
}
