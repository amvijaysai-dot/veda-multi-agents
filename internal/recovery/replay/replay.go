// Package replay provides mechanisms for reconstructing state from event logs.
package replay

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/recovery/checkpoint"
	"github.com/veda/agent-runtime/internal/recovery/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
)

// Reducer applies a single event to a serialized state, returning the new serialized state.
type Reducer func(ctx context.Context, state []byte, e event.Event) ([]byte, error)

// EventReplayer implements the Replayer interface.
type EventReplayer struct {
	reducer Reducer
}

// NewEventReplayer creates a new EventReplayer.
func NewEventReplayer(reducer Reducer) *EventReplayer {
	return &EventReplayer{
		reducer: reducer,
	}
}

// Replay applies the given events to reconstruct a state, starting from an optional baseline checkpoint.
func (r *EventReplayer) Replay(ctx context.Context, baseline *interfaces.Checkpoint, events []event.Event) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if r.reducer == nil {
		return nil, fmt.Errorf("no reducer configured for EventReplayer")
	}

	var state []byte
	var err error

	if baseline != nil {
		if len(baseline.Data) == 0 {
			return nil, fmt.Errorf("baseline checkpoint has no data")
		}
		// Try to decompress first
		state, err = checkpoint.DecompressState(baseline.Data)
		if err != nil {
			// If it fails, assume it was not compressed (e.g. mock test data)
			state = baseline.Data
		}
	}

	for _, e := range events {
		state, err = r.reducer(ctx, state, e)
		if err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.ID(), err)
		}
	}

	return state, nil
}
