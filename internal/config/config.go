package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration structure
type Config struct {
	General   GeneralConfig     `toml:"general"`
	Storage   map[string]string `toml:"storage"`
	Rules     map[string]string `toml:"rules"`
	Backup    BackupConfig      `toml:"backup"`
	Advanced  AdvancedConfig    `toml:"advanced"`
	AutoClean map[string]int    `toml:"auto_clean"`

	// Derived configuration
	ExtensionMap map[string]string `toml:"-"` // extension -> category
	Patterns     []PatternRule     `toml:"-"` // ordered list of pattern rules
}

type PatternType int

const (
	PatternTypeGlob PatternType = iota
	PatternTypeRegex
	PatternTypeFolder
)

type PatternRule struct {
	Category string
	Pattern  string
	Type     PatternType
}

type GeneralConfig struct {
	SourcePaths []string `toml:"source_paths"`
}

type BackupConfig struct {
	Enabled            bool     `toml:"enabled"`
	Destination        string   `toml:"destination"`
	ExcludeFileTypes   []string `toml:"exclude_file_types"`
	ExcludeDirectories []string `toml:"exclude_directories"`
}

type AdvancedConfig struct {
	UnknownFileHandler string `toml:"unknown_file_handler"`
	SmartLearnPrompt   string `toml:"smart_learn_prompt"`
	DefaultStoragePath string `toml:"default_storage_path"`
}

// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {
	// Expand home directory in path if needed
	expandedPath, err := expandPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		// Create default config
		cfg := DefaultConfig()
		if err := SaveConfig(expandedPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}

		// Post-process default configuration to expand paths
		if err := cfg.process(); err != nil {
			return nil, fmt.Errorf("failed to process default configuration: %w", err)
		}

		return cfg, nil
	}

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Post-process configuration
	if err := cfg.process(); err != nil {
		return nil, fmt.Errorf("failed to process configuration: %w", err)
	}

	return &cfg, nil
}

