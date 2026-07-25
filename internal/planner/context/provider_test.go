// Package context provides the planning context exchange mechanism.
package context

import (
	"context"
	"sync"
	"testing"
)

func TestBasicContextProvider_ProvideAndEmpty(t *testing.T) {
	cp := NewBasicContextProvider()
	ctx := context.Background()

	// Provide empty context
	c, err := cp.ProvidePlanningContext(ctx, "agent1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.AgentID != "agent1" {
		t.Errorf("expected agent1, got %s", c.AgentID)
	}
	if len(c.State) != 0 {
		t.Error("expected empty state")
	}

	err = cp.SetActiveGoal(ctx, "agent1", "goal1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, _ = cp.ProvidePlanningContext(ctx, "agent1")
	if c.ActiveGoalID != "goal1" {
		t.Errorf("expected goal1, got %s", c.ActiveGoalID)
	}
}

func TestBasicContextProvider_UpdatePlanningContext(t *testing.T) {
	cp := NewBasicContextProvider()
	ctx := context.Background()

	updates := map[string]string{
		"location": "kitchen",
		"holding":  "apple",
	}

	err := cp.UpdatePlanningContext(ctx, "agent1", updates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, _ := cp.ProvidePlanningContext(ctx, "agent1")
	if len(c.State) != 2 {
		t.Fatalf("expected 2 state entries, got %d", len(c.State))
	}
	if c.State["location"] != "kitchen" {
		t.Errorf("expected kitchen, got %s", c.State["location"])
	}

	// State isolation check: mutating the returned map should not affect internal state
	c.State["location"] = "living_room"

	c2, _ := cp.ProvidePlanningContext(ctx, "agent1")
	if c2.State["location"] != "kitchen" {
		t.Error("expected internal state to be protected from external mutation")
	}
}

func TestBasicContextProvider_Stubs(t *testing.T) {
	cp := NewBasicContextProvider()
	ctx := context.Background()

	caps, err := cp.GetPlanningCapabilities(ctx, "agent1")
	if err != nil || len(caps) == 0 {
		t.Error("failed to get capabilities stub")
	}

	models, err := cp.GetPlanningModels(ctx, "agent1")
	if err != nil || len(models) == 0 {
		t.Error("failed to get models stub")
	}

	res, err := cp.GetPlanningResources(ctx, "agent1")
	if err != nil || len(res) == 0 {
		t.Error("failed to get resources stub")
	}
}

func TestBasicContextProvider_EmptyAgentID(t *testing.T) {
	cp := NewBasicContextProvider()
	ctx := context.Background()

	_, err := cp.ProvidePlanningContext(ctx, "")
	if err == nil {
		t.Error("expected error for empty agentID")
	}

	err = cp.UpdatePlanningContext(ctx, "", map[string]string{})
	if err == nil {
		t.Error("expected error for empty agentID in update")
	}
}

func TestBasicContextProvider_Concurrency(t *testing.T) {
	cp := NewBasicContextProvider()
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cp.UpdatePlanningContext(ctx, "agent1", map[string]string{"k": "v"})
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cp.ProvidePlanningContext(ctx, "agent1")
		}()
	}
	wg.Wait()
}
