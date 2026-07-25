// Package interfaces defines the contracts for the recovery subsystem.
package interfaces

import (
	"context"
)

// Restorer is responsible for finding and loading saved checkpoints.
type Restorer interface {
	// Load retrieves a specific checkpoint by its ID.
	Load(ctx context.Context, checkpointID string) (*Checkpoint, error)
	// LoadLatest retrieves the most recent checkpoint for an agent.
	LoadLatest(ctx context.Context, agentID string) (*Checkpoint, error)
	// List returns all available checkpoints for an agent, ordered by time.
	List(ctx context.Context, agentID string) ([]*Checkpoint, error)
}
