// Package lifecycle implements the agent lifecycle subsystem for the VEDA Agent Runtime.
package lifecycle

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func validTestSpec() *spec.AgentSpec {
	return &spec.AgentSpec{
		ID:      "agent-creator-01",
		Name:    "Creator Test Agent",
		ModelID: "gpt-4",
	}
}

// ---------------------------------------------------------------------------
// Creator.Create — happy paths
// ---------------------------------------------------------------------------

func TestCreator_Create_ValidSpecReturnsInstance(t *testing.T) {
	creator := NewCreator()
	inst, err := creator.Create(context.Background(), validTestSpec())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if inst == nil {
		t.Fatal("expected non-nil instance, got nil")
	}
}

func TestCreator_Create_InstanceStartsInCreatingState(t *testing.T) {
	creator := NewCreator()
	inst, _ := creator.Create(context.Background(), validTestSpec())
	if inst.State() != runtime.AgentCreating {
		t.Errorf("expected initial state %v, got %v", runtime.AgentCreating, inst.State())
	}
}

func TestCreator_Create_InstanceIDMatchesSpec(t *testing.T) {
	s := validTestSpec()
	creator := NewCreator()
	inst, _ := creator.Create(context.Background(), s)
	if inst.ID() != s.ID {
		t.Errorf("expected instance ID %q, got %q", s.ID, inst.ID())
	}
}

func TestCreator_Create_NormalizesMaxIterations(t *testing.T) {
	s := validTestSpec()
	s.MaxIterations = 0
	creator := NewCreator()
	_, err := creator.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After Create, the spec should be normalized.
	if s.MaxIterations == 0 {
		t.Error("expected MaxIterations to be normalized from 0 to a positive default")
	}
}

func TestCreator_Create_SpecWithCapabilitiesSucceeds(t *testing.T) {
	s := validTestSpec()
	s.Capabilities = []string{"search", "calculator"}
	creator := NewCreator()
	inst, err := creator.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst == nil {
		t.Fatal("expected non-nil instance")
	}
}

// ---------------------------------------------------------------------------
// Creator.Create — invalid specs
// ---------------------------------------------------------------------------

func TestCreator_Create_NilSpecReturnsError(t *testing.T) {
	creator := NewCreator()
	_, err := creator.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil spec, got nil")
	}
}

func TestCreator_Create_EmptyIDReturnsError(t *testing.T) {
	s := validTestSpec()
	s.ID = ""
	creator := NewCreator()
	_, err := creator.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}
}

func TestCreator_Create_EmptyNameReturnsError(t *testing.T) {
	s := validTestSpec()
	s.Name = ""
	creator := NewCreator()
	_, err := creator.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for empty Name, got nil")
	}
}

func TestCreator_Create_EmptyModelIDReturnsError(t *testing.T) {
	s := validTestSpec()
	s.ModelID = ""
	creator := NewCreator()
	_, err := creator.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for empty ModelID, got nil")
	}
}

func TestCreator_Create_NegativeMaxIterationsReturnsError(t *testing.T) {
	s := validTestSpec()
	s.MaxIterations = -5
	creator := NewCreator()
	_, err := creator.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for negative MaxIterations, got nil")
	}
}

func TestCreator_Create_EmptyCapabilityEntryReturnsError(t *testing.T) {
	s := validTestSpec()
	s.Capabilities = []string{"valid", ""}
	creator := NewCreator()
	_, err := creator.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for empty capability entry, got nil")
	}
}

// ---------------------------------------------------------------------------
// Creator.Create — cancelled context
// ---------------------------------------------------------------------------

func TestCreator_Create_CancelledContextReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	creator := NewCreator()
	_, err := creator.Create(ctx, validTestSpec())
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
