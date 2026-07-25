// Package checkpoint provides implementations for creating and storing agent state checkpoints.
package checkpoint

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/veda/agent-runtime/internal/recovery/interfaces"
)

// generateUUID creates a random v4 UUID string.
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10

	buf := make([]byte, 36)
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])

	return string(buf)
}

// FileCheckpointer implements the Checkpointer interface using the local filesystem.
type FileCheckpointer struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFileCheckpointer creates a new FileCheckpointer.
func NewFileCheckpointer(baseDir string) (*FileCheckpointer, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &FileCheckpointer{
		baseDir: baseDir,
	}, nil
}

// Save captures the current state for an agent and stores it.
func (c *FileCheckpointer) Save(ctx context.Context, agentID string, state []byte, version string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	checkpointID := generateUUID()
	timestamp := time.Now().UTC()

	cp := &interfaces.Checkpoint{
		ID:        checkpointID,
		Timestamp: timestamp,
		AgentID:   agentID,
		Data:      state,
		Version:   version,
	}

	// Compress the state to save space
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(state); err != nil {
		return "", fmt.Errorf("failed to compress state: %w", err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip writer: %w", err)
	}

	cp.Data = buf.Bytes() // Store the compressed data

	encoded, err := json.Marshal(cp)
	if err != nil {
		return "", fmt.Errorf("failed to serialize checkpoint metadata: %w", err)
	}

	agentDir := filepath.Join(c.baseDir, agentID)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create agent directory: %w", err)
	}

	// File format: checkpoint_<timestamp>_<id>.json
	filename := fmt.Sprintf("checkpoint_%d_%s.json", timestamp.UnixNano(), checkpointID)
	path := filepath.Join(agentDir, filename)

	if err := os.WriteFile(path, encoded, 0600); err != nil {
		return "", fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return checkpointID, nil
}

// Delete removes a specific checkpoint by ID across all agent directories.
func (c *FileCheckpointer) Delete(ctx context.Context, checkpointID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// This is a naive implementation that searches all agent directories.
	// A more optimized implementation would index checkpoints.
	entries, err := os.ReadDir(c.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to delete
		}
		return fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentDir := filepath.Join(c.baseDir, entry.Name())
		files, err := os.ReadDir(agentDir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			// Search for the ID in the filename
			// Filename format: checkpoint_<timestamp>_<id>.json
			// The ID is 36 chars long (uuid) + 5 chars (.json) = 41 chars from end
			name := file.Name()
			if len(name) > 41 && name[len(name)-41:len(name)-5] == checkpointID {
				if err := os.Remove(filepath.Join(agentDir, name)); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove checkpoint file: %w", err)
				}
				return nil // Found and deleted
			}
		}
	}

	// Checkpoint not found, but we consider this a success (idempotent)
	return nil
}

// DeleteForAgent removes all checkpoints for a specific agent.
func (c *FileCheckpointer) DeleteForAgent(ctx context.Context, agentID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	agentDir := filepath.Join(c.baseDir, agentID)
	if err := os.RemoveAll(agentDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove agent directory: %w", err)
	}

	return nil
}

// DecompressState is a utility function to decompress checkpoint data.
func DecompressState(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer zr.Close()
	return io.ReadAll(zr)
}
