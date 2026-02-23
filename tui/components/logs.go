package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type LogsModel struct {
	Width    int
	Height   int
	Viewport viewport.Model
	Lines    []string
}

func NewLogsModel(width, height int) LogsModel {
	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(0, 1)

	return LogsModel{
		Width:    width,
		Height:   height,
		Viewport: vp,
		Lines:    []string{"Welcome to Janny Logs.", "Press 'd' to return to Dashboard."},
	}
}

func (m *LogsModel) AddLog(msg string) {
	m.Lines = append(m.Lines, fmt.Sprintf("[%s] %s", "INFO", msg))
	m.Viewport.SetContent(strings.Join(m.Lines, "\n"))
	m.Viewport.GotoBottom()
}

func (m LogsModel) View() string {
	return m.Viewport.View()
}
