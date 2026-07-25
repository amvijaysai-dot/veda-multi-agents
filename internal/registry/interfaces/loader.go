// Package interfaces defines the capability registry contracts.
package interfaces

import "context"

// CapabilitySource represents a location or mechanism from which a capability
// can be loaded (e.g., a file path, a URL, or an embedded asset).
type CapabilitySource struct {
	Type string // e.g., "filesystem", "remote", "builtin"
	URI  string // e.g., "file:///path/to/capability.json"
}

// LoadedCapability represents a capability that has been parsed and is ready
// for validation and registration.
type LoadedCapability struct {
	Metadata CapabilityMetadata
	// Source is where this capability was loaded from.
	Source CapabilitySource
	// ExecutablePath or reference to the binary/script to run.
	ExecutablePath string
}

// CapabilityLoader discovers and loads capabilities from configured sources.
type CapabilityLoader interface {
	// Load discovers and parses a capability from the provided source.
	Load(ctx context.Context, source CapabilitySource) (LoadedCapability, error)

	// Discover scans a root source (e.g., a directory) for all available capabilities
	// and returns their sources.
	Discover(ctx context.Context, rootSource CapabilitySource) ([]CapabilitySource, error)
}
