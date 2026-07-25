package coordinate

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/recovery/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
)

type mockCP struct {
	saveErr error
}

func (m *mockCP) Save(ctx context.Context, agentID string, state []byte, version string) (string, error) {
	if m.saveErr != nil {
		return "", m.saveErr
	}
	return "cp-1", nil
}
func (m *mockCP) Delete(ctx context.Context, checkpointID string) error    { return nil }
func (m *mockCP) DeleteForAgent(ctx context.Context, agentID string) error { return nil }

type mockRT struct {
	cp  *interfaces.Checkpoint
	err error
}

func (m *mockRT) Load(ctx context.Context, checkpointID string) (*interfaces.Checkpoint, error) {
	return m.cp, m.err
}
func (m *mockRT) LoadLatest(ctx context.Context, agentID string) (*interfaces.Checkpoint, error) {
	return m.cp, m.err
}
func (m *mockRT) List(ctx context.Context, agentID string) ([]*interfaces.Checkpoint, error) {
	return []*interfaces.Checkpoint{m.cp}, m.err
}

type mockVD struct {
	err error
}

func (m *mockVD) Validate(ctx context.Context, cp *interfaces.Checkpoint) error {
	return m.err
}

type mockRP struct {
	state []byte
	err   error
}

func (m *mockRP) Replay(ctx context.Context, baseline *interfaces.Checkpoint, events []event.Event) ([]byte, error) {
	return m.state, m.err
}

func TestDefaultCoordinator_Recover(t *testing.T) {
	ctx := context.Background()
	cp := &interfaces.Checkpoint{ID: "cp-1", Data: []byte("base")}

	rt := &mockRT{cp: cp}
	vd := &mockVD{}
	rp := &mockRP{state: []byte("recovered")}

	eventStore := func(ctx context.Context, agentID string, since *interfaces.Checkpoint) ([]event.Event, error) {
		return []event.Event{event.NewBaseEvent("e1", "test", "test")}, nil
	}

	coord := NewDefaultCoordinator(&mockCP{}, rt, vd, rp, eventStore)

	state, err := coord.Recover(ctx, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(state) != "recovered" {
		t.Errorf("expected 'recovered', got %s", string(state))
	}
}

func TestDefaultCoordinator_Recover_NoEvents(t *testing.T) {
	ctx := context.Background()
	cp := &interfaces.Checkpoint{ID: "cp-1", Data: []byte("base")}

	rt := &mockRT{cp: cp}
	vd := &mockVD{}
	rp := &mockRP{state: []byte("base")} // replayer decompresses it

	eventStore := func(ctx context.Context, agentID string, since *interfaces.Checkpoint) ([]event.Event, error) {
		return nil, nil
	}

	coord := NewDefaultCoordinator(&mockCP{}, rt, vd, rp, eventStore)

	state, err := coord.Recover(ctx, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(state) != "base" {
		t.Errorf("expected 'base', got %s", string(state))
	}
}

func TestDefaultCoordinator_CreateCheckpoint(t *testing.T) {
	ctx := context.Background()
	cpMock := &mockCP{}
	rt := &mockRT{
		cp: &interfaces.Checkpoint{
			ID: "cp-1", AgentID: "agent-1", Version: "v1", Timestamp: time.Now(), Data: []byte("data"),
		},
	}
	vd := &mockVD{}

	coord := NewDefaultCoordinator(cpMock, rt, vd, nil, nil)

	id, err := coord.CreateCheckpoint(ctx, "agent-1", []byte("data"), "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "cp-1" {
		t.Errorf("expected cp-1, got %s", id)
	}
}

func TestDefaultCoordinator_CreateCheckpoint_FailValidation(t *testing.T) {
	ctx := context.Background()
	cpMock := &mockCP{}
	rt := &mockRT{
		cp: &interfaces.Checkpoint{
			ID: "cp-1", AgentID: "agent-1", Version: "v1", Timestamp: time.Now(), Data: []byte("data"),
		},
	}
	vd := &mockVD{err: fmt.Errorf("invalid")}

	coord := NewDefaultCoordinator(cpMock, rt, vd, nil, nil)

	_, err := coord.CreateCheckpoint(ctx, "agent-1", []byte("data"), "v1")
	if err == nil {
		t.Fatal("expected error due to validation failure")
	}
}
