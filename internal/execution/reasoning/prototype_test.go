// Package reasoning provides the prototype ReasoningEngine implementation.
package reasoning

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/veda/agent-runtime/internal/execution/interfaces"
)

// ---------------------------------------------------------------------------
// Compile-time interface satisfaction.
// ---------------------------------------------------------------------------

var _ interfaces.ReasoningEngine = (*PrototypeReasoningEngine)(nil)
var _ LLMClient = (*MockLLMClient)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newEngine(responses ...string) *PrototypeReasoningEngine {
	return NewPrototypeReasoningEngine(&MockLLMClient{Responses: responses})
}

// ---------------------------------------------------------------------------
// NewPrototypeReasoningEngine
// ---------------------------------------------------------------------------

func TestNew_PanicsOnNilClient(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil LLMClient, got none")
		}
	}()
	NewPrototypeReasoningEngine(nil)
}

// ---------------------------------------------------------------------------
// Reason — FinalAnswer path
// ---------------------------------------------------------------------------

func TestReason_FinalAnswerResponse(t *testing.T) {
	eng := newEngine("FINAL: The sky is blue.")
	out, err := eng.Reason(context.Background(), "sys", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.FinalAnswer != "The sky is blue." {
		t.Errorf("expected FinalAnswer %q, got %q", "The sky is blue.", out.FinalAnswer)
	}
	if len(out.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %d", len(out.ToolCalls))
	}
}

// ---------------------------------------------------------------------------
// Reason — ToolCall path
// ---------------------------------------------------------------------------

func TestReason_SingleToolCall(t *testing.T) {
	eng := newEngine("TOOL: search|{\"query\":\"golang\"}")
	out, err := eng.Reason(context.Background(), "sys", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(out.ToolCalls))
	}
	if out.ToolCalls[0].ToolName != "search" {
		t.Errorf("expected tool %q, got %q", "search", out.ToolCalls[0].ToolName)
	}
	if out.ToolCalls[0].Input != `{"query":"golang"}` {
		t.Errorf("unexpected input %q", out.ToolCalls[0].Input)
	}
	if out.FinalAnswer != "" {
		t.Error("expected empty FinalAnswer for tool-call response")
	}
}

func TestReason_MultipleToolCalls(t *testing.T) {
	response := "TOOL: search|{\"q\":\"go\"}\nTOOL: calculator|{\"expr\":\"2+2\"}"
	eng := newEngine(response)
	out, err := eng.Reason(context.Background(), "sys", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.ToolCalls) != 2 {
		t.Fatalf("expected 2 tool calls, got %d", len(out.ToolCalls))
	}
	if out.ToolCalls[0].ToolName != "search" {
		t.Errorf("first tool should be search, got %q", out.ToolCalls[0].ToolName)
	}
	if out.ToolCalls[1].ToolName != "calculator" {
		t.Errorf("second tool should be calculator, got %q", out.ToolCalls[1].ToolName)
	}
	// Each call gets a unique ID.
	if out.ToolCalls[0].ID == out.ToolCalls[1].ID {
		t.Error("tool call IDs should be unique")
	}
}

func TestReason_ToolCallWithoutInputDefaultsToEmptyJSON(t *testing.T) {
	eng := newEngine("TOOL: noop")
	out, err := eng.Reason(context.Background(), "sys", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.ToolCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(out.ToolCalls))
	}
	if out.ToolCalls[0].Input != "{}" {
		t.Errorf("expected default input {}, got %q", out.ToolCalls[0].Input)
	}
}

// ---------------------------------------------------------------------------
// Reason — pure thought path
// ---------------------------------------------------------------------------

