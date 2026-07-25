// Package interfaces defines the contracts for all execution engine components.
package interfaces

import (
	"context"
	"testing"
)

// ---------------------------------------------------------------------------
// Interface compliance verification via compile-time type assertions.
// Any concrete type that claims to implement these interfaces will fail to
// compile here if its signature drifts from the contract.
// ---------------------------------------------------------------------------

// mockReasoningEngine is a minimal stub that satisfies ReasoningEngine.
type mockReasoningEngine struct{}

func (m *mockReasoningEngine) Reason(_ context.Context, _, _, _ string) (ReasoningOutput, error) {
	return ReasoningOutput{FinalAnswer: "ok"}, nil
}

// mockActingEngine is a minimal stub that satisfies ActingEngine.
type mockActingEngine struct{}

func (m *mockActingEngine) Act(_ context.Context, calls []ToolCall) ([]ToolResult, error) {
	results := make([]ToolResult, len(calls))
	for i, c := range calls {
		results[i] = ToolResult{CallID: c.ID, ToolName: c.ToolName, Output: "{}"}
	}
	return results, nil
}

// mockContextManager is a minimal stub that satisfies ContextManager.
type mockContextManager struct{}

func (m *mockContextManager) BuildPrompt(_ context.Context, _, _ string, _ []string) (string, error) {
	return "system-prompt", nil
}
func (m *mockContextManager) AppendObservation(history string, _ ToolResult) string {
	return history + "\n[observation]"
}

// mockToolOrchestrator is a minimal stub that satisfies ToolOrchestrator.
type mockToolOrchestrator struct{}

func (m *mockToolOrchestrator) Orchestrate(_ context.Context, calls []ToolCall) ([]ToolResult, error) {
	results := make([]ToolResult, len(calls))
	for i, c := range calls {
		results[i] = ToolResult{CallID: c.ID, ToolName: c.ToolName}
	}
	return results, nil
}

// mockIterationController is a minimal stub that satisfies IterationController.
type mockIterationController struct{}

func (m *mockIterationController) Decide(_ ReasoningOutput, _, _ int) IterationDecision {
	return Terminate
}
func (m *mockIterationController) Reset() {}

// mockErrorHandler is a minimal stub that satisfies ErrorHandler.
type mockErrorHandler struct{}

func (m *mockErrorHandler) Classify(_ error) ErrorClassification { return Fatal }
func (m *mockErrorHandler) ShouldRetry(_ error, _ int) bool      { return false }
func (m *mockErrorHandler) RecordFailure()                       {}
func (m *mockErrorHandler) RecordSuccess()                       {}
func (m *mockErrorHandler) IsCircuitOpen() bool                  { return false }

// mockResourceManager is a minimal stub that satisfies ResourceManager.
type mockResourceManager struct{}

func (m *mockResourceManager) RecordTokens(_ int)         {}
func (m *mockResourceManager) RecordToolCall()            {}
func (m *mockResourceManager) Snapshot() ResourceSnapshot { return ResourceSnapshot{} }
func (m *mockResourceManager) Reset(_ context.Context)    {}
func (m *mockResourceManager) CheckBudget() error         { return nil }

// mockObservabilityHooks is a minimal stub that satisfies ObservabilityHooks.
type mockObservabilityHooks struct{}

func (m *mockObservabilityHooks) OnTurnStart(_, _ string)                            {}
func (m *mockObservabilityHooks) OnReasoningStep(_ string, _ int, _ ReasoningOutput) {}
func (m *mockObservabilityHooks) OnToolCall(_ string, _ ToolCall, _ ToolResult)      {}
func (m *mockObservabilityHooks) OnTurnEnd(_ string, _ ExecutionResult, _ error)     {}

// ---------------------------------------------------------------------------
// Compile-time interface satisfaction assertions.
// ---------------------------------------------------------------------------

var (
	_ ReasoningEngine     = (*mockReasoningEngine)(nil)
	_ ActingEngine        = (*mockActingEngine)(nil)
	_ ContextManager      = (*mockContextManager)(nil)
	_ ToolOrchestrator    = (*mockToolOrchestrator)(nil)
	_ IterationController = (*mockIterationController)(nil)
	_ ErrorHandler        = (*mockErrorHandler)(nil)
	_ ResourceManager     = (*mockResourceManager)(nil)
	_ ObservabilityHooks  = (*mockObservabilityHooks)(nil)
)

