package security

import (
	"context"
	"testing"
)

func TestSandboxManager(t *testing.T) {
	sm := NewSandboxManager()

	policy := Policy{
		AllowedCapabilities: []string{"search", "read-file"},
		MaxNetworkRequests:  2,
	}

	agentID := "agent-123"
	sm.AttachPolicy(agentID, policy)

	// Test CheckCapability
	err := sm.CheckCapability(agentID, "search")
	if err != nil {
		t.Fatalf("expected nil error for allowed capability, got %v", err)
	}

	err = sm.CheckCapability(agentID, "write-file")
	if err == nil {
		t.Fatalf("expected error for denied capability")
	}

	err = sm.CheckCapability("unknown-agent", "search")
	if err == nil {
		t.Fatalf("expected error for unknown agent")
	}

	// Test CheckNetworkRequest
	ctx := context.Background()
	err = sm.CheckNetworkRequest(ctx, agentID)
	if err != nil {
		t.Fatalf("expected success for 1st request, got %v", err)
	}

	err = sm.CheckNetworkRequest(ctx, agentID)
	if err != nil {
		t.Fatalf("expected success for 2nd request, got %v", err)
	}

	err = sm.CheckNetworkRequest(ctx, agentID)
	if err == nil || err.Error() != "sandbox breach: network request limit exceeded for agent agent-123" {
		t.Fatalf("expected sandbox breach error, got %v", err)
	}
}

func TestValidateInput(t *testing.T) {
	cases := []struct {
		input string
		err   bool
	}{
		{"safe/path/file.txt", false},
		{"unsafe/../path", true},
		{"unsafe\\..\\path", true},
		{"..\\windows\\system32", true},
		{"normal text with no traversal", false},
		{"text with null \x00 byte", true},
	}

	for _, c := range cases {
		err := ValidateInput(c.input)
		if (err != nil) != c.err {
			t.Errorf("expected err=%v for input %q, got err=%v", c.err, c.input, err)
		}
	}
}
