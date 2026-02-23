package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evrenesat/janny/internal/backup"
	"github.com/evrenesat/janny/tui/components"
)

// commands use components.StatusMsg and components.DashboardMsg

func (m Model) organizeCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.org.Run(context.Background())
		if err != nil {
			return components.DashboardMsg(fmt.Sprintf("❌ ERROR: Error during organization:\n%v", err))
		}
		return components.DashboardMsg("✅ SUCCESS: All files organized according to rules.")
	}
}

func (m Model) backupCmd() tea.Cmd {
	return func() tea.Msg {
		if !m.cfg.Backup.Enabled {
			return components.DashboardMsg("⚠️ WARNING: Backup is disabled in configuration.")
		}

		var storagePaths []string
		for _, path := range m.cfg.Storage {
			storagePaths = append(storagePaths, path)
		}

		b := backup.New(&m.cfg.Backup, storagePaths, false)
		err := b.Run()
		if err != nil {
			return components.DashboardMsg(fmt.Sprintf("❌ ERROR: Error during backup:\n%v", err))
		}
		return components.DashboardMsg(fmt.Sprintf("✅ SUCCESS: Backup completed.\nDestination: %s", m.cfg.Backup.Destination))
	}
}

func (m Model) smartLearnCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.org.LearnSmart(context.Background(), m.configPath)
		if err != nil {
			return components.DashboardMsg(fmt.Sprintf("❌ ERROR: Error during learning process:\n%v", err))
		}
		return components.DashboardMsg("✅ SUCCESS: Unknown file types learned smartly and rules updated.")
	}
}
