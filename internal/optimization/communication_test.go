package optimization

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/types/event"
)

func TestEventBatcher(t *testing.T) {
	var count atomic.Int32
	batcher := NewEventBatcher(3, 100*time.Millisecond, func(events []event.Event) {
		count.Add(int32(len(events)))
	})

	// Add 2 events; should not trigger flush yet (maxSize = 3)
	batcher.Add(event.NewBaseEvent("1", "test", "test"))
	batcher.Add(event.NewBaseEvent("2", "test", "test"))

	time.Sleep(10 * time.Millisecond)
	if count.Load() != 0 {
		t.Fatalf("expected 0 events flushed, got %d", count.Load())
	}

	// Add 1 more event; should trigger flush
	batcher.Add(event.NewBaseEvent("3", "test", "test"))

	time.Sleep(10 * time.Millisecond) // Wait a little just in case
	if count.Load() != 3 {
		t.Fatalf("expected 3 events flushed, got %d", count.Load())
	}

	// Add 1 event and wait for timeout
	batcher.Add(event.NewBaseEvent("4", "test", "test"))
	time.Sleep(150 * time.Millisecond)

	if count.Load() != 4 {
		t.Fatalf("expected 4 events flushed after timeout, got %d", count.Load())
	}

	// Test manual flush
	batcher.Add(event.NewBaseEvent("5", "test", "test"))
	batcher.Flush()
	if count.Load() != 5 {
		t.Fatalf("expected 5 events flushed after manual flush, got %d", count.Load())
	}

	batcher.Stop()
}
