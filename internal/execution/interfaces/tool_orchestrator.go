// Package interfaces defines the contracts for all execution engine components.
package interfaces

import "context"

// ToolOrchestrator manages tool execution scheduling and result collection.
// It wraps the ActingEngine to add scheduling, timeout enforcement, and
// result validation logic on top of raw tool dispatch.
//
// In v0.4 the ToolOrchestrator is wired into the orchestrator as a thin wrapper
// around ActingEngine. Richer scheduling will be added in v0.7.
type ToolOrchestrator interface {
	// Orchestrate schedules and executes the provided tool calls, returning
	// their results. It enforces per-call timeouts and validates results before
	// returning them to the caller.
	//
	// Implementations must respect ctx cancellation.
	Orchestrate(ctx context.Context, calls []ToolCall) ([]ToolResult, error)
}
