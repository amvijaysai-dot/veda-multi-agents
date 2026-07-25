// Package longterm provides the long-term memory stub.
package longterm

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestInMemoryLongTerm_StoreAndRetrieve(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()

	err := mem.Store(ctx, "agent1", "key1", "val1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := mem.Retrieve(ctx, "agent1", "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "val1" {
		t.Errorf("expected val1, got %q", val)
	}

	// Unknown key
	_, err = mem.Retrieve(ctx, "agent1", "unknown")
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}

	// Unknown agent
	_, err = mem.Retrieve(ctx, "agent2", "key1")
	if err == nil {
		t.Error("expected error for unknown agent, got nil")
	}
}

func TestInMemoryLongTerm_EmptyParameters(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()

	err := mem.Store(ctx, "", "key1", "val1", 0)
	if err == nil {
		t.Error("expected error for empty agentID")
	}
	err = mem.Store(ctx, "agent1", "", "val1", 0)
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestInMemoryLongTerm_Delete(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "key1", "val1", 0)
	_ = mem.Delete(ctx, "agent1", "key1")

	_, err := mem.Retrieve(ctx, "agent1", "key1")
	if err == nil {
		t.Error("expected error for deleted key")
	}
}

func TestInMemoryLongTerm_Forget(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "key1", "val1", 0)

	// Forget acts like Delete in the stub
	err := mem.Forget(ctx, "agent1", "key1", "privacy request")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = mem.Retrieve(ctx, "agent1", "key1")
	if err == nil {
		t.Error("expected error for forgotten key")
	}
}

func TestInMemoryLongTerm_Scan(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "user_1", "u1", 0)
	_ = mem.Store(ctx, "agent1", "user_2", "u2", 0)
	_ = mem.Store(ctx, "agent1", "pref_1", "p1", 0)

	keys, err := mem.Scan(ctx, "agent1", "user_")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestInMemoryLongTerm_Query(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "k1", "likes apples", 0)
	_ = mem.Store(ctx, "agent1", "k2", "likes oranges", 0)
	_ = mem.Store(ctx, "agent1", "k3", "hates apples", 0)

	results, err := mem.Query(ctx, "agent1", "apples")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryLongTerm_TTL(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()

	_ = mem.Store(ctx, "agent1", "key1", "val1", time.Millisecond)

	// busy wait 2ms
	start := time.Now().UTC()
	for {
		if time.Now().UTC().Sub(start) > 2*time.Millisecond {
			break
		}
	}

	_, err := mem.Retrieve(ctx, "agent1", "key1")
	if err == nil {
		t.Error("expected expired key to return error")
	}
}

func TestInMemoryLongTerm_Concurrency(t *testing.T) {
	mem := NewInMemoryLongTerm()
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = mem.Store(ctx, "agent1", "key", "val", 0)
		}(i)
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = mem.Retrieve(ctx, "agent1", "key")
		}(i)
	}
	wg.Wait()
}
