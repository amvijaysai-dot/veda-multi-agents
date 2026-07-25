// Package loader provides capability loaders.
package loader

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

func TestFileSystemLoader_DiscoverAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	validJSON := `{
		"id": "test.cap",
		"version": "1.0.0",
		"name": "Test Cap",
		"description": "A test capability"
	}`

	invalidJSON := `{"id": "broken, no version"}`
	notJSON := `some plain text`

	err := os.WriteFile(filepath.Join(tmpDir, "valid.json"), []byte(validJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "text.txt"), []byte(notJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ldr := NewFileSystemLoader()
	ctx := context.Background()

	root := interfaces.CapabilitySource{
		Type: "filesystem",
		URI:  "file://" + tmpDir,
	}

	sources, err := ldr.Discover(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only find the valid.json, because invalid.json doesn't parse to a CapabilityMetadata
	// properly (it misses version but wait, json.Unmarshal won't fail if fields are missing, but
	// my Discover method checks `if json.Unmarshal() == nil && meta.ID != ""`
	// Let's see what Discover finds: valid.json has ID="test.cap", invalid.json has ID="broken, no version".
	// Both will be "discovered" by the basic check!
	if len(sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(sources))
	}

	// Now try to load them
	var validLoaded bool
	for _, src := range sources {
		cap, err := ldr.Load(ctx, src)
		if err == nil {
			if cap.Metadata.ID == "test.cap" {
				validLoaded = true
			}
		} else {
			// invalid.json should fail here because it lacks Version, and Load enforces it
			if err.Error() != "capability must have ID and Version" {
				t.Errorf("unexpected error message: %v", err)
			}
		}
	}

	if !validLoaded {
		t.Error("failed to load valid capability")
	}
}

func TestFileSystemLoader_UnsupportedType(t *testing.T) {
	ldr := NewFileSystemLoader()
	ctx := context.Background()

	_, err := ldr.Discover(ctx, interfaces.CapabilitySource{Type: "remote"})
	if err == nil {
		t.Error("expected error for unsupported type")
	}

	_, err = ldr.Load(ctx, interfaces.CapabilitySource{Type: "remote"})
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}
