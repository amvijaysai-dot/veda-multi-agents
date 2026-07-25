// Package observability provides integration between the execution engine and observability systems.
package observability

import (
	"errors"
	"testing"

	execinterfaces "github.com/veda/agent-runtime/internal/execution/interfaces"
	"github.com/veda/agent-runtime/internal/metrics"
	"github.com/veda/agent-runtime/internal/tracing"
)

func TestEngineHooks(t *testing.T) {
	tracer := tracing.NewInMemTracer()
	collector := metrics.NewCollector()
	hooks := NewEngineHooks(tracer, collector)

	agentID := "agent-x"

	// Turn Start
	hooks.OnTurnStart(agentID, "session-123")

	// Reasoning
	hooks.OnReasoningStep(agentID, 1, execinterfaces.ReasoningOutput{
		Thought: "thinking about fetching data",
	})

	// Tool Call
	hooks.OnToolCall(agentID, execinterfaces.ToolCall{ToolName: "search"}, execinterfaces.ToolResult{Err: nil})
	hooks.OnToolCall(agentID, execinterfaces.ToolCall{ToolName: "broken"}, execinterfaces.ToolResult{Err: errors.New("fail")})

	// Turn End
	hooks.OnTurnEnd(agentID, execinterfaces.ExecutionResult{IterationsUsed: 1}, nil)

	// Validate Traces
	spans := tracer.GetSpans()
	// Should have: agent.turn, search, broken
	if len(spans) != 3 {
		t.Errorf("expected 3 spans, got %d", len(spans))
	}

	var rootSpan *tracing.InMemSpan
	for _, s := range spans {
		if s.SpanID() == "span-1" {
			rootSpan = s
		}
	}
	if rootSpan == nil {
		t.Fatal("missing root span")
	}

	// Validate Metrics
	snap := collector.GatherSnapshot()
	if len(snap) != 4 {
		t.Fatalf("expected 4 metrics registered, got %d", len(snap))
	}

	// Check tool_calls
	var toolCallsMD *metrics.MetricData
	for _, md := range snap {
		if md.Name == "agent_tool_calls_total" {
			toolCallsMD = md
		}
	}
	if toolCallsMD == nil {
		t.Fatal("missing tool calls metric")
	}

	foundSuccess := false
	foundError := false
	toolCallsMD.Data.Range(func(key, value any) bool {
		sv := value.(*metrics.SeriesValue)
		if sv.Labels["status"] == "success" && sv.Labels["tool"] == "search" {
			foundSuccess = true
			if sv.Value != 1 {
				t.Errorf("expected 1 successful search, got %d", sv.Value)
			}
		}
		if sv.Labels["status"] == "error" && sv.Labels["tool"] == "broken" {
			foundError = true
			if sv.Value != 1 {
				t.Errorf("expected 1 error broken, got %d", sv.Value)
			}
		}
		return true
	})

	if !foundSuccess || !foundError {
		t.Error("missing tool call metrics data")
	}
}
