// Package impl provides the concrete implementation of the VEDA Agent Runtime kernel.
package impl

import (
	"fmt"
	"sync"

	"github.com/veda/agent-runtime/internal/kernel/interfaces"
)

// Registry manages the set of subsystems registered with the kernel.
// It provides thread-safe registration, retrieval, and removal of named subsystems.
//
// Registry is safe for concurrent use.
type Registry struct {
	mu         sync.RWMutex
	subsystems map[string]interfaces.Subsystem
	// names preserves insertion order for deterministic Init/Start/Stop sequencing.
	names []string
}

// newRegistry creates and returns a new empty Registry.
func newRegistry() *Registry {
	return &Registry{
		subsystems: make(map[string]interfaces.Subsystem),
	}
}

// Register adds a subsystem to the registry under the given name.
//
// Returns an error if:
//   - name is empty.
//   - subsystem is nil.
//   - A subsystem with the same name is already registered.
func (r *Registry) Register(name string, subsystem interfaces.Subsystem) error {
	if name == "" {
		return fmt.Errorf("subsystem name must not be empty")
	}
	if subsystem == nil {
		return fmt.Errorf("subsystem must not be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.subsystems[name]; exists {
		return fmt.Errorf("subsystem %q is already registered", name)
	}

	r.subsystems[name] = subsystem
	r.names = append(r.names, name)
	return nil
}

// Unregister removes the subsystem with the given name from the registry.
//
// Returns an error if no subsystem with that name is registered.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.subsystems[name]; !exists {
		return fmt.Errorf("subsystem %q is not registered", name)
	}

	delete(r.subsystems, name)

	// Remove from ordered names slice.
	for i, n := range r.names {
		if n == name {
			r.names = append(r.names[:i], r.names[i+1:]...)
			break
		}
	}
	return nil
}

// Get retrieves the subsystem registered under the given name.
//
// Returns an error if no subsystem with that name is registered.
func (r *Registry) Get(name string) (interfaces.Subsystem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sub, exists := r.subsystems[name]
	if !exists {
		return nil, fmt.Errorf("subsystem %q is not registered", name)
	}
	return sub, nil
}

// Names returns the ordered list of registered subsystem names.
// The order reflects the order in which subsystems were registered.
// This order is used for Init/Start (forward) and Stop (reverse) sequencing.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, len(r.names))
	copy(result, r.names)
	return result
}

// Len returns the number of subsystems currently registered.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.subsystems)
}
