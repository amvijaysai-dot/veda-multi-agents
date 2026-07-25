// Package interfaces defines the contracts for all execution engine components.
package interfaces

import "context"

// ReasoningEngine generates a reasoning step from the current prompt and context.
// It abstracts over the underlying LLM; concrete implementations use a real model
// client while the prototype uses a deterministic mock.
type ReasoningEngine interface {
	// Reason performs one reasoning step given the system prompt, history, and
	// the latest observation. It returns a ReasoningOutput describing the model's
	// thought, any tool calls it wants to make, and optionally a final answer.
	//
	// Implementations must respect ctx cancellation.
	Reason(ctx context.Context, systemPrompt, history, observation string) (ReasoningOutput, error)
}
