// Package interfaces defines the contracts for the planner subsystem.
package interfaces

import "context"

// PlanningContext represents the state and environment available to an agent
// during plan generation and evaluation.
type PlanningContext struct {
	AgentID      string
	ActiveGoalID string
	State        map[string]string
}

// ContextProvider facilitates the exchange of state between the planner and
// the rest of the runtime (e.g., memory and capability registries).
type ContextProvider interface {
	// ProvidePlanningContext retrieves the current agent state relevant for planning.
	ProvidePlanningContext(ctx context.Context, agentID string) (PlanningContext, error)

	// UpdatePlanningContext merges updates into the agent's planning context based
	// on plan progress.
	UpdatePlanningContext(ctx context.Context, planID string, updates map[string]string) error

	// GetPlanningCapabilities returns a list of capability names available for
	// the agent to use in plan execution. (In v0.6, this is a stub).
	GetPlanningCapabilities(ctx context.Context, agentID string) ([]string, error)

	// GetPlanningModels returns a list of model names available for plan execution.
	GetPlanningModels(ctx context.Context, agentID string) ([]string, error)

	// GetPlanningResources returns a list of resource names available for plan execution.
	GetPlanningResources(ctx context.Context, agentID string) ([]string, error)
}
