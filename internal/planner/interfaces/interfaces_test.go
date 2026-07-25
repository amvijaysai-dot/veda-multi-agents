// Package interfaces defines the contracts for the planner subsystem.
package interfaces

import (
	"context"
	"testing"
)

type mockGoalManager struct{}

func (m *mockGoalManager) SubmitGoal(_ context.Context, _ Goal) error           { return nil }
func (m *mockGoalManager) UpdateGoal(_ context.Context, _ string, _ Goal) error { return nil }
func (m *mockGoalManager) CancelGoal(_ context.Context, _ string) error         { return nil }
func (m *mockGoalManager) GetGoalStatus(_ context.Context, _ string) (GoalStatus, error) {
	return GoalPending, nil
}
func (m *mockGoalManager) ListGoals(_ context.Context, _ GoalFilter) ([]Goal, error) { return nil, nil }

type mockPlanManager struct{}

func (m *mockPlanManager) GetPlan(_ context.Context, _ string) (Plan, error)      { return Plan{}, nil }
func (m *mockPlanManager) ListPlans(_ context.Context, _ string) ([]Plan, error)  { return nil, nil }
func (m *mockPlanManager) UpdatePlan(_ context.Context, _ string, _ Plan) error   { return nil }
func (m *mockPlanManager) ExecutePlan(_ context.Context, _ string) error          { return nil }
func (m *mockPlanManager) AdjustPlan(_ context.Context, _ string, _ string) error { return nil }

type mockContextProvider struct{}

func (m *mockContextProvider) ProvidePlanningContext(_ context.Context, _ string) (PlanningContext, error) {
	return PlanningContext{}, nil
}
func (m *mockContextProvider) UpdatePlanningContext(_ context.Context, _ string, _ map[string]string) error {
	return nil
}
func (m *mockContextProvider) GetPlanningCapabilities(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockContextProvider) GetPlanningModels(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockContextProvider) GetPlanningResources(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

type mockFeedbackHandler struct{}

func (m *mockFeedbackHandler) ReportPlanOutcome(_ context.Context, _ PlanOutcome) error { return nil }
func (m *mockFeedbackHandler) ProvidePlanFeedback(_ context.Context, _, _ string) error { return nil }
func (m *mockFeedbackHandler) LearnFromPlanExecution(_ context.Context, _, _ string) error {
	return nil
}
func (m *mockFeedbackHandler) SuggestGoalRefinements(_ context.Context, _ string, _ []string) error {
	return nil
}

type mockPlanMonitor struct{}

func (m *mockPlanMonitor) MonitorPlanExecution(_ context.Context, _ string) (<-chan PlanProgress, error) {
	return nil, nil
}

// Compile-time interface compliance assertions.
var (
	_ GoalManager     = (*mockGoalManager)(nil)
	_ PlanManager     = (*mockPlanManager)(nil)
	_ ContextProvider = (*mockContextProvider)(nil)
	_ FeedbackHandler = (*mockFeedbackHandler)(nil)
	_ PlanMonitor     = (*mockPlanMonitor)(nil)
)

func TestInterfaces_Compliance(t *testing.T) {
	// A basic test to ensure the mock stubs can be invoked without panic
	ctx := context.Background()

	var gm GoalManager = &mockGoalManager{}
	_ = gm.SubmitGoal(ctx, Goal{})

	var pm PlanManager = &mockPlanManager{}
	_ = pm.ExecutePlan(ctx, "plan1")

	var cp ContextProvider = &mockContextProvider{}
	_, _ = cp.ProvidePlanningContext(ctx, "agent1")

	var fh FeedbackHandler = &mockFeedbackHandler{}
	_ = fh.ReportPlanOutcome(ctx, PlanOutcome{})

	var mon PlanMonitor = &mockPlanMonitor{}
	_, _ = mon.MonitorPlanExecution(ctx, "plan1")
}
