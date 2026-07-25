// Package plan provides the plan management implementation for the
// VEDA Agent Runtime planner subsystem.
package plan

import (
	"context"
	"fmt"
	"sync"
	"time"

	execinterfaces "github.com/veda/agent-runtime/internal/execution/interfaces"
	"github.com/veda/agent-runtime/internal/planner/interfaces"
)

// ExecutionDelegator represents the execution engine interface required
// by the planner to delegate plan execution.
type ExecutionDelegator interface {
	Execute(ctx context.Context, input execinterfaces.ExecutionInput) (execinterfaces.ExecutionResult, error)
}

// DelegatingPlanManager implements interfaces.PlanManager.
// It tracks plans in memory and delegates execution to the provided
// Execution Engine Delegator.
type DelegatingPlanManager struct {
	mu    sync.RWMutex
	plans map[string]interfaces.Plan

	delegator ExecutionDelegator
}

// NewDelegatingPlanManager creates a new DelegatingPlanManager.
func NewDelegatingPlanManager(delegator ExecutionDelegator) *DelegatingPlanManager {
	if delegator == nil {
		panic("plan: delegator cannot be nil")
	}
	return &DelegatingPlanManager{
		plans:     make(map[string]interfaces.Plan),
		delegator: delegator,
	}
}

// AddPlan is a helper to register a plan. (Not in the interface, but needed for setup).
func (m *DelegatingPlanManager) AddPlan(ctx context.Context, p interfaces.Plan) error {
	if p.ID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plans[p.ID]; exists {
		return fmt.Errorf("plan %q already exists", p.ID)
	}
	m.plans[p.ID] = p
	return nil
}

// GetPlan retrieves a specific plan by its ID.
func (m *DelegatingPlanManager) GetPlan(ctx context.Context, planID string) (interfaces.Plan, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.Plan{}, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.plans[planID]
	if !exists {
		return interfaces.Plan{}, fmt.Errorf("plan %q not found", planID)
	}
	return p, nil
}

// ListPlans returns all plans associated with the specified goal ID.
func (m *DelegatingPlanManager) ListPlans(ctx context.Context, goalID string) ([]interfaces.Plan, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []interfaces.Plan
	for _, p := range m.plans {
		if p.GoalID == goalID {
			result = append(result, p)
		}
	}
	return result, nil
}

// UpdatePlan modifies an existing plan.
func (m *DelegatingPlanManager) UpdatePlan(ctx context.Context, planID string, modifications interfaces.Plan) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.plans[planID]
	if !exists {
		return fmt.Errorf("plan %q not found", planID)
	}

	if modifications.Status != "" {
		p.Status = modifications.Status
	}
	if len(modifications.Steps) > 0 {
		p.Steps = modifications.Steps
	}

	m.plans[planID] = p
	return nil
}

// ExecutePlan delegates the execution of the plan to the execution engine.
// In v0.6, we map a plan execution to a single ReAct loop invocation.
func (m *DelegatingPlanManager) ExecutePlan(ctx context.Context, planID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// 1. Mark as executing
	err := m.UpdatePlan(ctx, planID, interfaces.Plan{Status: interfaces.PlanExecuting})
	if err != nil {
		return err
	}

	// 2. Fetch plan to extract instruction
	p, err := m.GetPlan(ctx, planID)
	if err != nil {
		return err
	}

	// Assemble instruction from steps
	instruction := "Execute plan steps:\n"
	for i, step := range p.Steps {
		instruction += fmt.Sprintf("%d. %s\n", i+1, step)
	}

	// 3. Delegate to execution engine
	input := execinterfaces.ExecutionInput{
		AgentID:       "planner-delegated-agent", // In a real integration, this comes from context
		SessionID:     fmt.Sprintf("plan-session-%s", planID),
		UserMessage:   instruction,
		MaxIterations: 10,
	}

	// We run it synchronously here. In a robust system, this might be async.
	_, execErr := m.delegator.Execute(ctx, input)

	// 4. Update status based on execution outcome
	finalStatus := interfaces.PlanCompleted
	if execErr != nil {
		finalStatus = interfaces.PlanFailed
	}

	// Ignore update error on completion to prioritize returning the execution result
	_ = m.UpdatePlan(ctx, planID, interfaces.Plan{Status: finalStatus})

	return execErr
}

// AdjustPlan modifies a plan in real-time based on execution feedback.
func (m *DelegatingPlanManager) AdjustPlan(ctx context.Context, planID string, feedback string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.plans[planID]
	if !exists {
		return fmt.Errorf("plan %q not found", planID)
	}

	// In v0.6, we simulate adjustment by simply appending the feedback as a new step constraint.
	p.Steps = append(p.Steps, fmt.Sprintf("[Adjusted based on feedback]: %s", feedback))
	m.plans[planID] = p
	return nil
}

// MonitorPlanExecution returns a channel receiving plan progress updates.
func (m *DelegatingPlanManager) MonitorPlanExecution(ctx context.Context, planID string) (<-chan interfaces.PlanProgress, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	_, exists := m.plans[planID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plan %q not found", planID)
	}

	// In v0.6, we return a closed/empty channel as true real-time event-driven
	// progress monitoring requires the EventBus which is an advanced feature.
	// This satisfies the interface contract for now.
	ch := make(chan interfaces.PlanProgress)

	go func() {
		defer close(ch)
		// Simulate immediate startup progress
		select {
		case ch <- interfaces.PlanProgress{
			PlanID:      planID,
			Status:      interfaces.PlanExecuting,
			CurrentStep: 0,
		}:
		case <-ctx.Done():
		case <-time.After(time.Millisecond * 100):
		}
	}()

	return ch, nil
}

// Compile-time interface checks
var (
	_ interfaces.PlanManager = (*DelegatingPlanManager)(nil)
	_ interfaces.PlanMonitor = (*DelegatingPlanManager)(nil)
)
