// Package logging provides structured logging capabilities for the runtime.
package logging

import (
	"context"

	"github.com/veda/agent-runtime/internal/tracing"
)

// TraceLogger wraps a standard Logger to automatically inject trace IDs
// from the context into the structured log fields.
type TraceLogger struct {
	base *Logger
}

// NewTraceLogger creates a new TraceLogger wrapping the provided base logger.
func NewTraceLogger(base *Logger) *TraceLogger {
	return &TraceLogger{
		base: base,
	}
}

// WithContext returns a new Logger that has trace fields injected if they exist in ctx.
func (t *TraceLogger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return t.base
	}

	span, ok := tracing.SpanFromContext(ctx)
	if !ok {
		return t.base
	}

	return t.base.WithFields(
		Field{Key: "trace_id", Value: span.TraceID()},
		Field{Key: "span_id", Value: span.SpanID()},
	)
}

// Trace logs a message at TRACE level, injecting trace fields from ctx.
func (t *TraceLogger) Trace(ctx context.Context, msg string, fields ...Field) {
	t.WithContext(ctx).Trace(msg, fields...)
}

// Debug logs a message at DEBUG level, injecting trace fields from ctx.
func (t *TraceLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	t.WithContext(ctx).Debug(msg, fields...)
}

// Info logs a message at INFO level, injecting trace fields from ctx.
func (t *TraceLogger) Info(ctx context.Context, msg string, fields ...Field) {
	t.WithContext(ctx).Info(msg, fields...)
}

// Warn logs a message at WARN level, injecting trace fields from ctx.
func (t *TraceLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	t.WithContext(ctx).Warn(msg, fields...)
}

// Error logs a message at ERROR level, injecting trace fields from ctx.
func (t *TraceLogger) Error(ctx context.Context, msg string, fields ...Field) {
	t.WithContext(ctx).Error(msg, fields...)
}

// Fatal logs a message at FATAL level and then exits, injecting trace fields from ctx.
func (t *TraceLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	t.WithContext(ctx).Fatal(msg, fields...)
}
