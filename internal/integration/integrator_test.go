package integration

import (
	"context"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/config"
	"github.com/veda/agent-runtime/internal/lifecycle/spec"
)

func TestRuntimeIntegrator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	integrator := NewRuntimeIntegrator()

	// Initialize with defaults (nil for subsystems which will be ignored or mocked in real scenarios)
	// For testing, just wiring the kernel is enough to prove the integrator logic.
	cfg := config.DefaultConfig()
	err := integrator.InitSubsystems(ctx, cfg, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to init subsystems: %v", err)
	}

	// Test Start
	err = integrator.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start integrator: %v", err)
	}

	// Test CreateAgent
	agentSpec := &spec.AgentSpec{
		ID:          "agent-123",
		Name:        "TestAgent",
		Description: "A test agent",
		ModelID:     "gpt-4",
	}
	id, err := integrator.CreateAgent(ctx, agentSpec)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty agent ID")
	}

	// GetAgent is removed or simplified since we don't have a kernel-level agent registry exposed


	// Test Stop
	err = integrator.Stop(ctx)
	if err != nil {
		t.Fatalf("failed to stop integrator: %v", err)
	}
}
