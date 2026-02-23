package components

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	title, desc string
	path        string
	isDir       bool
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
func (i item) GetPath() string     { return i.path }
func (i item) IsDir() bool         { return i.isDir }

type FileBrowserModel struct {
	List             list.Model
	Path             string
	Width            int
	Height           int
	SelectMode       bool                 // If true, user can "Select" the current item/directory
	SelectionHandler func(string) tea.Cmd // Function to call when a selection is made
}

func NewFileBrowserModel(initialPath string, width, height int) FileBrowserModel {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = "File Browser: " + initialPath

	m := FileBrowserModel{
		List:   l,
		Path:   initialPath,
		Width:  width,
		Height: height,
	}
	m.Refresh()
	return m
}

func (m *FileBrowserModel) Refresh() {
	entries, err := os.ReadDir(m.Path)
	if err != nil {
		m.List.Title = fmt.Sprintf("Error: %v", err)
		return
	}

	var items []list.Item
	// Add ".." entry if not at root
	if m.Path != "/" && m.Path != "C:\\" {
		items = append(items, item{title: "..", desc: "Go up one level", path: filepath.Dir(m.Path), isDir: true})
	}

	for _, entry := range entries {
		info, _ := entry.Info()
		desc := fmt.Sprintf("%d bytes", info.Size())
		title := entry.Name()
		if entry.IsDir() {
			title = "📁 " + title
			desc = "Directory"
		} else {
			title = "📄 " + title
		}
		items = append(items, item{
			title: title,
			desc:  desc,
			path:  filepath.Join(m.Path, entry.Name()),
			isDir: entry.IsDir(),
		})
	}
	m.List.SetItems(items)
	m.List.Title = "File Browser: " + m.Path
}

func (m FileBrowserModel) View() string {
	view := m.List.View()
	if m.SelectMode {
		footer := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true).
			Render("\n  Press 's' to select current directory: " + m.Path)
		return view + footer
	}
	return view
}

// In components/filebrowser.go, let's also update the List logic to handle 's' in SelectMode.
// But wait, the list usually handles its own updates. m.List.Update(msg).
// We should check if 's' was pressed BEFORE passing it to the list.
func (m FileBrowserModel) Update(msg tea.Msg) (FileBrowserModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.SelectMode && msg.String() == "s" {
			return m, func() tea.Msg {
				// We need the fieldIndex here. The Model knows it.
				// For now, let's just return the path.
				// The main model will attach the fieldIndex.
				return SelectPathMsg{Path: m.Path}
			}
		}
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}
