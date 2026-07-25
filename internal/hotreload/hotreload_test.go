package hotreload

import (
	"context"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/config"
	"github.com/veda/agent-runtime/internal/kernel/impl"
	registryinterfaces "github.com/veda/agent-runtime/internal/registry/interfaces"
)

// mockRegistry implements CapabilityRegistry for testing.
type mockRegistry struct {
	caps map[string]registryinterfaces.CapabilityMetadata
}

func (m *mockRegistry) Register(ctx context.Context, metadata registryinterfaces.CapabilityMetadata) error {
	m.caps[metadata.ID] = metadata
	return nil
}

func (m *mockRegistry) Deregister(ctx context.Context, id, version string) error {
	delete(m.caps, id)
	return nil
}

func (m *mockRegistry) Lookup(ctx context.Context, id, version string) (registryinterfaces.CapabilityMetadata, error) {
	return m.caps[id], nil
}

func (m *mockRegistry) List(ctx context.Context, filter map[string]string) ([]registryinterfaces.CapabilityMetadata, error) {
	return nil, nil
}

func TestConfigReloader(t *testing.T) {
	k := impl.NewKernel()

	// Start with default config
	initialCfg := config.DefaultConfig()
	initialCfg.RuntimeID = "initial"

	cr := NewConfigReloader(initialCfg, k)

	cfg := cr.GetConfig()
	if cfg.RuntimeID != "initial" {
		t.Fatalf("expected initial config")
	}

	// Reload config (loads from env/defaults which has "veda-runtime-default")
	err := cr.Reload()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	cfg = cr.GetConfig()
	if cfg.RuntimeID != "veda-runtime-default" {
		t.Fatalf("expected reloaded config, got %s", cfg.RuntimeID)
	}
}

func TestCapabilityReloader(t *testing.T) {
	k := impl.NewKernel()
	reg := &mockRegistry{caps: make(map[string]registryinterfaces.CapabilityMetadata)}
	cr := NewCapabilityReloader(reg, k)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	meta := registryinterfaces.CapabilityMetadata{
		ID:      "test-cap",
		Version: "1.0",
	}

	err := cr.ReloadCapability(ctx, meta)
	if err != nil {
		t.Fatalf("failed to reload capability: %v", err)
	}

	if _, ok := reg.caps["test-cap"]; !ok {
		t.Fatalf("capability not registered")
	}
}
