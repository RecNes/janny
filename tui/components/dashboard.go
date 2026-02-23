package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/evrenesat/janny/internal/config"
)

type DashboardModel struct {
	Cfg    *config.Config
	Width  int
	Height int
	Output string
}

func (m DashboardModel) View() string {
	if m.Cfg == nil {
		return "Loading configuration..."
	}

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true).
		MarginBottom(1)

	// Status Card
	statusCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(36)

	var statusContent strings.Builder
	statusContent.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("System Status")))
	statusContent.WriteString(fmt.Sprintf("Source Paths: %d\n", len(m.Cfg.General.SourcePaths)))
	statusContent.WriteString(fmt.Sprintf("Storage Categories: %d\n", len(m.Cfg.Storage)))
	statusContent.WriteString(fmt.Sprintf("Backup Enabled: %v\n", m.Cfg.Backup.Enabled))

	// Rules Summary Card
	rulesCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFA500")).
		Padding(1, 2).
		Width(36)

	var rulesContent strings.Builder
	rulesContent.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("Rules Summary")))
	rulesContent.WriteString(fmt.Sprintf("Extension Map: %d rules\n", len(m.Cfg.ExtensionMap)))
	rulesContent.WriteString(fmt.Sprintf("Patterns: %d\n", len(m.Cfg.Patterns)))
	rulesContent.WriteString(fmt.Sprintf("Auto Clean: %d categories\n", len(m.Cfg.AutoClean)))

	// Layout columns
	cards := lipgloss.JoinHorizontal(lipgloss.Top,
		statusCard.Render(statusContent.String()),
		rulesCard.Render(rulesContent.String()),
	)

	b.WriteString(cards)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(" Use hotkeys to start an action (o/b/s)"))
	b.WriteString("\n\n")

	// Output Box
	outputWidth := m.Width - 4
	if outputWidth < 76 {
		outputWidth = 76
	}

	// Calculate remaining height
	outputBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF00")).
		Padding(0, 1).
		Width(outputWidth).
		Height(m.Height - 16)

	outputTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render(" Operation Output ")

	content := m.Output
	if content == "" {
		content = "No operations performed yet."
	}

	b.WriteString(outputTitle + "\n")
	b.WriteString(outputBoxStyle.Render(content))

	return b.String()
}
