// Package acting provides the prototype ActingEngine implementation.
package acting

import (
	"context"
	"fmt"
)

// ToolRegistry dispatches a named tool call with a JSON-encoded input and
// returns the JSON-encoded output. The interface is kept minimal to allow
// easy substitution between the mock (v0.4) and the real capability registry (v0.7).
type ToolRegistry interface {
	// Execute invokes the tool identified by name with the given JSON input.
	// Returns the JSON-encoded output, or an error if the tool is unknown or
	// execution fails.
	//
	// Implementations must respect ctx cancellation.
	Execute(ctx context.Context, name, input string) (string, error)
}

// ToolHandler is a function that handles a specific tool call.
// It receives the raw JSON input and returns raw JSON output.
type ToolHandler func(ctx context.Context, input string) (string, error)

// MockToolRegistry is a deterministic ToolRegistry for use in tests.
// It dispatches calls to registered ToolHandler functions and tracks all
// executions for post-test assertion.
type MockToolRegistry struct {
	// handlers maps tool names to their handler functions.
	handlers map[string]ToolHandler

	// Executions records every (name, input) pair passed to Execute in call order.
	Executions []ToolExecution
}

// ToolExecution captures the details of a single tool call recorded by the mock.
type ToolExecution struct {
	Name  string
	Input string
}

// NewMockToolRegistry creates a MockToolRegistry with no registered handlers.
// Use RegisterTool to add tool handlers before running tests.
func NewMockToolRegistry() *MockToolRegistry {
	return &MockToolRegistry{handlers: make(map[string]ToolHandler)}
}

// RegisterTool registers a ToolHandler for the given tool name.
// Registering the same name twice overwrites the previous handler.
func (m *MockToolRegistry) RegisterTool(name string, handler ToolHandler) {
	m.handlers[name] = handler
}

// Execute dispatches the call to the registered handler. Returns an error if
// no handler is registered for name.
func (m *MockToolRegistry) Execute(ctx context.Context, name, input string) (string, error) {
	m.Executions = append(m.Executions, ToolExecution{Name: name, Input: input})
	handler, ok := m.handlers[name]
	if !ok {
		return "", fmt.Errorf("acting: tool %q not found in mock registry", name)
	}
	return handler(ctx, input)
}
