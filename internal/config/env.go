// Package config provides configuration loading and management for the VEDA Agent Runtime.
package config

import (
	"fmt"
	"os"
	"strings"
)

// envVarPrefix is the prefix applied to all environment variable names for this runtime.
const envVarPrefix = "VEDA_"

// Load reads configuration from environment variables and returns a populated Config.
//
// Environment variable names follow the VEDA_<FIELD> convention, where each field name
// is mapped to its uppercase environment variable counterpart. Missing variables fall
// back to the default values defined in DefaultConfig.
//
// Returns an error if any environment variable is present but cannot be parsed
// into the expected type (e.g., a non-numeric value for an integer field).
//
// Variable mapping:
//
//	VEDA_RUNTIME_ID                  → Config.RuntimeID
//	VEDA_LOG_LEVEL                   → Config.LogLevel
//	VEDA_KERNEL_SHUTDOWN_TIMEOUT_SEC → Config.KernelShutdownTimeoutSec
//	VEDA_KERNEL_SUSPEND_TIMEOUT_SEC  → Config.KernelSuspendTimeoutSec
//	VEDA_EVENT_BUS_BUFFER_SIZE       → Config.EventBusBufferSize
//	VEDA_MAX_SUBSYSTEMS              → Config.MaxSubsystems
//	VEDA_HEALTH_CHECK_INTERVAL_SEC   → Config.HealthCheckIntervalSec
func Load() (*Config, error) {
	cfg := DefaultConfig()

	if v := getenv("RUNTIME_ID"); v != "" {
		cfg.RuntimeID = v
	}

	if v := getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = strings.ToLower(strings.TrimSpace(v))
	}

	if v := getenv("KERNEL_SHUTDOWN_TIMEOUT_SEC"); v != "" {
		n, err := intField(v, cfg.KernelShutdownTimeoutSec)
		if err != nil {
			return nil, fmt.Errorf("VEDA_KERNEL_SHUTDOWN_TIMEOUT_SEC: %w", err)
		}
		cfg.KernelShutdownTimeoutSec = n
	}

	if v := getenv("KERNEL_SUSPEND_TIMEOUT_SEC"); v != "" {
		n, err := intField(v, cfg.KernelSuspendTimeoutSec)
		if err != nil {
			return nil, fmt.Errorf("VEDA_KERNEL_SUSPEND_TIMEOUT_SEC: %w", err)
		}
		cfg.KernelSuspendTimeoutSec = n
	}

	if v := getenv("EVENT_BUS_BUFFER_SIZE"); v != "" {
		n, err := intField(v, cfg.EventBusBufferSize)
		if err != nil {
			return nil, fmt.Errorf("VEDA_EVENT_BUS_BUFFER_SIZE: %w", err)
		}
		cfg.EventBusBufferSize = n
	}

	if v := getenv("MAX_SUBSYSTEMS"); v != "" {
		n, err := intField(v, cfg.MaxSubsystems)
		if err != nil {
			return nil, fmt.Errorf("VEDA_MAX_SUBSYSTEMS: %w", err)
		}
		cfg.MaxSubsystems = n
	}

	if v := getenv("HEALTH_CHECK_INTERVAL_SEC"); v != "" {
		n, err := intField(v, cfg.HealthCheckIntervalSec)
		if err != nil {
			return nil, fmt.Errorf("VEDA_HEALTH_CHECK_INTERVAL_SEC: %w", err)
		}
		cfg.HealthCheckIntervalSec = n
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// LoadWithDefaults is a convenience wrapper that always returns a valid Config,
// using environment variables where present and defaults elsewhere.
// If any environment variable is malformed, defaults are used for that field
// rather than returning an error.
func LoadWithDefaults() *Config {
	cfg, err := Load()
	if err != nil {
		// On parse failures, return pure defaults rather than a partially-loaded config.
		return DefaultConfig()
	}
	return cfg
}

// getenv retrieves the value of the environment variable with the VEDA_ prefix
// prepended to the given suffix. Returns an empty string if not set.
func getenv(suffix string) string {
	return os.Getenv(envVarPrefix + suffix)
}

// LookupEnv is exported for testing; it looks up a VEDA_-prefixed variable and
// reports whether it was set.
func LookupEnv(suffix string) (string, bool) {
	return os.LookupEnv(envVarPrefix + suffix)
}
