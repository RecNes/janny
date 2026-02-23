package main

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#00FFFF") // Cyan
	SecondaryColor = lipgloss.Color("#FF00FF") // Magenta
	SuccessColor   = lipgloss.Color("#00FF00") // Green
	ErrorColor     = lipgloss.Color("#FF0000") // Red
	WarningColor   = lipgloss.Color("#FFA500") // Orange
	HighlightColor = lipgloss.Color("#7D56F4") // Purple-ish

	// Stays consistent with terminal theme but adds borders/padding
	BoxStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(HighlightColor)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(HighlightColor).
			Padding(0, 1).
			Bold(true)

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Padding(0, 1)

	TabStyle = lipgloss.NewStyle().
			Padding(0, 2)

	ActiveTabStyle = TabStyle.
			Foreground(PrimaryColor).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(PrimaryColor).
			Bold(true)

	InactiveTabStyle = TabStyle.
			Foreground(lipgloss.Color("#888888"))

	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			MarginBottom(1)

	KeywordStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Background(lipgloss.Color("#222222")).
			Bold(true).
			PaddingLeft(1)

	NormalStyle = lipgloss.NewStyle().
			PaddingLeft(1)

	TableHeadStyle = lipgloss.NewStyle().
			Foreground(HighlightColor).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false)
)
