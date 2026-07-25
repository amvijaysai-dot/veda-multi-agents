// Package interfaces defines the contracts for the planner subsystem in the
// VEDA Agent Runtime.
//
// These interfaces govern goal management, plan execution, context exchange,
// and feedback integration between the planner and the execution engine.
//
// Dependency rule: this package depends only on standard library types. It must
// not import any concrete implementation packages.
package interfaces

import "context"

// GoalStatus represents the current state of a goal in the planner system.
type GoalStatus string

const (
	GoalPending   GoalStatus = "PENDING"
	GoalActive    GoalStatus = "ACTIVE"
	GoalCompleted GoalStatus = "COMPLETED"
	GoalFailed    GoalStatus = "FAILED"
	GoalCancelled GoalStatus = "CANCELLED"
)

// Goal represents a target objective that the planner aims to achieve.
type Goal struct {
	ID          string
	Description string
	Status      GoalStatus
}

// GoalFilter defines criteria for filtering goals in ListGoals.
// A zero value in a field means no filtering is applied for that criteria.
type GoalFilter struct {
	Status GoalStatus
}

// GoalManager handles the submission, tracking, and retrieval of agent goals.
type GoalManager interface {
	// SubmitGoal submits a new goal for planning consideration.
	SubmitGoal(ctx context.Context, goal Goal) error

	// UpdateGoal modifies an existing goal (e.g., changing its description or status).
	UpdateGoal(ctx context.Context, goalID string, modifications Goal) error

	// CancelGoal aborts a submitted goal.
	CancelGoal(ctx context.Context, goalID string) error

	// GetGoalStatus returns the current status of the specified goal.
	GetGoalStatus(ctx context.Context, goalID string) (GoalStatus, error)

	// ListGoals returns all goals matching the provided filter criteria.
	ListGoals(ctx context.Context, filter GoalFilter) ([]Goal, error)
}
