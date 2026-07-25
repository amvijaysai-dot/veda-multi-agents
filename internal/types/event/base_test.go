package event

import (
	"reflect"
	"testing"
	"time"
)

func TestNewBaseEvent(t *testing.T) {
	e := NewBaseEvent("evt-001", TypeAgentCreated, "test-source")

	if e.ID() != "evt-001" {
		t.Errorf("expected ID 'evt-001', got %q", e.ID())
	}
	if e.Type() != TypeAgentCreated {
		t.Errorf("expected Type 'agent.created', got %q", e.Type())
	}
	if e.Source() != "test-source" {
		t.Errorf("expected Source 'test-source', got %q", e.Source())
	}
	if e.Timestamp().IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if e.AgentID() != "" {
		t.Errorf("expected empty AgentID, got %q", e.AgentID())
	}
	if e.SessionID() != "" {
		t.Errorf("expected empty SessionID, got %q", e.SessionID())
	}
	if e.CorrelationID() != "" {
		t.Errorf("expected empty CorrelationID, got %q", e.CorrelationID())
	}
	if e.CausationID() != "" {
		t.Errorf("expected empty CausationID, got %q", e.CausationID())
	}
	if e.Payload() != nil {
		t.Error("expected nil payload")
	}
}

func TestNewBaseEventWithOptions(t *testing.T) {
	payload := map[string]string{"key": "value"}
	e := NewBaseEvent(
		"evt-002",
		TypeTurnStarted,
		"execution",
		WithAgentID("agent-1"),
		WithSessionID("session-1"),
		WithCorrelationID("corr-1"),
		WithCausationID("cause-1"),
		WithPayload(payload),
	)

	if e.AgentID() != "agent-1" {
		t.Errorf("expected AgentID 'agent-1', got %q", e.AgentID())
	}
	if e.SessionID() != "session-1" {
		t.Errorf("expected SessionID 'session-1', got %q", e.SessionID())
	}
	if e.CorrelationID() != "corr-1" {
		t.Errorf("expected CorrelationID 'corr-1', got %q", e.CorrelationID())
	}
	if e.CausationID() != "cause-1" {
		t.Errorf("expected CausationID 'cause-1', got %q", e.CausationID())
	}
	if !reflect.DeepEqual(e.Payload(), payload) {
		t.Error("expected payload to match")
	}
}

func TestBaseEventImplementsEvent(t *testing.T) {
	var _ Event = (*BaseEvent)(nil)
}

func TestTimestampIsUTC(t *testing.T) {
	e := NewBaseEvent("evt-003", TypeAgentCreated, "test")
	loc := e.Timestamp().Location()
	if loc != time.UTC {
		t.Errorf("expected UTC timezone, got %v", loc)
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType Type
		expected  string
	}{
		{TypeAgentCreated, "agent.created"},
		{TypeTurnStarted, "turn.started"},
		{TypeToolInvoked, "tool.invoked"},
		{TypeMemoryStored, "memory.stored"},
		{TypeModelLoaded, "model.loaded"},
		{TypeAccessGranted, "security.access.granted"},
		{TypeCPUExceeded, "resource.cpu.exceeded"},
		{TypeHeartbeat, "health.heartbeat"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
