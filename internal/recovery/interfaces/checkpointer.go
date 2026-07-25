// Package interfaces defines the contracts for the recovery subsystem.
package interfaces

import (
	"context"
	"time"
)

// Checkpoint represents a point-in-time snapshot of the application state.
type Checkpoint struct {
	// ID is the unique identifier of the checkpoint.
	ID string `json:"id"`
	// Timestamp is the time the checkpoint was created.
	Timestamp time.Time `json:"timestamp"`
	// AgentID is the identifier of the agent whose state is checkpointed.
	AgentID string `json:"agent_id"`
	// Data contains the serialized state.
	Data []byte `json:"data"`
	// Version identifies the schema version of the data.
	Version string `json:"version"`
	// Metadata contains optional diagnostic or contextual information.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Checkpointer is responsible for creating and saving state checkpoints.
type Checkpointer interface {
	// Save captures the current state for an agent and stores it.
	Save(ctx context.Context, agentID string, state []byte, version string) (string, error)
	// Delete removes a specific checkpoint by ID.
	Delete(ctx context.Context, checkpointID string) error
	// DeleteForAgent removes all checkpoints for a specific agent.
	DeleteForAgent(ctx context.Context, agentID string) error
}
