// Package reasoning provides the prototype ReasoningEngine implementation.
package reasoning

import "context"

// LLMClient is the minimal interface the reasoning engine requires from an LLM
// backend. It abstracts the underlying model client so the engine can be tested
// deterministically without real network calls.
type LLMClient interface {
	// Complete sends the assembled prompt to the LLM and returns its raw text
	// response. Implementations must respect ctx cancellation.
	Complete(ctx context.Context, prompt string) (string, error)
}

// MockLLMClient is a deterministic LLMClient for use in unit tests.
// It returns a pre-configured sequence of responses, cycling back to the last
// response once the sequence is exhausted.
type MockLLMClient struct {
	// Responses is the ordered list of responses to return on successive calls.
	// Must have at least one entry.
	Responses []string

	// CallCount tracks the number of times Complete has been called.
	CallCount int

	// ForcedError, when non-nil, is returned as the error for every Complete call.
	ForcedError error
}

// Complete returns the next response in the configured sequence.
// If ForcedError is set it is returned immediately.
func (m *MockLLMClient) Complete(_ context.Context, _ string) (string, error) {
	m.CallCount++
	if m.ForcedError != nil {
		return "", m.ForcedError
	}
	if len(m.Responses) == 0 {
		return "", nil
	}
	idx := m.CallCount - 1
	if idx >= len(m.Responses) {
		idx = len(m.Responses) - 1 // use last response once exhausted
	}
	return m.Responses[idx], nil
}
