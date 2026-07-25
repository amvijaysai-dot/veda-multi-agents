// Package interfaces defines the contracts for the memory subsystem in the
// VEDA Agent Runtime.
//
// These interfaces govern short-term (session) and long-term (persistent) memory,
// as well as consolidation, sharing, and privacy (PII scrubbing).
//
// Dependency rule: this package depends only on standard library types. It must
// not import any concrete implementation packages.
package interfaces

import "context"

// ShortTermMemory manages ephemeral, session-scoped data for an agent.
// It is intended for working memory during a single conversation or task execution.
type ShortTermMemory interface {
	// Store saves a value for the given agent and session.
	Store(ctx context.Context, agentID, sessionID, key, value string) error

	// Retrieve fetches a value by key for the given agent and session.
	// Returns an error if the key does not exist.
	Retrieve(ctx context.Context, agentID, sessionID, key string) (string, error)

	// Delete removes a key from the session memory.
	Delete(ctx context.Context, agentID, sessionID, key string) error

	// List returns all keys matching the given prefix for the session.
	List(ctx context.Context, agentID, sessionID, prefix string) ([]string, error)

	// Clear removes all data for the specified session.
	Clear(ctx context.Context, agentID, sessionID string) error

	// PersistenceHint marks a key as important for long-term storage consideration.
	// The consolidation manager will use this hint when evaluating what to keep.
	PersistenceHint(ctx context.Context, agentID, sessionID, key string) error
}
