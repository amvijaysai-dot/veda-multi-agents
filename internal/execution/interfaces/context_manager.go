// Package interfaces defines the contracts for all execution engine components.
package interfaces

import "context"

// ContextManager builds and maintains the prompt context for LLM reasoning steps.
//
// In v0.4 the memory integration is stubbed; a basic in-memory history buffer is
// maintained per turn. Real memory reads (ShortTermMemory, LongTermMemory) will be
// wired in v0.5.
type ContextManager interface {
	// BuildPrompt constructs the system prompt for the given agent and session.
	// It incorporates static configuration, the list of available tool names, and
	// any context retrieved from memory (stubbed in v0.4).
	//
	// Implementations must respect ctx cancellation.
	BuildPrompt(ctx context.Context, agentID, sessionID string, tools []string) (string, error)

	// AppendObservation merges a tool result observation into the running history
	// buffer for the current turn and returns the updated history string.
	AppendObservation(history string, result ToolResult) string
}
