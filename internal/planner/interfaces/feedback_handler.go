// Package interfaces defines the contracts for the planner subsystem.
package interfaces

import "context"

// PlanOutcome represents the final result of executing a plan.
type PlanOutcome struct {
	PlanID  string
	Success bool
	Result  string
	Error   error
}

// FeedbackHandler processes execution outcomes and feedback to improve future planning.
type FeedbackHandler interface {
	// ReportPlanOutcome records the outcome of a plan execution.
	ReportPlanOutcome(ctx context.Context, outcome PlanOutcome) error

	// ProvidePlanFeedback records qualitative feedback on a plan's execution.
	ProvidePlanFeedback(ctx context.Context, planID string, feedback string) error

	// LearnFromPlanExecution extracts reusable learnings from a plan execution.
	LearnFromPlanExecution(ctx context.Context, planID string, experience string) error

	// SuggestGoalRefinements suggests improvements to a goal based on planning experience.
	SuggestGoalRefinements(ctx context.Context, goalID string, suggestions []string) error
}
