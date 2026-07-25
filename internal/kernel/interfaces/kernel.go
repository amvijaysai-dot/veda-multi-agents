// Package interfaces defines the core contracts for the VEDA Agent Runtime kernel.
// All components that interact with the kernel must depend only on these interfaces,
// never on concrete implementations in kernel/impl.
package interfaces

import (
	"context"

	"github.com/veda/agent-runtime/internal/types/event"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// Kernel defines the core lifecycle and orchestration interface for the VEDA Agent Runtime.
// Implementations of this interface manage subsystem registration, event routing, and
// the overall operational state of the runtime.
//
// Lifecycle contract:
//
//	Init → Start → (Suspend ↔ Resume) → Stop
type Kernel interface {
	// Init initializes the kernel with the provided context.
	// It configures core services but does not begin processing work.
	// Init must be called exactly once before any other method.
	Init(ctx context.Context) error

	// Start begins kernel operation, initializing and starting all registered subsystems
	// in dependency order. Start must be called after Init succeeds.
	Start(ctx context.Context) error

	// Stop initiates a graceful shutdown, stopping all subsystems in reverse initialization
	// order. It waits for ongoing operations to complete before returning.
	// Stop is idempotent; calling it more than once is safe.
	Stop(ctx context.Context) error

	// Suspend temporarily pauses kernel operation. New work is rejected but
	// in-flight operations are allowed to reach safe completion points.
	// Suspend must be called after Start and before Stop.
	Suspend(ctx context.Context) error

	// Resume resumes kernel operation after a successful Suspend.
	Resume(ctx context.Context) error

	// GetStatus returns the current operational status of the kernel.
	// This method is safe to call from any goroutine at any time.
	GetStatus() runtime.RuntimeStatus

	// RegisterSubsystem registers a named subsystem with the kernel.
	// Returns an error if a subsystem with the same name is already registered.
	// Registration must occur before Init is called.
	RegisterSubsystem(name string, subsystem Subsystem) error

	// UnregisterSubsystem removes a previously registered subsystem by name.
	// Returns an error if no subsystem with that name exists, or if the kernel
	// is in a state that does not permit subsystem removal.
	UnregisterSubsystem(name string) error

	// GetSubsystem retrieves a registered subsystem by name.
	// Returns an error if no subsystem with that name exists.
	GetSubsystem(name string) (Subsystem, error)

	// PublishEvent publishes an event to the kernel's internal event bus.
	// Delivery to subscribers is asynchronous and best-effort for non-critical events.
	PublishEvent(evt event.Event) error

	// SubscribeToEvent registers a handler function for events of the specified type.
	// Handlers are invoked asynchronously and must not block for extended periods.
	// Returns a SubscriptionID that must be retained for later unsubscription.
	SubscribeToEvent(eventType event.Type, handler func(event.Event)) (SubscriptionID, error)

	// UnsubscribeFromEvent removes the subscription identified by the given SubscriptionID.
	// If the ID is not found, the call is a no-op.
	UnsubscribeFromEvent(id SubscriptionID) error
}
