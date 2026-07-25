// Package observability provides integration between the execution engine and observability systems.
package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	execinterfaces "github.com/veda/agent-runtime/internal/execution/interfaces"
	"github.com/veda/agent-runtime/internal/metrics"
	"github.com/veda/agent-runtime/internal/tracing"
)

// EngineHooks implements execinterfaces.ObservabilityHooks.
type EngineHooks struct {
	mu        sync.Mutex
	tracer    tracing.Tracer
	collector *metrics.Collector

	// Pre-registered metrics
	turnsStarted   metrics.Counter
	turnsEnded     metrics.Counter
	toolCalls      metrics.Counter
	reasoningSteps metrics.Counter

	// Active state tracking
	activeSpans map[string]tracing.Span
}

// NewEngineHooks creates a new EngineHooks.
func NewEngineHooks(tracer tracing.Tracer, collector *metrics.Collector) *EngineHooks {
	h := &EngineHooks{
		tracer:      tracer,
		collector:   collector,
		activeSpans: make(map[string]tracing.Span),
	}

	if collector != nil {
		h.turnsStarted = collector.RegisterCounter("agent_turns_started_total", "Total number of agent turns started")
		h.turnsEnded = collector.RegisterCounter("agent_turns_ended_total", "Total number of agent turns ended")
		h.toolCalls = collector.RegisterCounter("agent_tool_calls_total", "Total number of tool calls made")
		h.reasoningSteps = collector.RegisterCounter("agent_reasoning_steps_total", "Total number of reasoning steps completed")
	}

	return h
}

// OnTurnStart is called at the beginning of each agent turn.
func (h *EngineHooks) OnTurnStart(agentID, sessionID string) {
	if h.collector != nil {
		h.turnsStarted.Inc(metrics.Labels{"agent_id": agentID})
	}

	if h.tracer != nil {
		h.mu.Lock()
		defer h.mu.Unlock()

		ctx := context.Background()
		_, span := h.tracer.Start(ctx, "agent.turn")
		span.SetAttribute("agent_id", agentID)
		span.SetAttribute("session_id", sessionID)
		h.activeSpans[agentID] = span
	}
}

// OnReasoningStep is called after each successful reasoning step.
func (h *EngineHooks) OnReasoningStep(agentID string, iteration int, output execinterfaces.ReasoningOutput) {
	if h.collector != nil {
		h.reasoningSteps.Inc(metrics.Labels{"agent_id": agentID})
	}

	if h.tracer != nil {
		h.mu.Lock()
		defer h.mu.Unlock()
		
		if span, ok := h.activeSpans[agentID]; ok {
			actionType := "tools"
			if output.FinalAnswer != "" {
				actionType = "final_answer"
			}
			span.AddEvent("reasoning_step", map[string]string{
				"iteration": fmt.Sprintf("%d", iteration),
				"type":      actionType,
			})
		}
	}
}

// OnToolCall is called after each tool call completes.
func (h *EngineHooks) OnToolCall(agentID string, call execinterfaces.ToolCall, result execinterfaces.ToolResult) {
	if h.collector != nil {
		status := "success"
		if result.Err != nil {
			status = "error"
		}
		h.toolCalls.Inc(metrics.Labels{
			"agent_id": agentID,
			"tool":     call.ToolName,
			"status":   status,
		})
	}

	if h.tracer != nil {
		h.mu.Lock()
		defer h.mu.Unlock()
		
		if parentSpan, ok := h.activeSpans[agentID]; ok {
			ctx := tracing.ContextWithSpan(context.Background(), parentSpan)
			_, span := h.tracer.Start(ctx, "tool.call")
			span.SetAttribute("tool_name", call.ToolName)
			
			if result.Err != nil {
				span.SetAttribute("error", result.Err.Error())
			} else {
				// Record duration if we had it, but for v0.8 we just mark end immediately as this 
				// hook fires *after* the tool call completes. Realistically, we'd wrap the call.
				span.SetAttribute("success", "true")
			}
			span.End()
		}
	}
}

// OnTurnEnd is called when the turn completes, whether successfully or not.
func (h *EngineHooks) OnTurnEnd(agentID string, result execinterfaces.ExecutionResult, err error) {
	if h.collector != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		h.turnsEnded.Inc(metrics.Labels{
			"agent_id": agentID,
			"status":   status,
		})
	}

	if h.tracer != nil {
		h.mu.Lock()
		defer h.mu.Unlock()
		
		if span, ok := h.activeSpans[agentID]; ok {
			if err != nil {
				span.SetAttribute("error", err.Error())
			}
			
			// We can record token usage on turn end if it's available in result or context
			// For v0.8 just note the end.
			span.AddEvent("turn_end", map[string]string{
				"iterations": fmt.Sprintf("%d", result.IterationsUsed),
				"time":       time.Now().Format(time.RFC3339),
			})
			span.End()
			delete(h.activeSpans, agentID)
		}
	}
}

var _ execinterfaces.ObservabilityHooks = (*EngineHooks)(nil)
