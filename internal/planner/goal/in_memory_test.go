// Package goal provides the goal management implementation.
package goal

import (
	"context"
	"sync"
	"testing"

	"github.com/veda/agent-runtime/internal/planner/interfaces"
)

func TestInMemoryGoalManager_SubmitAndGet(t *testing.T) {
	gm := NewInMemoryGoalManager()
	ctx := context.Background()

	g := interfaces.Goal{
		ID:          "g1",
		Description: "test goal",
		Status:      interfaces.GoalActive,
	}

	err := gm.SubmitGoal(ctx, g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, err := gm.GetGoalStatus(ctx, "g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != interfaces.GoalActive {
		t.Errorf("expected ACTIVE, got %v", status)
	}
}

func TestInMemoryGoalManager_SubmitDuplicate(t *testing.T) {
	gm := NewInMemoryGoalManager()
	ctx := context.Background()

	g := interfaces.Goal{ID: "g1"}
	_ = gm.SubmitGoal(ctx, g)

	err := gm.SubmitGoal(ctx, g)
	if err == nil {
		t.Error("expected error on duplicate goal submission")
	}
}

func TestInMemoryGoalManager_UpdateGoal(t *testing.T) {
	gm := NewInMemoryGoalManager()
	ctx := context.Background()

	_ = gm.SubmitGoal(ctx, interfaces.Goal{ID: "g1", Status: interfaces.GoalPending})

	err := gm.UpdateGoal(ctx, "g1", interfaces.Goal{
		Description: "updated description",
		Status:      interfaces.GoalCompleted,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gm.mu.RLock()
	g := gm.goals["g1"]
	gm.mu.RUnlock()

	if g.Status != interfaces.GoalCompleted {
		t.Errorf("expected COMPLETED, got %v", g.Status)
	}
	if g.Description != "updated description" {
		t.Errorf("expected updated description, got %v", g.Description)
	}

	err = gm.UpdateGoal(ctx, "unknown", interfaces.Goal{})
	if err == nil {
		t.Error("expected error updating unknown goal")
	}
}

func TestInMemoryGoalManager_CancelGoal(t *testing.T) {
	gm := NewInMemoryGoalManager()
	ctx := context.Background()

	_ = gm.SubmitGoal(ctx, interfaces.Goal{ID: "g1"})

	err := gm.CancelGoal(ctx, "g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, _ := gm.GetGoalStatus(ctx, "g1")
	if status != interfaces.GoalCancelled {
		t.Errorf("expected CANCELLED, got %v", status)
	}
}

func TestInMemoryGoalManager_ListGoals(t *testing.T) {
	gm := NewInMemoryGoalManager()
	ctx := context.Background()

	_ = gm.SubmitGoal(ctx, interfaces.Goal{ID: "g1", Status: interfaces.GoalPending})
	_ = gm.SubmitGoal(ctx, interfaces.Goal{ID: "g2", Status: interfaces.GoalActive})
	_ = gm.SubmitGoal(ctx, interfaces.Goal{ID: "g3", Status: interfaces.GoalPending})

	// List all
	all, err := gm.ListGoals(ctx, interfaces.GoalFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 goals, got %d", len(all))
	}

	// Filter by Pending
	pending, err := gm.ListGoals(ctx, interfaces.GoalFilter{Status: interfaces.GoalPending})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("expected 2 pending goals, got %d", len(pending))
	}
}

func TestInMemoryGoalManager_Concurrency(t *testing.T) {
	gm := NewInMemoryGoalManager()
	ctx := context.Background()
	var wg sync.WaitGroup

	_ = gm.SubmitGoal(ctx, interfaces.Goal{ID: "g1"})

	// Concurrent reads and updates
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = gm.UpdateGoal(ctx, "g1", interfaces.Goal{Status: interfaces.GoalActive})
		}()
	}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = gm.GetGoalStatus(ctx, "g1")
		}()
	}
	wg.Wait()
}
