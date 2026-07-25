// Package memory provides memory lifecycle integration.
package memory

import (
	"context"

	"github.com/veda/agent-runtime/internal/lifecycle"
	"github.com/veda/agent-runtime/internal/lifecycle/instance"
	"github.com/veda/agent-runtime/internal/memory/interfaces"
)

// MemoryLifecycleManager bridges the memory subsystem with the agent lifecycle.
// It provides hooks that can be registered with the lifecycle Initializer and Terminator.
type MemoryLifecycleManager struct {
	shortTerm    interfaces.ShortTermMemory
	consolidator interfaces.ConsolidationManager
}

// NewMemoryLifecycleManager creates a new MemoryLifecycleManager.
func NewMemoryLifecycleManager(
	shortTerm interfaces.ShortTermMemory,
	consolidator interfaces.ConsolidationManager,
) *MemoryLifecycleManager {
	if shortTerm == nil {
		panic("memory: shortTerm cannot be nil")
	}
	// consolidator is optional (if nil, no consolidation happens on terminate)
	return &MemoryLifecycleManager{
		shortTerm:    shortTerm,
		consolidator: consolidator,
	}
}

// InitHook returns a lifecycle.InitHook that prepares memory for a new agent.
// In v0.5, since memory is mostly lazily allocated per session, this hook is a no-op
// but serves as an integration point for future initialization requirements (e.g. warming up caches).
func (m *MemoryLifecycleManager) InitHook() lifecycle.InitHook {
	return func(ctx context.Context, inst *instance.AgentInstance) error {
		// No eager initialization required for in-memory stubs.
		return nil
	}
}

// CleanupHook returns a lifecycle.CleanupHook that performs memory consolidation
// and cleanup when an agent is terminated.
// Note: AgentInstance does not track active sessions in v0.5. For testing
// integration, we consolidate a default "default-session" and clear it.
func (m *MemoryLifecycleManager) CleanupHook() lifecycle.CleanupHook {
	return func(ctx context.Context, inst *instance.AgentInstance) error {
		agentID := inst.ID()
		// In a real system, we would query the active sessions for this agent.
		// For v0.5 stub integration, we assume a single "default-session".
		sessionID := "default-session"

		if m.consolidator != nil {
			// Ignore errors during cleanup as per Terminator contract
			_ = m.consolidator.Consolidate(ctx, agentID, sessionID)
		}

		// Clear short-term memory for the session
		_ = m.shortTerm.Clear(ctx, agentID, sessionID)

		return nil
	}
}
