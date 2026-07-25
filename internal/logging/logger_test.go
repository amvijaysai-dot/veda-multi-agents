package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewDefaultLogger(t *testing.T) {
	l := NewDefault()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	if l.level != LevelInfo {
		t.Errorf("expected default level INFO, got %v", l.level)
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New(LevelDebug, &buf)

	l.Debug("debug message")
	if !strings.Contains(buf.String(), "DEBUG") {
		t.Error("expected DEBUG level in output")
	}
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("expected message in output")
	}

	buf.Reset()
	l.Trace("trace message")
	if buf.Len() > 0 {
		t.Error("expected no output for TRACE when level is DEBUG")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := New(LevelWarn, &buf)

	l.Info("should be filtered")
	if buf.Len() > 0 {
		t.Error("expected INFO to be filtered when level is WARN")
	}

	l.Warn("warning message")
	if !strings.Contains(buf.String(), "WARN") {
		t.Error("expected WARN level in output")
	}
}

func TestLoggerStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(LevelInfo, &buf)

	l.Info("test message", Field{"key1", "value1"}, Field{"key2", 42})

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("expected message in output")
	}
	if !strings.Contains(output, "key1") {
		t.Error("expected key1 in output")
	}
	if !strings.Contains(output, "value1") {
		t.Error("expected value1 in output")
	}
	if !strings.Contains(output, "key2") {
		t.Error("expected key2 in output")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(LevelInfo, &buf)
	l = l.WithFields(Field{"component", "test"}, Field{"module", "logging"})

	l.Info("contextual message")

	output := buf.String()
	if !strings.Contains(output, "component") {
		t.Error("expected component field in output")
	}
	if !strings.Contains(output, "module") {
		t.Error("expected module field in output")
	}
}

func TestLoggerLevelSetting(t *testing.T) {
	var buf bytes.Buffer
	l := New(LevelInfo, &buf)

	if l.Level() != LevelInfo {
		t.Errorf("expected LevelInfo, got %v", l.Level())
	}

	l.SetLevel(LevelDebug)
	if l.Level() != LevelDebug {
		t.Errorf("expected LevelDebug, got %v", l.Level())
	}
}

func TestLoggerAllLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New(LevelTrace, &buf)

	l.Trace("trace msg")
	if !strings.Contains(buf.String(), "TRACE") {
		t.Error("expected TRACE in output")
	}

	buf.Reset()
	l.Debug("debug msg")
	if !strings.Contains(buf.String(), "DEBUG") {
		t.Error("expected DEBUG in output")
	}

	buf.Reset()
	l.Info("info msg")
	if !strings.Contains(buf.String(), "INFO") {
		t.Error("expected INFO in output")
	}

	buf.Reset()
	l.Warn("warn msg")
	if !strings.Contains(buf.String(), "WARN") {
		t.Error("expected WARN in output")
	}

	buf.Reset()
	l.Error("error msg")
	if !strings.Contains(buf.String(), "ERROR") {
		t.Error("expected ERROR in output")
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelTrace, "TRACE"},
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLevelUnknown(t *testing.T) {
	var unknown Level = 99
	if got := unknown.String(); got != "UNKNOWN" {
		t.Errorf("expected 'UNKNOWN', got %q", got)
	}
}
