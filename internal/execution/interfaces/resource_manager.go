// Package interfaces defines the contracts for all execution engine components.
package interfaces

import "context"

// ResourceSnapshot captures a point-in-time view of resource usage.
type ResourceSnapshot struct {
	// TokensUsed is the cumulative LLM token count for the current turn.
	TokensUsed int

	// ToolCallsUsed is the cumulative number of tool calls in the current turn.
	ToolCallsUsed int
}

// ResourceManager tracks and enforces resource budgets during execution.
// It prevents runaway agents from consuming unbounded LLM tokens or making
// unlimited tool calls.
//
// In v0.4 the resource manager tracks only token and tool-call counts.
// CPU and memory tracking will be integrated in later milestones.
type ResourceManager interface {
	// RecordTokens adds n to the current turn's token usage counter.
	RecordTokens(n int)

	// RecordToolCall increments the current turn's tool-call counter.
	RecordToolCall()

	// Snapshot returns a point-in-time copy of current resource usage.
	Snapshot() ResourceSnapshot

	// Reset clears all usage counters for a new turn.
	Reset(ctx context.Context)

	// CheckBudget returns an error if any resource budget has been exceeded.
	CheckBudget() error
}

// ObservabilityHooks provides callbacks for integrating the execution engine
// with tracing, metrics, and logging systems.
//
// In v0.4 a no-op implementation is used. Real implementations will emit
// OpenTelemetry spans and Prometheus metrics in v0.8.
type ObservabilityHooks interface {
	// OnTurnStart is called at the beginning of each agent turn.
	OnTurnStart(agentID, sessionID string)

	// OnReasoningStep is called after each successful reasoning step.
	OnReasoningStep(agentID string, iteration int, output ReasoningOutput)

	// OnToolCall is called after each tool call completes.
	OnToolCall(agentID string, call ToolCall, result ToolResult)

	// OnTurnEnd is called when the turn completes, whether successfully or not.
	OnTurnEnd(agentID string, result ExecutionResult, err error)
}
