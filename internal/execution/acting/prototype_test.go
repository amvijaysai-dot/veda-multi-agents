// Package acting provides the prototype ActingEngine implementation.
package acting

import (
	"context"
	"errors"
	"testing"

	"github.com/veda/agent-runtime/internal/execution/interfaces"
)

// ---------------------------------------------------------------------------
// Compile-time interface satisfaction.
// ---------------------------------------------------------------------------

var _ interfaces.ActingEngine = (*PrototypeActingEngine)(nil)
var _ ToolRegistry = (*MockToolRegistry)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newActingEngine() (*PrototypeActingEngine, *MockToolRegistry) {
	reg := NewMockToolRegistry()
	eng := NewPrototypeActingEngine(reg)
	return eng, reg
}

func echoHandler(_ context.Context, input string) (string, error) {
	return input, nil
}

func failHandler(_ context.Context, _ string) (string, error) {
	return "", errors.New("tool failed")
}

// ---------------------------------------------------------------------------
// NewPrototypeActingEngine
// ---------------------------------------------------------------------------

func TestNew_PanicsOnNilRegistry(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil ToolRegistry, got none")
		}
	}()
	NewPrototypeActingEngine(nil)
}

// ---------------------------------------------------------------------------
// Act — empty calls
// ---------------------------------------------------------------------------

func TestAct_EmptyCallsReturnsEmptyResults(t *testing.T) {
	eng, _ := newActingEngine()
	results, err := eng.Act(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// Act — single successful tool call
// ---------------------------------------------------------------------------

func TestAct_SingleSuccessfulCall(t *testing.T) {
	eng, reg := newActingEngine()
	reg.RegisterTool("echo", echoHandler)

	calls := []interfaces.ToolCall{{ID: "1", ToolName: "echo", Input: `{"msg":"hi"}`}}
	results, err := eng.Act(context.Background(), calls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("unexpected tool error: %v", results[0].Err)
	}
	if results[0].Output != `{"msg":"hi"}` {
		t.Errorf("unexpected output: %q", results[0].Output)
	}
	if results[0].CallID != "1" {
		t.Errorf("expected CallID %q, got %q", "1", results[0].CallID)
	}
}

// ---------------------------------------------------------------------------
// Act — multiple calls, all succeed
// ---------------------------------------------------------------------------

func TestAct_MultipleCallsAllSucceed(t *testing.T) {
	eng, reg := newActingEngine()
	reg.RegisterTool("tool-a", echoHandler)
	reg.RegisterTool("tool-b", echoHandler)

	calls := []interfaces.ToolCall{
		{ID: "1", ToolName: "tool-a", Input: "{}"},
		{ID: "2", ToolName: "tool-b", Input: "{}"},
	}
	results, err := eng.Act(context.Background(), calls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// Act — individual tool failure does not abort remaining calls
// ---------------------------------------------------------------------------

func TestAct_IndividualFailureDoesNotAbort(t *testing.T) {
	eng, reg := newActingEngine()
	reg.RegisterTool("fail", failHandler)
	reg.RegisterTool("echo", echoHandler)

	calls := []interfaces.ToolCall{
		{ID: "1", ToolName: "fail", Input: "{}"},
		{ID: "2", ToolName: "echo", Input: `{"ok":true}`},
	}
	results, err := eng.Act(context.Background(), calls)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected error in first result")
	}
	if results[1].Err != nil {
		t.Errorf("expected success in second result, got: %v", results[1].Err)
	}
}

// ---------------------------------------------------------------------------
// Act — unknown tool
// ---------------------------------------------------------------------------

func TestAct_UnknownToolRecordedAsError(t *testing.T) {
	eng, _ := newActingEngine()
	calls := []interfaces.ToolCall{{ID: "1", ToolName: "unknown", Input: "{}"}}
	results, err := eng.Act(context.Background(), calls)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected error for unknown tool, got nil")
	}
}

// ---------------------------------------------------------------------------
// Act — context cancellation
// ---------------------------------------------------------------------------

func TestAct_CancelledContextBeforeExecutionReturnsError(t *testing.T) {
	eng, reg := newActingEngine()
	reg.RegisterTool("echo", echoHandler)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := eng.Act(ctx, []interfaces.ToolCall{{ID: "1", ToolName: "echo", Input: "{}"}})
	if err == nil {
		t.Fatal("expected error for pre-cancelled context, got nil")
	}
}

// ---------------------------------------------------------------------------
// MockToolRegistry
// ---------------------------------------------------------------------------

func TestMockRegistry_RecordsExecutions(t *testing.T) {
	reg := NewMockToolRegistry()
	reg.RegisterTool("search", echoHandler)

	_, _ = reg.Execute(context.Background(), "search", `{"q":"go"}`)
	_, _ = reg.Execute(context.Background(), "search", `{"q":"rust"}`)

	if len(reg.Executions) != 2 {
		t.Fatalf("expected 2 recorded executions, got %d", len(reg.Executions))
	}
	if reg.Executions[0].Input != `{"q":"go"}` {
		t.Errorf("unexpected first execution input: %q", reg.Executions[0].Input)
	}
}

func TestMockRegistry_UnknownToolReturnsError(t *testing.T) {
	reg := NewMockToolRegistry()
	_, err := reg.Execute(context.Background(), "ghost", "{}")
	if err == nil {
		t.Fatal("expected error for unknown tool, got nil")
	}
}

func TestMockRegistry_OverwritesPreviousHandler(t *testing.T) {
	reg := NewMockToolRegistry()
	reg.RegisterTool("echo", echoHandler)
	reg.RegisterTool("echo", failHandler)

	_, err := reg.Execute(context.Background(), "echo", "{}")
	if err == nil {
		t.Fatal("expected overwritten failHandler to be used, got nil error")
	}
}
