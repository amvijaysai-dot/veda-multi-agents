// Package config provides configuration loading and management for the VEDA Agent Runtime.
// Configuration is sourced from environment variables with support for typed access
// and default values. This package is designed to be used at startup before any
// subsystem is initialized.
package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration values loaded from the environment.
// Fields are exported to allow direct access where needed; the Load function
// is the primary entry point for populating a Config.
type Config struct {
	// Runtime settings
	RuntimeID string // Unique identifier for this runtime instance (VEDA_RUNTIME_ID)
	LogLevel  string // Minimum log level: trace, debug, info, warn, error (VEDA_LOG_LEVEL)

	// Kernel settings
	KernelShutdownTimeoutSec int // Graceful shutdown timeout in seconds (VEDA_KERNEL_SHUTDOWN_TIMEOUT_SEC)
	KernelSuspendTimeoutSec  int // Suspend quiesce timeout in seconds (VEDA_KERNEL_SUSPEND_TIMEOUT_SEC)

	// Event bus settings
	EventBusBufferSize int // Buffer size for async event delivery (VEDA_EVENT_BUS_BUFFER_SIZE)

	// Subsystem settings
	MaxSubsystems int // Maximum number of subsystems allowed (VEDA_MAX_SUBSYSTEMS)

	// Health check settings
	HealthCheckIntervalSec int // Interval between health checks in seconds (VEDA_HEALTH_CHECK_INTERVAL_SEC)
}

// DefaultConfig returns a Config populated with safe default values.
// These defaults are designed to be suitable for local development and testing.
func DefaultConfig() *Config {
	return &Config{
		RuntimeID:                "veda-runtime-default",
		LogLevel:                 "info",
		KernelShutdownTimeoutSec: 30,
		KernelSuspendTimeoutSec:  10,
		EventBusBufferSize:       1024,
		MaxSubsystems:            64,
		HealthCheckIntervalSec:   15,
	}
}

// Validate checks that the Config contains valid values.
// Returns a descriptive error if any field violates its constraints.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.RuntimeID) == "" {
		return fmt.Errorf("config: RuntimeID must not be empty")
	}

	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("config: LogLevel %q is not valid; must be one of: trace, debug, info, warn, error", c.LogLevel)
	}

	if c.KernelShutdownTimeoutSec <= 0 {
		return fmt.Errorf("config: KernelShutdownTimeoutSec must be > 0, got %d", c.KernelShutdownTimeoutSec)
	}

	if c.KernelSuspendTimeoutSec <= 0 {
		return fmt.Errorf("config: KernelSuspendTimeoutSec must be > 0, got %d", c.KernelSuspendTimeoutSec)
	}

	if c.EventBusBufferSize <= 0 {
		return fmt.Errorf("config: EventBusBufferSize must be > 0, got %d", c.EventBusBufferSize)
	}

	if c.MaxSubsystems <= 0 {
		return fmt.Errorf("config: MaxSubsystems must be > 0, got %d", c.MaxSubsystems)
	}

	if c.HealthCheckIntervalSec <= 0 {
		return fmt.Errorf("config: HealthCheckIntervalSec must be > 0, got %d", c.HealthCheckIntervalSec)
	}

	return nil
}

// KernelShutdownTimeout returns the shutdown timeout as a time.Duration.
func (c *Config) KernelShutdownTimeout() time.Duration {
	return time.Duration(c.KernelShutdownTimeoutSec) * time.Second
}

// KernelSuspendTimeout returns the suspend timeout as a time.Duration.
func (c *Config) KernelSuspendTimeout() time.Duration {
	return time.Duration(c.KernelSuspendTimeoutSec) * time.Second
}

// HealthCheckInterval returns the health check interval as a time.Duration.
func (c *Config) HealthCheckInterval() time.Duration {
	return time.Duration(c.HealthCheckIntervalSec) * time.Second
}

// intField is a helper used in env loading to parse integer fields.
func intField(raw string, defaultVal int) (int, error) {
	if raw == "" {
		return defaultVal, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("expected an integer, got %q: %w", raw, err)
	}
	return v, nil
}
