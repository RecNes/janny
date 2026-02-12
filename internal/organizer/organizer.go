package organizer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/evrenesat/janny/internal/config"
	"github.com/evrenesat/janny/internal/external"
)

type Organizer struct {
	config  *config.Config
	handler *external.Handler
	dryRun  bool
	verbose bool
}

func New(cfg *config.Config, handler *external.Handler, dryRun bool, verbose bool) *Organizer {
	return &Organizer{
		config:  cfg,
		handler: handler,
		dryRun:  dryRun,
		verbose: verbose,
	}
}

// FileAction represents a decision made by the organizer for a single file
type FileAction struct {
	SourceDir string
	Filename  string
	Category  string // e.g. "images" or empty if unknown
	TargetDir string // Full target directory path
	DestName  string // Destination filename (usually same as Filename)
	Reason    string // For skips or errors
	Skip      bool
}

// Run executes the organization process
func (o *Organizer) Run(ctx context.Context) error {
	for _, sourcePath := range o.config.General.SourcePaths {
		if o.verbose {
			fmt.Printf("Scanning directory: %s\n", sourcePath)
		}

		actions, err := o.planDirectory(ctx, sourcePath)
		if err != nil {
			return fmt.Errorf("error planning %s: %w", sourcePath, err)
		}

		// Always print the tree in dry-run, or if verbose in actual run
		// User requested this structured output primarily for dry-run validation
		if o.dryRun || o.verbose {
			o.printTree(sourcePath, actions)
		}

		if !o.dryRun {
			if err := o.executeActions(actions); err != nil {
				return fmt.Errorf("error executing actions for %s: %w", sourcePath, err)
			}
		}
	}

	// Auto-clean after organization
	if !o.dryRun {
		if err := o.Clean(ctx); err != nil {
			return fmt.Errorf("error during auto-clean: %w", err)
		}
	}

	return nil
}

// planDirectory scans a directory and returns a list of actions to take
func (o *Organizer) planDirectory(ctx context.Context, sourcePath string) ([]FileAction, error) {
	var actions []FileAction

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return nil, err
	}

	for _, d := range entries {
		path := filepath.Join(sourcePath, d.Name())
		action, err := o.planFile(ctx, path, d.IsDir())
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Error planning file %s: %v\n", d.Name(), err)
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func (o *Organizer) planFile(ctx context.Context, path string, isDir bool) (FileAction, error) {
	filename := filepath.Base(path)
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	action := FileAction{
		SourceDir: filepath.Dir(path),
		Filename:  filename,
		DestName:  filename,
	}

	if ext == "" {
		// Even without extension, it might match a pattern (e.g. "README")
		// So we shouldn't skip yet, unless we want to enforce extensions for simple rules.
		// Use empty ext for extension map lookup (which will likely fail), but check patterns.
	}

	// 1. Check Patterns (Priority)
	for _, rule := range o.config.Patterns {
		matched := false
		var err error

		switch rule.Type {
		case config.PatternTypeGlob:
			if isDir {
				continue
			}
			matched, err = filepath.Match(rule.Pattern, filename)
			if err != nil {
				// Invalid glob pattern in config, log and continue?
				fmt.Fprintf(os.Stderr, "Invalid glob pattern '%s': %v\n", rule.Pattern, err)
				continue
			}
		case config.PatternTypeRegex:
			if isDir {
				continue
			}
			matched, err = regexp.MatchString(rule.Pattern, filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid regex pattern '%s': %v\n", rule.Pattern, err)
				continue
			}
		case config.PatternTypeFolder:
			if !isDir {
				continue
			}
			matched, err = filepath.Match(rule.Pattern, filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid folder pattern '%s': %v\n", rule.Pattern, err)
				continue
			}
		}

		if matched {
			if o.verbose {
				fmt.Printf("Entry '%s' matched pattern '%s' -> %s\n", filename, rule.Pattern, rule.Category)
			}
			// Found a match!
			return o.createAction(path, filename, rule.Category)
		}
	}

	if isDir {
		// Directories only match specific folder patterns, not extensions
		return FileAction{
			SourceDir: filepath.Dir(path),
			Filename:  filename,
			DestName:  filename,
			Skip:      true,
			Reason:    "directory not matched by any folder pattern",
		}, nil
	}

	// 2. Check Extension Map
	if ext == "" {
		// If no extension and no pattern matched, skip
		action := FileAction{
			SourceDir: filepath.Dir(path),
			Filename:  filename,
			DestName:  filename,
			Skip:      true,
			Reason:    "no extension and no pattern match",
		}
		return action, nil
	}

	category, found := o.config.ExtensionMap[ext]
	if !found {
		if o.handler != nil {
			var err error
			category, err = o.handler.Classify(ctx, path)
			if err != nil {
				action.Skip = true
				action.Reason = fmt.Sprintf("external handler failed: %v", err)
				return action, nil
			}
			if category == "" {
				action.Skip = true
				action.Reason = "external handler returned no category"
				return action, nil
			}
			// Verify category exists
			if _, ok := o.config.Storage[category]; !ok {
				action.Skip = true
				action.Reason = fmt.Sprintf("unknown category '%s' from handler", category)
				return action, nil
			}
		} else {
			action.Skip = true
			action.Reason = "unknown extension"
			return action, nil
		}
	}

	return o.createAction(path, filename, category)
}

func (o *Organizer) createAction(path, filename, category string) (FileAction, error) {
	action := FileAction{
		SourceDir: filepath.Dir(path),
		Filename:  filename,
		DestName:  filename,
		Category:  category,
		TargetDir: o.config.Storage[category],
	}

	// Check for conflicts and resolve filename
	targetPath := filepath.Join(action.TargetDir, action.DestName)
	if _, err := os.Stat(targetPath); err == nil {
		// Conflict detected, find a new name
		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		ext := filepath.Ext(filename)
		for i := 1; ; i++ {
			newFilename := fmt.Sprintf("%s_%d%s", name, i, ext)
			targetPath = filepath.Join(action.TargetDir, newFilename)
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				action.DestName = newFilename
				break
			}
		}
	}

	return action, nil
}

