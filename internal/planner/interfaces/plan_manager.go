// Package interfaces defines the contracts for the planner subsystem.
package interfaces

import "context"

// PlanStatus represents the current state of a plan.
type PlanStatus string

const (
	PlanPending   PlanStatus = "PENDING"
	PlanExecuting PlanStatus = "EXECUTING"
	PlanCompleted PlanStatus = "COMPLETED"
	PlanFailed    PlanStatus = "FAILED"
)

// Plan represents a sequence of steps or a structured approach to achieve a Goal.
type Plan struct {
	ID     string
	GoalID string
	Steps  []string
	Status PlanStatus
}

// PlanManager handles the lifecycle, execution, and modification of plans.
type PlanManager interface {
	// GetPlan retrieves a specific plan by its ID.
	GetPlan(ctx context.Context, planID string) (Plan, error)

	// ListPlans returns all plans associated with the specified goal ID.
	ListPlans(ctx context.Context, goalID string) ([]Plan, error)

	// UpdatePlan modifies an existing plan (e.g., adding steps or changing status).
	UpdatePlan(ctx context.Context, planID string, modifications Plan) error

	// ExecutePlan delegates the execution of the plan to the execution engine.
	ExecutePlan(ctx context.Context, planID string) error

	// AdjustPlan modifies a plan in real-time based on execution feedback.
	AdjustPlan(ctx context.Context, planID string, feedback string) error
}
