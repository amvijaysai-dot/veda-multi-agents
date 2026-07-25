// Package impl provides the concrete implementation of the VEDA Agent Runtime kernel.
// External packages must not import this package directly; they should depend only on
// the interfaces defined in kernel/interfaces.
package impl

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/veda/agent-runtime/internal/kernel/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// Kernel is the concrete implementation of the interfaces.Kernel contract.
// It manages the lifecycle of the runtime, coordinates subsystem initialization
// and shutdown via the Sequencer, and routes events through an internal event bus.
//
// Kernel is safe for concurrent use.
type Kernel struct {
	mu        sync.RWMutex
	status    runtime.RuntimeStatus
	registry  *Registry
	sequencer *Sequencer
	eventBus  *eventBus
}

// NewKernel creates and returns a new Kernel instance in the Uninitialized state.
// The returned Kernel has no subsystems registered; callers must register them
// before calling Init.
func NewKernel() *Kernel {
	r := newRegistry()
	return &Kernel{
		status:    runtime.StatusUninitialized,
		registry:  r,
		sequencer: newSequencer(r),
		eventBus:  newEventBus(),
	}
}

// Compile-time assertion: *Kernel must satisfy the Kernel interface.
var _ interfaces.Kernel = (*Kernel)(nil)

// GetStatus returns the current operational status of the kernel.
// This method is safe to call from any goroutine at any time.
func (k *Kernel) GetStatus() runtime.RuntimeStatus {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.status
}

// setStatus transitions the kernel to a new status.
// Callers must hold k.mu (write lock).
func (k *Kernel) setStatus(s runtime.RuntimeStatus) {
	k.status = s
}

// RegisterSubsystem registers a named subsystem with the kernel.
// Registration must occur before Init is called; registering after Init returns
// an error to prevent partial initialization races.
//
// Returns an error if:
//   - The kernel has already been initialized.
//   - A subsystem with the same name is already registered.
func (k *Kernel) RegisterSubsystem(name string, subsystem interfaces.Subsystem) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.status != runtime.StatusUninitialized {
		return fmt.Errorf("cannot register subsystem %q after Init has been called (status: %s)", name, k.status)
	}
	return k.registry.Register(name, subsystem)
}

// UnregisterSubsystem removes a previously registered subsystem by name.
// Removal is only permitted when the kernel is in Uninitialized or Terminated state.
// Returns an error if no subsystem with that name exists, or if the kernel lifecycle
// state does not permit removal.
func (k *Kernel) UnregisterSubsystem(name string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.status != runtime.StatusUninitialized && k.status != runtime.StatusTerminated {
		return fmt.Errorf(
			"cannot unregister subsystem %q in state %s; removal is only allowed in Uninitialized or Terminated state",
			name, k.status,
		)
	}
	return k.registry.Unregister(name)
}

// GetSubsystem retrieves a registered subsystem by name.
// Returns an error if no subsystem with that name exists.
func (k *Kernel) GetSubsystem(name string) (interfaces.Subsystem, error) {
	return k.registry.Get(name)
}

// PublishEvent publishes an event to the kernel's internal event bus.
func (k *Kernel) PublishEvent(evt event.Event) error {
	return k.eventBus.publish(evt)
}

// SubscribeToEvent registers a handler function for events of the specified type.
// Returns a SubscriptionID that must be retained for later use with UnsubscribeFromEvent.
func (k *Kernel) SubscribeToEvent(eventType event.Type, handler func(event.Event)) (interfaces.SubscriptionID, error) {
	return k.eventBus.subscribe(eventType, handler)
}

// UnsubscribeFromEvent removes the subscription identified by the given SubscriptionID.
// If the ID is not found, the call is a no-op.
func (k *Kernel) UnsubscribeFromEvent(id interfaces.SubscriptionID) error {
	return k.eventBus.unsubscribe(id)
}

// ---------------------------------------------------------------------------
// eventBus — internal, in-process event dispatcher
// ---------------------------------------------------------------------------

// subscription pairs a handler with its stable identifier.
type subscription struct {
	id      interfaces.SubscriptionID
	handler func(event.Event)
}

// eventBus is a lightweight, in-process event dispatcher used by the Kernel.
// Subscriptions are identified by opaque SubscriptionIDs generated from an atomic
// counter, eliminating the need for unreliable function pointer comparisons.
//
// It is intentionally simple for v0.2; a production event bus (channel-based,
// async delivery) will replace it in a later milestone.
type eventBus struct {
	mu            sync.RWMutex
	subscriptions map[event.Type][]subscription
	stopped       bool
	nextID        atomic.Uint64
}

// newEventBus creates and returns a new eventBus.
func newEventBus() *eventBus {
	return &eventBus{
		subscriptions: make(map[event.Type][]subscription),
	}
}

// publish dispatches an event to all registered handlers for its type.
// Each handler is invoked synchronously in this v0.2 implementation.
//
// NOTE(tech-debt): This is synchronous for v0.2 simplicity. A production
// implementation must use buffered channels for async, non-blocking delivery
// per the architecture's at-least-once and backpressure guarantees.
func (b *eventBus) publish(evt event.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.stopped {
		return fmt.Errorf("event bus is stopped; cannot publish event type %s", evt.Type())
	}

	for _, sub := range b.subscriptions[evt.Type()] {
		sub.handler(evt)
	}
	return nil
}

// subscribe registers a handler for the given event type and returns a stable
// SubscriptionID. The ID is generated from an atomic counter and is unique
// across the lifetime of this eventBus instance.
func (b *eventBus) subscribe(eventType event.Type, handler func(event.Event)) (interfaces.SubscriptionID, error) {
	if handler == nil {
		return 0, fmt.Errorf("handler must not be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		return 0, fmt.Errorf("event bus is stopped; cannot subscribe to event type %s", eventType)
	}

	id := interfaces.SubscriptionID(b.nextID.Add(1))
	b.subscriptions[eventType] = append(b.subscriptions[eventType], subscription{id: id, handler: handler})
	return id, nil
}

// unsubscribe removes the subscription with the given SubscriptionID.
// If the ID is not found, the call is a no-op and returns nil.
func (b *eventBus) unsubscribe(id interfaces.SubscriptionID) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for eventType, subs := range b.subscriptions {
		for i, sub := range subs {
			if sub.id == id {
				b.subscriptions[eventType] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}
	// Not found is a no-op per the interface contract.
	return nil
}

// stop marks the event bus as stopped, preventing further publish/subscribe operations.
func (b *eventBus) stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopped = true
}

// start marks the event bus as active.
func (b *eventBus) start() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopped = false
}
