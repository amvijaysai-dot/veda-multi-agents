// Package memory provides memory lifecycle integration.
package memory

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/lifecycle"
	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	"github.com/veda/agent-runtime/internal/memory/consolidation"
	"github.com/veda/agent-runtime/internal/memory/long_term"
	"github.com/veda/agent-runtime/internal/memory/privacy"
	"github.com/veda/agent-runtime/internal/memory/short_term"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

func TestMemoryLifecycleIntegration_FullCycle(t *testing.T) {
	ctx := context.Background()
	
	// Setup memory components
	stMem := shortterm.NewInMemoryShortTerm(0)
	ltMem := longterm.NewInMemoryLongTerm()
	scrubber := privacy.NewRegexScrubber(nil, "[MASKED]")
	
	// Wrap consolidator to verify it was called
	consolidator := consolidation.NewBasicConsolidator(stMem, ltMem, scrubber, 0)
	
	manager := NewMemoryLifecycleManager(stMem, consolidator)
	
	// Prepare lifecycle components
	initializer := lifecycle.NewInitializer(manager.InitHook())
	terminator := lifecycle.NewTerminator(manager.CleanupHook())
	creator := lifecycle.NewCreator()
	
	// 1. Create and Initialize agent
	agentSpec := &spec.AgentSpec{
		ID:            "agent-mem-test",
		Name:          "Memory Test Agent",
		ModelID:       "test-model",
		MaxIterations: 5,
	}
	
	inst, err := creator.Create(ctx, agentSpec)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	err = initializer.Initialize(ctx, inst)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}
	
	if inst.State() != runtime.AgentReady {
		t.Fatalf("Expected AgentReady, got %v", inst.State())
	}
	
	// 2. Simulate agent execution by storing memories
	sessionID := "default-session"
	
	_ = stMem.Store(ctx, inst.ID(), sessionID, "temp_key", "temporary data")
	_ = stMem.Store(ctx, inst.ID(), sessionID, "persist_key", "my email is test@example.com")
	_ = stMem.PersistenceHint(ctx, inst.ID(), sessionID, "persist_key")
	
	// 3. Terminate agent (triggers cleanup and consolidation)
	// Terminate handles transitioning to AgentStopping
	err = terminator.Terminate(ctx, inst)
	if err != nil {
		t.Fatalf("Failed to terminate agent: %v", err)
	}
	
	if inst.State() != runtime.AgentTerminated {
		t.Fatalf("Expected AgentTerminated, got %v", inst.State())
	}

	// 4. Verify results
	// Short-term memory should be cleared
	_, err = stMem.Retrieve(ctx, inst.ID(), sessionID, "persist_key")
	if err == nil {
		t.Error("Expected short-term memory to be cleared, but key was found")
	}

	// Long-term memory should contain the consolidated, scrubbed value
	val, err := ltMem.Retrieve(ctx, inst.ID(), "persist_key")
	if err != nil {
		t.Errorf("Expected key in long-term memory, got error: %v", err)
	}
	if val != "my email is [MASKED]" {
		t.Errorf("Expected scrubbed value 'my email is [MASKED]', got %q", val)
	}

	// Un-hinted memory should not be consolidated
	_, err = ltMem.Retrieve(ctx, inst.ID(), "temp_key")
	if err == nil {
		t.Error("Expected un-hinted key to not be in long-term memory")
	}
}

func TestNewMemoryLifecycleManager_PanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil shortTerm, got none")
		}
	}()
	NewMemoryLifecycleManager(nil, nil)
}
