// Package registry provides the capability registry implementation.
package registry

import (
	"context"
	"sync"
	"testing"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

func TestInMemoryRegistry_RegisterAndLookup(t *testing.T) {
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	meta := interfaces.CapabilityMetadata{
		ID:      "cap.fs.read",
		Version: "1.0.0",
		Name:    "File Reader",
	}

	err := reg.Register(ctx, meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Exact lookup
	fetched, err := reg.Lookup(ctx, "cap.fs.read", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "File Reader" {
		t.Errorf("expected File Reader, got %q", fetched.Name)
	}

	// Implicit highest version lookup
	_ = reg.Register(ctx, interfaces.CapabilityMetadata{
		ID:      "cap.fs.read",
		Version: "1.1.0",
		Name:    "File Reader V2",
	})

	fetchedHighest, err := reg.Lookup(ctx, "cap.fs.read", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetchedHighest.Version != "1.1.0" {
		t.Errorf("expected highest version 1.1.0, got %q", fetchedHighest.Version)
	}
}

func TestInMemoryRegistry_RegisterDuplicate(t *testing.T) {
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	meta := interfaces.CapabilityMetadata{ID: "c1", Version: "1"}
	_ = reg.Register(ctx, meta)

	err := reg.Register(ctx, meta)
	if err == nil {
		t.Error("expected error on duplicate registration")
	}
}

func TestInMemoryRegistry_Deregister(t *testing.T) {
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	_ = reg.Register(ctx, interfaces.CapabilityMetadata{ID: "c1", Version: "1"})
	_ = reg.Register(ctx, interfaces.CapabilityMetadata{ID: "c1", Version: "2"})

	err := reg.Deregister(ctx, "c1", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = reg.Lookup(ctx, "c1", "1")
	if err == nil {
		t.Error("expected error looking up deregistered capability")
	}

	// Should still find version 2
	_, err = reg.Lookup(ctx, "c1", "2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Deregister last version
	err = reg.Deregister(ctx, "c1", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Map should be cleaned up
	reg.mu.RLock()
	_, exists := reg.capabilities["c1"]
	reg.mu.RUnlock()
	if exists {
		t.Error("expected map key to be deleted when empty")
	}
}

func TestInMemoryRegistry_List(t *testing.T) {
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	_ = reg.Register(ctx, interfaces.CapabilityMetadata{ID: "c1", Version: "1", Name: "Tool A"})
	_ = reg.Register(ctx, interfaces.CapabilityMetadata{ID: "c2", Version: "1", Name: "Tool B"})
	_ = reg.Register(ctx, interfaces.CapabilityMetadata{ID: "c2", Version: "2", Name: "Tool B upgraded"})

	// List all
	all, err := reg.List(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 items, got %d", len(all))
	}

	// List filtered by ID
	filtered, _ := reg.List(ctx, map[string]string{"id": "c2"})
	if len(filtered) != 2 {
		t.Errorf("expected 2 items for c2, got %d", len(filtered))
	}

	// List filtered by Name
	filteredName, _ := reg.List(ctx, map[string]string{"name": "upgraded"})
	if len(filteredName) != 1 {
		t.Errorf("expected 1 item, got %d", len(filteredName))
	}
}

func TestInMemoryRegistry_Concurrency(t *testing.T) {
	reg := NewInMemoryRegistry()
	ctx := context.Background()
	var wg sync.WaitGroup

	// Start concurrent reads and writes
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = reg.Register(ctx, interfaces.CapabilityMetadata{
				ID:      "cap.sync",
				Version: "1.0.0",
			})
			_, _ = reg.Lookup(ctx, "cap.sync", "")
			_, _ = reg.List(ctx, nil)
		}(i)
	}

	wg.Wait()
}
