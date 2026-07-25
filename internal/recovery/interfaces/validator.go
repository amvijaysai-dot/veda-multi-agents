// Package interfaces defines the contracts for the recovery subsystem.
package interfaces

import (
	"context"
)

// Validator ensures the integrity and consistency of a Checkpoint.
type Validator interface {
	// Validate checks if the provided checkpoint data is valid and uncorrupted.
	Validate(ctx context.Context, cp *Checkpoint) error
}
