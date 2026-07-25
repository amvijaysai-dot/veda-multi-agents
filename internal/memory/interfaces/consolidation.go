// Package interfaces defines the contracts for the memory subsystem.
package interfaces

import "context"

// ConsolidationManager handles the movement of data from ephemeral short-term
// memory into persistent long-term storage.
type ConsolidationManager interface {
	// Consolidate processes all short-term memories for the given agent and session.
	// It evaluates which data should be persisted (e.g. via PersistenceHints),
	// scrubs it using the PrivacyManager, and stores it in LongTermMemory.
	// This is typically called at the end of an agent turn or when a session is closed.
	Consolidate(ctx context.Context, agentID, sessionID string) error
}
