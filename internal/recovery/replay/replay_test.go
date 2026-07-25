package replay

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/recovery/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
)

func TestEventReplayer_Replay(t *testing.T) {
	reducer := func(ctx context.Context, state []byte, e event.Event) ([]byte, error) {
		if e.Type() == "test.fail" {
			return nil, fmt.Errorf("simulated failure")
		}

		res := string(state)
		if res != "" {
			res += ","
		}
		res += e.ID()
		return []byte(res), nil
	}

	replayer := NewEventReplayer(reducer)
	ctx := context.Background()

	// Test with no baseline
	events := []event.Event{
		event.NewBaseEvent("e1", "test.event", "test"),
		event.NewBaseEvent("e2", "test.event", "test"),
	}

	state, err := replayer.Replay(ctx, nil, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(state) != "e1,e2" {
		t.Errorf("expected e1,e2 got %s", string(state))
	}

	// Test with baseline
	baseline := &interfaces.Checkpoint{
		Data:      []byte("base"),
		Timestamp: time.Now(),
	}

	state, err = replayer.Replay(ctx, baseline, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(state) != "base,e1,e2" {
		t.Errorf("expected base,e1,e2 got %s", string(state))
	}

	// Test failure
	events = append(events, event.NewBaseEvent("e3", "test.fail", "test"))
	_, err = replayer.Replay(ctx, baseline, events)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestEventReplayer_NoReducer(t *testing.T) {
	replayer := NewEventReplayer(nil)
	_, err := replayer.Replay(context.Background(), nil, nil)
	if err == nil {
		t.Error("expected error with no reducer")
	}
}
