package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/evrenesat/janny/internal/config"
)

// Handler manages external command execution for unknown files
type Handler struct {
	command string
	config  *config.Config
}

func New(command string, cfg *config.Config) *Handler {
	return &Handler{
		command: command,
		config:  cfg,
	}
}

// Classify executes the external command with the file path and returns the category
func (h *Handler) Classify(ctx context.Context, filePath string) (string, error) {
	if h.command == "" {
		return "", nil
	}

	// Create command with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second) // 30s timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", h.command+" "+filePath)

	// Prepare stdin
	if h.config != nil {
		jsonData, err := json.Marshal(h.config)
		if err == nil {
			cmd.Stdin = bytes.NewReader(jsonData)
		} else {
			// Log error but continue? Or just ignore?
			// For now, let's just ignore if marshalling fails, though it shouldn't.
			// Ideally we should log it.
		}
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Capture stderr too just in case

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("external command failed: %v, output: %s", err, out.String())
	}

	category := strings.TrimSpace(out.String())
	return category, nil
}
