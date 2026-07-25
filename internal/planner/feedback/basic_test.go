// Package feedback provides the feedback and logging mechanisms.
package feedback

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/veda/agent-runtime/internal/planner/interfaces"
)

func TestBasicFeedbackHandler_ReportPlanOutcome(t *testing.T) {
	fh := NewBasicFeedbackHandler()
	ctx := context.Background()

	err := fh.ReportPlanOutcome(ctx, interfaces.PlanOutcome{
		PlanID:  "p1",
		Success: true,
		Result:  "done",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = fh.ReportPlanOutcome(ctx, interfaces.PlanOutcome{
		PlanID:  "p2",
		Success: false,
		Error:   errors.New("failed"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outcomes := fh.GetOutcomes()
	if len(outcomes) != 2 {
		t.Errorf("expected 2 outcomes, got %d", len(outcomes))
	}
	if outcomes[0].PlanID != "p1" || outcomes[1].PlanID != "p2" {
		t.Error("outcomes recorded incorrectly")
	}

	err = fh.ReportPlanOutcome(ctx, interfaces.PlanOutcome{})
	if err == nil {
		t.Error("expected error for empty plan id")
	}
}

func TestBasicFeedbackHandler_ProvidePlanFeedback(t *testing.T) {
	fh := NewBasicFeedbackHandler()
	ctx := context.Background()

	err := fh.ProvidePlanFeedback(ctx, "p1", "good plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = fh.ProvidePlanFeedback(ctx, "p1", "could be faster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fbs := fh.GetFeedbacks("p1")
	if len(fbs) != 2 {
		t.Fatalf("expected 2 feedbacks, got %d", len(fbs))
	}
	if fbs[0] != "good plan" || fbs[1] != "could be faster" {
		t.Error("feedbacks recorded incorrectly")
	}

	err = fh.ProvidePlanFeedback(ctx, "", "bad")
	if err == nil {
		t.Error("expected error for empty plan id")
	}
}

func TestBasicFeedbackHandler_LearnAndRefine(t *testing.T) {
	fh := NewBasicFeedbackHandler()
	ctx := context.Background()

	err := fh.LearnFromPlanExecution(ctx, "p1", "always use standard library")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = fh.SuggestGoalRefinements(ctx, "g1", []string{"be more specific", "add deadlines"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fh.mu.RLock()
	defer fh.mu.RUnlock()

	if len(fh.learnings["p1"]) != 1 {
		t.Error("learning not recorded")
	}
	if len(fh.refinements["g1"]) != 2 {
		t.Error("refinements not recorded")
	}
}

func TestBasicFeedbackHandler_Concurrency(t *testing.T) {
	fh := NewBasicFeedbackHandler()
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = fh.ReportPlanOutcome(ctx, interfaces.PlanOutcome{PlanID: "p1"})
			_ = fh.ProvidePlanFeedback(ctx, "p1", "fb")
		}()
	}
	wg.Wait()

	if len(fh.GetOutcomes()) != 20 {
		t.Error("concurrent outcomes missed")
	}
	if len(fh.GetFeedbacks("p1")) != 20 {
		t.Error("concurrent feedbacks missed")
	}
}