func (o *Organizer) printTree(root string, actions []FileAction) {
	// If no actions, don't print anything for this directory?
	// Or print just the root? Let's print root if there are relevant actions.

	// Filter out "no extension" or "unknown extension" skips if not verbose?
	// User said "output is too noisy".
	// "skipping unknown file" is noise if I have many files.
	// Let's only print actual moves, or skips that are significant (errors).

	validActions := make([]FileAction, 0, len(actions))
	for _, a := range actions {
		if !a.Skip {
			validActions = append(validActions, a)
		}
	}

	if len(validActions) == 0 {
		return
	}

	fmt.Printf("%s\n", root)
	for i, action := range validActions {
		prefix := "├── "
		if i == len(validActions)-1 {
			prefix = "└── "
		}

		// Format: filename -> [category]
		// Or: filename -> [category]/newname if renamed

		out := fmt.Sprintf("%s%s -> [%s]", prefix, action.Filename, action.Category)
		if action.DestName != action.Filename {
			out += fmt.Sprintf("/%s", action.DestName)
		}

		fmt.Println(out)
	}
}

func (o *Organizer) executeActions(actions []FileAction) error {
	for _, action := range actions {
		if action.Skip {
			continue
		}

		// Ensure target directory exists
		if err := os.MkdirAll(action.TargetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", action.TargetDir, err)
		}

		srcPath := filepath.Join(action.SourceDir, action.Filename)
		dstPath := filepath.Join(action.TargetDir, action.DestName)

		if o.verbose {
			fmt.Printf("Moving %s -> %s\n", srcPath, dstPath)
		}

		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s: %w", srcPath, err)
		}
	}
	return nil
}

// Learn scans the target directories and updates the configuration with inferred rules
func (o *Organizer) Learn(configPath string) error {
	learnedCount := 0

	for category, targetDir := range o.config.Storage {
		if o.verbose {
			fmt.Printf("Learning from category '%s' in %s\n", category, targetDir)
		}

		err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// If target directory doesn't exist, just skip it
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
			if ext == "" {
				return nil
			}

			// Check if this extension is already known
			if _, known := o.config.ExtensionMap[ext]; known {
				return nil
			}

			// It's a new extension! Add it to the rules.
			if o.verbose {
				fmt.Printf("Found new extension '%s' in category '%s'\n", ext, category)
			}

			// Update the Rules map
			currentRules := o.config.Rules[category]
			if currentRules == "" {
				o.config.Rules[category] = ext
			} else {
				o.config.Rules[category] = currentRules + "," + ext
			}

			// Update the in-memory ExtensionMap
			o.config.ExtensionMap[ext] = category
			learnedCount++

			return nil
		})

		if err != nil {
			return fmt.Errorf("error learning from %s: %w", targetDir, err)
		}
	}

	if learnedCount > 0 {
		fmt.Printf("Learned %d new file extension rules.\n", learnedCount)
		// Save the updated configuration
		if err := config.SaveConfig(configPath, o.config); err != nil {
			return fmt.Errorf("failed to save updated configuration: %w", err)
		}
		fmt.Printf("Configuration saved to %s\n", configPath)
	} else {
		fmt.Println("No new rules learned.")
	}

	return nil
}

// Clean performs auto-deletion of old files based on configuration
func (o *Organizer) Clean(ctx context.Context) error {
	for category, days := range o.config.AutoClean {
		targetDir, ok := o.config.Storage[category]
		if !ok {
			// Config validation should catch this, but safe check
			if o.verbose {
				fmt.Printf("Auto-clean skipped for category '%s': no storage path defined\n", category)
			}
			continue
		}

		if o.verbose {
			fmt.Printf("Cleaning category '%s' (older than %d days) in %s\n", category, days, targetDir)
		}

		cutoff := time.Now().AddDate(0, 0, -days)

		entries, err := os.ReadDir(targetDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read directory %s for cleaning: %w", targetDir, err)
		}

		cleanedCount := 0
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get info for %s: %v\n", entry.Name(), err)
				continue
			}

			if info.ModTime().Before(cutoff) {
				fullPath := filepath.Join(targetDir, entry.Name())
				if o.verbose {
					fmt.Printf("Deleting old item: %s (Modified: %s)\n", fullPath, info.ModTime().Format(time.RFC3339))
				}

				if err := os.RemoveAll(fullPath); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to delete %s: %v\n", fullPath, err)
				} else {
					cleanedCount++
				}
			}
		}

		if cleanedCount > 0 && o.verbose {
			fmt.Printf("Deleted %d old items from %s\n", cleanedCount, targetDir)
		}
	}
	return nil
}
