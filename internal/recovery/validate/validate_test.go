package validate

import (
	"context"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/recovery/interfaces"
)

func TestStateValidator_Validate(t *testing.T) {
	validator := NewStateValidator()
	ctx := context.Background()

	validCP := &interfaces.Checkpoint{
		ID:        "cp-1",
		AgentID:   "agent-1",
		Version:   "v1",
		Timestamp: time.Now().UTC(),
		Data:      []byte("some-data"),
	}
	AttachChecksum(validCP)

	if err := validator.Validate(ctx, validCP); err != nil {
		t.Errorf("expected valid checkpoint to pass, got error: %v", err)
	}

	tests := []struct {
		name string
		cp   *interfaces.Checkpoint
	}{
		{"nil checkpoint", nil},
		{"empty ID", &interfaces.Checkpoint{AgentID: "a", Version: "v1", Timestamp: time.Now(), Data: []byte("d")}},
		{"empty AgentID", &interfaces.Checkpoint{ID: "i", Version: "v1", Timestamp: time.Now(), Data: []byte("d")}},
		{"empty Version", &interfaces.Checkpoint{ID: "i", AgentID: "a", Timestamp: time.Now(), Data: []byte("d")}},
		{"empty Data", &interfaces.Checkpoint{ID: "i", AgentID: "a", Version: "v1", Timestamp: time.Now()}},
		{"zero Timestamp", &interfaces.Checkpoint{ID: "i", AgentID: "a", Version: "v1", Data: []byte("d")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validator.Validate(ctx, tt.cp); err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}

	// Test corrupted data
	corruptedCP := &interfaces.Checkpoint{
		ID:        "cp-2",
		AgentID:   "agent-1",
		Version:   "v1",
		Timestamp: time.Now().UTC(),
		Data:      []byte("good-data"),
	}
	AttachChecksum(corruptedCP)

	// Mutate data
	corruptedCP.Data = []byte("bad-data")
	if err := validator.Validate(ctx, corruptedCP); err == nil {
		t.Error("expected error for corrupted checksum, got nil")
	}
}
