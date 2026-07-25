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
// Integration tests for the full agent lifecycle:
//   Create → Initialize → (work) → Terminate
//
// These tests wire together Creator, Initializer, and Terminator to verify
// the complete sequence, state transitions at each step, and error handling
// paths, matching acceptance criteria V0.3.06.
// ---------------------------------------------------------------------------

// makeSpec returns a valid AgentSpec for integration tests.
func makeSpec(id string) *spec.AgentSpec {
	return &spec.AgentSpec{
		ID:            id,
		Name:          "Integration Test Agent",
		ModelID:       "gpt-4",
		MaxIterations: 5,
		Capabilities:  []string{"search"},
		Config:        map[string]string{"env": "test"},
	}
}

// ---------------------------------------------------------------------------
// Full happy-path lifecycle: Create → Initialize → Terminate
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_CreateInitializeTerminate(t *testing.T) {
	ctx := context.Background()
	creator := NewCreator()
	initializer := NewInitializer()
	terminator := NewTerminator()

	// 1. Create
	inst, err := creator.Create(ctx, makeSpec("int-agent-01"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if inst.State() != runtime.AgentCreating {
		t.Errorf("after Create: expected %v, got %v", runtime.AgentCreating, inst.State())
	}

	// 2. Initialize
	if err := initializer.Initialize(ctx, inst); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if inst.State() != runtime.AgentReady {
		t.Errorf("after Initialize: expected %v, got %v", runtime.AgentReady, inst.State())
	}

	// 3. Terminate
	if err := terminator.Terminate(ctx, inst); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("after Terminate: expected %v, got %v", runtime.AgentTerminated, inst.State())
	}
}

// ---------------------------------------------------------------------------
// Lifecycle with init hooks (simulating stubbed dependency setup)
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_WithInitHooks(t *testing.T) {
	ctx := context.Background()
	hookOrder := make([]string, 0, 3)

	// Simulated stubbed hooks — memory allocation, capability binding, model loading.
	memoryHook := func(_ context.Context, _ *instance.AgentInstance) error {
		hookOrder = append(hookOrder, "memory")
		return nil
	}
	capHook := func(_ context.Context, _ *instance.AgentInstance) error {
		hookOrder = append(hookOrder, "capability")
		return nil
	}
	modelHook := func(_ context.Context, _ *instance.AgentInstance) error {
		hookOrder = append(hookOrder, "model")
		return nil
	}

	creator := NewCreator()
	initializer := NewInitializer(memoryHook, capHook, modelHook)
	terminator := NewTerminator()

	inst, _ := creator.Create(ctx, makeSpec("int-agent-hooks"))
	if err := initializer.Initialize(ctx, inst); err != nil {
		t.Fatalf("Initialize with hooks failed: %v", err)
	}
	if err := terminator.Terminate(ctx, inst); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}

	// Verify hook execution order.
	expected := []string{"memory", "capability", "model"}
	if len(hookOrder) != len(expected) {
		t.Fatalf("expected %d hooks to run, got %d", len(expected), len(hookOrder))
	}
	for i, name := range expected {
		if hookOrder[i] != name {
			t.Errorf("hook[%d]: expected %q, got %q", i, name, hookOrder[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Lifecycle with cleanup hooks (simulating stubbed capability unbinding)
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_WithCleanupHooks(t *testing.T) {
	ctx := context.Background()
	cleanupRan := false

	cleanupHook := func(_ context.Context, _ *instance.AgentInstance) error {
		cleanupRan = true
		return nil
	}

	creator := NewCreator()
	initializer := NewInitializer()
	terminator := NewTerminator(cleanupHook)

	inst, _ := creator.Create(ctx, makeSpec("int-agent-cleanup"))
	_ = initializer.Initialize(ctx, inst)

	if err := terminator.Terminate(ctx, inst); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}
	if !cleanupRan {
		t.Error("expected cleanup hook to run during termination")
	}
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected Terminated, got %v", inst.State())
	}
}

// ---------------------------------------------------------------------------
// Error handling: initialization failure
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_InitializationFailure(t *testing.T) {
	ctx := context.Background()
	creator := NewCreator()
	initializer := NewInitializer(func(_ context.Context, _ *instance.AgentInstance) error {
		return errors.New("model unavailable")
	})

	inst, _ := creator.Create(ctx, makeSpec("int-agent-init-fail"))

	err := initializer.Initialize(ctx, inst)
	if err == nil {
		t.Fatal("expected initialization to fail, got nil")
	}
	// Agent should be terminated with an error message.
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected Terminated after init failure, got %v", inst.State())
	}
	if inst.ErrorMessage() == "" {
		t.Error("expected non-empty error message after init failure")
	}
}

// ---------------------------------------------------------------------------
// Error handling: termination with failing cleanup hooks
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_TerminationWithFailingCleanup(t *testing.T) {
	ctx := context.Background()
	creator := NewCreator()
	initializer := NewInitializer()
	terminator := NewTerminator(func(_ context.Context, _ *instance.AgentInstance) error {
		return errors.New("memory release failed")
	})

	inst, _ := creator.Create(ctx, makeSpec("int-agent-term-fail"))
	_ = initializer.Initialize(ctx, inst)

	err := terminator.Terminate(ctx, inst)
	if err == nil {
		t.Fatal("expected error from failing cleanup hook, got nil")
	}
	// Despite cleanup failure, agent must be Terminated.
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected Terminated despite cleanup error, got %v", inst.State())
	}
}

// ---------------------------------------------------------------------------
// Multiple independent agents created concurrently
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_MultipleAgentsConcurrently(t *testing.T) {
	const numAgents = 10
	ctx := context.Background()
	results := make(chan error, numAgents)

	for i := 0; i < numAgents; i++ {
		agentID := spec.AgentSpec{
			ID:      string(rune('a'+i)) + "-concurrent",
			Name:    "Concurrent Agent",
			ModelID: "gpt-4",
		}
		go func(s spec.AgentSpec) {
			creator := NewCreator()
			initializer := NewInitializer()
			terminator := NewTerminator()

			inst, err := creator.Create(ctx, &s)
			if err != nil {
				results <- err
				return
			}
			if err := initializer.Initialize(ctx, inst); err != nil {
				results <- err
				return
			}
			if err := terminator.Terminate(ctx, inst); err != nil {
				results <- err
				return
			}
			if inst.State() != runtime.AgentTerminated {
				results <- errors.New("not terminated")
				return
			}
			results <- nil
		}(agentID)
	}

	for i := 0; i < numAgents; i++ {
		if err := <-results; err != nil {
			t.Errorf("concurrent agent failed: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Verify state cannot be skipped in the lifecycle
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_CannotSkipInitialization(t *testing.T) {
	ctx := context.Background()
	creator := NewCreator()
	terminator := NewTerminator()

	inst, _ := creator.Create(ctx, makeSpec("int-agent-skip"))
	// Attempt to terminate without initializing — agent is in Creating, not Ready.
	err := terminator.Terminate(ctx, inst)
	if err == nil {
		t.Fatal("expected error when terminating an un-initialized (Creating) agent, got nil")
	}
}

// ---------------------------------------------------------------------------
// Simulate busy → stopping transition (agent was doing work)
// ---------------------------------------------------------------------------

func TestLifecycle_Integration_BusyThenTerminate(t *testing.T) {
	ctx := context.Background()
	creator := NewCreator()
	initializer := NewInitializer()
	terminator := NewTerminator()

	inst, _ := creator.Create(ctx, makeSpec("int-agent-busy"))
	_ = initializer.Initialize(ctx, inst)

	// Simulate that agent is executing a turn.
	if err := inst.TransitionTo(runtime.AgentBusy); err != nil {
		t.Fatalf("unexpected error moving to Busy: %v", err)
	}

	// Terminate from Busy state should succeed.
	if err := terminator.Terminate(ctx, inst); err != nil {
		t.Fatalf("Terminate from Busy failed: %v", err)
	}
	if inst.State() != runtime.AgentTerminated {
		t.Errorf("expected Terminated after Busy→Stop, got %v", inst.State())
	}
}
