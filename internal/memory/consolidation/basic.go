// Package consolidation provides mechanisms to move data from short-term
// memory to long-term memory.
package consolidation

import (
	"context"
	"fmt"
	"time"

	"github.com/veda/agent-runtime/internal/memory/interfaces"
)

// CandidateProvider defines the method required to extract data hinted for
// persistence from a short-term memory store.
type CandidateProvider interface {
	GetConsolidationCandidates(agentID, sessionID string) map[string]string
}

// BasicConsolidator implements interfaces.ConsolidationManager.
// It retrieves hinted items from short-term memory, scrubs them for PII,
// and saves them to long-term memory.
type BasicConsolidator struct {
	shortTerm CandidateProvider
	longTerm  interfaces.LongTermMemory
	privacy   interfaces.PrivacyManager

	// retentionTTL is the time-to-live applied to consolidated items in long-term storage.
	// If zero, items are stored indefinitely.
	retentionTTL time.Duration
}

// NewBasicConsolidator creates a new BasicConsolidator.
func NewBasicConsolidator(
	shortTerm CandidateProvider,
	longTerm interfaces.LongTermMemory,
	privacy interfaces.PrivacyManager,
	retentionTTL time.Duration,
) *BasicConsolidator {
	if shortTerm == nil {
		panic("consolidation: shortTerm provider cannot be nil")
	}
	if longTerm == nil {
		panic("consolidation: longTerm memory cannot be nil")
	}
	if privacy == nil {
		panic("consolidation: privacy manager cannot be nil")
	}
	return &BasicConsolidator{
		shortTerm:    shortTerm,
		longTerm:     longTerm,
		privacy:      privacy,
		retentionTTL: retentionTTL,
	}
}

// Consolidate processes all hinted short-term memories for the given session.
func (c *BasicConsolidator) Consolidate(ctx context.Context, agentID, sessionID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if agentID == "" || sessionID == "" {
		return fmt.Errorf("agentID and sessionID must not be empty")
	}

	candidates := c.shortTerm.GetConsolidationCandidates(agentID, sessionID)
	if len(candidates) == 0 {
		return nil // nothing to consolidate
	}

	for key, value := range candidates {
		// Scrub PII before storing
		scrubbed, err := c.privacy.Scrub(ctx, value)
		if err != nil {
			// Skip item if scrubbing fails, do not halt consolidation
			continue
		}

		err = c.longTerm.Store(ctx, agentID, key, scrubbed, c.retentionTTL)
		if err != nil {
			return fmt.Errorf("failed to store %q in long-term memory: %w", key, err)
		}
	}

	return nil
}

var _ interfaces.ConsolidationManager = (*BasicConsolidator)(nil)
