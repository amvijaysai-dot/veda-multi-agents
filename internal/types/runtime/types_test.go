package runtime

import "testing"

func TestRuntimeStatusString(t *testing.T) {
	tests := []struct {
		status   RuntimeStatus
		expected string
	}{
		{StatusUninitialized, "uninitialized"},
		{StatusInitializing, "initializing"},
		{StatusReady, "ready"},
		{StatusBusy, "busy"},
		{StatusSuspending, "suspending"},
		{StatusSuspended, "suspended"},
		{StatusResuming, "resuming"},
		{StatusShuttingDown, "shutting_down"},
		{StatusTerminated, "terminated"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRuntimeStatusUnknown(t *testing.T) {
	var unknown RuntimeStatus = 99
	if got := unknown.String(); got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}

func TestAgentStateString(t *testing.T) {
	tests := []struct {
		state    AgentState
		expected string
	}{
		{AgentNonExistent, "non_existent"},
		{AgentCreating, "creating"},
		{AgentInitializing, "initializing"},
		{AgentReady, "ready"},
		{AgentBusy, "busy"},
		{AgentSuspending, "suspending"},
		{AgentSuspended, "suspended"},
		{AgentResuming, "resuming"},
		{AgentCheckpointing, "checkpointing"},
		{AgentRecovering, "recovering"},
		{AgentStopping, "stopping"},
		{AgentTerminated, "terminated"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestAgentStateUnknown(t *testing.T) {
	var unknown AgentState = 99
	if got := unknown.String(); got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}

func TestVersionStruct(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("expected Version{1, 2, 3}, got {%d, %d, %d}", v.Major, v.Minor, v.Patch)
	}
}

func TestRuntimeStatusOrder(t *testing.T) {
	// Verify the enum order matches expected lifecycle
	if StatusUninitialized >= StatusInitializing {
		t.Error("StatusUninitialized should be before StatusInitializing")
	}
	if StatusInitializing >= StatusReady {
		t.Error("StatusInitializing should be before StatusReady")
	}
	if StatusReady >= StatusBusy {
		t.Error("StatusReady should be before StatusBusy")
	}
	if StatusShuttingDown >= StatusTerminated {
		t.Error("StatusShuttingDown should be before StatusTerminated")
	}
}
