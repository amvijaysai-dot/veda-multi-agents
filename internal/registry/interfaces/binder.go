// Package interfaces defines the capability registry contracts.
package interfaces

import "context"

// BindingContext encapsulates the agent-specific state and permissions
// required to safely execute a capability.
type BindingContext struct {
	AgentID   string
	SessionID string
	// Workspace is the sandbox path for filesystem isolation.
	Workspace string
	// AllowedPermissions is the subset of the agent's permissions granted to this binding.
	AllowedPermissions []string
}

// ExecutableCapability is a capability bound to an agent context, ready to be invoked.
// It wraps the execution logic and translates generic tool invocation into the
// specific protocol (e.g., calling an external binary, making an HTTP request).
type ExecutableCapability interface {
	// Execute runs the capability with the provided JSON payload.
	Execute(ctx context.Context, inputJSON string) (string, error)

	// Close cleans up any resources held by the bound capability.
	Close(ctx context.Context) error
}

// CapabilityBinder prepares a registered capability for execution by an agent.
type CapabilityBinder interface {
	// Bind links a capability to a specific agent context, verifying permissions
	// and preparing the execution sandbox.
	Bind(ctx context.Context, metadata CapabilityMetadata, bindCtx BindingContext) (ExecutableCapability, error)

	// Unbind cleans up a bound capability, releasing sandbox resources.
	Unbind(ctx context.Context, agentID, capabilityID string) error
}
