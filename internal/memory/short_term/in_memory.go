// Package shortterm provides the short-term memory implementation for the
// VEDA Agent Runtime.
package shortterm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/veda/agent-runtime/internal/memory/interfaces"
)

// entry represents a single stored value with its persistence hint state
// and expiration time.
type entry struct {
	value           string
	persistenceHint bool
	expiresAt       time.Time
}

// InMemoryShortTerm implements interfaces.ShortTermMemory using a thread-safe
// nested map structure: agentID -> sessionID -> key -> entry.
// It supports basic TTL expiration via lazy eviction on read.
type InMemoryShortTerm struct {
	mu sync.RWMutex
	// stores agentID -> sessionID -> key -> entry
	data map[string]map[string]map[string]entry

	// defaultTTL is applied to all stored entries. If zero, entries do not expire.
	defaultTTL time.Duration
}

// NewInMemoryShortTerm creates a new InMemoryShortTerm with the specified default TTL.
func NewInMemoryShortTerm(defaultTTL time.Duration) *InMemoryShortTerm {
	return &InMemoryShortTerm{
		data:       make(map[string]map[string]map[string]entry),
		defaultTTL: defaultTTL,
	}
}

// ensureSessionMap creates the nested maps if they don't exist.
// Must be called with a write lock held.
func (m *InMemoryShortTerm) ensureSessionMap(agentID, sessionID string) map[string]entry {
	if m.data[agentID] == nil {
		m.data[agentID] = make(map[string]map[string]entry)
	}
	if m.data[agentID][sessionID] == nil {
		m.data[agentID][sessionID] = make(map[string]entry)
	}
	return m.data[agentID][sessionID]
}

// Store saves a value for the given agent and session.
func (m *InMemoryShortTerm) Store(ctx context.Context, agentID, sessionID, key, value string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if agentID == "" || sessionID == "" || key == "" {
		return fmt.Errorf("agentID, sessionID, and key must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sessionData := m.ensureSessionMap(agentID, sessionID)

	// Preserve existing persistence hint if updating
	hint := false
	if existing, ok := sessionData[key]; ok {
		hint = existing.persistenceHint
	}

	var expiresAt time.Time
	if m.defaultTTL > 0 {
		expiresAt = time.Now().UTC().Add(m.defaultTTL)
	}

	sessionData[key] = entry{
		value:           value,
		persistenceHint: hint,
		expiresAt:       expiresAt,
	}
	return nil
}

// Retrieve fetches a value by key. Applies lazy eviction for expired entries.
func (m *InMemoryShortTerm) Retrieve(ctx context.Context, agentID, sessionID, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	m.mu.RLock()
	sessionData, exists := m.getSessionData(agentID, sessionID)
	if !exists {
		m.mu.RUnlock()
		return "", fmt.Errorf("key %q not found", key)
	}
	ent, ok := sessionData[key]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}

	if !ent.expiresAt.IsZero() && time.Now().UTC().After(ent.expiresAt) {
		// Lazy eviction
		_ = m.Delete(ctx, agentID, sessionID, key)
		return "", fmt.Errorf("key %q not found (expired)", key)
	}

	return ent.value, nil
}

// getSessionData is a lock-free helper to navigate the nested maps.
// Caller must hold RLock or Lock.
func (m *InMemoryShortTerm) getSessionData(agentID, sessionID string) (map[string]entry, bool) {
	agentData, ok := m.data[agentID]
	if !ok {
		return nil, false
	}
	sessionData, ok := agentData[sessionID]
	return sessionData, ok
}

// Delete removes a key from the session memory.
func (m *InMemoryShortTerm) Delete(ctx context.Context, agentID, sessionID, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sessionData, exists := m.getSessionData(agentID, sessionID)
	if exists {
		delete(sessionData, key)
	}
	return nil
}

// List returns all keys matching the given prefix for the session.
// Expired entries are not returned, but are not eagerly evicted here to keep List fast.
func (m *InMemoryShortTerm) List(ctx context.Context, agentID, sessionID, prefix string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionData, exists := m.getSessionData(agentID, sessionID)
	if !exists {
		return []string{}, nil
	}

	now := time.Now().UTC()
	var keys []string
	for k, ent := range sessionData {
		if strings.HasPrefix(k, prefix) {
			if !ent.expiresAt.IsZero() && now.After(ent.expiresAt) {
				continue
			}
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// Clear removes all data for the specified session.
func (m *InMemoryShortTerm) Clear(ctx context.Context, agentID, sessionID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if agentData, ok := m.data[agentID]; ok {
		delete(agentData, sessionID)
	}
	return nil
}

// PersistenceHint marks a key as important for long-term storage consideration.
func (m *InMemoryShortTerm) PersistenceHint(ctx context.Context, agentID, sessionID, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sessionData, exists := m.getSessionData(agentID, sessionID)
	if !exists {
		return fmt.Errorf("key %q not found", key)
	}

	ent, ok := sessionData[key]
	if !ok {
		return fmt.Errorf("key %q not found", key)
	}

	// Update hint
	ent.persistenceHint = true
	sessionData[key] = ent
	return nil
}

// GetConsolidationCandidates is an internal helper used by the consolidation manager
// to fetch all valid (unexpired) entries that have a persistence hint.
func (m *InMemoryShortTerm) GetConsolidationCandidates(agentID, sessionID string) map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	candidates := make(map[string]string)
	sessionData, exists := m.getSessionData(agentID, sessionID)
	if !exists {
		return candidates
	}

	now := time.Now().UTC()
	for k, ent := range sessionData {
		if ent.persistenceHint {
			if !ent.expiresAt.IsZero() && now.After(ent.expiresAt) {
				continue
			}
			candidates[k] = ent.value
		}
	}
	return candidates
}

var _ interfaces.ShortTermMemory = (*InMemoryShortTerm)(nil)
