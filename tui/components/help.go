package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HelpModel struct {
	Width  int
	Height int
}

func (m HelpModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true).
		Underline(true).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	renderRow := func(k, d string) string {
		return fmt.Sprintf("  %s %s", keyStyle.Render(strings.TrimSpace(k)), descStyle.Render(d))
	}

	b.WriteString(titleStyle.Render(" GLOBAL SHORTCUTS ") + "\n\n")
	b.WriteString(renderRow("d      ", "Switch to Dashboard") + "\n")
	b.WriteString(renderRow("c      ", "Switch to Configuration") + "\n")
	b.WriteString(renderRow("l      ", "Switch to Log screen") + "\n")
	b.WriteString(renderRow("f      ", "Switch to File browser") + "\n")
	b.WriteString(renderRow("h      ", "Switch to Help screen") + "\n")
	b.WriteString(renderRow("q      ", "Quit") + "\n\n")

	b.WriteString(titleStyle.Render(" DASHBOARD ") + "\n\n")
	b.WriteString(renderRow("o      ", "Organize files manually") + "\n")
	b.WriteString(renderRow("b      ", "Start backup sync") + "\n")
	b.WriteString(renderRow("s      ", "Run Smart Learn") + "\n\n")

	b.WriteString(titleStyle.Render(" ABOUT JANNY ") + "\n\n")
	b.WriteString("  Janny is an automated file organizer.\n")
	b.WriteString("  How it works:\n")
	b.WriteString("  1. Scan: Iterates through source paths.\n")
	b.WriteString("  2. Match: Uses extension maps, regex, or globs.\n")
	b.WriteString("  3. Move: Places files in categorized storage.\n")
	b.WriteString("  4. Clean: (Optional) Deletes old file versions.\n")
	b.WriteString("  5. Backup: (Optional) Syncs storage to destination.\n")

	return b.String()
}