func TestReason_PureThought(t *testing.T) {
	eng := newEngine("I need to search for information first.")
	out, err := eng.Reason(context.Background(), "sys", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Thought == "" {
		t.Error("expected non-empty Thought")
	}
	if len(out.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %d", len(out.ToolCalls))
	}
	if out.FinalAnswer != "" {
		t.Error("expected empty FinalAnswer")
	}
}

// ---------------------------------------------------------------------------
// Reason — prompt components
// ---------------------------------------------------------------------------

func TestReason_IncludesSystemPromptInCallCount(t *testing.T) {
	mock := &MockLLMClient{Responses: []string{"FINAL: done"}}
	eng := NewPrototypeReasoningEngine(mock)
	_, _ = eng.Reason(context.Background(), "Be helpful.", "prev", "result")
	if mock.CallCount != 1 {
		t.Errorf("expected 1 LLM call, got %d", mock.CallCount)
	}
}

// ---------------------------------------------------------------------------
// Reason — LLM error propagation
// ---------------------------------------------------------------------------

func TestReason_LLMErrorPropagates(t *testing.T) {
	mock := &MockLLMClient{ForcedError: errors.New("model unavailable")}
	eng := NewPrototypeReasoningEngine(mock)
	_, err := eng.Reason(context.Background(), "sys", "", "")
	if err == nil {
		t.Fatal("expected error when LLM fails, got nil")
	}
}

// ---------------------------------------------------------------------------
// Reason — context cancellation
// ---------------------------------------------------------------------------

func TestReason_CancelledContextReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	eng := newEngine("FINAL: ignored")
	_, err := eng.Reason(ctx, "sys", "", "")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

// ---------------------------------------------------------------------------
// MockLLMClient
// ---------------------------------------------------------------------------

func TestMockLLMClient_ReturnsSequenceInOrder(t *testing.T) {
	m := &MockLLMClient{Responses: []string{"first", "second", "third"}}
	for i, want := range []string{"first", "second", "third"} {
		got, err := m.Complete(context.Background(), "prompt")
		if err != nil {
			t.Fatalf("step %d: unexpected error: %v", i, err)
		}
		if got != want {
			t.Errorf("step %d: expected %q, got %q", i, want, got)
		}
	}
}

func TestMockLLMClient_RepeatsLastResponseOnExhaustion(t *testing.T) {
	m := &MockLLMClient{Responses: []string{"only"}}
	for i := 0; i < 5; i++ {
		got, err := m.Complete(context.Background(), "prompt")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if got != "only" {
			t.Errorf("call %d: expected %q, got %q", i, "only", got)
		}
	}
}

func TestMockLLMClient_ForcedErrorOverridesResponse(t *testing.T) {
	m := &MockLLMClient{
		Responses:   []string{"never returned"},
		ForcedError: errors.New("forced"),
	}
	_, err := m.Complete(context.Background(), "prompt")
	if err == nil {
		t.Fatal("expected forced error, got nil")
	}
}

func TestMockLLMClient_TrackCallCount(t *testing.T) {
	m := &MockLLMClient{Responses: []string{"r"}}
	for i := 0; i < 3; i++ {
		_, _ = m.Complete(context.Background(), "p")
	}
	if m.CallCount != 3 {
		t.Errorf("expected CallCount=3, got %d", m.CallCount)
	}
}

// ---------------------------------------------------------------------------
// buildPrompt internal helper
// ---------------------------------------------------------------------------

func TestBuildPrompt_IncludesAllComponents(t *testing.T) {
	prompt := buildPrompt("sys", "hist", "obs")
	if !strings.Contains(prompt, "sys") {
		t.Error("prompt should contain system prompt")
	}
	if !strings.Contains(prompt, "hist") {
		t.Error("prompt should contain history")
	}
	if !strings.Contains(prompt, "obs") {
		t.Error("prompt should contain observation")
	}
}

func TestBuildPrompt_EmptyComponentsOmitted(t *testing.T) {
	prompt := buildPrompt("sys", "", "")
	if strings.Contains(prompt, "History:") {
		t.Error("empty history should not appear in prompt")
	}
	if strings.Contains(prompt, "Observation:") {
		t.Error("empty observation should not appear in prompt")
	}
}
