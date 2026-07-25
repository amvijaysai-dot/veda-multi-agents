// Package registry provides the capability registry implementation for the
// VEDA Agent Runtime.
package registry

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

// InMemoryRegistry implements interfaces.CapabilityRegistry using a thread-safe
// map. It supports semantic versioning implicitly by sorting versions.
type InMemoryRegistry struct {
	mu           sync.RWMutex
	capabilities map[string]map[string]interfaces.CapabilityMetadata // id -> version -> metadata
}

// NewInMemoryRegistry creates a new InMemoryRegistry.
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		capabilities: make(map[string]map[string]interfaces.CapabilityMetadata),
	}
}

// Register adds a capability to the registry.
func (r *InMemoryRegistry) Register(ctx context.Context, metadata interfaces.CapabilityMetadata) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if metadata.ID == "" || metadata.Version == "" {
		return fmt.Errorf("capability ID and Version are required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.capabilities[metadata.ID]; !ok {
		r.capabilities[metadata.ID] = make(map[string]interfaces.CapabilityMetadata)
	}

	if _, exists := r.capabilities[metadata.ID][metadata.Version]; exists {
		return fmt.Errorf("capability %q version %q is already registered", metadata.ID, metadata.Version)
	}

	r.capabilities[metadata.ID][metadata.Version] = metadata
	return nil
}

// Deregister removes a capability from the registry.
func (r *InMemoryRegistry) Deregister(ctx context.Context, id, version string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" || version == "" {
		return fmt.Errorf("capability ID and Version are required for deregistration")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	versions, ok := r.capabilities[id]
	if !ok {
		return fmt.Errorf("capability %q not found", id)
	}

	if _, exists := versions[version]; !exists {
		return fmt.Errorf("capability %q version %q not found", id, version)
	}

	delete(versions, version)

	// Cleanup map if empty
	if len(versions) == 0 {
		delete(r.capabilities, id)
	}
	return nil
}

// Lookup finds a capability by ID.
// If version is empty, it returns the highest available version (lexicographical sort for simplicity in v0.7).
func (r *InMemoryRegistry) Lookup(ctx context.Context, id, version string) (interfaces.CapabilityMetadata, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.CapabilityMetadata{}, err
	}
	if id == "" {
		return interfaces.CapabilityMetadata{}, fmt.Errorf("capability ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.capabilities[id]
	if !ok || len(versions) == 0 {
		return interfaces.CapabilityMetadata{}, fmt.Errorf("capability %q not found", id)
	}

	if version != "" {
		meta, exists := versions[version]
		if !exists {
			return interfaces.CapabilityMetadata{}, fmt.Errorf("capability %q version %q not found", id, version)
		}
		return meta, nil
	}

	// Find highest version if not specified
	var keys []string
	for k := range versions {
		keys = append(keys, k)
	}

	// Note: In a production system (v0.8+), this should use proper semantic versioning sort
	// (e.g. hashicorp/go-version). For v0.7, a simple string sort suffices for the contract.
	sort.Strings(keys)
	highestVersion := keys[len(keys)-1]

	return versions[highestVersion], nil
}

// List returns all registered capabilities matching an optional filter.
func (r *InMemoryRegistry) List(ctx context.Context, filter map[string]string) ([]interfaces.CapabilityMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []interfaces.CapabilityMetadata

	for id, versions := range r.capabilities {
		for ver, meta := range versions {
			if matchesFilter(meta, filter) {
				// We still append them all, a caller might want all versions.
				result = append(result, meta)
			}
			// In v0.7, if they filter just by ID, we return all versions for that ID.
			_ = id
			_ = ver
		}
	}

	return result, nil
}

func matchesFilter(meta interfaces.CapabilityMetadata, filter map[string]string) bool {
	if len(filter) == 0 {
		return true
	}

	for k, v := range filter {
		switch strings.ToLower(k) {
		case "id":
			if meta.ID != v {
				return false
			}
		case "name":
			if !strings.Contains(strings.ToLower(meta.Name), strings.ToLower(v)) {
				return false
			}
		case "version":
			if meta.Version != v {
				return false
			}
		case "author":
			if meta.Author != v {
				return false
			}
		}
	}
	return true
}

var _ interfaces.CapabilityRegistry = (*InMemoryRegistry)(nil)
