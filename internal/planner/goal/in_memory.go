// Package goal provides the in-memory goal management implementation for the
// VEDA Agent Runtime planner subsystem.
package goal

import (
	"context"
	"fmt"
	"sync"

	"github.com/veda/agent-runtime/internal/planner/interfaces"
)

// InMemoryGoalManager implements interfaces.GoalManager using a thread-safe map.
type InMemoryGoalManager struct {
	mu    sync.RWMutex
	goals map[string]interfaces.Goal
}

// NewInMemoryGoalManager creates a new InMemoryGoalManager.
func NewInMemoryGoalManager() *InMemoryGoalManager {
	return &InMemoryGoalManager{
		goals: make(map[string]interfaces.Goal),
	}
}

// SubmitGoal submits a new goal for planning consideration.
func (m *InMemoryGoalManager) SubmitGoal(ctx context.Context, goal interfaces.Goal) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if goal.ID == "" {
		return fmt.Errorf("goal ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.goals[goal.ID]; exists {
		return fmt.Errorf("goal with ID %q already exists", goal.ID)
	}

	// Default status if not provided
	if goal.Status == "" {
		goal.Status = interfaces.GoalPending
	}

	m.goals[goal.ID] = goal
	return nil
}

// UpdateGoal modifies an existing goal.
func (m *InMemoryGoalManager) UpdateGoal(ctx context.Context, goalID string, modifications interfaces.Goal) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	goal, exists := m.goals[goalID]
	if !exists {
		return fmt.Errorf("goal %q not found", goalID)
	}

	if modifications.Description != "" {
		goal.Description = modifications.Description
	}
	if modifications.Status != "" {
		goal.Status = modifications.Status
	}

	m.goals[goalID] = goal
	return nil
}

// CancelGoal aborts a submitted goal.
func (m *InMemoryGoalManager) CancelGoal(ctx context.Context, goalID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	goal, exists := m.goals[goalID]
	if !exists {
		return fmt.Errorf("goal %q not found", goalID)
	}

	goal.Status = interfaces.GoalCancelled
	m.goals[goalID] = goal
	return nil
}

// GetGoalStatus returns the current status of the specified goal.
func (m *InMemoryGoalManager) GetGoalStatus(ctx context.Context, goalID string) (interfaces.GoalStatus, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	goal, exists := m.goals[goalID]
	if !exists {
		return "", fmt.Errorf("goal %q not found", goalID)
	}

	return goal.Status, nil
}

// ListGoals returns all goals matching the provided filter criteria.
func (m *InMemoryGoalManager) ListGoals(ctx context.Context, filter interfaces.GoalFilter) ([]interfaces.Goal, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []interfaces.Goal
	for _, goal := range m.goals {
		if filter.Status != "" && goal.Status != filter.Status {
			continue
		}
		result = append(result, goal)
	}

	return result, nil
}

var _ interfaces.GoalManager = (*InMemoryGoalManager)(nil)
