// Package tracing provides distributed tracing interfaces for the VEDA Agent Runtime.
package tracing

import (
	"context"
	"time"
)

// Span represents a single operation within a trace.
type Span interface {
	// End completes the span, recording its duration.
	End()

	// AddEvent records an event at the current time within the span.
	AddEvent(name string, attributes map[string]string)

	// SetAttribute adds a key-value pair to the span.
	SetAttribute(key, value string)

	// TraceID returns the globally unique identifier for the entire trace.
	TraceID() string

	// SpanID returns the unique identifier for this specific span.
	SpanID() string
}

// Tracer defines the interface for creating and managing spans.
type Tracer interface {
	// Start begins a new span. If the context contains a parent span,
	// the new span will be a child of that parent.
	// Returns a new context containing the child span, and the span itself.
	Start(ctx context.Context, operationName string) (context.Context, Span)
}

type contextKey struct{}

var spanContextKey = contextKey{}

// ContextWithSpan injects a span into a context.
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, spanContextKey, span)
}

// SpanFromContext extracts a span from a context, if one exists.
func SpanFromContext(ctx context.Context) (Span, bool) {
	span, ok := ctx.Value(spanContextKey).(Span)
	return span, ok
}

// Event represents a point-in-time occurrence within a span.
type Event struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]string
}
