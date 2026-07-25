package validation

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/security"
)

// TestBoundaryValidation ensures inputs at the edge are strictly validated.
func TestBoundaryValidation(t *testing.T) {
	err := security.ValidateInput("../../../etc/passwd")
	if err == nil {
		t.Fatal("expected validation to fail for path traversal")
	}

	err = security.ValidateInput("normal-input-data")
	if err != nil {
		t.Fatalf("expected validation to succeed, got %v", err)
	}
}

// TestIsolationEffectiveness validates that the sandbox manager isolates capabilities.
func TestIsolationEffectiveness(t *testing.T) {
	sm := security.NewSandboxManager()

	sm.AttachPolicy("agent-a", security.Policy{
		AllowedCapabilities: []string{"read"},
		MaxNetworkRequests:  1,
	})
	sm.AttachPolicy("agent-b", security.Policy{
		AllowedCapabilities: []string{"write"},
		MaxNetworkRequests:  0, // no network
	})

	// Agent A testing isolation
	err := sm.CheckCapability("agent-a", "write")
	if err == nil {
		t.Fatal("expected agent-a to be denied 'write'")
	}
	err = sm.CheckNetworkRequest(context.Background(), "agent-a")
	if err != nil {
		t.Fatalf("expected agent-a to succeed on 1st network request: %v", err)
	}
	err = sm.CheckNetworkRequest(context.Background(), "agent-a")
	if err == nil {
		t.Fatal("expected agent-a to fail on 2nd network request (limit 1)")
	}

	// Agent B testing isolation
	err = sm.CheckCapability("agent-b", "read")
	if err == nil {
		t.Fatal("expected agent-b to be denied 'read'")
	}
}
