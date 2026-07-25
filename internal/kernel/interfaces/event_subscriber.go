// Package interfaces defines the core contracts for the VEDA Agent Runtime kernel.
package interfaces

import (
	"github.com/veda/agent-runtime/internal/types/event"
)

// SubscriptionID is an opaque token returned by SubscribeToEvent.
// It uniquely identifies a subscription and is used to unsubscribe the handler
// without relying on function pointer comparisons, which are not reliable in Go.
type SubscriptionID uint64

// EventSubscriber defines the interface for subscribing to runtime events.
// Components that need to react to events should depend on this interface rather
// than the full Kernel interface, following the Interface Segregation Principle.
type EventSubscriber interface {
	// SubscribeToEvent registers a handler function for events of the specified type.
	// The handler is invoked for each matching event published on the event bus.
	// Handlers must be non-blocking; long-running work should be dispatched to a goroutine.
	// Returns a stable SubscriptionID that the caller must retain for later unsubscription.
	// Returns an error if subscription fails (e.g., event bus is stopped or handler is nil).
	SubscribeToEvent(eventType event.Type, handler func(event.Event)) (SubscriptionID, error)

	// UnsubscribeFromEvent removes the subscription identified by the given SubscriptionID.
	// If the ID is not found, the call is a no-op and returns nil.
	// Returns an error if the event bus is in an invalid state.
	UnsubscribeFromEvent(id SubscriptionID) error
}
