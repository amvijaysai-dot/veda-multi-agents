// Package interfaces defines the contracts for the capability registry subsystem
// in the VEDA Agent Runtime.
//
// These interfaces govern the discovery, validation, registration, loading, and
// binding of capabilities (tools) to agent instances.
package interfaces

// CapabilityMetadata describes a capability's identity, schema, and dependencies.
type CapabilityMetadata struct {
	// ID is the globally unique identifier for this capability.
	ID string `json:"id"`
	// Version is the semantic version of the capability.
	Version string `json:"version"`
	// Name is the human-readable name.
	Name string `json:"name"`
	// Description provides details on what the capability does.
	Description string `json:"description"`
	// Author represents the creator or maintainer.
	Author string `json:"author"`
	// License is the usage license.
	License string `json:"license"`

	// InputSchema defines the expected JSON schema for invocation arguments.
	InputSchema string `json:"input_schema"`
	// OutputSchema defines the JSON schema for the return value.
	OutputSchema string `json:"output_schema"`

	// RequiredPermissions lists the sandbox permissions required to execute.
	RequiredPermissions []string `json:"required_permissions"`
	// Dependencies lists the IDs of other capabilities or models this depends on.
	Dependencies []string `json:"dependencies"`
}
