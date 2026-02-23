package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evrenesat/janny/internal/config"
)

type ConfigField struct {
	Label    string
	Input    string // For selection fields, we store the current selection here
	TI       textinput.Model
	Key      string // identifier
	IsPath   bool   // if true, browsing is enabled
	IsSelect bool   // if true, it's a dropdown/select
	Options  []string
	Selected int
}

type ConfigModel struct {
	Cfg        *config.Config
	ConfigPath string
	Width      int
	Height     int
	Fields     []ConfigField
	Cursor     int // which field is selected
	Ready      bool
	Viewport   viewport.Model // kept for potential large views
}

func NewConfigModel(cfg *config.Config, path string, width, height int) ConfigModel {
	m := ConfigModel{
		Cfg:        cfg,
		ConfigPath: path,
		Width:      width,
		Height:     height,
	}

	// 1. Source Paths
	for i, p := range cfg.General.SourcePaths {
		ti := textinput.New()
		ti.SetValue(p)
		m.Fields = append(m.Fields, ConfigField{
			Label:  fmt.Sprintf("Source %d", i+1),
			TI:     ti,
			Key:    fmt.Sprintf("source:%d", i),
			IsPath: true,
		})
	}

	// 2. Backup Destination
	ti := textinput.New()
	ti.SetValue(cfg.Backup.Destination)
	m.Fields = append(m.Fields, ConfigField{
		Label:  "Backup Dest",
		TI:     ti,
		Key:    "backup_dest",
		IsPath: true,
	})

	// 3. Default Storage
	ti = textinput.New()
	ti.SetValue(cfg.Advanced.DefaultStoragePath)
	m.Fields = append(m.Fields, ConfigField{
		Label:  "Default Storage",
		TI:     ti,
		Key:    "default_storage",
		IsPath: true,
	})

	// 4. Rules (Mapping) - Dropdown/Cycle selection
	// For each rule, we show Category (Select) and Pattern (Input)
	// For now, let's just show one example for clarity
	var categories []string
	for k := range cfg.Storage {
		categories = append(categories, k)
	}

	for k, v := range cfg.Rules {
		// Category (Select)
		m.Fields = append(m.Fields, ConfigField{
			Label:    fmt.Sprintf("Rule %s Category", k),
			Key:      fmt.Sprintf("rule_cat:%s", k),
			IsSelect: true,
			Options:  categories,
			Input:    k,
		})
		// Pattern (Input)
		ti := textinput.New()
		ti.SetValue(v)
		m.Fields = append(m.Fields, ConfigField{
			Label: fmt.Sprintf("Rule %s Patterns", k),
			TI:    ti,
			Key:   fmt.Sprintf("rule_pat:%s", k),
		})
	}

	if len(m.Fields) > 0 {
		if !m.Fields[0].IsSelect {
			m.Fields[0].TI.Focus()
		}
	}

	return m
}

func (m *ConfigModel) SetValue(index int, value string) {
	if index >= 0 && index < len(m.Fields) {
		if m.Fields[index].IsSelect {
			m.Fields[index].Input = value
		} else {
			m.Fields[index].TI.SetValue(value)
		}
	}
}

func (m ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if !m.Fields[m.Cursor].IsSelect {
				m.Fields[m.Cursor].TI.Blur()
			}
			m.Cursor = (m.Cursor + 1) % len(m.Fields)
			if !m.Fields[m.Cursor].IsSelect {
				m.Fields[m.Cursor].TI.Focus()
			}
		case "shift+tab":
			if !m.Fields[m.Cursor].IsSelect {
				m.Fields[m.Cursor].TI.Blur()
			}
			m.Cursor = (m.Cursor - 1 + len(m.Fields)) % len(m.Fields)
			if !m.Fields[m.Cursor].IsSelect {
				m.Fields[m.Cursor].TI.Focus()
			}
		case " ":
			if m.Fields[m.Cursor].IsSelect {
				m.Fields[m.Cursor].Selected = (m.Fields[m.Cursor].Selected + 1) % len(m.Fields[m.Cursor].Options)
				m.Fields[m.Cursor].Input = m.Fields[m.Cursor].Options[m.Fields[m.Cursor].Selected]
			}
		case "ctrl+b":
			if m.Fields[m.Cursor].IsPath {
				return m, func() tea.Msg {
					return BrowseMsg{FieldIndex: m.Cursor}
				}
			}
		}
	}

	if !m.Fields[m.Cursor].IsSelect {
		var cmd tea.Cmd
		m.Fields[m.Cursor].TI, cmd = m.Fields[m.Cursor].TI.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ConfigModel) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true).Render(" --- CONFIGURATION FORM --- ") + "\n\n")

	for i, field := range m.Fields {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.Cursor {
			cursor = "> "
			style = style.Foreground(lipgloss.Color("#FFD700")).Bold(true)
		}

		if field.IsSelect {
			b.WriteString(fmt.Sprintf("%s%s: [ %s ] (Space to cycle)\n", cursor, style.Render(field.Label), field.Input))
		} else {
			b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, style.Render(field.Label), field.TI.View()))
		}
	}

	b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(" Tab/Shift+Tab: switch • Enter: edit • Ctrl+S: save"))

	return b.String()
}
