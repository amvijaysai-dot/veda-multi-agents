// Package restore provides implementations for loading and restoring agent state checkpoints.
package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/veda/agent-runtime/internal/recovery/interfaces"
)

// FileRestorer implements the Restorer interface using the local filesystem.
type FileRestorer struct {
	baseDir string
}

// NewFileRestorer creates a new FileRestorer.
func NewFileRestorer(baseDir string) *FileRestorer {
	return &FileRestorer{
		baseDir: baseDir,
	}
}

// Load retrieves a specific checkpoint by its ID.
func (r *FileRestorer) Load(ctx context.Context, checkpointID string) (*interfaces.Checkpoint, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentDir := filepath.Join(r.baseDir, entry.Name())
		files, err := os.ReadDir(agentDir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// Filename format: checkpoint_<timestamp>_<id>.json
			name := file.Name()
			if len(name) > 41 && name[len(name)-41:len(name)-5] == checkpointID {
				return r.readFile(filepath.Join(agentDir, name))
			}
		}
	}

	return nil, fmt.Errorf("checkpoint %s not found", checkpointID)
}

// LoadLatest retrieves the most recent checkpoint for an agent.
func (r *FileRestorer) LoadLatest(ctx context.Context, agentID string) (*interfaces.Checkpoint, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	checkpoints, err := r.List(ctx, agentID)
	if err != nil {
		return nil, err
	}

	if len(checkpoints) == 0 {
		return nil, fmt.Errorf("no checkpoints found for agent %s", agentID)
	}

	return checkpoints[len(checkpoints)-1], nil
}

// List returns all available checkpoints for an agent, ordered by time.
func (r *FileRestorer) List(ctx context.Context, agentID string) ([]*interfaces.Checkpoint, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	agentDir := filepath.Join(r.baseDir, agentID)
	files, err := os.ReadDir(agentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*interfaces.Checkpoint{}, nil
		}
		return nil, fmt.Errorf("failed to read agent directory: %w", err)
	}

	var fileNames []string
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "checkpoint_") {
			continue
		}
		fileNames = append(fileNames, file.Name())
	}

	// Sort files by name which sorts by timestamp due to timestamp in filename
	sort.Strings(fileNames)

	var checkpoints []*interfaces.Checkpoint
	for _, name := range fileNames {
		path := filepath.Join(agentDir, name)
		cp, err := r.readFile(path)
		if err != nil {
			// Skip corrupted files or log them
			continue
		}
		checkpoints = append(checkpoints, cp)
	}

	return checkpoints, nil
}

func (r *FileRestorer) readFile(path string) (*interfaces.Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var cp interfaces.Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &cp, nil
}
