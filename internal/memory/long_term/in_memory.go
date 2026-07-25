// Package longterm provides the in-memory stub implementation for long-term storage.
// This is a temporary implementation for v0.5 and will be replaced by a robust
// vector DB or key-value store integration in later milestones.
package longterm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/veda/agent-runtime/internal/memory/interfaces"
)

type entry struct {
	value     string
	expiresAt time.Time
}

// InMemoryLongTerm implements interfaces.LongTermMemory using a thread-safe
// map structure: agentID -> key -> entry.
type InMemoryLongTerm struct {
	mu sync.RWMutex
	// stores agentID -> key -> entry
	data map[string]map[string]entry
}

// NewInMemoryLongTerm creates a new InMemoryLongTerm stub.
func NewInMemoryLongTerm() *InMemoryLongTerm {
	return &InMemoryLongTerm{
		data: make(map[string]map[string]entry),
	}
}

// ensureAgentMap creates the agent map if it doesn't exist.
// Must be called with a write lock held.
func (m *InMemoryLongTerm) ensureAgentMap(agentID string) map[string]entry {
	if m.data[agentID] == nil {
		m.data[agentID] = make(map[string]entry)
	}
	return m.data[agentID]
}

// Store saves persistent data with an optional TTL.
func (m *InMemoryLongTerm) Store(ctx context.Context, agentID, key, value string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if agentID == "" || key == "" {
		return fmt.Errorf("agentID and key must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	agentData := m.ensureAgentMap(agentID)

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().UTC().Add(ttl)
	}

	agentData[key] = entry{
		value:     value,
		expiresAt: expiresAt,
	}
	return nil
}

// Retrieve fetches persistent data by key.
func (m *InMemoryLongTerm) Retrieve(ctx context.Context, agentID, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	m.mu.RLock()
	agentData, ok := m.data[agentID]
	if !ok {
		m.mu.RUnlock()
		return "", fmt.Errorf("key %q not found", key)
	}
	ent, ok := agentData[key]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}

	if !ent.expiresAt.IsZero() && time.Now().UTC().After(ent.expiresAt) {
		_ = m.Delete(ctx, agentID, key)
		return "", fmt.Errorf("key %q not found (expired)", key)
	}

	return ent.value, nil
}

// Delete removes persistent data by key.
func (m *InMemoryLongTerm) Delete(ctx context.Context, agentID, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if agentData, ok := m.data[agentID]; ok {
		delete(agentData, key)
	}
	return nil
}

// Query performs a simple substring match on the stored values.
func (m *InMemoryLongTerm) Query(ctx context.Context, agentID, query string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if query == "" {
		return []string{}, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	agentData, ok := m.data[agentID]
	if !ok {
		return []string{}, nil
	}

	now := time.Now().UTC()
	var results []string
	for _, ent := range agentData {
		if !ent.expiresAt.IsZero() && now.After(ent.expiresAt) {
			continue
		}
		if strings.Contains(ent.value, query) {
			results = append(results, ent.value)
		}
	}
	return results, nil
}

// Scan returns all keys matching the given prefix.
func (m *InMemoryLongTerm) Scan(ctx context.Context, agentID, prefix string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	agentData, ok := m.data[agentID]
	if !ok {
		return []string{}, nil
	}

	now := time.Now().UTC()
	var keys []string
	for k, ent := range agentData {
		if strings.HasPrefix(k, prefix) {
			if !ent.expiresAt.IsZero() && now.After(ent.expiresAt) {
				continue
			}
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// Forget removes persistent data and explicitly logs the reason (simulated here).
func (m *InMemoryLongTerm) Forget(ctx context.Context, agentID, key, reason string) error {
	// In a real implementation this would write to an audit log.
	// For this stub, we just delegate to Delete.
	return m.Delete(ctx, agentID, key)
}

var _ interfaces.LongTermMemory = (*InMemoryLongTerm)(nil)
