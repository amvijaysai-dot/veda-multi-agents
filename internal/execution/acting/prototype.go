// Package acting provides the prototype ActingEngine implementation for the
// VEDA Agent Runtime execution engine.
//
// In v0.4, the acting engine dispatches tool calls to a ToolRegistry. The mock
// registry returns deterministic results for testing; real capability dispatch
// will be wired via the Capability Registry in v0.7.
//
// Dependency rule: this package imports execution/interfaces only.
// It must not import execution/reasoning, execution/context, or execution/iteration.
package acting

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/execution/interfaces"
)

// PrototypeActingEngine executes tool calls by dispatching them to an injected
// ToolRegistry. All calls are executed and their results are collected regardless
// of individual failures, consistent with the ActingEngine contract.
type PrototypeActingEngine struct {
	registry ToolRegistry
}

// NewPrototypeActingEngine creates and returns a PrototypeActingEngine backed by
// the provided ToolRegistry.
func NewPrototypeActingEngine(registry ToolRegistry) *PrototypeActingEngine {
	if registry == nil {
		panic("acting: ToolRegistry must not be nil")
	}
	return &PrototypeActingEngine{registry: registry}
}

// Act executes all provided tool calls via the ToolRegistry and returns their
// results. Individual tool failures are recorded in ToolResult.Err; a top-level
// error is returned only for systemic issues (context cancellation).
//
// Act implements interfaces.ActingEngine.
func (e *PrototypeActingEngine) Act(
	ctx context.Context,
	calls []interfaces.ToolCall,
) ([]interfaces.ToolResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("acting: context cancelled before execution: %w", err)
	}

	results := make([]interfaces.ToolResult, 0, len(calls))
	for _, call := range calls {
		// Re-check cancellation between tool calls for responsiveness.
		if err := ctx.Err(); err != nil {
			return results, fmt.Errorf("acting: context cancelled during execution: %w", err)
		}

		output, err := e.registry.Execute(ctx, call.ToolName, call.Input)
		results = append(results, interfaces.ToolResult{
			CallID:   call.ID,
			ToolName: call.ToolName,
			Output:   output,
			Err:      err,
		})
	}

	return results, nil
}
