// Package tracing provides distributed tracing capabilities.
package tracing

import (
	"context"
	"testing"
)

func TestTracer_StartAndEndSpan(t *testing.T) {
	tracer := NewInMemTracer()
	ctx := context.Background()

	ctx2, span := tracer.Start(ctx, "root-op")
	if span.TraceID() != "trace-1" {
		t.Errorf("expected trace-1, got %s", span.TraceID())
	}
	if span.SpanID() != "span-1" {
		t.Errorf("expected span-1, got %s", span.SpanID())
	}

	span.SetAttribute("key1", "val1")
	span.AddEvent("something-happened", map[string]string{"foo": "bar"})

	// Verify propagation
	_, childSpan := tracer.Start(ctx2, "child-op")
	if childSpan.TraceID() != "trace-1" {
		t.Errorf("expected child to inherit trace-1, got %s", childSpan.TraceID())
	}
	if childSpan.SpanID() != "span-2" {
		t.Errorf("expected span-2, got %s", childSpan.SpanID())
	}

	imChild := childSpan.(*InMemSpan)
	if imChild.parentSpanID != "span-1" {
		t.Errorf("expected parent span-1, got %s", imChild.parentSpanID)
	}

	childSpan.End()
	span.End()

	imSpan := span.(*InMemSpan)
	if imSpan.endTime.IsZero() {
		t.Error("expected end time to be set")
	}
	if imSpan.attributes["key1"] != "val1" {
		t.Error("expected attribute to be saved")
	}
	if len(imSpan.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(imSpan.events))
	}
	if imSpan.events[0].Name != "something-happened" {
		t.Errorf("expected event name, got %s", imSpan.events[0].Name)
	}

	// Verify attribute setting after end is ignored
	span.SetAttribute("late", "no")
	if _, ok := imSpan.attributes["late"]; ok {
		t.Error("expected attribute set after End() to be ignored")
	}
}

func TestContextExtraction(t *testing.T) {
	ctx := context.Background()
	_, ok := SpanFromContext(ctx)
	if ok {
		t.Error("expected no span in empty context")
	}

	tracer := NewInMemTracer()
	ctx2, span := tracer.Start(ctx, "test")

	extracted, ok := SpanFromContext(ctx2)
	if !ok {
		t.Fatal("expected to find span in context")
	}
	if extracted.SpanID() != span.SpanID() {
		t.Error("extracted span mismatch")
	}
}
