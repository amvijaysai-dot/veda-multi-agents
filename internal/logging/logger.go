// Package logging provides structured logging capabilities for the runtime.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	// LevelTrace is the most detailed log level, typically disabled in production.
	LevelTrace Level = -1
	// LevelDebug is for diagnostic information, typically disabled in production.
	LevelDebug Level = 0
	// LevelInfo is for general information about system operation.
	LevelInfo Level = 1
	// LevelWarn is for warning about potentially harmful situations.
	LevelWarn Level = 2
	// LevelError is for error events that might still allow the application to continue.
	LevelError Level = 3
	// LevelFatal is for severe errors causing premature termination.
	LevelFatal Level = 4
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Field represents a key-value pair in a structured log entry.
type Field struct {
	Key   string
	Value interface{}
}

// Logger provides structured logging capabilities.
type Logger struct {
	mu     sync.Mutex
	level  Level
	writer io.Writer
	logger *log.Logger
	fields []Field
}

// New creates a new Logger with the given level and writer.
func New(level Level, writer io.Writer) *Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return &Logger{
		level:  level,
		writer: writer,
		logger: log.New(writer, "", 0),
	}
}

// NewDefault creates a new Logger with default settings (INFO level, stdout).
func NewDefault() *Logger {
	return New(LevelInfo, os.Stdout)
}

// WithFields returns a new Logger with the given fields added to the context.
func (l *Logger) WithFields(fields ...Field) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &Logger{
		level:  l.level,
		writer: l.writer,
		logger: l.logger,
		fields: newFields,
	}
}

// SetLevel sets the minimum log level for this logger.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Level returns the current minimum log level.
func (l *Logger) Level() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Trace logs a message at TRACE level.
func (l *Logger) Trace(msg string, fields ...Field) {
	l.log(LevelTrace, msg, fields...)
}

// Debug logs a message at DEBUG level.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

// Info logs a message at INFO level.
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

// Warn logs a message at WARN level.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

// Error logs a message at ERROR level.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

// Fatal logs a message at FATAL level and then calls os.Exit(1).
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(LevelFatal, msg, fields...)
	os.Exit(1)
}

// log writes a structured log entry if the level is enabled.
func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	// Build structured log entry
	entry := make(map[string]interface{})
	entry["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = level.String()
	entry["message"] = msg

	// Add logger fields
	for _, f := range l.fields {
		entry[f.Key] = f.Value
	}

	// Add message fields
	for _, f := range fields {
		entry[f.Key] = f.Value
	}

	// Format as JSON-like string
	line := formatEntry(entry)
	l.logger.Println(line)
}

// formatEntry formats a log entry as a JSON-like string.
func formatEntry(entry map[string]interface{}) string {
	line := "{"
	first := true
	for k, v := range entry {
		if !first {
			line += ", "
		}
		first = false
		line += fmt.Sprintf("%q: %v", k, formatValue(v))
	}
	line += "}"
	return line
}

// formatValue formats a value for inclusion in a log entry.
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case error:
		return fmt.Sprintf("%q", val.Error())
	default:
		return fmt.Sprintf("%v", val)
	}
}
