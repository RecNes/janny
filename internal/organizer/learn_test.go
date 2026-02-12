package organizer

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evrenesat/janny/internal/config"
)

// Helper to capture stdout
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestOrganizer_Learn_DryRun_ListRules(t *testing.T) {
	// Setup temp structure
	tmpDir, err := os.MkdirTemp(".", "janny_learn_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storageDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Create a file with unknown extension
	if err := os.WriteFile(filepath.Join(storageDir, "test.xyz"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Storage: map[string]string{
			"docs": storageDir,
		},
		Rules:        map[string]string{},
		ExtensionMap: map[string]string{},
	}

	org := New(cfg, nil, true, false) // DryRun = true

	output := captureStdout(func() {
		if err := org.Learn(filepath.Join(tmpDir, "config.toml")); err != nil {
			t.Fatalf("Learn() failed: %v", err)
		}
	})

	if !strings.Contains(output, "Would learn 1 new file extension rules") {
		t.Errorf("Expected output to mention 1 new rule, got:\n%s", output)
	}
	if !strings.Contains(output, ".xyz -> docs") {
		t.Errorf("Expected output to list .xyz -> docs rule, got:\n%s", output)
	}
	if !strings.Contains(output, "Configuration NOT saved") {
		t.Errorf("Expected output to say Configuration NOT saved, got:\n%s", output)
	}
}

func TestOrganizer_Learn_BackupConfig(t *testing.T) {
	// Setup temp structure
	tmpDir, err := os.MkdirTemp(".", "janny_learn_backup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")
	// Create dummy config file
	if err := os.WriteFile(configPath, []byte("[general]\n"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	storageDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}
	// Create a file with unknown extension
	if err := os.WriteFile(filepath.Join(storageDir, "test.abc"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Storage: map[string]string{
			"docs": storageDir,
		},
		Rules:        map[string]string{},
		ExtensionMap: map[string]string{},
	}

	org := New(cfg, nil, false, false) // DryRun = false

	output := captureStdout(func() {
		if err := org.Learn(configPath); err != nil {
			t.Fatalf("Learn() failed: %v", err)
		}
	})

	if !strings.Contains(output, "Created config backup") {
		t.Errorf("Expected output to confirm backup creation, got:\n%s", output)
	}

	// Verify backup file exists
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	backupFound := false
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "config.toml.") && strings.HasSuffix(f.Name(), ".bak") {
			backupFound = true
			break
		}
	}

	if !backupFound {
		t.Errorf("Backup file not found in %s", tmpDir)
	}
}
