// Package lifecycle provides capability lifecycle coordination.
package lifecycle

import (
	"context"
	"fmt"
	"testing"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

type mockLoader struct {
	discoverErr error
	loadErr     error
}

func (m *mockLoader) Discover(ctx context.Context, root interfaces.CapabilitySource) ([]interfaces.CapabilitySource, error) {
	if m.discoverErr != nil {
		return nil, m.discoverErr
	}
	return []interfaces.CapabilitySource{
		{Type: "mock", URI: "mock://cap1"},
		{Type: "mock", URI: "mock://cap2"},
	}, nil
}

func (m *mockLoader) Load(ctx context.Context, src interfaces.CapabilitySource) (interfaces.LoadedCapability, error) {
	if m.loadErr != nil {
		return interfaces.LoadedCapability{}, m.loadErr
	}
	return interfaces.LoadedCapability{
		Metadata: interfaces.CapabilityMetadata{ID: src.URI, Version: "1.0"},
	}, nil
}

type mockValidator struct {
	valid bool
}

func (m *mockValidator) Validate(ctx context.Context, cap interfaces.LoadedCapability) (interfaces.ValidationResult, error) {
	if !m.valid {
		return interfaces.ValidationResult{IsValid: false, Errors: []string{"invalid"}}, nil
	}
	return interfaces.ValidationResult{IsValid: true}, nil
}

type mockRegistry struct {
	registered int
}

func (m *mockRegistry) Register(ctx context.Context, metadata interfaces.CapabilityMetadata) error {
	m.registered++
	return nil
}
func (m *mockRegistry) Deregister(ctx context.Context, id, version string) error { return nil }
func (m *mockRegistry) Lookup(ctx context.Context, id, version string) (interfaces.CapabilityMetadata, error) {
	return interfaces.CapabilityMetadata{}, nil
}
func (m *mockRegistry) List(ctx context.Context, filter map[string]string) ([]interfaces.CapabilityMetadata, error) {
	return nil, nil
}

func TestLifecycleManager_BootSuccess(t *testing.T) {
	loader := &mockLoader{}
	validator := &mockValidator{valid: true}
	registry := &mockRegistry{}

	mgr := NewLifecycleManager(loader, validator, registry)
	ctx := context.Background()

	registered, errs := mgr.Boot(ctx, interfaces.CapabilitySource{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if registered != 2 {
		t.Errorf("expected 2 registered capabilities, got %d", registered)
	}
	if registry.registered != 2 {
		t.Errorf("expected registry to have 2 entries, got %d", registry.registered)
	}
}

func TestLifecycleManager_BootDiscoveryFail(t *testing.T) {
	loader := &mockLoader{discoverErr: fmt.Errorf("scan failed")}
	validator := &mockValidator{valid: true}
	registry := &mockRegistry{}

	mgr := NewLifecycleManager(loader, validator, registry)
	ctx := context.Background()

	registered, errs := mgr.Boot(ctx, interfaces.CapabilitySource{})
	if len(errs) != 1 || errs[0].Error() != "discovery failed: scan failed" {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if registered != 0 {
		t.Errorf("expected 0 registered, got %d", registered)
	}
}

func TestLifecycleManager_BootValidationFail(t *testing.T) {
	loader := &mockLoader{}
	validator := &mockValidator{valid: false}
	registry := &mockRegistry{}

	mgr := NewLifecycleManager(loader, validator, registry)
	ctx := context.Background()

	registered, errs := mgr.Boot(ctx, interfaces.CapabilitySource{})
	if len(errs) != 2 {
		t.Fatalf("expected 2 validation errors, got %d", len(errs))
	}
	if registered != 0 {
		t.Errorf("expected 0 registered, got %d", registered)
	}
}
