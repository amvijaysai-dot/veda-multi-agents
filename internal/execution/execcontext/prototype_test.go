// Package execcontext provides the prototype ContextManager implementation.
package execcontext

import (
	"context"
	"strings"
	"testing"

	"github.com/veda/agent-runtime/internal/execution/interfaces"
)

// ---------------------------------------------------------------------------
// Compile-time interface satisfaction.
// ---------------------------------------------------------------------------

var _ interfaces.ContextManager = (*PrototypeContextManager)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newCM(desc string) *PrototypeContextManager {
	return NewPrototypeContextManager(desc, 0)
}

// ---------------------------------------------------------------------------
// NewPrototypeContextManager — defaults
// ---------------------------------------------------------------------------

func TestNew_ZeroMaxHistoryDefaulted(t *testing.T) {
	cm := NewPrototypeContextManager("desc", 0)
	if cm.maxHistoryEntries != 50 {
		t.Errorf("expected default maxHistoryEntries=50, got %d", cm.maxHistoryEntries)
	}
}

func TestNew_NegativeMaxHistoryDefaulted(t *testing.T) {
	cm := NewPrototypeContextManager("desc", -5)
	if cm.maxHistoryEntries != 50 {
		t.Errorf("expected default maxHistoryEntries=50, got %d", cm.maxHistoryEntries)
	}
}

func TestNew_PositiveMaxHistoryPreserved(t *testing.T) {
	cm := NewPrototypeContextManager("desc", 10)
	if cm.maxHistoryEntries != 10 {
		t.Errorf("expected maxHistoryEntries=10, got %d", cm.maxHistoryEntries)
	}
}

// ---------------------------------------------------------------------------
// BuildPrompt — happy paths
// ---------------------------------------------------------------------------

func TestBuildPrompt_ContainsAgentID(t *testing.T) {
	cm := newCM("")
	prompt, err := cm.BuildPrompt(context.Background(), "agent-007", "session-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "agent-007") {
		t.Error("prompt should contain agent ID")
	}
}

func TestBuildPrompt_ContainsSessionID(t *testing.T) {
	cm := newCM("")
	prompt, err := cm.BuildPrompt(context.Background(), "a1", "my-session", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "my-session") {
		t.Error("prompt should contain session ID")
	}
}

func TestBuildPrompt_ContainsDescription(t *testing.T) {
	cm := newCM("You are a helpful assistant.")
	prompt, err := cm.BuildPrompt(context.Background(), "a1", "s1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "You are a helpful assistant.") {
		t.Error("prompt should contain agent description")
	}
}

func TestBuildPrompt_ContainsToolList(t *testing.T) {
	cm := newCM("")
	tools := []string{"search", "calculator"}
	prompt, err := cm.BuildPrompt(context.Background(), "a1", "s1", tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tool := range tools {
		if !strings.Contains(prompt, tool) {
			t.Errorf("prompt should contain tool %q", tool)
		}
	}
}

func TestBuildPrompt_NoToolsOmitsToolSection(t *testing.T) {
	cm := newCM("")
	prompt, err := cm.BuildPrompt(context.Background(), "a1", "s1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(prompt, "Available tools:") {
		t.Error("prompt should not contain tool section when no tools given")
	}
}

func TestBuildPrompt_IncludesMemoryStub(t *testing.T) {
	cm := newCM("")
	prompt, err := cm.BuildPrompt(context.Background(), "a1", "s1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "Memory") {
		t.Error("prompt should reference memory section (even if stubbed)")
	}
}

// ---------------------------------------------------------------------------
// BuildPrompt — validation
// ---------------------------------------------------------------------------

func TestBuildPrompt_EmptyAgentIDReturnsError(t *testing.T) {
	cm := newCM("")
	_, err := cm.BuildPrompt(context.Background(), "", "s1", nil)
	if err == nil {
		t.Fatal("expected error for empty agentID, got nil")
	}
}

// ---------------------------------------------------------------------------
// BuildPrompt — context cancellation
// ---------------------------------------------------------------------------

func TestBuildPrompt_CancelledContextReturnsError(t *testing.T) {
	cm := newCM("")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := cm.BuildPrompt(ctx, "a1", "s1", nil)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

// ---------------------------------------------------------------------------
// AppendObservation
// ---------------------------------------------------------------------------

func TestAppendObservation_EmptyHistoryReturnsEntry(t *testing.T) {
	cm := newCM("")
	result := interfaces.ToolResult{ToolName: "search", Output: `{"result":"go"}`}
	history := cm.AppendObservation("", result)
	if history == "" {
		t.Error("expected non-empty history after first observation")
	}
	if !strings.Contains(history, "search") {
		t.Error("history should contain tool name")
	}
}

func TestAppendObservation_AppendsToExistingHistory(t *testing.T) {
	cm := newCM("")
	r1 := interfaces.ToolResult{ToolName: "tool-a", Output: "out-a"}
	r2 := interfaces.ToolResult{ToolName: "tool-b", Output: "out-b"}
	h := cm.AppendObservation("", r1)
	h = cm.AppendObservation(h, r2)
	if !strings.Contains(h, "tool-a") {
		t.Error("history should contain first observation")
	}
	if !strings.Contains(h, "tool-b") {
		t.Error("history should contain second observation")
	}
}

func TestAppendObservation_ErrorResultFormatsError(t *testing.T) {
	cm := newCM("")
	errResult := interfaces.ToolResult{ToolName: "bad-tool", Err: context.DeadlineExceeded}
	history := cm.AppendObservation("", errResult)
	if !strings.Contains(history, "Error") {
		t.Error("error result should include 'Error' in history entry")
	}
}

func TestAppendObservation_EnforcesMaxHistoryEntries(t *testing.T) {
	cm := NewPrototypeContextManager("", 3)
	h := ""
	for i := 0; i < 10; i++ {
		h = cm.AppendObservation(h, interfaces.ToolResult{ToolName: "t", Output: "o"})
	}
	lines := strings.Split(h, "\n")
	if len(lines) > 3 {
		t.Errorf("expected at most 3 history lines, got %d", len(lines))
	}
}
