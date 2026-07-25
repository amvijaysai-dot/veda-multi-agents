// Package hotreload provides runtime reloading capabilities for configuration and capabilities.
package hotreload

import (
	"fmt"
	"sync"
	"time"

	"github.com/veda/agent-runtime/internal/config"
	"github.com/veda/agent-runtime/internal/kernel/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
)

// ConfigReloader manages a thread-safe reference to the current configuration
// and provides mechanisms to reload it at runtime.
type ConfigReloader struct {
	mu     sync.RWMutex
	cfg    *config.Config
	kernel interfaces.Kernel
}

// NewConfigReloader initializes a ConfigReloader with a baseline configuration.
func NewConfigReloader(initial *config.Config, kernel interfaces.Kernel) *ConfigReloader {
	if initial == nil {
		initial = config.DefaultConfig()
	}
	return &ConfigReloader{
		cfg:    initial,
		kernel: kernel,
	}
}

// GetConfig returns the currently active configuration in a thread-safe manner.
func (cr *ConfigReloader) GetConfig() *config.Config {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.cfg
}

// Reload attempts to load a new configuration, validates it, and applies it
// if successful. On failure, the old configuration is retained (rollback).
func (cr *ConfigReloader) Reload() error {
	newCfg, err := config.Load()
	if err != nil {
		// Rollback implicitly by not applying newCfg
		return fmt.Errorf("failed to load new configuration: %w", err)
	}

	cr.mu.Lock()
	cr.cfg = newCfg
	cr.mu.Unlock()

	// Notify affected components if kernel is attached
	if cr.kernel != nil {
		evt := event.NewBaseEvent(
			fmt.Sprintf("cfg-reload-%d", time.Now().UnixNano()),
			event.Type("config.reloaded"), // Extend event types conceptually
			"ConfigReloader",
		)
		_ = cr.kernel.PublishEvent(evt)
	}

	return nil
}
