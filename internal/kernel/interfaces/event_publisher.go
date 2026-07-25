// Package interfaces defines the core contracts for the VEDA Agent Runtime kernel.
package interfaces

import (
	"github.com/veda/agent-runtime/internal/types/event"
)

// EventPublisher defines the interface for publishing events to the runtime event bus.
// Components that need to emit events should depend on this interface rather than
// the full Kernel interface, following the Interface Segregation Principle.
type EventPublisher interface {
	// PublishEvent publishes an event to the event bus.
	// The event is delivered asynchronously to all registered subscribers.
	// Returns an error if the event cannot be accepted for delivery (e.g., bus is stopped).
	PublishEvent(evt event.Event) error
}
