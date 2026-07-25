// Package validate provides validation logic for state checkpoints.
package validate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/veda/agent-runtime/internal/recovery/interfaces"
)

// StateValidator implements the Validator interface.
type StateValidator struct{}

// NewStateValidator creates a new StateValidator.
func NewStateValidator() *StateValidator {
	return &StateValidator{}
}

// Validate checks if the provided checkpoint data is valid and uncorrupted.
func (v *StateValidator) Validate(ctx context.Context, cp *interfaces.Checkpoint) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if cp == nil {
		return fmt.Errorf("checkpoint is nil")
	}

	if cp.ID == "" {
		return fmt.Errorf("checkpoint ID is empty")
	}
	if cp.AgentID == "" {
		return fmt.Errorf("checkpoint AgentID is empty")
	}
	if cp.Version == "" {
		return fmt.Errorf("checkpoint Version is empty")
	}
	if cp.Timestamp.IsZero() || cp.Timestamp.After(time.Now().UTC().Add(time.Hour)) {
		return fmt.Errorf("checkpoint Timestamp is invalid or too far in the future")
	}
	if len(cp.Data) == 0 {
		return fmt.Errorf("checkpoint Data is empty")
	}

	// Verify checksum if present in metadata
	if expectedHash, ok := cp.Metadata["sha256"]; ok {
		hash := sha256.Sum256(cp.Data)
		actualHash := hex.EncodeToString(hash[:])
		if actualHash != expectedHash {
			return fmt.Errorf("checkpoint checksum mismatch: expected %s, got %s", expectedHash, actualHash)
		}
	}

	return nil
}

// AttachChecksum computes the SHA-256 hash of the data and attaches it to Metadata.
// This is typically called by the Checkpointer before saving.
func AttachChecksum(cp *interfaces.Checkpoint) {
	if cp.Metadata == nil {
		cp.Metadata = make(map[string]string)
	}
	hash := sha256.Sum256(cp.Data)
	cp.Metadata["sha256"] = hex.EncodeToString(hash[:])
}
