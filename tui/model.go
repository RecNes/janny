package main

import (
	"os"

	"github.com/evrenesat/janny/internal/config"
	"github.com/evrenesat/janny/internal/organizer"
	"github.com/evrenesat/janny/tui/components"
)

type ActiveView int

const (
	ViewDashboard ActiveView = iota
	ViewConfig
	ViewLogs
	ViewFileBrowser
	ViewHelp
)

type Model struct {
	activeView ActiveView
	width      int
	height     int
	cfg        *config.Config
	configPath string
	org        *organizer.Organizer

	// State
	lastError string
	statusMsg string
	quitting  bool

	// Component models
	dashboard   components.DashboardModel
	help        components.HelpModel
	logs        components.LogsModel
	configView  components.ConfigModel
	fileBrowser components.FileBrowserModel

	browsingFieldIndex int // index of the config field being updated
}

func NewModel(cfg *config.Config, configPath string, org *organizer.Organizer) Model {
	home, _ := os.UserHomeDir()
	return Model{
		activeView: ViewDashboard,
		cfg:        cfg,
		configPath: configPath,
		org:        org,
		dashboard: components.DashboardModel{
			Cfg: cfg,
		},
		logs:        components.NewLogsModel(80, 20),
		configView:  components.NewConfigModel(cfg, configPath, 80, 20),
		fileBrowser: components.NewFileBrowserModel(home, 80, 20),
	}
}
