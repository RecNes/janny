package organizer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/evrenesat/janny/internal/config"
	"github.com/evrenesat/janny/internal/external"
)

func TestOrganizer_PlanFile(t *testing.T) {
	// Setup config
	cfg := &config.Config{
		Storage: map[string]string{
			"docs":    "/tmp/docs",
			"reports": "/tmp/reports",
			"images":  "/tmp/images",
		},
		Rules: map[string]string{
			// These populate ExtensionMap and Patterns
			"docs":    "txt",
			"reports": "glob:*report*",
			"images":  "regex:^img_\\d+\\.png$",
		},
	}
	// Manually populate derived fields as process() would
	// Since ExtensionMap and Patterns are unexported or internal package specific?
	// Config struct in internal/config/config.go: Patterns is exported (Capital P).
	// ExtensionMap is also exported (Capital E).
	// So direct assignment should work if package is imported correctly.
	// Ah, in test file package is `organizer`, strict visibility applies.
	cfg.ExtensionMap = map[string]string{
		"txt": "docs",
	}
	cfg.Patterns = []config.PatternRule{
		{Category: "images", Type: config.PatternTypeRegex, Pattern: "^img_\\d+\\.png$"},
		// PatternTypeGlob is exported?
		// internal/config/config.go: PatternTypeGlob is exported.

		// PatternRule fields are exported.
		{Category: "reports", Type: config.PatternTypeGlob, Pattern: "*report*"},
		{Category: "docs", Type: config.PatternTypeFolder, Pattern: "my_*"},
	}

	org := New(cfg, nil, false, false)
	ctx := context.Background()

	tests := []struct {
		filename     string
		isDir        bool
		wantCategory string
		wantSkip     bool
	}{
		{"file.txt", false, "docs", false},                // Extension match
		{"final_report.pdf", false, "reports", false},     // Glob match
		{"img_123.png", false, "images", false},           // Regex match
		{"img_abc.png", false, "", true},                  // Regex mismatch
		{"quarterly_report.txt", false, "reports", false}, // Pattern priority
		{"unknown.xyz", false, "", true},                  // No match
		{"my_folder", true, "docs", false},                // Folder match (my_*)
		{"other_folder", true, "", true},                  // Folder mismatch
	}

	for _, tt := range tests {
		path := filepath.Join("/tmp/source", tt.filename)
		action, err := org.planFile(ctx, path, tt.isDir)
		if err != nil {
			t.Errorf("planFile(%s) unexpected error: %v", tt.filename, err)
			continue
		}

		if tt.wantSkip {
			if !action.Skip {
				t.Errorf("planFile(%s) skipped = false, want true", tt.filename)
			}
		} else {
			if action.Skip {
				t.Errorf("planFile(%s) skipped = true, want false (reason: %s)", tt.filename, action.Reason)
			}
			if action.Category != tt.wantCategory {
				t.Errorf("planFile(%s) category = %s, want %s", tt.filename, action.Category, tt.wantCategory)
			}
		}
	}
}

// Test external handler (mocking logic if possible or just checking fallback)
type mockHandler struct {
	external.Handler
}

// Actually external.Handler is a struct, not interface. We can't mock it easily unless we extract interface.
// But we can test standard matching without handler.
