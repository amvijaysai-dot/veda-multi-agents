// Package lifecycle implements the agent lifecycle subsystem for the VEDA Agent Runtime.
package lifecycle

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/lifecycle/instance"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// InitHook is a function that is called as part of the agent initialization
// sequence. Hooks are executed in the order they are registered. If any hook
// returns an error, initialization is aborted and the agent is moved to a
// terminal error state.
//
// Each hook receives the context and the agent instance being initialized.
// Hooks must not call Initializer.Initialize recursively.
//
// In Milestone v0.3, hooks for memory allocation, capability binding, and model
// loading are stubbed. They will be replaced with real implementations in later
// milestones.
type InitHook func(ctx context.Context, inst *instance.AgentInstance) error

// Initializer performs the agent initialization sequence defined in Section 4
// of the architecture specification (Initialization phase).
//
// The sequence:
//  1. Transition agent to AgentInitializing.
//  2. Run each registered InitHook in order.
//  3. On success, transition agent to AgentReady.
//  4. On any hook failure, call inst.SetError and return an error.
//
// Initializer is safe to call from multiple goroutines for different agent
// instances; the caller is responsible for ensuring each instance is only
// initialized once.
type Initializer struct {
	hooks []InitHook
}

// NewInitializer creates and returns a new Initializer.
// If hooks is nil or empty, the initializer runs no extra steps — it simply
// transitions the agent to Ready after successfully moving it to Initializing.
// This is the correct behavior for v0.3 where all dependency hooks are stubbed.
func NewInitializer(hooks ...InitHook) *Initializer {
	return &Initializer{hooks: hooks}
}

// Initialize runs the initialization sequence for the provided AgentInstance.
// The instance must be in the AgentCreating state when this method is called;
// it is the caller's (Creator's) responsibility to ensure this precondition.
//
// On success:
//   - The agent is in the AgentReady state.
//   - Returns nil.
//
// On failure:
//   - The agent is moved to AgentTerminated via inst.SetError.
//   - Returns a non-nil, descriptive error.
func (i *Initializer) Initialize(ctx context.Context, inst *instance.AgentInstance) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("initializer.Initialize: context cancelled before start: %w", err)
	}

	// Step 1: Transition to Initializing.
	if err := inst.TransitionTo(runtime.AgentInitializing); err != nil {
		return fmt.Errorf("initializer.Initialize: cannot start initialization: %w", err)
	}

	// Step 2: Run each registered hook.
	for idx, hook := range i.hooks {
		if err := ctx.Err(); err != nil {
			msg := fmt.Sprintf("context cancelled after hook %d", idx)
			inst.SetError(msg)
			return fmt.Errorf("initializer.Initialize: %s: %w", msg, err)
		}

		if err := hook(ctx, inst); err != nil {
			msg := fmt.Sprintf("initialization hook %d failed: %v", idx, err)
			inst.SetError(msg)
			return fmt.Errorf("initializer.Initialize: %s", msg)
		}
	}

	// Step 3: All hooks passed — transition to Ready.
	if err := inst.TransitionTo(runtime.AgentReady); err != nil {
		msg := fmt.Sprintf("cannot transition to Ready after initialization: %v", err)
		inst.SetError(msg)
		return fmt.Errorf("initializer.Initialize: %s", msg)
	}

	return nil
}
