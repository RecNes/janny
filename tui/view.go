package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Header
	header := m.renderHeader()

	// Content
	var content string
	switch m.activeView {
	case ViewDashboard:
		content = m.dashboard.View()
	case ViewConfig:
		content = m.configView.View()
	case ViewLogs:
		content = m.logs.View()
	case ViewHelp:
		content = m.help.View()
	case ViewFileBrowser:
		content = m.fileBrowser.View()
	default:
		content = "Unknown View"
	}

	// Vertically center content and pad
	content = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4). // Subtract header/footer height
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m Model) renderHeader() string {
	title := HeaderStyle.Render(" 🗂  JANNY TUI ")

	tabs := []struct {
		name string
		key  string
	}{
		{"Dashboard", "D"},
		{"Config", "C"},
		{"Logs", "L"},
		{"Files", "F"},
		{"Help", "H"},
	}

	var renderedTabs []string

	for i, t := range tabs {
		// Highlight the hotkey letter with an underline
		displayName := fmt.Sprintf("%s%s",
			lipgloss.NewStyle().Underline(true).Render(string(t.name[0])),
			t.name[1:],
		)

		if int(m.activeView) == i {
			renderedTabs = append(renderedTabs, ActiveTabStyle.Render(displayName))
		} else {
			renderedTabs = append(renderedTabs, InactiveTabStyle.Render(displayName))
		}
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	return lipgloss.JoinHorizontal(lipgloss.Top, title, tabRow)
}

func (m Model) renderFooter() string {
	status := "Ready"
	if m.statusMsg != "" {
		status = m.statusMsg
	}
	if m.lastError != "" {
		status = ErrorStyle.Render(fmt.Sprintf("Error: %s", m.lastError))
	}

	return FooterStyle.Render(" " + status)
}

// Stub for ErrorStyle since it wasn't in styles.go yet or I need to ensure it's there
var ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
