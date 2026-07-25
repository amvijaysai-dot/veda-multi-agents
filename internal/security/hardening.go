// Package security provides boundaries and validation for agent runtime operations.
package security

import (
	"context"
	"fmt"
	"sync"
)

// Policy defines what an agent is allowed to do.
type Policy struct {
	AllowedCapabilities []string
	MaxFiles            int
	MaxNetworkRequests  int
}

// SandboxManager enforces isolation and least privilege.
type SandboxManager struct {
	mu       sync.RWMutex
	policies map[string]Policy
	usage    map[string]*Usage
}

// Usage tracks resource usage per agent for sandbox constraints.
type Usage struct {
	FilesOpen       int
	NetworkRequests int
}

// NewSandboxManager initializes a security sandbox manager.
func NewSandboxManager() *SandboxManager {
	return &SandboxManager{
		policies: make(map[string]Policy),
		usage:    make(map[string]*Usage),
	}
}

// AttachPolicy binds a security policy to an agent ID.
func (sm *SandboxManager) AttachPolicy(agentID string, policy Policy) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.policies[agentID] = policy
	if _, exists := sm.usage[agentID]; !exists {
		sm.usage[agentID] = &Usage{}
	}
}

// CheckCapability checks if the agent is authorized to use the capability.
func (sm *SandboxManager) CheckCapability(agentID string, capID string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	policy, exists := sm.policies[agentID]
	if !exists {
		return fmt.Errorf("access denied: no policy found for agent %s", agentID)
	}

	for _, c := range policy.AllowedCapabilities {
		if c == capID || c == "*" {
			return nil
		}
	}
	return fmt.Errorf("access denied: agent %s is not allowed to use %s", agentID, capID)
}

// CheckNetwork Request checks and increments network usage, returning error if exceeded.
func (sm *SandboxManager) CheckNetworkRequest(ctx context.Context, agentID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	policy, ok := sm.policies[agentID]
	if !ok {
		return fmt.Errorf("access denied: no policy found for agent %s", agentID)
	}

	usage := sm.usage[agentID]
	if policy.MaxNetworkRequests > 0 && usage.NetworkRequests >= policy.MaxNetworkRequests {
		return fmt.Errorf("sandbox breach: network request limit exceeded for agent %s", agentID)
	}

	usage.NetworkRequests++
	return nil
}