// ---------------------------------------------------------------------------
// Runtime tests: verify interface method behaviours on mock stubs.
// ---------------------------------------------------------------------------

func TestReasoningEngineInterface_ComplianceAndBehaviour(t *testing.T) {
	eng := &mockReasoningEngine{}
	out, err := eng.Reason(context.Background(), "sys", "hist", "obs")
	if err != nil {
		t.Fatalf("Reason returned unexpected error: %v", err)
	}
	if out.FinalAnswer != "ok" {
		t.Errorf("expected FinalAnswer %q, got %q", "ok", out.FinalAnswer)
	}
}

func TestActingEngineInterface_ComplianceAndBehaviour(t *testing.T) {
	eng := &mockActingEngine{}
	calls := []ToolCall{{ID: "1", ToolName: "search", Input: `{}`}}
	results, err := eng.Act(context.Background(), calls)
	if err != nil {
		t.Fatalf("Act returned unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].CallID != "1" {
		t.Errorf("expected CallID %q, got %q", "1", results[0].CallID)
	}
}

func TestContextManagerInterface_ComplianceAndBehaviour(t *testing.T) {
	cm := &mockContextManager{}
	prompt, err := cm.BuildPrompt(context.Background(), "agent-1", "session-1", []string{"search"})
	if err != nil {
		t.Fatalf("BuildPrompt returned unexpected error: %v", err)
	}
	if prompt == "" {
		t.Error("expected non-empty system prompt")
	}
	updated := cm.AppendObservation("history", ToolResult{Output: "result"})
	if updated == "history" {
		t.Error("expected AppendObservation to modify history")
	}
}

func TestIterationControllerInterface_ComplianceAndBehaviour(t *testing.T) {
	ctrl := &mockIterationController{}
	decision := ctrl.Decide(ReasoningOutput{FinalAnswer: "done"}, 1, 10)
	if decision != Terminate {
		t.Errorf("expected Terminate, got %v", decision)
	}
	ctrl.Reset() // must not panic
}

func TestErrorHandlerInterface_ComplianceAndBehaviour(t *testing.T) {
	h := &mockErrorHandler{}
	h.RecordFailure()
	h.RecordSuccess()
	if h.IsCircuitOpen() {
		t.Error("expected circuit to be closed on mock")
	}
	if h.ShouldRetry(nil, 1) {
		t.Error("expected mock to return false for ShouldRetry")
	}
}

func TestIterationDecision_String(t *testing.T) {
	cases := []struct {
		d    IterationDecision
		want string
	}{
		{Continue, "continue"},
		{Terminate, "terminate"},
		{IterationDecision(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.d.String(); got != tc.want {
			t.Errorf("IterationDecision(%d).String() = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestErrorClassification_String(t *testing.T) {
	cases := []struct {
		c    ErrorClassification
		want string
	}{
		{Recoverable, "recoverable"},
		{Fatal, "fatal"},
		{ErrorClassification(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.c.String(); got != tc.want {
			t.Errorf("ErrorClassification(%d).String() = %q, want %q", tc.c, got, tc.want)
		}
	}
}

func TestResourceManagerInterface_Compliance(t *testing.T) {
	rm := &mockResourceManager{}
	rm.RecordTokens(100)
	rm.RecordToolCall()
	snap := rm.Snapshot()
	if snap.TokensUsed != 0 { // mock returns zero-value
		t.Errorf("expected zero-value snapshot from mock, got %+v", snap)
	}
	rm.Reset(context.Background())
	if err := rm.CheckBudget(); err != nil {
		t.Errorf("expected nil from mock CheckBudget, got: %v", err)
	}
}

func TestObservabilityHooksInterface_Compliance(t *testing.T) {
	// Verify no-op hooks don't panic.
	h := &mockObservabilityHooks{}
	h.OnTurnStart("agent-1", "session-1")
	h.OnReasoningStep("agent-1", 1, ReasoningOutput{})
	h.OnToolCall("agent-1", ToolCall{}, ToolResult{})
	h.OnTurnEnd("agent-1", ExecutionResult{}, nil)
}
