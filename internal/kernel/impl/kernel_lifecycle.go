// Package impl provides the concrete implementation of the VEDA Agent Runtime kernel.
package impl

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/types/runtime"
)

// Start begins kernel operation by starting all registered subsystems in registration order.
//
// Start must be called after Init returns nil. It transitions the kernel status from
// Ready to Busy during startup, then back to Ready once all subsystems are running.
//
// Subsystem start sequencing is delegated to the Sequencer, which handles rollback
// on partial failure.
func (k *Kernel) Start(ctx context.Context) error {
	k.mu.Lock()
	if k.status != runtime.StatusReady {
		k.mu.Unlock()
		return fmt.Errorf("kernel.Start called in unexpected state %s; must be called from %s",
			k.status, runtime.StatusReady)
	}
	k.setStatus(runtime.StatusBusy)
	k.mu.Unlock()

	// Start the event bus so that subsystems can publish events during startup.
	k.eventBus.start()

	// Delegate start sequencing to the Sequencer. It handles rollback on partial failure.
	if err := k.sequencer.StartAll(ctx); err != nil {
		// Roll back to Ready to allow a retry or a clean Stop.
		k.mu.Lock()
		k.setStatus(runtime.StatusReady)
		k.mu.Unlock()
		return fmt.Errorf("kernel.Start: %w", err)
	}

	k.mu.Lock()
	k.setStatus(runtime.StatusReady)
	k.mu.Unlock()

	return nil
}

// Stop initiates a graceful shutdown, stopping all subsystems in reverse initialization order.
//
// Stop transitions the kernel to ShuttingDown, delegates stop sequencing to the Sequencer,
// and then transitions to Terminated. It is safe to call Stop multiple times; subsequent
// calls after the first are no-ops.
//
// Returns a combined error if any subsystem fails to stop cleanly; all subsystems
// are stopped regardless of individual failures (the Sequencer aggregates errors).
func (k *Kernel) Stop(ctx context.Context) error {
	k.mu.Lock()
	if k.status == runtime.StatusTerminated || k.status == runtime.StatusShuttingDown {
		k.mu.Unlock()
		return nil // idempotent
	}
	k.setStatus(runtime.StatusShuttingDown)
	k.mu.Unlock()

	// Stop the event bus before stopping subsystems to prevent new events during shutdown.
	k.eventBus.stop()

	// Delegate stop sequencing to the Sequencer. It stops all subsystems in reverse
	// registration order and aggregates errors rather than short-circuiting.
	stopErr := k.sequencer.StopAll(ctx)

	k.mu.Lock()
	k.setStatus(runtime.StatusTerminated)
	k.mu.Unlock()

	if stopErr != nil {
		return fmt.Errorf("kernel.Stop: %w", stopErr)
	}
	return nil
}

// Suspend temporarily pauses kernel operation.
//
// The kernel transitions to Suspending, then to Suspended. New work is rejected
// while suspended. Returns an error if the kernel is not in a state that allows suspension.
func (k *Kernel) Suspend(ctx context.Context) error {
	k.mu.Lock()
	if k.status != runtime.StatusReady && k.status != runtime.StatusBusy {
		k.mu.Unlock()
		return fmt.Errorf("kernel.Suspend called in unexpected state %s; must be Ready or Busy", k.status)
	}
	k.setStatus(runtime.StatusSuspending)
	k.mu.Unlock()

	// In v0.2, suspension is a logical state change only.
	// Future milestones will add actual quiescing of subsystems.
	_ = ctx // ctx will be used for subsystem quiescing in later milestones

	k.mu.Lock()
	k.setStatus(runtime.StatusSuspended)
	k.mu.Unlock()

	return nil
}

// Resume resumes kernel operation after a successful Suspend call.
//
// Returns an error if the kernel is not in the Suspended state.
func (k *Kernel) Resume(ctx context.Context) error {
	k.mu.Lock()
	if k.status != runtime.StatusSuspended {
		k.mu.Unlock()
		return fmt.Errorf("kernel.Resume called in unexpected state %s; must be called from Suspended", k.status)
	}
	k.setStatus(runtime.StatusResuming)
	k.mu.Unlock()

	// In v0.2, resumption is a logical state change only.
	// Future milestones will add actual resumption of subsystems.
	_ = ctx // ctx will be used for subsystem resumption in later milestones

	k.mu.Lock()
	k.setStatus(runtime.StatusReady)
	k.mu.Unlock()

	return nil
}
