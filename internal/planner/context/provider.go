// Package context provides the planning context exchange mechanism for the
// VEDA Agent Runtime planner subsystem.
package context

import (
	"context"
	"fmt"
	"sync"

	"github.com/veda/agent-runtime/internal/planner/interfaces"
)

// BasicContextProvider implements interfaces.ContextProvider.
// It maintains the planning state for agents and provides stubs for registries
// that will be introduced in later milestones.
type BasicContextProvider struct {
	mu       sync.RWMutex
	contexts map[string]interfaces.PlanningContext
}

// NewBasicContextProvider creates a new BasicContextProvider.
func NewBasicContextProvider() *BasicContextProvider {
	return &BasicContextProvider{
		contexts: make(map[string]interfaces.PlanningContext),
	}
}

// ProvidePlanningContext retrieves the current agent state relevant for planning.
func (p *BasicContextProvider) ProvidePlanningContext(ctx context.Context, agentID string) (interfaces.PlanningContext, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.PlanningContext{}, err
	}
	if agentID == "" {
		return interfaces.PlanningContext{}, fmt.Errorf("agentID cannot be empty")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	c, exists := p.contexts[agentID]
	if !exists {
		// Return an empty context initialized for the agent
		return interfaces.PlanningContext{
			AgentID: agentID,
			State:   make(map[string]string),
		}, nil
	}

	// Create a deep copy of the state map to prevent external mutation
	stateCopy := make(map[string]string, len(c.State))
	for k, v := range c.State {
		stateCopy[k] = v
	}

	return interfaces.PlanningContext{
		AgentID:      c.AgentID,
		ActiveGoalID: c.ActiveGoalID,
		State:        stateCopy,
	}, nil
}

// UpdatePlanningContext merges updates into the agent's planning context.
// In this basic implementation, we tie context updates to the agent ID rather than
// complex plan-scoped isolation.
func (p *BasicContextProvider) UpdatePlanningContext(ctx context.Context, agentID string, updates map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if agentID == "" {
		return fmt.Errorf("agentID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	c, exists := p.contexts[agentID]
	if !exists {
		c = interfaces.PlanningContext{
			AgentID: agentID,
			State:   make(map[string]string),
		}
	}
	if c.State == nil {
		c.State = make(map[string]string)
	}

	for k, v := range updates {
		c.State[k] = v
	}

	p.contexts[agentID] = c
	return nil
}

// GetPlanningCapabilities returns a list of capability names available.
// Stubbed for v0.6 until Capability Registry (v0.7) is implemented.
func (p *BasicContextProvider) GetPlanningCapabilities(ctx context.Context, agentID string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []string{"file_system_read", "file_system_write", "http_request"}, nil
}

// GetPlanningModels returns a list of model names available.
// Stubbed for v0.6 until Model Integration (v0.8) is implemented.
func (p *BasicContextProvider) GetPlanningModels(ctx context.Context, agentID string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []string{"gpt-4-stub", "claude-3-stub"}, nil
}

// GetPlanningResources returns a list of resource names available.
func (p *BasicContextProvider) GetPlanningResources(ctx context.Context, agentID string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []string{"local_filesystem", "network_access"}, nil
}

// SetActiveGoal is a helper to update the active goal for an agent context.
func (p *BasicContextProvider) SetActiveGoal(ctx context.Context, agentID, goalID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	c, exists := p.contexts[agentID]
	if !exists {
		c = interfaces.PlanningContext{
			AgentID: agentID,
			State:   make(map[string]string),
		}
	}
	c.ActiveGoalID = goalID
	p.contexts[agentID] = c
	return nil
}

var _ interfaces.ContextProvider = (*BasicContextProvider)(nil)
