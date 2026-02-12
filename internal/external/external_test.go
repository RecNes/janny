package external

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/evrenesat/janny/internal/config"
)

func TestHandler_Classify_WithConfig(t *testing.T) {
	// Create a dummy config
	cfg := &config.Config{
		Rules: map[string]string{
			"test_cat": "test_ext",
		},
		Storage: map[string]string{
			"test_cat": "/tmp/test",
		},
	}

	// Use 'cat' as the command to echo stdin to stdout
	h := New("cat", cfg)

	// Execute Classify
	// The command will be "cat <filepath>"
	// But we are interested in stdin.
	// Wait, "cat <filepath>" will read the file, not stdin unless we use "-" or no args.
	// The current implementation executes "sh -c command filepath".
	// So it executes "sh -c cat filepath". This will read the file!
	// We want testing that STDIN is passed.
	// If the command ignores stdin, we can't verify it with "cat filepath".
	// We need a command that reads stdin.
	// We can use a custom shell command that ignores arguments and reads stdin.
	// command = "cat < /dev/stdin #"
	// "sh -c cat < /dev/stdin # filepath" -> this comments out the filepath.

	// Let's try to use a command that prints stdin.
	// "cat" reads from stdin if no file args.
	// But we Pass filepath as arg.
	// So "cat filepath" reads file.

	// We can use a script that just reads stdin.
	// h := New("read LINE; echo $LINE; #", cfg)
	// command + " " + filePath
	// "read LINE; echo $LINE; # filepath"

	// Let's try that.
	// Note: read in sh might behave differently.
	// Let's use "grep .", or "head -n 1" from stdin?
	// No, that requires ignoring the argument.

	// How about using python or similar if available?
	// Or just "cat; echo" and ignore the trailing arg?
	// "cat" will try to open the trailing arg.

	// We can inject a comment char `#` in the command if we want to ignore the filepath?
	// The implementation does: exec.CommandContext(ctx, "sh", "-c", h.command+" "+filePath)

	// If I set command to `cat #`, then it becomes `cat # filepath`.
	// `sh -c "cat # filepath"` -> `cat` runs without args, so it reads stdin!

	h = New("cat #", cfg)

	ctx := context.Background()
	out, err := h.Classify(ctx, "dummy_path")
	if err != nil {
		t.Fatalf("Classify failed: %v", err)
	}

	// The output should be the JSON config
	// Since JSON marshalling is not deterministic with map key order, we should unmarshal it back to check.
	var receivedConfig config.Config
	if err := json.Unmarshal([]byte(out), &receivedConfig); err != nil {
		t.Fatalf("Failed to parsing output as JSON: %v. Output: %s", err, out)
	}

	// Check if data matches
	if len(receivedConfig.Rules) != 1 || receivedConfig.Rules["test_cat"] != "test_ext" {
		t.Errorf("Config mismatch. Rules: %v", receivedConfig.Rules)
	}
	if len(receivedConfig.Storage) != 1 || receivedConfig.Storage["test_cat"] != "/tmp/test" {
		t.Errorf("Config mismatch. Storage: %v", receivedConfig.Storage)
	}
}
