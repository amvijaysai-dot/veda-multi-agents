// Package shortterm provides the short-term memory implementation.
package shortterm

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestInMemoryShortTerm_StoreAndRetrieve(t *testing.T) {
	mem := NewInMemoryShortTerm(0)
	ctx := context.Background()

	err := mem.Store(ctx, "agent1", "session1", "key1", "val1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := mem.Retrieve(ctx, "agent1", "session1", "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "val1" {
		t.Errorf("expected val1, got %q", val)
	}

	// Unknown key
	_, err = mem.Retrieve(ctx, "agent1", "session1", "unknown")
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}

	// Unknown agent
	_, err = mem.Retrieve(ctx, "agent2", "session1", "key1")
	if err == nil {
		t.Error("expected error for unknown agent, got nil")
	}
}

func TestInMemoryShortTerm_EmptyParameters(t *testing.T) {
	mem := NewInMemoryShortTerm(0)
	ctx := context.Background()

	err := mem.Store(ctx, "", "session1", "key1", "val1")
	if err == nil {
		t.Error("expected error for empty agentID")
	}
	err = mem.Store(ctx, "agent1", "", "key1", "val1")
	if err == nil {
		t.Error("expected error for empty sessionID")
	}
	err = mem.Store(ctx, "agent1", "session1", "", "val1")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestInMemoryShortTerm_Delete(t *testing.T) {
	mem := NewInMemoryShortTerm(0)
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "session1", "key1", "val1")
	_ = mem.Delete(ctx, "agent1", "session1", "key1")

	_, err := mem.Retrieve(ctx, "agent1", "session1", "key1")
	if err == nil {
		t.Error("expected error for deleted key")
	}

	// Deleting non-existent key should not panic or error
	err = mem.Delete(ctx, "agent1", "session1", "key1")
	if err != nil {
		t.Errorf("unexpected error deleting non-existent key: %v", err)
	}
}

func TestInMemoryShortTerm_Clear(t *testing.T) {
	mem := NewInMemoryShortTerm(0)
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "session1", "key1", "val1")
	_ = mem.Store(ctx, "agent1", "session1", "key2", "val2")
	_ = mem.Store(ctx, "agent1", "session2", "key1", "val3")

	err := mem.Clear(ctx, "agent1", "session1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = mem.Retrieve(ctx, "agent1", "session1", "key1")
	if err == nil {
		t.Error("expected session1 keys to be cleared")
	}

	// session2 should remain intact
	val, err := mem.Retrieve(ctx, "agent1", "session2", "key1")
	if err != nil || val != "val3" {
		t.Error("expected session2 to remain intact")
	}
}

func TestInMemoryShortTerm_List(t *testing.T) {
	mem := NewInMemoryShortTerm(0)
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "session1", "prefix_1", "v1")
	_ = mem.Store(ctx, "agent1", "session1", "prefix_2", "v2")
	_ = mem.Store(ctx, "agent1", "session1", "other_3", "v3")

	keys, err := mem.List(ctx, "agent1", "session1", "prefix_")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}

	has1, has2 := false, false
	for _, k := range keys {
		if k == "prefix_1" {
			has1 = true
		}
		if k == "prefix_2" {
			has2 = true
		}
	}
	if !has1 || !has2 {
		t.Error("missing expected keys in list result")
	}
}

func TestInMemoryShortTerm_TTL(t *testing.T) {
	// TTL of 1 millisecond
	mem := NewInMemoryShortTerm(time.Millisecond)
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "session1", "key1", "val1")

	// Wait for TTL to expire using a busy wait to avoid Defender false positive.
	start := time.Now().UTC()
	for {
		if time.Now().UTC().Sub(start) > 2*time.Millisecond {
			break
		}
	}

	// Retrieve should trigger lazy eviction
	_, err := mem.Retrieve(ctx, "agent1", "session1", "key1")
	if err == nil {
		t.Error("expected expired key to return error")
	}

	// Internal state should no longer have the key (evicted)
	mem.mu.RLock()
	_, exists := mem.data["agent1"]["session1"]["key1"]
	mem.mu.RUnlock()
	if exists {
		t.Error("expected key to be deleted via lazy eviction")
	}
}

func TestInMemoryShortTerm_PersistenceHint(t *testing.T) {
	mem := NewInMemoryShortTerm(0)
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "session1", "key1", "val1")

	err := mem.PersistenceHint(ctx, "agent1", "session1", "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Update value, should preserve hint
	_ = mem.Store(ctx, "agent1", "session1", "key1", "val2")

	candidates := mem.GetConsolidationCandidates("agent1", "session1")
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates["key1"] != "val2" {
		t.Errorf("expected val2, got %q", candidates["key1"])
	}
}

func TestInMemoryShortTerm_Concurrency(t *testing.T) {
	mem := NewInMemoryShortTerm(0)
	ctx := context.Background()
	var wg sync.WaitGroup

	// 10 goroutines storing, 10 retrieving simultaneously
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = mem.Store(ctx, "agent1", "session1", "key", "val")
		}(i)
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = mem.Retrieve(ctx, "agent1", "session1", "key")
		}(i)
	}
	wg.Wait()
}
