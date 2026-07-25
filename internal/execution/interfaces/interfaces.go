// Package interfaces defines the contracts for all execution engine components
// in the VEDA Agent Runtime.
//
// These interfaces are the frozen public contracts for the ReAct execution loop.
// Implementations live in their respective sub-packages (reasoning, acting, etc.)
// and must not be imported from this package.
//
// Dependency rule: this package depends only on the standard library.
// It must not import any execution implementation package or any internal package
// beyond standard Go types.
package interfaces

// ExecutionInput is the top-level input to the ReAct orchestrator for a single
// agent turn.
type ExecutionInput struct {
	// AgentID is the identifier of the agent executing this turn.
	AgentID string

	// SessionID scopes the execution to a conversational session.
	SessionID string

	// UserMessage is the user-facing input that triggered this turn.
	UserMessage string

	// SystemPrompt is the pre-built system prompt provided by the ContextManager.
	SystemPrompt string

	// MaxIterations caps the number of ReAct loop cycles per turn.
	MaxIterations int
}

// ExecutionResult is the outcome of a complete agent turn.
type ExecutionResult struct {
	// FinalAnswer is the agent's terminal response to the user message.
	FinalAnswer string

	// IterationsUsed records how many ReAct cycles were executed.
	IterationsUsed int

	// ToolResults contains every tool result produced during the turn.
	ToolResults []ToolResult
}

// ToolCall describes a single tool invocation requested by the reasoning engine.
type ToolCall struct {
	// ID is a unique identifier for this call within the current turn.
	ID string

	// ToolName is the name of the tool to invoke.
	ToolName string

	// Input is the raw JSON-encoded input payload for the tool.
	Input string
}

// ToolResult carries the outcome of executing a single ToolCall.
type ToolResult struct {
	// CallID matches the ToolCall.ID this result belongs to.
	CallID string

	// ToolName is echoed from the originating ToolCall.
	ToolName string

	// Output is the raw JSON-encoded output from the tool.
	Output string

	// Err is non-nil if the tool execution failed.
	Err error
}

// ReasoningOutput is the result produced by ReasoningEngine for a single step.
type ReasoningOutput struct {
	// Thought is the model's reasoning text (the "Think" step in ReAct).
	Thought string

	// ToolCalls is the list of tool invocations the model requests.
	// An empty slice means the model has produced a final answer.
	ToolCalls []ToolCall

	// FinalAnswer is non-empty when the model has produced a terminal response
	// and no further ReAct iterations are needed.
	FinalAnswer string
}
