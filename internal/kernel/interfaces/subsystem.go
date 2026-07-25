// Package interfaces defines the core contracts for the VEDA Agent Runtime kernel.
package interfaces

import "context"

// Subsystem defines the lifecycle contract that all kernel subsystems must implement.
// Subsystems are discrete components (e.g., event bus, scheduler, lifecycle manager)
// that are registered with the kernel and managed through its lifecycle.
//
// Lifecycle contract:
//
//	Init → Start → Stop
//
// Methods must be called in order; calling Start before Init is an error.
type Subsystem interface {
	// Init initializes the subsystem's internal state.
	// It should set up data structures, validate configuration, and prepare for operation
	// but must not start background goroutines or begin processing.
	// Returns an error if initialization fails.
	Init(ctx context.Context) error

	// Start begins subsystem operation after Init has completed.
	// It may launch background goroutines and start processing.
	// Returns an error if the subsystem fails to start.
	Start(ctx context.Context) error

	// Stop initiates a graceful shutdown of the subsystem.
	// It should stop accepting new work, drain in-progress work, stop all goroutines,
	// and release all resources before returning.
	// Stop is idempotent; calling it more than once is safe.
	Stop(ctx context.Context) error
}
