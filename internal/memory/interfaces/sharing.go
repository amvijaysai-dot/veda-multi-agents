// Package interfaces defines the contracts for the memory subsystem.
package interfaces

import "context"

// SharingManager handles the controlled sharing of memory data between different agents.
type SharingManager interface {
	// Share grants the targetAgentID access to the memory stored at key by agentID.
	// This relies on the security subsystem in later milestones to enforce access control.
	Share(ctx context.Context, agentID, key, targetAgentID string) error
}
