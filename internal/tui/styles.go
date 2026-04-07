package tui

import "github.com/charmbracelet/lipgloss"

// Fixed colors — avoids OSC terminal queries that inject escape sequences
// into text input on some terminals (WSL2, Windows Terminal).
var (
	Accent    = lipgloss.Color("#7C6AFF")
	Subtle    = lipgloss.Color("#5C5C5C")
	TextColor = lipgloss.Color("#FAFAFA")
	DimColor  = lipgloss.Color("#666666")
	Green     = lipgloss.Color("#25D198")
	Red       = lipgloss.Color("#FF6690")
	Yellow    = lipgloss.Color("#FFCC4D")
)

// Shared styles.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Accent).
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Subtle)

	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Accent)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(DimColor).
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	DimStyle = lipgloss.NewStyle().
			Foreground(DimColor)

	SuccessStyle = lipgloss.NewStyle().Foreground(Green)
	ErrorStyle   = lipgloss.NewStyle().Foreground(Red)
	WarnStyle    = lipgloss.NewStyle().Foreground(Yellow)
)
