// Package interfaces defines the contracts for the planner subsystem.
package interfaces

import "context"

// PlanProgress represents an incremental update during plan execution.
type PlanProgress struct {
	PlanID       string
	CurrentStep  int
	TotalSteps   int
	Status       PlanStatus
	LatestOutput string
}

// PlanMonitor allows observation of plan execution in real-time.
type PlanMonitor interface {
	// MonitorPlanExecution returns a channel that receives real-time progress
	// updates for the specified plan. The channel is closed when execution finishes.
	MonitorPlanExecution(ctx context.Context, planID string) (<-chan PlanProgress, error)
}
