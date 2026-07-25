// Package binder provides the capability binding mechanisms.
package binder

import (
	"context"
	"fmt"
	"sync"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

// ContextualBinder implements interfaces.CapabilityBinder.
type ContextualBinder struct {
	mu          sync.RWMutex
	boundAgents map[string]map[string]interfaces.ExecutableCapability // agentID -> capID -> executable
}

// NewContextualBinder creates a new ContextualBinder.
func NewContextualBinder() *ContextualBinder {
	return &ContextualBinder{
		boundAgents: make(map[string]map[string]interfaces.ExecutableCapability),
	}
}

// Bind links a capability to a specific agent context, verifying permissions
// and returning an executable closure.
func (b *ContextualBinder) Bind(ctx context.Context, metadata interfaces.CapabilityMetadata, bindCtx interfaces.BindingContext) (interfaces.ExecutableCapability, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if metadata.ID == "" {
		return nil, fmt.Errorf("capability ID required")
	}
	if bindCtx.AgentID == "" {
		return nil, fmt.Errorf("agent ID required")
	}

	// Verify permissions: ensure the agent has all permissions required by the capability
	allowedSet := make(map[string]bool)
	for _, p := range bindCtx.AllowedPermissions {
		allowedSet[p] = true
	}

	for _, reqPerm := range metadata.RequiredPermissions {
		if !allowedSet[reqPerm] {
			return nil, fmt.Errorf("binding denied: agent lacks required permission %q", reqPerm)
		}
	}

	exec := &MockExecutable{
		capID:   metadata.ID,
		agentID: bindCtx.AgentID,
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.boundAgents[bindCtx.AgentID]; !exists {
		b.boundAgents[bindCtx.AgentID] = make(map[string]interfaces.ExecutableCapability)
	}
	b.boundAgents[bindCtx.AgentID][metadata.ID] = exec

	return exec, nil
}

// Unbind cleans up a bound capability.
func (b *ContextualBinder) Unbind(ctx context.Context, agentID, capabilityID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	agentCaps, exists := b.boundAgents[agentID]
	if !exists {
		return nil // Already unbound or never bound
	}

	exec, bound := agentCaps[capabilityID]
	if bound {
		_ = exec.Close(ctx)
		delete(agentCaps, capabilityID)
	}

	if len(agentCaps) == 0 {
		delete(b.boundAgents, agentID)
	}

	return nil
}

// -- MockExecutable for V0.7 (Stub until real RPC/Execution integration in future) --

// MockExecutable represents a capability bound in memory.
type MockExecutable struct {
	capID   string
	agentID string
	closed  bool
	mu      sync.RWMutex
}

// Execute simulates running the capability.
func (m *MockExecutable) Execute(ctx context.Context, inputJSON string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return "", fmt.Errorf("capability %s is unbound/closed for agent %s", m.capID, m.agentID)
	}

	// Mock execution
	return fmt.Sprintf(`{"status":"success","executed":"%s"}`, m.capID), nil
}

// Close simulates cleanup.
func (m *MockExecutable) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

var _ interfaces.CapabilityBinder = (*ContextualBinder)(nil)
var _ interfaces.ExecutableCapability = (*MockExecutable)(nil)
