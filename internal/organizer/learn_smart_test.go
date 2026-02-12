package organizer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/evrenesat/janny/internal/config"
	"github.com/evrenesat/janny/internal/external"
)

func TestOrganizer_LearnSmart(t *testing.T) {
	// Setup temporary directory locally to avoid permission issues
	localTemp := "../../temp"
	if _, err := os.Stat(localTemp); os.IsNotExist(err) {
		os.MkdirAll(localTemp, 0755)
	}

	tempDir, err := os.MkdirTemp(localTemp, "janny_smart_learn_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directory structure
	sourceDir := filepath.Join(tempDir, "source")
	storageDir := filepath.Join(tempDir, "storage")
	configPath := filepath.Join(tempDir, "config.toml")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Create dummy files with unknown extensions
	createFile(t, sourceDir, "file1.xyz")
	createFile(t, sourceDir, "file2.abc")
	createFile(t, storageDir, "file3.def") // Test scanning storage too

	// Create a mock external handler script
	// This script reads stdin, verifies prompt is first line, and returns expected JSON
	handlerScript := filepath.Join(tempDir, "mock_handler.sh")
	scriptContent := `#!/bin/sh
read prompt
read payload

if [ "$prompt" != "Test Prompt" ]; then
  echo "Expected prompt 'Test Prompt', got '$prompt'" >&2
  exit 1
fi

# We could verify payload contains expected strings if we wanted, 
# but simply consuming it and returning valid JSON is enough for now.
# Real check is that prompt was first line.

cat <<EOF
{
  "rules": {
    "experiments": "xyz",
    "misc": "abc,def"
  }
}
EOF
`

	if err := os.WriteFile(handlerScript, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create mock handler script: %v", err)
	}

	// Create Config
	cfg := &config.Config{
		General: config.GeneralConfig{
			SourcePaths: []string{sourceDir},
		},
		Storage: map[string]string{
			"documents": storageDir,
		},
		Rules: map[string]string{
			"documents": "txt",
		},
		Advanced: config.AdvancedConfig{
			UnknownFileHandler: handlerScript,
			SmartLearnPrompt:   "Test Prompt",
			DefaultStoragePath: filepath.Join(tempDir, "organized"),
		},
		ExtensionMap: map[string]string{
			"txt": "documents",
		},
	}

	// Save initial config
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Initialize Organizer
	handler := external.New(handlerScript, cfg)
	org := New(cfg, handler, false, true) // verbose=true

	// Run LearnSmart
	ctx := context.Background()
	if err := org.LearnSmart(ctx, configPath); err != nil {
		t.Fatalf("LearnSmart failed: %v", err)
	}

	// Verify Config Updates in Memory
	if cfg.Rules["experiments"] != "xyz" {
		t.Errorf("Expected rule 'experiments'='xyz', got '%s'", cfg.Rules["experiments"])
	}
	if val, ok := cfg.Rules["misc"]; !ok || (val != "abc,def" && val != "def,abc") {
		t.Errorf("Expected rule 'misc'='abc,def', got '%s'", cfg.Rules["misc"])
	}

	// Verify storage creation from default path
	expectedExperiments := filepath.Join(tempDir, "organized", "experiments")
	if cfg.Storage["experiments"] != expectedExperiments {
		t.Errorf("Expected storage 'experiments'='%s', got '%s'", expectedExperiments, cfg.Storage["experiments"])
	}

	// Verify Config File Updates
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	if loadedCfg.Rules["experiments"] != "xyz" {
		t.Errorf("Persisted config missing 'experiments' rule")
	}
}

func createFile(t *testing.T, dir, name string) {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("dummy content"), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", path, err)
	}
}
