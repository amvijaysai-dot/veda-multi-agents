// Package config provides tests for the environment-based configuration loader.
package config

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// DefaultConfig tests
// ---------------------------------------------------------------------------

func TestDefaultConfig_FieldsAreNonZero(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.RuntimeID == "" {
		t.Error("expected non-empty RuntimeID in default config")
	}
	if cfg.LogLevel == "" {
		t.Error("expected non-empty LogLevel in default config")
	}
	if cfg.KernelShutdownTimeoutSec <= 0 {
		t.Errorf("expected KernelShutdownTimeoutSec > 0, got %d", cfg.KernelShutdownTimeoutSec)
	}
	if cfg.KernelSuspendTimeoutSec <= 0 {
		t.Errorf("expected KernelSuspendTimeoutSec > 0, got %d", cfg.KernelSuspendTimeoutSec)
	}
	if cfg.EventBusBufferSize <= 0 {
		t.Errorf("expected EventBusBufferSize > 0, got %d", cfg.EventBusBufferSize)
	}
	if cfg.MaxSubsystems <= 0 {
		t.Errorf("expected MaxSubsystems > 0, got %d", cfg.MaxSubsystems)
	}
	if cfg.HealthCheckIntervalSec <= 0 {
		t.Errorf("expected HealthCheckIntervalSec > 0, got %d", cfg.HealthCheckIntervalSec)
	}
}

func TestDefaultConfig_Validates(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("DefaultConfig().Validate() returned unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Config.Validate tests
// ---------------------------------------------------------------------------

func TestValidate_EmptyRuntimeIDFails(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RuntimeID = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty RuntimeID, got nil")
	}
}

func TestValidate_InvalidLogLevelFails(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LogLevel = "verbose"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid LogLevel, got nil")
	}
}

func TestValidate_ValidLogLevels(t *testing.T) {
	validLevels := []string{"trace", "debug", "info", "warn", "error"}
	for _, lvl := range validLevels {
		cfg := DefaultConfig()
		cfg.LogLevel = lvl
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate() returned error for valid LogLevel %q: %v", lvl, err)
		}
	}
}

func TestValidate_ZeroShutdownTimeoutFails(t *testing.T) {
	cfg := DefaultConfig()
	cfg.KernelShutdownTimeoutSec = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero KernelShutdownTimeoutSec, got nil")
	}
}

func TestValidate_NegativeEventBusFails(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EventBusBufferSize = -1
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative EventBusBufferSize, got nil")
	}
}

func TestValidate_ZeroMaxSubsystemsFails(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxSubsystems = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero MaxSubsystems, got nil")
	}
}

// ---------------------------------------------------------------------------
// Duration accessor tests
// ---------------------------------------------------------------------------

func TestKernelShutdownTimeout_ReturnsDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.KernelShutdownTimeoutSec = 30
	got := cfg.KernelShutdownTimeout()
	if got != 30*time.Second {
		t.Errorf("KernelShutdownTimeout() = %v, want %v", got, 30*time.Second)
	}
}

func TestKernelSuspendTimeout_ReturnsDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.KernelSuspendTimeoutSec = 10
	got := cfg.KernelSuspendTimeout()
	if got != 10*time.Second {
		t.Errorf("KernelSuspendTimeout() = %v, want %v", got, 10*time.Second)
	}
}

func TestHealthCheckInterval_ReturnsDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HealthCheckIntervalSec = 15
	got := cfg.HealthCheckInterval()
	if got != 15*time.Second {
		t.Errorf("HealthCheckInterval() = %v, want %v", got, 15*time.Second)
	}
}

// ---------------------------------------------------------------------------
// Load from environment tests
// ---------------------------------------------------------------------------

func TestLoad_DefaultsWhenNoEnvVars(t *testing.T) {
	// Ensure no VEDA_ env vars are set in the test environment
	// (rely on DefaultConfig values).
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.RuntimeID == "" {
		t.Error("expected non-empty RuntimeID from Load()")
	}
}

func TestLoad_ReadsRuntimeIDFromEnv(t *testing.T) {
	t.Setenv("VEDA_RUNTIME_ID", "test-runtime-42")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.RuntimeID != "test-runtime-42" {
		t.Errorf("expected RuntimeID %q, got %q", "test-runtime-42", cfg.RuntimeID)
	}
}

func TestLoad_ReadsLogLevelFromEnv(t *testing.T) {
	t.Setenv("VEDA_LOG_LEVEL", "DEBUG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel %q (lowercased), got %q", "debug", cfg.LogLevel)
	}
}

func TestLoad_ReadsIntFieldFromEnv(t *testing.T) {
	t.Setenv("VEDA_KERNEL_SHUTDOWN_TIMEOUT_SEC", "60")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.KernelShutdownTimeoutSec != 60 {
		t.Errorf("expected KernelShutdownTimeoutSec %d, got %d", 60, cfg.KernelShutdownTimeoutSec)
	}
}

func TestLoad_InvalidIntEnvVarReturnsError(t *testing.T) {
	t.Setenv("VEDA_KERNEL_SHUTDOWN_TIMEOUT_SEC", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for non-numeric int env var, got nil")
	}
}

func TestLoad_ReadsEventBusBufferSizeFromEnv(t *testing.T) {
	t.Setenv("VEDA_EVENT_BUS_BUFFER_SIZE", "2048")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.EventBusBufferSize != 2048 {
		t.Errorf("expected EventBusBufferSize %d, got %d", 2048, cfg.EventBusBufferSize)
	}
}

func TestLoad_ReadsMaxSubsystemsFromEnv(t *testing.T) {
	t.Setenv("VEDA_MAX_SUBSYSTEMS", "128")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.MaxSubsystems != 128 {
		t.Errorf("expected MaxSubsystems %d, got %d", 128, cfg.MaxSubsystems)
	}
}

func TestLoad_ReadsHealthCheckIntervalFromEnv(t *testing.T) {
	t.Setenv("VEDA_HEALTH_CHECK_INTERVAL_SEC", "30")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.HealthCheckIntervalSec != 30 {
		t.Errorf("expected HealthCheckIntervalSec %d, got %d", 30, cfg.HealthCheckIntervalSec)
	}
}

func TestLoadWithDefaults_NeverReturnsNil(t *testing.T) {
	// Set an invalid env var to trigger the error path in Load.
	t.Setenv("VEDA_KERNEL_SHUTDOWN_TIMEOUT_SEC", "invalid")

	cfg := LoadWithDefaults()
	if cfg == nil {
		t.Fatal("LoadWithDefaults() returned nil; must always return a valid config")
	}
}

func TestLoad_InvalidLogLevelFromEnvFails(t *testing.T) {
	t.Setenv("VEDA_LOG_LEVEL", "super-verbose")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid log level in env var, got nil")
	}
}

// ---------------------------------------------------------------------------
// intField helper tests
// ---------------------------------------------------------------------------

func TestIntField_EmptyStringReturnsDefault(t *testing.T) {
	got, err := intField("", 42)
	if err != nil {
		t.Fatalf("intField(\"\", 42) returned unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("expected default value 42, got %d", got)
	}
}

func TestIntField_ValidStringReturnsParsedValue(t *testing.T) {
	got, err := intField("100", 0)
	if err != nil {
		t.Fatalf("intField(\"100\", 0) returned unexpected error: %v", err)
	}
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestIntField_InvalidStringReturnsError(t *testing.T) {
	_, err := intField("abc", 0)
	if err == nil {
		t.Fatal("expected error for non-numeric string, got nil")
	}
}
