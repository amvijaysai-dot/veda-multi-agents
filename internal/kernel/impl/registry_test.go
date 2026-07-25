// Package impl provides tests for the kernel implementation.
package impl

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/kernel/interfaces"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// noopSubsystem is a minimal Subsystem implementation for registry testing.
type noopSubsystem struct {
	name string
}

func (n *noopSubsystem) Init(_ context.Context) error  { return nil }
func (n *noopSubsystem) Start(_ context.Context) error { return nil }
func (n *noopSubsystem) Stop(_ context.Context) error  { return nil }

// newNoop creates a named noopSubsystem.
func newNoop(name string) interfaces.Subsystem {
	return &noopSubsystem{name: name}
}

// ---------------------------------------------------------------------------
// Registry tests
// ---------------------------------------------------------------------------

func TestRegistry_Register_Success(t *testing.T) {
	r := newRegistry()
	if err := r.Register("alpha", newNoop("alpha")); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}
	if r.Len() != 1 {
		t.Errorf("expected Len() == 1, got %d", r.Len())
	}
}

func TestRegistry_Register_EmptyNameReturnsError(t *testing.T) {
	r := newRegistry()
	err := r.Register("", newNoop(""))
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestRegistry_Register_NilSubsystemReturnsError(t *testing.T) {
	r := newRegistry()
	err := r.Register("alpha", nil)
	if err == nil {
		t.Fatal("expected error for nil subsystem, got nil")
	}
}

func TestRegistry_Register_DuplicateNameReturnsError(t *testing.T) {
	r := newRegistry()
	_ = r.Register("alpha", newNoop("alpha"))

	err := r.Register("alpha", newNoop("alpha2"))
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
}

func TestRegistry_Get_Success(t *testing.T) {
	r := newRegistry()
	sub := newNoop("alpha")
	_ = r.Register("alpha", sub)

	got, err := r.Get("alpha")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if got != sub {
		t.Error("Get() returned a different subsystem than what was registered")
	}
}

func TestRegistry_Get_UnknownNameReturnsError(t *testing.T) {
	r := newRegistry()
	_, err := r.Get("unknown")
	if err == nil {
		t.Fatal("expected error for unknown subsystem, got nil")
	}
}

func TestRegistry_Unregister_Success(t *testing.T) {
	r := newRegistry()
	_ = r.Register("alpha", newNoop("alpha"))

	if err := r.Unregister("alpha"); err != nil {
		t.Fatalf("Unregister() unexpected error: %v", err)
	}
	if r.Len() != 0 {
		t.Errorf("expected Len() == 0 after unregister, got %d", r.Len())
	}
}

func TestRegistry_Unregister_UnknownNameReturnsError(t *testing.T) {
	r := newRegistry()
	err := r.Unregister("unknown")
	if err == nil {
		t.Fatal("expected error for unknown subsystem, got nil")
	}
}

func TestRegistry_Names_PreservesInsertionOrder(t *testing.T) {
	r := newRegistry()
	names := []string{"alpha", "beta", "gamma"}
	for _, n := range names {
		if err := r.Register(n, newNoop(n)); err != nil {
			t.Fatalf("Register(%q) failed: %v", n, err)
		}
	}

	got := r.Names()
	if len(got) != len(names) {
		t.Fatalf("Names() returned %d names, expected %d", len(got), len(names))
	}
	for i, want := range names {
		if got[i] != want {
			t.Errorf("Names()[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestRegistry_Names_UpdatesAfterUnregister(t *testing.T) {
	r := newRegistry()
	_ = r.Register("alpha", newNoop("alpha"))
	_ = r.Register("beta", newNoop("beta"))
	_ = r.Register("gamma", newNoop("gamma"))

	_ = r.Unregister("beta")

	got := r.Names()
	if len(got) != 2 {
		t.Fatalf("expected 2 names after unregister, got %d", len(got))
	}
	if got[0] != "alpha" || got[1] != "gamma" {
		t.Errorf("unexpected names after unregister: %v", got)
	}
}

func TestRegistry_Len_Empty(t *testing.T) {
	r := newRegistry()
	if r.Len() != 0 {
		t.Errorf("expected Len() == 0 for empty registry, got %d", r.Len())
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := newRegistry()

	// Register subsystems concurrently.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 10; i++ {
			name := "sub"
			_ = r.Register(name, newNoop(name))
			_ = r.Unregister(name)
		}
	}()

	// Concurrently read names.
	go func() {
		for i := 0; i < 10; i++ {
			_ = r.Names()
			_ = r.Len()
		}
	}()

	<-done
}
