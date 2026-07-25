package validation

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/lifecycle"
	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	shortterm "github.com/veda/agent-runtime/internal/memory/short_term"
)

// BenchmarkAgentCreation validates that agent creation takes <50ms.
func BenchmarkAgentCreation(b *testing.B) {
	creator := lifecycle.NewCreator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := &spec.AgentSpec{
			ID:      "bench-agent",
			Name:    "Bench Agent",
			ModelID: "gpt-bench",
		}
		_, err := creator.Create(ctx, s)
		if err != nil {
			b.Fatalf("failed to create agent: %v", err)
		}
	}
}

// BenchmarkMemoryOperations validates that memory operations take <1ms.
func BenchmarkMemoryOperations(b *testing.B) {
	store := shortterm.NewInMemoryShortTerm(0)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		key := "key-1"
		value := "test memory content"
		b.StartTimer()

		_ = store.Store(ctx, "agent-1", "session-1", key, value)
		_, _ = store.Retrieve(ctx, "agent-1", "session-1", key)
	}
}
