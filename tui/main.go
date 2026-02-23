package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evrenesat/janny/internal/config"
	"github.com/evrenesat/janny/internal/external"
	"github.com/evrenesat/janny/internal/organizer"
)

func main() {
	// 1. Load Config
	// Use default config path if not specified
	configPath := "~/.config/janny/config.toml"
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup Handler and Organizer
	handler := external.New(cfg.Advanced.UnknownFileHandler, cfg)
	org := organizer.New(cfg, handler, false, false)

	// 3. Initialize Model
	m := NewModel(cfg, configPath, org)

	// 4. Run Program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
