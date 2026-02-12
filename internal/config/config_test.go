package config

import (
	"testing"
)

func TestConfigProcess_Patterns(t *testing.T) {
	// Setup a config with various rules
	cfg := &Config{
		Storage: map[string]string{
			"docs":    "/tmp/docs",
			"images":  "/tmp/images",
			"reports": "/tmp/reports",
			"misc":    "/tmp/misc",
		},
		Rules: map[string]string{
			"docs":    "pdf,txt",
			"images":  "glob:*.jpg, thumb_*.png",
			"reports": "regex:^report_\\d{4}\\.pdf$",
			"misc":    "*readme*",
		},
	}

	if err := cfg.process(); err != nil {
		t.Fatalf("process() failed: %v", err)
	}

	// Verify ExtensionMap
	expectedExts := map[string]string{
		"pdf": "docs",
		"txt": "docs",
	}
	for ext, cat := range expectedExts {
		if got := cfg.ExtensionMap[ext]; got != cat {
			t.Errorf("ExtensionMap[%s] = %s, want %s", ext, got, cat)
		}
	}

	// Verify Patterns
	// We expect specific patterns.
	// Order depends on category sorting:
	// "docs", "images", "misc", "reports" (alphabetical)

	// "docs": "pdf,txt" -> No patterns
	// "images": "glob:*.jpg", "thumb_*.png" (implicit glob? No, thumb_*.png contains *)
	// "misc": "*readme*" (implicit glob)
	// "reports": "regex:..."

	// Expected Patterns:
	// From "images":
	//   Type: Glob, Pattern: "*.jpg"
	//   Type: Glob, Pattern: "thumb_*.png"
	// From "misc":
	//   Type: Glob, Pattern: "*readme*"
	// From "reports":
	//   Type: Regex, Pattern: "^report_\\d{4}\\.pdf$"

	expectedPatterns := []PatternRule{
		{Category: "images", Type: PatternTypeGlob, Pattern: "*.jpg"},
		{Category: "images", Type: PatternTypeGlob, Pattern: "thumb_*.png"},
		{Category: "misc", Type: PatternTypeGlob, Pattern: "*readme*"},
		{Category: "reports", Type: PatternTypeRegex, Pattern: "^report_\\d{4}\\.pdf$"},
	}

	// Sort actual patterns for comparison or check existence?
	// The implementation sorts categories, then iterates rules in order.
	// So order should be stable if we know rule order.
	// But "Rules" map iteration order in "process" loop over categories is sorted by me.
	// But inside "images" rules "glob:*.jpg, thumb_*.png", split by comma preserves order.

	// However, we want to verify we have exactly these patterns.
	if len(cfg.Patterns) != len(expectedPatterns) {
		t.Fatalf("Got %d patterns, want %d", len(cfg.Patterns), len(expectedPatterns))
	}

	// Helper to find a pattern
	findPattern := func(target PatternRule) bool {
		for _, p := range cfg.Patterns {
			if p.Category == target.Category && p.Type == target.Type && p.Pattern == target.Pattern {
				return true
			}
		}
		return false
	}

	for _, p := range expectedPatterns {
		if !findPattern(p) {
			t.Errorf("Missing expected pattern: %+v", p)
		}
	}
}

func TestConfigProcess_Determinism(t *testing.T) {
	// Test that category processing order is deterministic (alphabetical)
	// Create rules where multiple categories claim the same extension/pattern?
	// Config struct Maps don't allow duplicate keys.
	// But we can test that Patterns list is populated in alphabetical order of categories.

	cfg := &Config{
		Storage: map[string]string{
			"zebra": "/tmp/z",
			"alpha": "/tmp/a",
		},
		Rules: map[string]string{
			"zebra": "glob:*",
			"alpha": "glob:*",
		},
	}
	// We need config to initialize maps
	cfg.ExtensionMap = make(map[string]string)
	cfg.Patterns = make([]PatternRule, 0)

	if err := cfg.process(); err != nil {
		t.Fatalf("process() failed: %v", err)
	}

	if len(cfg.Patterns) != 2 {
		t.Fatalf("Expected 2 patterns, got %d", len(cfg.Patterns))
	}

	if cfg.Patterns[0].Category != "alpha" {
		t.Errorf("First pattern should be from 'alpha', got '%s'", cfg.Patterns[0].Category)
	}
	if cfg.Patterns[1].Category != "zebra" {
		t.Errorf("Second pattern should be from 'zebra', got '%s'", cfg.Patterns[1].Category)
	}
}
