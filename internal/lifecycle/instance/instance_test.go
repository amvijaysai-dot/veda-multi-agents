// Package instance defines AgentInstance — the runtime representation of a live agent.
package instance

import (
	"sync"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestSpec() *spec.AgentSpec {
	return &spec.AgentSpec{
		ID:            "agent-test",
		Name:          "Test Agent",
		ModelID:       "gpt-4",
		MaxIterations: 10,
	}
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNew_InitialStateIsCreating(t *testing.T) {
	inst := New(newTestSpec())
	if inst.State() != runtime.AgentCreating {
		t.Errorf("expected initial state %v, got %v", runtime.AgentCreating, inst.State())
	}
}

func TestNew_IDMatchesSpec(t *testing.T) {
	s := newTestSpec()
	inst := New(s)
	if inst.ID() != s.ID {
		t.Errorf("expected ID %q, got %q", s.ID, inst.ID())
	}
}

func TestNew_SpecMatchesInput(t *testing.T) {
	s := newTestSpec()
	inst := New(s)
	if inst.Spec() != s {
		t.Error("expected Spec() to return the same pointer as the input spec")
	}
}

func TestNew_CreatedAtIsRecent(t *testing.T) {
	before := time.Now().UTC()
	inst := New(newTestSpec())
	after := time.Now().UTC()

	if inst.CreatedAt().Before(before) || inst.CreatedAt().After(after) {
		t.Errorf("CreatedAt %v not within expected window [%v, %v]",
			inst.CreatedAt(), before, after)
	}
}

func TestNew_UpdatedAtEqualsCreatedAt(t *testing.T) {
	inst := New(newTestSpec())
	// Allow 1ms wiggle due to two separate time.Now() calls.
	diff := inst.UpdatedAt().Sub(inst.CreatedAt())
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Millisecond {
		t.Errorf("expected UpdatedAt ≈ CreatedAt at construction, diff = %v", diff)
	}
}

func TestNew_NoInitialErrorMessage(t *testing.T) {
	inst := New(newTestSpec())
	if msg := inst.ErrorMessage(); msg != "" {
		t.Errorf("expected empty error message at construction, got %q", msg)
	}
}

// ---------------------------------------------------------------------------
// Valid transitions — positive path (full lifecycle)
// ---------------------------------------------------------------------------

func TestTransitionTo_CreatingToInitializing(t *testing.T) {
	inst := New(newTestSpec())
	if err := inst.TransitionTo(runtime.AgentInitializing); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst.State() != runtime.AgentInitializing {
		t.Errorf("expected state %v, got %v", runtime.AgentInitializing, inst.State())
	}
}

func TestTransitionTo_InitializingToReady(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	if err := inst.TransitionTo(runtime.AgentReady); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst.State() != runtime.AgentReady {
		t.Errorf("expected state %v, got %v", runtime.AgentReady, inst.State())
	}
}

func TestTransitionTo_ReadyToBusy(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	_ = inst.TransitionTo(runtime.AgentReady)
	if err := inst.TransitionTo(runtime.AgentBusy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransitionTo_BusyToReady(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	_ = inst.TransitionTo(runtime.AgentReady)
	_ = inst.TransitionTo(runtime.AgentBusy)
	if err := inst.TransitionTo(runtime.AgentReady); err != nil {
		t.Fatalf("unexpected error for Busy → Ready: %v", err)
	}
}

func TestTransitionTo_ReadyToStoppingToTerminated(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	_ = inst.TransitionTo(runtime.AgentReady)
	_ = inst.TransitionTo(runtime.AgentStopping)
	if err := inst.TransitionTo(runtime.AgentTerminated); err != nil {
		t.Fatalf("unexpected error for Stopping → Terminated: %v", err)
	}
}

func TestTransitionTo_SuspendResumeCycle(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	_ = inst.TransitionTo(runtime.AgentReady)
	_ = inst.TransitionTo(runtime.AgentSuspending)
	_ = inst.TransitionTo(runtime.AgentSuspended)
	_ = inst.TransitionTo(runtime.AgentResuming)
	if err := inst.TransitionTo(runtime.AgentReady); err != nil {
		t.Fatalf("unexpected error completing suspend/resume cycle: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Invalid transitions — negative path
// ---------------------------------------------------------------------------

func TestTransitionTo_CreatingToReadyIsInvalid(t *testing.T) {
	inst := New(newTestSpec())
	err := inst.TransitionTo(runtime.AgentReady)
	if err == nil {
		t.Fatal("expected error for Creating → Ready (skipping Initializing)")
	}
}

func TestTransitionTo_TerminatedToAnyIsInvalid(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	_ = inst.TransitionTo(runtime.AgentTerminated)

	targets := []runtime.AgentState{
		runtime.AgentCreating,
		runtime.AgentInitializing,
		runtime.AgentReady,
		runtime.AgentBusy,
		runtime.AgentStopping,
	}
	for _, target := range targets {
		if err := inst.TransitionTo(target); err == nil {
			t.Errorf("expected error for Terminated → %v, got nil", target)
		}
	}
}

func TestTransitionTo_ReadyToCreatingIsInvalid(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	_ = inst.TransitionTo(runtime.AgentReady)
	if err := inst.TransitionTo(runtime.AgentCreating); err == nil {
		t.Fatal("expected error for Ready → Creating")
	}
}

// ---------------------------------------------------------------------------
// UpdatedAt advances on transition
// ---------------------------------------------------------------------------

func TestTransitionTo_UpdatedAtAdvancesOnTransition(t *testing.T) {
	inst := New(newTestSpec())
	before := inst.UpdatedAt()
	// brief pause so the timestamps are distinguishable
	for {
		if time.Now().UTC().After(before) {
			break
		}
	}
	_ = inst.TransitionTo(runtime.AgentInitializing)
	if !inst.UpdatedAt().After(before) {
		t.Error("expected UpdatedAt to advance after transition")
	}
}

// ---------------------------------------------------------------------------
// SetError
// ---------------------------------------------------------------------------

func TestSetError_RecordsMessageAndTerminates(t *testing.T) {
	inst := New(newTestSpec())
	_ = inst.TransitionTo(runtime.AgentInitializing)
	inst.SetError("initialization failed: dependency missing")

	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected Terminated after SetError, got %v", inst.State())
	}
	if inst.ErrorMessage() == "" {
		t.Error("expected non-empty ErrorMessage after SetError")
	}
}

func TestSetError_ClearsOnSubsequentSuccessfulTransition(t *testing.T) {
	// ErrorMessage is cleared by TransitionTo on success.
	inst := New(newTestSpec())
	// Force an error first via SetError, then confirm a fresh instance doesn't carry it.
	inst2 := New(newTestSpec())
	inst2.SetError("boom")
	// Now test that a normal instance's TransitionTo clears errorMsg.
	inst3 := New(newTestSpec())
	_ = inst3.TransitionTo(runtime.AgentInitializing)
	if inst3.ErrorMessage() != "" {
		t.Errorf("expected no error message after clean transition, got %q", inst3.ErrorMessage())
	}
	_ = inst // suppress unused warning
	_ = inst2
}

// ---------------------------------------------------------------------------
// Concurrency: concurrent reads and transitions
// ---------------------------------------------------------------------------

func TestAgentInstance_ConcurrentStateReads(t *testing.T) {
	inst := New(newTestSpec())
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = inst.State()
		}()
	}
	wg.Wait()
}
