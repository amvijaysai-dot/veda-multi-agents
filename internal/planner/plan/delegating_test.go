// Package plan provides plan management implementation.
package plan

import (
	"context"
	"errors"
	"testing"
	"time"

	execinterfaces "github.com/veda/agent-runtime/internal/execution/interfaces"
	"github.com/veda/agent-runtime/internal/planner/interfaces"
)

type mockOrchestrator struct {
	lastInput execinterfaces.ExecutionInput
	err       error
}

func (m *mockOrchestrator) Execute(_ context.Context, input execinterfaces.ExecutionInput) (execinterfaces.ExecutionResult, error) {
	m.lastInput = input
	if m.err != nil {
		return execinterfaces.ExecutionResult{}, m.err
	}
	return execinterfaces.ExecutionResult{FinalAnswer: "success"}, nil
}

func TestNewDelegatingPlanManager_PanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil orchestrator")
		}
	}()
	NewDelegatingPlanManager(nil)
}

func TestDelegatingPlanManager_CRUD(t *testing.T) {
	ctx := context.Background()
	orch := &mockOrchestrator{}
	pm := NewDelegatingPlanManager(orch)

	p := interfaces.Plan{
		ID:     "p1",
		GoalID: "g1",
		Status: interfaces.PlanPending,
	}

	err := pm.AddPlan(ctx, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get
	fetched, err := pm.GetPlan(ctx, "p1")
	if err != nil || fetched.ID != "p1" {
		t.Error("failed to get plan")
	}

	// List
	_ = pm.AddPlan(ctx, interfaces.Plan{ID: "p2", GoalID: "g1"})
	_ = pm.AddPlan(ctx, interfaces.Plan{ID: "p3", GoalID: "g2"})

	plans, err := pm.ListPlans(ctx, "g1")
	if err != nil || len(plans) != 2 {
		t.Error("failed to list plans correctly")
	}

	// Update
	err = pm.UpdatePlan(ctx, "p1", interfaces.Plan{Status: interfaces.PlanExecuting})
	if err != nil {
		t.Error("failed to update plan")
	}
	fetched, _ = pm.GetPlan(ctx, "p1")
	if fetched.Status != interfaces.PlanExecuting {
		t.Error("expected status to be executing")
	}
}

func TestDelegatingPlanManager_ExecutePlan_Success(t *testing.T) {
	ctx := context.Background()
	orch := &mockOrchestrator{}
	pm := NewDelegatingPlanManager(orch)

	_ = pm.AddPlan(ctx, interfaces.Plan{
		ID:     "p1",
		Steps:  []string{"step 1", "step 2"},
		Status: interfaces.PlanPending,
	})

	err := pm.ExecutePlan(ctx, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify orchestrator received correct input
	if orch.lastInput.UserMessage == "" {
		t.Error("expected orchestrator to receive steps in user message")
	}

	// Verify plan status updated to completed
	fetched, _ := pm.GetPlan(ctx, "p1")
	if fetched.Status != interfaces.PlanCompleted {
		t.Errorf("expected COMPLETED, got %v", fetched.Status)
	}
}

func TestDelegatingPlanManager_ExecutePlan_Failure(t *testing.T) {
	ctx := context.Background()
	orch := &mockOrchestrator{err: errors.New("execution failed")}
	pm := NewDelegatingPlanManager(orch)

	_ = pm.AddPlan(ctx, interfaces.Plan{ID: "p1", Status: interfaces.PlanPending})

	err := pm.ExecutePlan(ctx, "p1")
	if err == nil {
		t.Error("expected execution error to be propagated")
	}

	// Verify plan status updated to failed
	fetched, _ := pm.GetPlan(ctx, "p1")
	if fetched.Status != interfaces.PlanFailed {
		t.Errorf("expected FAILED, got %v", fetched.Status)
	}
}

func TestDelegatingPlanManager_AdjustPlan(t *testing.T) {
	ctx := context.Background()
	orch := &mockOrchestrator{}
	pm := NewDelegatingPlanManager(orch)

	_ = pm.AddPlan(ctx, interfaces.Plan{ID: "p1", Steps: []string{"step 1"}})

	err := pm.AdjustPlan(ctx, "p1", "do it better")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, _ := pm.GetPlan(ctx, "p1")
	if len(fetched.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(fetched.Steps))
	}
	if fetched.Steps[1] != "[Adjusted based on feedback]: do it better" {
		t.Error("feedback step not appended correctly")
	}
}

func TestDelegatingPlanManager_MonitorPlanExecution(t *testing.T) {
	ctx := context.Background()
	orch := &mockOrchestrator{}
	pm := NewDelegatingPlanManager(orch)

	_ = pm.AddPlan(ctx, interfaces.Plan{ID: "p1"})

	ch, err := pm.MonitorPlanExecution(ctx, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case progress, ok := <-ch:
		if !ok {
			t.Fatal("channel closed before progress received")
		}
		if progress.PlanID != "p1" {
			t.Error("wrong plan id in progress")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for simulated progress")
	}
}
