package backup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/evrenesat/janny/internal/config"
)

type Backup struct {
	config      *config.BackupConfig
	sourcePaths []string
	dryRun      bool
}

func New(cfg *config.BackupConfig, sourcePaths []string, dryRun bool) *Backup {
	return &Backup{
		config:      cfg,
		sourcePaths: sourcePaths,
		dryRun:      dryRun,
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

	for _, source := range b.sourcePaths {
		if err := b.sync(source); err != nil {
			return fmt.Errorf("backup failed for source %s: %w", source, err)
		}
	}

	return nil
}

func (b *Backup) sync(source string) error {
	args := []string{"-av", "--delete"}

	// Add exclusions
	for _, exclude := range b.config.Exclude {
		args = append(args, "--exclude", exclude)
	}

	args = append(args, source, b.config.Destination)

	if b.dryRun {
		fmt.Printf("[DRY RUN] rsync %s\n", args)
		return nil
	}

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync command failed: %w", err)
	}

	return nil
}
