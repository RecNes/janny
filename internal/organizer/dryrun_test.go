package organizer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/evrenesat/janny/internal/config"
)

func TestDryRun_LearnAndClean(t *testing.T) {
	// Setup temporary directory locally to avoid permission issues
	if err := os.MkdirAll("temp", 0755); err != nil {
		t.Fatal(err)
	}
	tmpDir, err := os.MkdirTemp("temp", "janny-dryrun-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directories for Learn/Clean
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.Mkdir(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a dummy config file
	configFile := filepath.Join(tmpDir, "config.toml")

	// Create a file for "Learn" to find (unknown extension .xyz)
	if err := os.WriteFile(filepath.Join(docsDir, "test.xyz"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file for "AutoClean" to find (old file)
	oldFile := filepath.Join(docsDir, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}
	// Set mtime to 31 days ago
	oldTime := time.Now().AddDate(0, 0, -31)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Setup Config object
	cfg := &config.Config{
		Storage: map[string]string{
			"docs": docsDir,
		},
		ExtensionMap: map[string]string{
			"txt": "docs",
		},
		Rules: map[string]string{
			"docs": "txt",
		},
		AutoClean: map[string]int{
			"docs": 30, // Clean files older than 30 days
		},
	}
	// Write initial config to disk so Learn has something to read/update
	if err := config.SaveConfig(configFile, cfg); err != nil {
		t.Fatal(err)
	}

	// Initialize Organizer with DryRun = true
	org := New(cfg, nil, true, true) // dryRun=true, verbose=true

	// Test 1: Learn with DryRun
	// Should NOT update config file with .xyz
	if err := org.Learn(configFile); err != nil {
		t.Fatalf("Learn failed: %v", err)
	}

	// Reload config to verify it hasn't changed
	loadedCfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}
	if _, ok := loadedCfg.ExtensionMap["xyz"]; ok {
		t.Error("DryRun Learn FAILED: Config was updated with 'xyz' extension")
	}

	// Test 2: Clean with DryRun
	// Should NOT delete old.txt
	if err := org.Clean(context.Background()); err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	if _, err := os.Stat(oldFile); os.IsNotExist(err) {
		t.Error("DryRun Clean FAILED: Old file was deleted")
	}
}
