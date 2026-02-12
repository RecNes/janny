package backup

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/evrenesat/janny/internal/config"
)

func TestBackup_Run_DryRun_Exclusions(t *testing.T) {
	// Setup capture of stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.BackupConfig{
		Enabled:            true,
		Destination:        "/tmp/backup",
		ExcludeFileTypes:   []string{"pdf", "jpg"},
		ExcludeDirectories: []string{"node_modules", "temp"},
	}

	// Create temp source directory in current dir to avoid permission issues
	sourceDir, err := os.MkdirTemp(".", "temp_test_backup_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	b := New(cfg, []string{sourceDir}, true) // true for dryRun

	err = b.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains rsync command with expected exclusions
	expectedParts := []string{
		"rsync",
		"-av", "--delete",
		"--exclude *.pdf",
		"--exclude *.jpg",
		"--exclude node_modules/",
		"--exclude temp/",
		sourceDir,
		"/tmp/backup",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output missing expected part: %s. Got: %s", part, output)
		}
	}
}

func TestBackup_Run_Disabled(t *testing.T) {
	cfg := &config.BackupConfig{
		Enabled: false,
	}
	b := New(cfg, []string{"/tmp/source"}, true)
	if err := b.Run(); err != nil {
		t.Errorf("Run() failed when disabled: %v", err)
	}
}
