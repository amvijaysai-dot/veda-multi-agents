// Package tracing provides distributed tracing capabilities.
package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemSpan implements the Span interface.
type InMemSpan struct {
	mu           sync.Mutex
	traceID      string
	spanID       string
	parentSpanID string
	operation    string
	startTime    time.Time
	endTime      time.Time
	attributes   map[string]string
	events       []Event
	ended        bool
}

func (s *InMemSpan) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ended {
		s.endTime = time.Now()
		s.ended = true
	}
}

func (s *InMemSpan) AddEvent(name string, attributes map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended {
		return
	}

	// Copy attributes to avoid external mutation
	attrs := make(map[string]string, len(attributes))
	for k, v := range attributes {
		attrs[k] = v
	}

	s.events = append(s.events, Event{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: attrs,
	})
}

func (s *InMemSpan) SetAttribute(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ended {
		s.attributes[key] = value
	}
}

func (s *InMemSpan) TraceID() string {
	return s.traceID
}

func (s *InMemSpan) SpanID() string {
	return s.spanID
}

// InMemTracer implements the Tracer interface.
type InMemTracer struct {
	mu           sync.Mutex
	spanCounter  int
	traceCounter int
	// completedSpans stores spans that have ended for test verification
	completedSpans []*InMemSpan
}

// NewInMemTracer creates a new InMemTracer.
func NewInMemTracer() *InMemTracer {
	return &InMemTracer{
		completedSpans: make([]*InMemSpan, 0),
	}
}

func (t *InMemTracer) Start(ctx context.Context, operationName string) (context.Context, Span) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.spanCounter++
	spanID := fmt.Sprintf("span-%d", t.spanCounter)

	var traceID string
	var parentSpanID string

	if parent, ok := SpanFromContext(ctx); ok {
		traceID = parent.TraceID()
		parentSpanID = parent.SpanID()
	} else {
		t.traceCounter++
		traceID = fmt.Sprintf("trace-%d", t.traceCounter)
	}

	span := &InMemSpan{
		traceID:      traceID,
		spanID:       spanID,
		parentSpanID: parentSpanID,
		operation:    operationName,
		startTime:    time.Now(),
		attributes:   make(map[string]string),
	}

	// We don't append to completedSpans here. A real tracer might use a callback
	// or processor on End(), but for v0.8 memory simplicity we can just track them
	// all directly or rely on the caller maintaining a reference if they need it.
	// For testing, we'll track them.
	t.completedSpans = append(t.completedSpans, span)

	return ContextWithSpan(ctx, span), span
}

// GetSpans returns a copy of all tracked spans for testing/verification.
func (t *InMemTracer) GetSpans() []*InMemSpan {
	t.mu.Lock()
	defer t.mu.Unlock()
	spans := make([]*InMemSpan, len(t.completedSpans))
	copy(spans, t.completedSpans)
	return spans
}

var _ Span = (*InMemSpan)(nil)
var _ Tracer = (*InMemTracer)(nil)
