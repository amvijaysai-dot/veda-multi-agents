// Package interfaces defines the contracts for the recovery subsystem.
package interfaces

import (
	"context"
)

// RecoveryCoordinator orchestrates the full recovery lifecycle.
type RecoveryCoordinator interface {
	// Recover attempts to restore an agent to a valid state by loading the latest
	// checkpoint and optionally replaying subsequent events.
	Recover(ctx context.Context, agentID string) ([]byte, error)

	// CreateCheckpoint orchestrates the creation and validation of a new checkpoint.
	CreateCheckpoint(ctx context.Context, agentID string, state []byte, version string) (string, error)
}
