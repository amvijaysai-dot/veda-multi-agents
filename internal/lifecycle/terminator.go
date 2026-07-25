// Package lifecycle implements the agent lifecycle subsystem for the VEDA Agent Runtime.
package lifecycle

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/lifecycle/instance"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// CleanupHook is a function called as part of the agent termination sequence.
// Hooks are executed in the order they are registered (typically reverse capability
// binding order). Errors from cleanup hooks are collected and reported but do not
// abort the termination sequence — the agent will always reach AgentTerminated.
//
// In Milestone v0.3, hooks for capability unbinding and memory release are
// stubbed. They will be replaced with real implementations in later milestones.
type CleanupHook func(ctx context.Context, inst *instance.AgentInstance) error

// Terminator performs the agent termination and resource cleanup sequence
// defined in Section 4 of the architecture specification (Shutdown phase).
//
// The sequence:
//  1. Transition agent to AgentStopping.
//  2. Run each registered CleanupHook in order, collecting any errors.
//  3. Transition agent to AgentTerminated.
//  4. Return a combined error if any hook failed, nil otherwise.
//
// Unlike Initializer, Terminator does not abort on the first hook failure.
// All cleanup hooks always run regardless of individual failures, since
// partial cleanup would leave resources in an inconsistent state.
//
// Terminator is safe to call from multiple goroutines for different agent
// instances; the caller is responsible for ensuring each instance is only
// terminated once.
type Terminator struct {
	hooks []CleanupHook
}

// NewTerminator creates and returns a new Terminator.
// If hooks is nil or empty, the terminator runs no extra cleanup steps.
func NewTerminator(hooks ...CleanupHook) *Terminator {
	return &Terminator{hooks: hooks}
}

// Terminate runs the termination sequence for the provided AgentInstance.
// The instance must be in a state that permits transition to AgentStopping:
// AgentReady, AgentBusy, or AgentSuspended. The caller is responsible for
// ensuring the agent has reached one of these states before calling Terminate.
//
// On success (all hooks pass):
//   - The agent is in the AgentTerminated state.
//   - Returns nil.
//
// On partial failure (some hooks return errors):
//   - The agent is still moved to AgentTerminated.
//   - Returns a combined error describing all hook failures.
//
// On failure to transition to AgentStopping:
//   - Returns an error; the agent state is unchanged.
func (t *Terminator) Terminate(ctx context.Context, inst *instance.AgentInstance) error {
	// Step 1: Transition to Stopping.
	if err := inst.TransitionTo(runtime.AgentStopping); err != nil {
		return fmt.Errorf("terminator.Terminate: cannot begin termination: %w", err)
	}

	// Step 2: Run all cleanup hooks, collecting errors.
	// We do NOT short-circuit on error; all hooks must have a chance to run
	// so that partial resource release is avoided where possible.
	var hookErrors []error
	for idx, hook := range t.hooks {
		if err := hook(ctx, inst); err != nil {
			hookErrors = append(hookErrors, fmt.Errorf("cleanup hook %d: %w", idx, err))
		}
	}

	// Step 3: Transition to Terminated regardless of hook outcomes.
	if err := inst.TransitionTo(runtime.AgentTerminated); err != nil {
		// This should not happen given the state machine, but guard defensively.
		return fmt.Errorf("terminator.Terminate: cannot complete termination: %w", err)
	}

	// Step 4: Report aggregated hook errors if any.
	if len(hookErrors) > 0 {
		return fmt.Errorf("terminator.Terminate: %d cleanup hook(s) failed: %w",
			len(hookErrors), joinErrors(hookErrors))
	}

	return nil
}

// joinErrors combines multiple errors into a single error. It is a minimal
// implementation sufficient for v0.3; a future milestone may use errors.Join
// (Go 1.20+) if the project upgrades its minimum Go version.
func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	msg := errs[0].Error()
	for _, e := range errs[1:] {
		msg += "; " + e.Error()
	}
	return fmt.Errorf("%s", msg)
}
