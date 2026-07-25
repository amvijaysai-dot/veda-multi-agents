package optimization

import (
	"context"
	"errors"
	"sync"
)

// ResourcePool manages a bounded pool of reusable resources (e.g., connections).
type ResourcePool[T any] struct {
	mu        sync.Mutex
	resources []T
	factory   func(context.Context) (T, error)
	closer    func(T) error
	maxSize   int
	sem       chan struct{}
}

// NewResourcePool creates a new bounded resource pool.
func NewResourcePool[T any](maxSize int, factory func(context.Context) (T, error), closer func(T) error) *ResourcePool[T] {
	return &ResourcePool[T]{
		resources: make([]T, 0, maxSize),
		factory:   factory,
		closer:    closer,
		maxSize:   maxSize,
		sem:       make(chan struct{}, maxSize),
	}
}

// Acquire gets a resource from the pool or creates a new one if the pool isn't full.
// Blocks until a resource is available or context is canceled.
func (p *ResourcePool[T]) Acquire(ctx context.Context) (T, error) {
	var zero T

	// Wait for a slot
	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		return zero, ctx.Err()
	}

	p.mu.Lock()
	if len(p.resources) > 0 {
		res := p.resources[len(p.resources)-1]
		p.resources = p.resources[:len(p.resources)-1]
		p.mu.Unlock()
		return res, nil
	}
	p.mu.Unlock()

	// Need to create a new resource
	res, err := p.factory(ctx)
	if err != nil {
		// Release slot
		<-p.sem
		return zero, err
	}

	return res, nil
}

// Release returns a resource to the pool.
func (p *ResourcePool[T]) Release(res T) {
	p.mu.Lock()
	p.resources = append(p.resources, res)
	p.mu.Unlock()
	<-p.sem // Free a slot
}

// Close destroys all pooled resources.
func (p *ResourcePool[T]) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for _, res := range p.resources {
		if err := p.closer(res); err != nil {
			errs = append(errs, err)
		}
	}
	p.resources = nil

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
