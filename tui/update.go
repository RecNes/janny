package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evrenesat/janny/tui/components"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case components.DashboardMsg:
		m.dashboard.Output = string(msg)
		m.statusMsg = "Operation complete."
		return m, nil

	case components.StatusMsg:
		m.statusMsg = string(msg)
		return m, nil

	case components.BrowseMsg:
		m.browsingFieldIndex = msg.FieldIndex
		m.fileBrowser.SelectMode = true
		m.activeView = ViewFileBrowser
		return m, nil

	case components.SelectPathMsg:
		// Attach the stored field index
		msg.FieldIndex = m.browsingFieldIndex
		m.configView.SetValue(msg.FieldIndex, msg.Path)
		m.activeView = ViewConfig
		m.fileBrowser.SelectMode = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "d":
			m.activeView = ViewDashboard
		case "c":
			m.activeView = ViewConfig
		case "l":
			m.activeView = ViewLogs
		case "f":
			m.activeView = ViewFileBrowser
		case "h":
			m.activeView = ViewHelp

		case "o":
			if m.activeView == ViewDashboard {
				m.statusMsg = "Organizing..."
				return m, m.organizeCmd()
			}
		case "b":
			if m.activeView == ViewDashboard {
				m.statusMsg = "Backing up..."
				return m, m.backupCmd()
			}
		case "s":
			if m.activeView == ViewDashboard {
				m.statusMsg = "Learning..."
				return m, m.smartLearnCmd()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update components
		m.dashboard.Width = msg.Width
		m.dashboard.Height = msg.Height - 4
		m.help.Width = msg.Width
		m.help.Height = msg.Height - 4

		m.logs.Width = msg.Width
		m.logs.Height = msg.Height - 4
		m.logs.Viewport.Width = msg.Width
		m.logs.Viewport.Height = msg.Height - 4

		m.configView.Width = msg.Width
		m.configView.Height = msg.Height - 4

		m.fileBrowser.Width = msg.Width
		m.fileBrowser.Height = msg.Height - 4
		m.fileBrowser.List.SetSize(msg.Width, msg.Height-4)
	}

	// Delegate update to sub-models if needed
	switch m.activeView {
	case ViewLogs:
		var logsCmd tea.Cmd
		m.logs.Viewport, logsCmd = m.logs.Viewport.Update(msg)
		cmd = tea.Batch(cmd, logsCmd)
	case ViewConfig:
		var configCmd tea.Cmd
		m.configView, configCmd = m.configView.Update(msg)
		cmd = tea.Batch(cmd, configCmd)
	case ViewFileBrowser:
		// Navigation logic
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			if i, ok := m.fileBrowser.List.SelectedItem().(interface {
				GetPath() string
				IsDir() bool
			}); ok {
				if i.IsDir() {
					m.fileBrowser.Path = i.GetPath()
					m.fileBrowser.Refresh()
				}
			}
		}

		var fbCmd tea.Cmd
		m.fileBrowser, fbCmd = m.fileBrowser.Update(msg)
		cmd = tea.Batch(cmd, fbCmd)
	}

	return m, cmd
}
