// Package interfaces defines the capability registry contracts.
package interfaces

import "context"

// CapabilityRegistry is the central repository for available capabilities.
// It tracks registered capabilities and their metadata, handling versioning
// and compatibility checks.
type CapabilityRegistry interface {
	// Register adds a capability to the registry.
	// It returns an error if a capability with the same ID and version already exists.
	Register(ctx context.Context, metadata CapabilityMetadata) error

	// Deregister removes a capability from the registry.
	Deregister(ctx context.Context, id, version string) error

	// Lookup finds a capability by ID.
	// If version is empty, it returns the highest available semantic version.
	Lookup(ctx context.Context, id, version string) (CapabilityMetadata, error)

	// List returns all registered capabilities matching an optional filter.
	// A nil or empty filter returns all capabilities.
	List(ctx context.Context, filter map[string]string) ([]CapabilityMetadata, error)
}
