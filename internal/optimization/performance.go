package optimization

import (
	"sync"

	"github.com/veda/agent-runtime/internal/types/event"
)

// EventPool provides a reusable pool of BaseEvent objects to minimize allocation
// overhead and GC pressure during high-throughput event publishing.
type EventPool struct {
	pool sync.Pool
}

// NewEventPool creates a new EventPool.
func NewEventPool() *EventPool {
	return &EventPool{
		pool: sync.Pool{
			New: func() interface{} {
				// Allocate a blank event
				return event.NewBaseEvent("", "", "")
			},
		},
	}
}

// Get retrieves an event from the pool and initializes it.
func (p *EventPool) Get(id string, eventType event.Type, source string, opts ...event.EventOption) *event.BaseEvent {
	e := p.pool.Get().(*event.BaseEvent)

	// Reset state
	*e = *event.NewBaseEvent(id, eventType, source, opts...)

	return e
}

// Put returns the event to the pool. Callers must not use the event after putting it back.
func (p *EventPool) Put(e *event.BaseEvent) {
	// Clear references to help GC if payload is large
	*e = *event.NewBaseEvent("", "", "")
	p.pool.Put(e)
}

// ByteSlicePool provides a pool for byte slices to reduce allocation during serialization.
type ByteSlicePool struct {
	pool sync.Pool
}

// NewByteSlicePool creates a pool of byte slices with an initial capacity.
func NewByteSlicePool(capacity int) *ByteSlicePool {
	return &ByteSlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, capacity)
				return &b
			},
		},
	}
}

// Get retrieves a byte slice from the pool.
func (p *ByteSlicePool) Get() *[]byte {
	return p.pool.Get().(*[]byte)
}

// Put returns the byte slice to the pool, resetting its length.
func (p *ByteSlicePool) Put(b *[]byte) {
	*b = (*b)[:0]
	p.pool.Put(b)
}
