package optimization

import (
	"context"
	"testing"
	"time"
)

func TestEventPool(t *testing.T) {
	pool := NewEventPool()

	e := pool.Get("123", "test.event", "source1")
	if e.ID() != "123" {
		t.Fatalf("expected ID 123, got %s", e.ID())
	}

	pool.Put(e)

	e2 := pool.Get("456", "test.event2", "source2")
	if e2.ID() != "456" {
		t.Fatalf("expected ID 456, got %s", e2.ID())
	}
}

func TestByteSlicePool(t *testing.T) {
	pool := NewByteSlicePool(1024)

	b := pool.Get()
	if cap(*b) != 1024 {
		t.Fatalf("expected cap 1024, got %d", cap(*b))
	}

	*b = append(*b, []byte("hello")...)
	pool.Put(b)

	b2 := pool.Get()
	if len(*b2) != 0 {
		t.Fatalf("expected length 0, got %d", len(*b2))
	}
}

type mockConn struct {
	id int
}

func TestResourcePool(t *testing.T) {
	ctx := context.Background()
	counter := 0
	factory := func(ctx context.Context) (*mockConn, error) {
		counter++
		return &mockConn{id: counter}, nil
	}
	closer := func(c *mockConn) error { return nil }

	pool := NewResourcePool(2, factory, closer)

	// Acquire 1
	c1, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("failed to acquire: %v", err)
	}
	if c1.id != 1 {
		t.Fatalf("expected id 1, got %d", c1.id)
	}

	// Acquire 2
	c2, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("failed to acquire: %v", err)
	}

	// Pool is full. Context timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err = pool.Acquire(ctxTimeout)
	if err == nil {
		t.Fatalf("expected timeout error")
	}

	// Release 1 and acquire again
	pool.Release(c1)

	c3, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("failed to acquire: %v", err)
	}

	// Should reuse the released connection
	if c3.id != 1 {
		t.Fatalf("expected id 1 (reused), got %d", c3.id)
	}

	pool.Release(c3)
	pool.Release(c2)

	err = pool.Close()
	if err != nil {
		t.Fatalf("failed to close pool: %v", err)
	}
}
