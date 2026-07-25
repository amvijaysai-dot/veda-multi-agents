// Package interfaces defines the contracts for the recovery subsystem.
package interfaces

import (
	"context"

	"github.com/veda/agent-runtime/internal/types/event"
)

// Replayer reconstructs state by applying a sequence of events.
type Replayer interface {
	// Replay applies the given events to reconstruct a state, starting from
	// an optional baseline checkpoint.
	Replay(ctx context.Context, baseline *Checkpoint, events []event.Event) ([]byte, error)
}
