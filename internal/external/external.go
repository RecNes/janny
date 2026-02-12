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

type SmartLearnPayload struct {
	UnknownExtensions []string          `json:"unknown_extensions"`
	Rules             map[string]string `json:"rules"`
}

type SmartLearnResponse struct {
	Rules map[string]string `json:"rules"`
}

// SmartLearn executes the external command with a batch of unknown extensions
func (h *Handler) SmartLearn(ctx context.Context, extensions []string, prompt string) (*SmartLearnResponse, error) {
	if h.command == "" {
		return nil, fmt.Errorf("no external handler command configured")
	}

	payload := SmartLearnPayload{
		UnknownExtensions: extensions,
		Rules:             h.config.Rules,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create command with longer timeout for batch
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second) // 2 minutes timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", h.command)

	// Prepare stdin: Prompt + Newline + JSON
	// Ensure JSON is compact (json.Marshal does this by default)
	input := prompt + "\n" + string(jsonData)
	cmd.Stdin = strings.NewReader(input)

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("external command failed: %v, stderr: %s", err, errOut.String())
	}

	var response SmartLearnResponse
	if err := json.Unmarshal(out.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nOutput was:\n%s", err, out.String())
	}

	return &response, nil
}
