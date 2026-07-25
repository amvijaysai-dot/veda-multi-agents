// Package feedback provides the feedback and logging mechanisms for the
// VEDA Agent Runtime planner subsystem.
package feedback

import (
	"context"
	"fmt"
	"sync"

	"github.com/veda/agent-runtime/internal/planner/interfaces"
)

// BasicFeedbackHandler implements interfaces.FeedbackHandler.
// In v0.6, this is a basic implementation that stores outcomes and feedback
// in memory to be queried or logged. In later milestones, this will integrate
// with the observability and vector memory subsystems to form a learning corpus.
type BasicFeedbackHandler struct {
	mu          sync.RWMutex
	outcomes    []interfaces.PlanOutcome
	feedbacks   map[string][]string // planID -> list of feedback strings
	learnings   map[string][]string // planID -> list of learnings
	refinements map[string][]string // goalID -> list of refinements
}

// NewBasicFeedbackHandler creates a new BasicFeedbackHandler.
func NewBasicFeedbackHandler() *BasicFeedbackHandler {
	return &BasicFeedbackHandler{
		outcomes:    make([]interfaces.PlanOutcome, 0),
		feedbacks:   make(map[string][]string),
		learnings:   make(map[string][]string),
		refinements: make(map[string][]string),
	}
}

// ReportPlanOutcome records the outcome of a plan execution.
func (h *BasicFeedbackHandler) ReportPlanOutcome(ctx context.Context, outcome interfaces.PlanOutcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if outcome.PlanID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.outcomes = append(h.outcomes, outcome)
	return nil
}

// ProvidePlanFeedback records qualitative feedback on a plan's execution.
func (h *BasicFeedbackHandler) ProvidePlanFeedback(ctx context.Context, planID string, feedback string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if planID == "" || feedback == "" {
		return fmt.Errorf("plan ID and feedback cannot be empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.feedbacks[planID] = append(h.feedbacks[planID], feedback)
	return nil
}

// LearnFromPlanExecution extracts reusable learnings from a plan execution.
func (h *BasicFeedbackHandler) LearnFromPlanExecution(ctx context.Context, planID string, experience string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if planID == "" || experience == "" {
		return fmt.Errorf("plan ID and experience cannot be empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.learnings[planID] = append(h.learnings[planID], experience)
	return nil
}

// SuggestGoalRefinements suggests improvements to a goal based on planning experience.
func (h *BasicFeedbackHandler) SuggestGoalRefinements(ctx context.Context, goalID string, suggestions []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if goalID == "" || len(suggestions) == 0 {
		return fmt.Errorf("goal ID and suggestions cannot be empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.refinements[goalID] = append(h.refinements[goalID], suggestions...)
	return nil
}

// -- Helpers for testing/querying --

// GetOutcomes returns a copy of all recorded plan outcomes.
func (h *BasicFeedbackHandler) GetOutcomes() []interfaces.PlanOutcome {
	h.mu.RLock()
	defer h.mu.RUnlock()

	outcomes := make([]interfaces.PlanOutcome, len(h.outcomes))
	copy(outcomes, h.outcomes)
	return outcomes
}

// GetFeedbacks returns all feedback provided for a specific plan.
func (h *BasicFeedbackHandler) GetFeedbacks(planID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	fb, exists := h.feedbacks[planID]
	if !exists {
		return nil
	}
	ret := make([]string, len(fb))
	copy(ret, fb)
	return ret
}

var _ interfaces.FeedbackHandler = (*BasicFeedbackHandler)(nil)
