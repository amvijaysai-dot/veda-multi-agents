// Package binder provides capability binding logic.
package binder

import (
	"context"
	"strings"
	"testing"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

func TestContextualBinder_BindSuccess(t *testing.T) {
	b := NewContextualBinder()
	ctx := context.Background()

	meta := interfaces.CapabilityMetadata{
		ID:                  "cap.fs.read",
		RequiredPermissions: []string{"fs:read", "network:outbound"},
	}

	bindCtx := interfaces.BindingContext{
		AgentID:            "agent-1",
		AllowedPermissions: []string{"fs:read", "network:outbound", "extra:perm"},
	}

	exec, err := b.Bind(ctx, meta, bindCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exec == nil {
		t.Fatal("expected executable, got nil")
	}

	// Test execution
	res, err := exec.Execute(ctx, `{"path": "/"}`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !strings.Contains(res, "success") {
		t.Error("unexpected execution result")
	}

	// Test unbind
	err = b.Unbind(ctx, "agent-1", "cap.fs.read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Executing after unbind should fail
	_, err = exec.Execute(ctx, `{}`)
	if err == nil {
		t.Error("expected error executing unbound capability")
	}
}

func TestContextualBinder_PermissionDenied(t *testing.T) {
	b := NewContextualBinder()
	ctx := context.Background()

	meta := interfaces.CapabilityMetadata{
		ID:                  "cap.fs.write",
		RequiredPermissions: []string{"fs:write"},
	}

	bindCtx := interfaces.BindingContext{
		AgentID:            "agent-1",
		AllowedPermissions: []string{"fs:read"}, // missing fs:write
	}

	_, err := b.Bind(ctx, meta, bindCtx)
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
	if !strings.Contains(err.Error(), "denied") {
		t.Errorf("expected denial message, got: %v", err)
	}
}