// SaveConfig writes the configuration to the specified path
func SaveConfig(path string, cfg *Config) error {
	expandedPath, err := expandPath(path)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(expandedPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	// Candidates for source paths
	candidates := []string{"~/Downloads", "~/Desktop"}
	var validSources []string

	for _, path := range candidates {
		expanded, err := expandPath(path)
		if err == nil {
			if _, err := os.Stat(expanded); err == nil {
				validSources = append(validSources, path)
			}
		}
	}

	// If no candidates exist, default to current directory or warn
	if len(validSources) == 0 {
		// Fallback to home if nothing else exists, though unlikely
		validSources = []string{"~/"}
	}

	return &Config{
		General: GeneralConfig{
			SourcePaths: validSources,
		},
		Storage: map[string]string{
			"documents": "~/Documents/Janny/Documents",
			"images":    "~/Documents/Janny/Images",
			"archives":  "~/Documents/Janny/Archives",
			"music":     "~/Documents/Janny/Music",
			"videos":    "~/Documents/Janny/Videos",
		},
		Rules: map[string]string{
			"documents": "pdf,doc,docx,txt,md,xls,xlsx,ppt,pptx",
			"images":    "jpg,jpeg,png,gif,bmp,svg,webp",
			"archives":  "zip,tar,gz,7z,rar",
			"music":     "mp3,wav,flac,m4a",
			"videos":    "mp4,mkv,avi,mov",
		},
		Backup: BackupConfig{
			Enabled:            false,
			Destination:        "",
			ExcludeFileTypes:   []string{},
			ExcludeDirectories: []string{},
		},
		Advanced: AdvancedConfig{
			UnknownFileHandler: "",
			SmartLearnPrompt:   "You are a file organization assistant. I will provide a list of file extensions that I have found in my chaotic directories, along with my current organization rules and storage paths. Your task is to propose new rules for these extensions. \n\nOutput ONLY a JSON object with the following structure:\n{\n  \"rules\": {\n    \"category_name\": \"ext1,ext2\"\n  }\n}\n\nReuse existing categories if they fit. Create new categories only if necessary. Config is: ",
			DefaultStoragePath: "~/Documents/Janny/Organized",
		},
	}
}

// process handles path expansion and building the extension map
func (c *Config) process() error {
	// Expand source paths
	for i, path := range c.General.SourcePaths {
		expanded, err := expandPath(path)
		if err != nil {
			return fmt.Errorf("failed to expand source path %s: %w", path, err)
		}
		c.General.SourcePaths[i] = expanded
	}

	// Expand storage paths
	for key, path := range c.Storage {
		expanded, err := expandPath(path)
		if err != nil {
			return fmt.Errorf("failed to expand storage path for %s: %w", key, err)
		}
		c.Storage[key] = expanded
	}

	// Expand backup destination
	if c.Backup.Destination != "" {
		expanded, err := expandPath(c.Backup.Destination)
		if err != nil {
			return fmt.Errorf("failed to expand backup destination: %w", err)
		}
		c.Backup.Destination = expanded
	}

	// Expand default storage path
	if c.Advanced.DefaultStoragePath != "" {
		expanded, err := expandPath(c.Advanced.DefaultStoragePath)
		if err != nil {
			return fmt.Errorf("failed to expand default storage path: %w", err)
		}
		c.Advanced.DefaultStoragePath = expanded
	}

	// Build extension map and patterns
	c.ExtensionMap = make(map[string]string)
	c.Patterns = make([]PatternRule, 0)

	// Get sorted keys for determinism
	categories := make([]string, 0, len(c.Rules))
	for k := range c.Rules {
		categories = append(categories, k)
	}
	// Sort categories to ensure consistent order of pattern evaluation
	// If a file matches multiple patterns from different categories, the first one wins
	// The user requirement said: "if multiple patterns from different categories match the same file, the winner is determined by alphabetical order of the category name"
	// So we should append patterns in alphabetical order of categories.
	// Since we iterate sequentially in Organizer, the first match wins.
	// So sorting categories alphabetically here achieves that.
	sort.Strings(categories)

	for _, category := range categories {
		// Verify category exists in storage
		if _, ok := c.Storage[category]; !ok {
			return fmt.Errorf("rule references unknown storage category: %s", category)
		}

		extensions := c.Rules[category]
		// Split comma-separated rules
		rules := strings.Split(extensions, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			// Check for Regex
			if strings.HasPrefix(rule, "regex:") {
				pattern := strings.TrimPrefix(rule, "regex:")
				c.Patterns = append(c.Patterns, PatternRule{
					Category: category,
					Pattern:  pattern,
					Type:     PatternTypeRegex,
				})
				continue
			}

			// Check for Glob
			// Explicit glob: prefix
			if strings.HasPrefix(rule, "glob:") {
				pattern := strings.TrimPrefix(rule, "glob:")
				c.Patterns = append(c.Patterns, PatternRule{
					Category: category,
					Pattern:  pattern,
					Type:     PatternTypeGlob,
				})
				continue
			}

			// Check for Folder
			if strings.HasPrefix(rule, "folder:") {
				pattern := strings.TrimPrefix(rule, "folder:")
				c.Patterns = append(c.Patterns, PatternRule{
					Category: category,
					Pattern:  pattern,
					Type:     PatternTypeFolder,
				})
				continue
			}

			// Implicit glob: contains *, ?, [, ]
			if strings.ContainsAny(rule, "*?[]") {
				c.Patterns = append(c.Patterns, PatternRule{
					Category: category,
					Pattern:  rule,
					Type:     PatternTypeGlob,
				})
				continue
			}

			// Otherwise, it's a simple extension
			cleanExt := strings.TrimPrefix(rule, ".")
			c.ExtensionMap[strings.ToLower(cleanExt)] = category
		}
	}

	return nil
}

// expandPath expands the ~ to the user's home directory
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, path[2:]), nil
}
