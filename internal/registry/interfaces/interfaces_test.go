// Package interfaces defines the capability registry contracts.
package interfaces

import (
	"context"
	"testing"
)

type mockRegistry struct{}

func (m *mockRegistry) Register(_ context.Context, _ CapabilityMetadata) error { return nil }
func (m *mockRegistry) Deregister(_ context.Context, _, _ string) error        { return nil }
func (m *mockRegistry) Lookup(_ context.Context, _, _ string) (CapabilityMetadata, error) {
	return CapabilityMetadata{}, nil
}
func (m *mockRegistry) List(_ context.Context, _ map[string]string) ([]CapabilityMetadata, error) {
	return nil, nil
}

type mockLoader struct{}

func (m *mockLoader) Load(_ context.Context, _ CapabilitySource) (LoadedCapability, error) {
	return LoadedCapability{}, nil
}
func (m *mockLoader) Discover(_ context.Context, _ CapabilitySource) ([]CapabilitySource, error) {
	return nil, nil
}

type mockValidator struct{}

func (m *mockValidator) Validate(_ context.Context, _ LoadedCapability) (ValidationResult, error) {
	return ValidationResult{}, nil
}

type mockBinder struct{}

func (m *mockBinder) Bind(_ context.Context, _ CapabilityMetadata, _ BindingContext) (ExecutableCapability, error) {
	return nil, nil
}
func (m *mockBinder) Unbind(_ context.Context, _, _ string) error { return nil }

type mockExecutable struct{}

func (m *mockExecutable) Execute(_ context.Context, _ string) (string, error) { return "", nil }
func (m *mockExecutable) Close(_ context.Context) error                       { return nil }

var (
	_ CapabilityRegistry   = (*mockRegistry)(nil)
	_ CapabilityLoader     = (*mockLoader)(nil)
	_ CapabilityValidator  = (*mockValidator)(nil)
	_ CapabilityBinder     = (*mockBinder)(nil)
	_ ExecutableCapability = (*mockExecutable)(nil)
)

func TestInterfaces_Compliance(t *testing.T) {
	ctx := context.Background()
	var reg CapabilityRegistry = &mockRegistry{}
	_ = reg.Register(ctx, CapabilityMetadata{ID: "test"})

	var ldr CapabilityLoader = &mockLoader{}
	_, _ = ldr.Load(ctx, CapabilitySource{})

	var val CapabilityValidator = &mockValidator{}
	_, _ = val.Validate(ctx, LoadedCapability{})

	var bnd CapabilityBinder = &mockBinder{}
	_, _ = bnd.Bind(ctx, CapabilityMetadata{}, BindingContext{})

	var exec ExecutableCapability = &mockExecutable{}
	_, _ = exec.Execute(ctx, "{}")
}
