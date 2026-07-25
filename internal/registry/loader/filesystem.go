// Package loader provides capability loaders from various sources.
package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

// FileSystemLoader implements interfaces.CapabilityLoader for local files.
type FileSystemLoader struct{}

// NewFileSystemLoader creates a new FileSystemLoader.
func NewFileSystemLoader() *FileSystemLoader {
	return &FileSystemLoader{}
}

// Load reads and parses a capability from a JSON file.
func (l *FileSystemLoader) Load(ctx context.Context, source interfaces.CapabilitySource) (interfaces.LoadedCapability, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.LoadedCapability{}, err
	}

	if source.Type != "filesystem" {
		return interfaces.LoadedCapability{}, fmt.Errorf("unsupported source type %q", source.Type)
	}

	path := strings.TrimPrefix(source.URI, "file://")
	data, err := os.ReadFile(path)
	if err != nil {
		return interfaces.LoadedCapability{}, fmt.Errorf("failed to read capability file: %w", err)
	}

	var meta interfaces.CapabilityMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return interfaces.LoadedCapability{}, fmt.Errorf("failed to parse capability json: %w", err)
	}

	if meta.ID == "" || meta.Version == "" {
		return interfaces.LoadedCapability{}, fmt.Errorf("capability must have ID and Version")
	}

	return interfaces.LoadedCapability{
		Metadata:       meta,
		Source:         source,
		ExecutablePath: path, // For v0.7, the executable path is the manifest itself.
	}, nil
}

// Discover scans a directory for .json capability manifests.
func (l *FileSystemLoader) Discover(ctx context.Context, rootSource interfaces.CapabilitySource) ([]interfaces.CapabilitySource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if rootSource.Type != "filesystem" {
		return nil, fmt.Errorf("unsupported root source type %q", rootSource.Type)
	}

	rootPath := strings.TrimPrefix(rootSource.URI, "file://")
	var sources []interfaces.CapabilitySource

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			// Basic heuristic: check if it parses as a CapabilityMetadata
			data, parseErr := os.ReadFile(path)
			if parseErr != nil {
				return nil // skip unreadable
			}
			var meta interfaces.CapabilityMetadata
			if json.Unmarshal(data, &meta) == nil && meta.ID != "" {
				// URI should ideally be formatted cleanly for cross-platform,
				// but here we maintain simple file:// path
				sources = append(sources, interfaces.CapabilitySource{
					Type: "filesystem",
					URI:  "file://" + path,
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return sources, nil
}

var _ interfaces.CapabilityLoader = (*FileSystemLoader)(nil)
