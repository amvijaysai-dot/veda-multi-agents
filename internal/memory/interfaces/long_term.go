// Package interfaces defines the contracts for the memory subsystem.
package interfaces

import (
	"context"
	"time"
)

// LongTermMemory manages persistent data that survives across sessions.
type LongTermMemory interface {
	// Store saves persistent data with an optional time-to-live.
	// A zero ttl means the data does not expire.
	Store(ctx context.Context, agentID, key, value string, ttl time.Duration) error

	// Retrieve fetches persistent data by key.
	// Returns an error if the key does not exist or has expired.
	Retrieve(ctx context.Context, agentID, key string) (string, error)

	// Delete removes persistent data by key.
	Delete(ctx context.Context, agentID, key string) error

	// Query performs a simple search against the persistent storage.
	// In v0.5, this is a basic substring match; future milestones will enhance
	// it with semantic search capabilities.
	Query(ctx context.Context, agentID, query string) ([]string, error)

	// Scan returns all keys matching the given prefix.
	Scan(ctx context.Context, agentID, prefix string) ([]string, error)

	// Forget removes persistent data and explicitly logs the reason (e.g. privacy, retention).
	Forget(ctx context.Context, agentID, key, reason string) error
}
