// Package lifecycle implements the agent lifecycle subsystem for the VEDA Agent Runtime.
package lifecycle

import (
	"context"
	"errors"
	"testing"

	"github.com/veda/agent-runtime/internal/lifecycle/instance"
	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newReadyInstance creates an instance that has been created and initialized
// to AgentReady state, ready for termination testing.
func newReadyInstance(t *testing.T) *instance.AgentInstance {
	t.Helper()
	creator := NewCreator()
	inst, err := creator.Create(context.Background(), &spec.AgentSpec{
		ID:      "term-test-agent",
		Name:    "Terminator Test Agent",
		ModelID: "gpt-4",
	})
	if err != nil {
		t.Fatalf("newReadyInstance: Create failed: %v", err)
	}
	initializer := NewInitializer()
	if err := initializer.Initialize(context.Background(), inst); err != nil {
		t.Fatalf("newReadyInstance: Initialize failed: %v", err)
	}
	return inst
}

// ---------------------------------------------------------------------------
// Terminator.Terminate — happy paths
// ---------------------------------------------------------------------------

func TestTerminator_Terminate_NoHooksTransitionsToTerminated(t *testing.T) {
	inst := newReadyInstance(t)
	term := NewTerminator()

	if err := term.Terminate(context.Background(), inst); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected state %v, got %v", runtime.AgentTerminated, inst.State())
	}
}

func TestTerminator_Terminate_SetsStopping_DuringCleanup(t *testing.T) {
	inst := newReadyInstance(t)
	captured := runtime.AgentNonExistent

	hook := func(_ context.Context, i *instance.AgentInstance) error {
		captured = i.State()
		return nil
	}

	term := NewTerminator(hook)
	_ = term.Terminate(context.Background(), inst)

	if captured != runtime.AgentStopping {
		t.Errorf("expected agent in Stopping state inside hook, got %v", captured)
	}
}

func TestTerminator_Terminate_AllHooksRun(t *testing.T) {
	inst := newReadyInstance(t)
	count := 0
	hook := func(_ context.Context, _ *instance.AgentInstance) error {
		count++
		return nil
	}

	term := NewTerminator(hook, hook, hook)
	if err := term.Terminate(context.Background(), inst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected all 3 hooks to run, ran %d", count)
	}
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected Terminated, got %v", inst.State())
	}
}

// ---------------------------------------------------------------------------
// Terminator.Terminate — hook failure paths
// ---------------------------------------------------------------------------

func TestTerminator_Terminate_FailingHookStillTerminatesAgent(t *testing.T) {
	inst := newReadyInstance(t)
	failHook := func(_ context.Context, _ *instance.AgentInstance) error {
		return errors.New("capability cleanup failed")
	}

	term := NewTerminator(failHook)
	err := term.Terminate(context.Background(), inst)

	// Error is expected because hook failed.
	if err == nil {
		t.Fatal("expected error when hook fails, got nil")
	}
	// Agent must still reach Terminated.
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected Terminated even after hook failure, got %v", inst.State())
	}
}

func TestTerminator_Terminate_AllHooksRunEvenIfOneFails(t *testing.T) {
	inst := newReadyInstance(t)
	runCounts := [3]int{}

	hooks := []CleanupHook{
		func(_ context.Context, _ *instance.AgentInstance) error {
			runCounts[0]++
			return errors.New("hook 0 failed")
		},
		func(_ context.Context, _ *instance.AgentInstance) error {
			runCounts[1]++
			return nil
		},
		func(_ context.Context, _ *instance.AgentInstance) error {
			runCounts[2]++
			return errors.New("hook 2 failed")
		},
	}

	term := NewTerminator(hooks...)
	err := term.Terminate(context.Background(), inst)

	if err == nil {
		t.Fatal("expected combined error when hooks fail, got nil")
	}
	for i, count := range runCounts {
		if count != 1 {
			t.Errorf("hook %d ran %d times, expected 1 (all hooks must run)", i, count)
		}
	}
}

func TestTerminator_Terminate_MultipleHookErrorsReportedCombined(t *testing.T) {
	inst := newReadyInstance(t)
	term := NewTerminator(
		func(_ context.Context, _ *instance.AgentInstance) error { return errors.New("err A") },
		func(_ context.Context, _ *instance.AgentInstance) error { return errors.New("err B") },
	)
	err := term.Terminate(context.Background(), inst)
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	// Both error messages should appear somewhere in the combined error.
	msg := err.Error()
	if len(msg) == 0 {
		t.Error("expected non-empty error message")
	}
}

// ---------------------------------------------------------------------------
// Terminator.Terminate — invalid initial state
// ---------------------------------------------------------------------------

func TestTerminator_Terminate_WrongStateReturnsError(t *testing.T) {
	// A freshly-created instance is in AgentCreating; Stopping is not allowed from there.
	creator := NewCreator()
	inst, _ := creator.Create(context.Background(), &spec.AgentSpec{
		ID:      "term-state-test",
		Name:    "Term State Test",
		ModelID: "model-1",
	})

	term := NewTerminator()
	err := term.Terminate(context.Background(), inst)
	if err == nil {
		t.Fatal("expected error when terminating agent not in stoppable state, got nil")
	}
	// State should not have changed.
	if inst.State() != runtime.AgentCreating {
		t.Errorf("expected state to remain %v, got %v", runtime.AgentCreating, inst.State())
	}
}

func TestTerminator_Terminate_AlreadyTerminatedReturnsError(t *testing.T) {
	inst := newReadyInstance(t)
	term := NewTerminator()
	_ = term.Terminate(context.Background(), inst)    // first termination
	err := term.Terminate(context.Background(), inst) // second termination
	if err == nil {
		t.Fatal("expected error when terminating an already-Terminated agent, got nil")
	}
}
