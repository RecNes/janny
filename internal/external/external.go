package external

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Handler manages external command execution for unknown files
type Handler struct {
	command string
}

func New(command string) *Handler {
	return &Handler{
		command: command,
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
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Capture stderr too just in case

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("external command failed: %v, output: %s", err, out.String())
	}

	category := strings.TrimSpace(out.String())
	return category, nil
}
