package organizer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
		// Skip subdirectories
		if d.IsDir() {
			continue
		}

		path := filepath.Join(sourcePath, d.Name())
		action, err := o.planFile(ctx, path)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Error planning file %s: %v\n", d.Name(), err)
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func (o *Organizer) planFile(ctx context.Context, path string) (FileAction, error) {
	filename := filepath.Base(path)
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	action := FileAction{
		SourceDir: filepath.Dir(path),
		Filename:  filename,
		DestName:  filename,
	}

	if ext == "" {
		action.Skip = true
		action.Reason = "no extension"
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

	action.Category = category
	action.TargetDir = o.config.Storage[category]

	// Check for conflicts and resolve filename
	targetPath := filepath.Join(action.TargetDir, action.DestName)
	if _, err := os.Stat(targetPath); err == nil {
		// Conflict detected, find a new name
		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		for i := 1; ; i++ {
			newFilename := fmt.Sprintf("%s_%d.%s", name, i, ext)
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
