// Package reasoning provides the prototype ReasoningEngine implementation for
// the VEDA Agent Runtime execution engine.
//
// In v0.4, the reasoning engine interacts with a mock LLM that produces
// deterministic, configurable responses for testing. Real LLM client integration
// (Model Interface) will be wired in a later milestone.
//
// Dependency rule: this package imports execution/interfaces only.
// It must not import execution/acting, execution/context, or execution/iteration.
package reasoning

import (
	"context"
	"fmt"
	"strings"

	"github.com/veda/agent-runtime/internal/execution/interfaces"
)

// PrototypeReasoningEngine is a ReasoningEngine implementation that delegates
// LLM interaction to an injected LLMClient. In production, the client is a real
// model gateway; in tests it is a MockLLMClient for deterministic results.
type PrototypeReasoningEngine struct {
	client LLMClient
}

// NewPrototypeReasoningEngine creates and returns a PrototypeReasoningEngine
// backed by the provided LLMClient.
func NewPrototypeReasoningEngine(client LLMClient) *PrototypeReasoningEngine {
	if client == nil {
		panic("reasoning: LLMClient must not be nil")
	}
	return &PrototypeReasoningEngine{client: client}
}

// Reason performs one reasoning step by constructing a prompt from the inputs
// and delegating to the underlying LLMClient. The response is parsed into a
// ReasoningOutput according to the following conventions:
//
//   - If the response starts with "FINAL:" the remainder is the FinalAnswer.
//   - If the response contains "TOOL:" lines they are parsed as ToolCalls.
//   - Otherwise the whole response is stored as the Thought.
//
// Reason respects ctx cancellation: it returns ctx.Err() wrapped in an error
// if the context is already cancelled before the LLM call begins.
func (e *PrototypeReasoningEngine) Reason(
	ctx context.Context,
	systemPrompt, history, observation string,
) (interfaces.ReasoningOutput, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.ReasoningOutput{}, fmt.Errorf("reasoning: context cancelled: %w", err)
	}

	prompt := buildPrompt(systemPrompt, history, observation)
	response, err := e.client.Complete(ctx, prompt)
	if err != nil {
		return interfaces.ReasoningOutput{}, fmt.Errorf("reasoning: LLM call failed: %w", err)
	}

	return parseResponse(response), nil
}

// buildPrompt assembles the full text prompt from components.
func buildPrompt(systemPrompt, history, observation string) string {
	var b strings.Builder
	if systemPrompt != "" {
		b.WriteString("System: ")
		b.WriteString(systemPrompt)
		b.WriteString("\n\n")
	}
	if history != "" {
		b.WriteString("History:\n")
		b.WriteString(history)
		b.WriteString("\n\n")
	}
	if observation != "" {
		b.WriteString("Observation: ")
		b.WriteString(observation)
		b.WriteString("\n\n")
	}
	b.WriteString("Thought:")
	return b.String()
}

// parseResponse converts the raw LLM text into a structured ReasoningOutput.
//
// Supported response formats:
//
//	"FINAL: <answer>"          → FinalAnswer is set; no ToolCalls.
//	"TOOL: <name>|<input>"     → one ToolCall per such line; Thought is remainder.
//	anything else              → stored verbatim in Thought.
func parseResponse(response string) interfaces.ReasoningOutput {
	response = strings.TrimSpace(response)

	// Check for final answer.
	if after, ok := cutPrefix(response, "FINAL:"); ok {
		return interfaces.ReasoningOutput{
			Thought:     "",
			FinalAnswer: strings.TrimSpace(after),
		}
	}

	// Parse TOOL: lines.
	var calls []interfaces.ToolCall
	var thoughtLines []string
	callID := 1
	for _, line := range strings.Split(response, "\n") {
		trimmed := strings.TrimSpace(line)
		if after, ok := cutPrefix(trimmed, "TOOL:"); ok {
			parts := strings.SplitN(strings.TrimSpace(after), "|", 2)
			name := strings.TrimSpace(parts[0])
			input := "{}"
			if len(parts) == 2 {
				input = strings.TrimSpace(parts[1])
			}
			calls = append(calls, interfaces.ToolCall{
				ID:       fmt.Sprintf("call-%d", callID),
				ToolName: name,
				Input:    input,
			})
			callID++
		} else {
			thoughtLines = append(thoughtLines, line)
		}
	}

	return interfaces.ReasoningOutput{
		Thought:   strings.TrimSpace(strings.Join(thoughtLines, "\n")),
		ToolCalls: calls,
	}
}

// cutPrefix is a Go 1.18-compatible equivalent of strings.CutPrefix (1.20+).
func cutPrefix(s, prefix string) (string, bool) {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):], true
	}
	return "", false
}
