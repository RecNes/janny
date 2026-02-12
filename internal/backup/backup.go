package backup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/evrenesat/janny/internal/config"
)

type Backup struct {
	config       *config.BackupConfig
	storagePaths []string
	dryRun       bool
}

func New(cfg *config.BackupConfig, storagePaths []string, dryRun bool) *Backup {
	return &Backup{
		config:       cfg,
		storagePaths: storagePaths,
		dryRun:       dryRun,
	}
}

func (b *Backup) Run() error {
	if !b.config.Enabled {
		return nil
	}

	if b.config.Destination == "" {
		return fmt.Errorf("backup enabled but no destination path configured")
	}

	// Ensure destination directory exists
	if !b.dryRun {
		if err := os.MkdirAll(b.config.Destination, 0755); err != nil {
			return fmt.Errorf("failed to create backup destination %s: %w", b.config.Destination, err)
		}
	}

	for _, source := range b.storagePaths {
		if err := b.sync(source); err != nil {
			return fmt.Errorf("backup failed for source %s: %w", source, err)
		}
	}

	return nil
}

func (b *Backup) sync(source string) error {
	args := []string{"-av", "--delete"}

	if b.dryRun {
		args = append(args, "-n")
	}

	// Add file type exclusions
	for _, ext := range b.config.ExcludeFileTypes {
		// Ensure we're excluding based on extension pattern
		if !strings.HasPrefix(ext, "*.") {
			if strings.HasPrefix(ext, ".") {
				ext = "*" + ext
			} else {
				ext = "*." + ext
			}
		}
		args = append(args, "--exclude", ext)
	}

	// Add directory exclusions
	for _, dir := range b.config.ExcludeDirectories {
		// Ensure directory pattern ends with slash to match only directories
		if !strings.HasSuffix(dir, "/") {
			dir = dir + "/"
		}
		args = append(args, "--exclude", dir)
	}

	args = append(args, source, b.config.Destination)

	if b.dryRun {
		fmt.Printf("[DRY RUN] rsync %s\n", args)
	}

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync command failed: %w", err)
	}

	return nil
}
