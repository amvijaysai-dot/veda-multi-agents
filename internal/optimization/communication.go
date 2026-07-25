// Package optimization provides utilities to enhance the performance and resource efficiency of the runtime.
package optimization

import (
	"context"
	"sync"
	"time"

	"github.com/veda/agent-runtime/internal/types/event"
)

// EventBatcher collects events and dispatches them in bulk to reduce lock contention
// on the internal event bus.
type EventBatcher struct {
	mu         sync.Mutex
	buffer     []event.Event
	maxSize    int
	timeout    time.Duration
	flushTimer *time.Timer
	dispatcher func([]event.Event)
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewEventBatcher creates a new EventBatcher that flushes when maxSize is reached
// or timeout elapses. The dispatcher function is called with the accumulated events.
func NewEventBatcher(maxSize int, timeout time.Duration, dispatcher func([]event.Event)) *EventBatcher {
	ctx, cancel := context.WithCancel(context.Background())
	eb := &EventBatcher{
		buffer:     make([]event.Event, 0, maxSize),
		maxSize:    maxSize,
		timeout:    timeout,
		dispatcher: dispatcher,
		ctx:        ctx,
		cancel:     cancel,
	}

	eb.flushTimer = time.AfterFunc(timeout, eb.flushOnTimer)
	eb.flushTimer.Stop() // Start stopped until first item is added
	return eb
}

// Add adds an event to the batch. If the batch reaches maxSize, it is flushed synchronously.
func (eb *EventBatcher) Add(e event.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	select {
	case <-eb.ctx.Done():
		return // Batcher is stopped
	default:
	}

	if len(eb.buffer) == 0 {
		// First item added to empty buffer, start the timer
		eb.flushTimer.Reset(eb.timeout)
	}

	eb.buffer = append(eb.buffer, e)

	if len(eb.buffer) >= eb.maxSize {
		eb.flushLocked()
	}
}

// flushOnTimer is called when the timeout triggers.
func (eb *EventBatcher) flushOnTimer() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	select {
	case <-eb.ctx.Done():
		return // Stopped
	default:
	}

	eb.flushLocked()
}

// flushLocked performs the actual flush. Caller must hold the mutex.
func (eb *EventBatcher) flushLocked() {
	if len(eb.buffer) == 0 {
		return
	}

	// Make a copy for dispatching so we can reset our buffer immediately
	batch := make([]event.Event, len(eb.buffer))
	copy(batch, eb.buffer)
	eb.buffer = eb.buffer[:0]

	// Stop timer since buffer is now empty. It will be restarted on next Add.
	eb.flushTimer.Stop()

	// Dispatch synchronously or asynchronously depending on caller needs.
	// We dispatch synchronously here; the dispatcher itself can launch a goroutine if needed.
	eb.dispatcher(batch)
}

// Flush forces an immediate flush of any buffered events.
func (eb *EventBatcher) Flush() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.flushLocked()
}

// Stop shuts down the batcher and prevents further adds or flushes.
func (eb *EventBatcher) Stop() {
	eb.cancel()
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.flushTimer != nil {
		eb.flushTimer.Stop()
	}
	eb.buffer = nil
}
