// Package impl provides the concrete implementation of the VEDA Agent Runtime kernel.
package impl

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/types/runtime"
)

// Init initializes the kernel, transitioning it from Uninitialized to Ready.
//
// Init performs the following steps:
//  1. Validates the kernel is in the expected state.
//  2. Transitions status to Initializing.
//  3. Delegates subsystem initialization (in registration order) to the Sequencer.
//  4. Transitions status to Ready on success.
//
// If any subsystem fails to initialize, the Sequencer rolls back already-initialized
// subsystems in reverse order before the error is returned. The kernel does not have
// a well-defined state after a partial initialization failure; callers should treat
// this as a fatal startup error.
func (k *Kernel) Init(ctx context.Context) error {
	k.mu.Lock()
	if k.status != runtime.StatusUninitialized {
		k.mu.Unlock()
		return fmt.Errorf("kernel.Init called in unexpected state %s; must be called from %s",
			k.status, runtime.StatusUninitialized)
	}
	k.setStatus(runtime.StatusInitializing)
	k.mu.Unlock()

	// Delegate initialization sequencing to the Sequencer. It handles rollback
	// on partial failure — no duplicate logic required here.
	if err := k.sequencer.InitAll(ctx); err != nil {
		return fmt.Errorf("kernel.Init: %w", err)
	}

	k.mu.Lock()
	k.setStatus(runtime.StatusReady)
	k.mu.Unlock()

	return nil
}
