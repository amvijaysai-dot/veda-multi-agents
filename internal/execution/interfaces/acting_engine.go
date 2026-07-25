// Package interfaces defines the contracts for all execution engine components.
package interfaces

import "context"

// ActingEngine executes tool calls produced by the ReasoningEngine.
// It dispatches each ToolCall to the appropriate handler and returns results.
//
// In v0.4 the acting engine uses a mock tool registry for deterministic testing.
// Real capability dispatch will be wired in v0.7 (Capability Registry milestone).
type ActingEngine interface {
	// Act executes the provided tool calls and returns their results.
	//
	// Implementations must attempt to execute every call even if some fail;
	// individual failures are reported in ToolResult.Err rather than as a
	// top-level error. A top-level error is returned only for systemic issues
	// (e.g. context cancellation, internal panic recovery).
	//
	// Implementations must respect ctx cancellation.
	Act(ctx context.Context, calls []ToolCall) ([]ToolResult, error)
}
