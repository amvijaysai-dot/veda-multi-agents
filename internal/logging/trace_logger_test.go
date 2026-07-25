// Package logging provides structured logging capabilities for the runtime.
package logging

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/veda/agent-runtime/internal/tracing"
)

func TestTraceLogger(t *testing.T) {
	var buf bytes.Buffer
	base := New(LevelDebug, &buf)
	tlog := NewTraceLogger(base)

	ctx := context.Background()
	tracer := tracing.NewInMemTracer()

	ctxWithSpan, _ := tracer.Start(ctx, "test-op")

	tlog.Info(ctxWithSpan, "hello trace")

	out := buf.String()
	if !strings.Contains(out, "hello trace") {
		t.Error("missing message")
	}
	if !strings.Contains(out, `"trace_id"`) {
		t.Error("missing trace_id field")
	}
	if !strings.Contains(out, `"span_id"`) {
		t.Error("missing span_id field")
	}
}

func TestTraceLogger_EmptyContext(t *testing.T) {
	var buf bytes.Buffer
	base := New(LevelDebug, &buf)
	tlog := NewTraceLogger(base)

	ctx := context.Background() // no span
	tlog.Info(ctx, "hello no trace")

	out := buf.String()
	if !strings.Contains(out, "hello no trace") {
		t.Error("missing message")
	}
	if strings.Contains(out, `"trace_id"`) {
		t.Error("should not have trace_id field")
	}
}
