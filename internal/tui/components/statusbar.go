package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xoai/sage-wiki/internal/tui"
)

// KeyHint represents a key binding hint for the status bar.
type KeyHint struct {
	Key  string
	Help string
}

// StatusBar renders context-sensitive key hints at the bottom.
type StatusBar struct {
	hints []KeyHint
	width int
	info  string // optional right-aligned info text
}

// NewStatusBar creates a status bar.
func NewStatusBar(width int) StatusBar {
	return StatusBar{width: width}
}

// SetHints updates the displayed key hints.
func (s *StatusBar) SetHints(hints []KeyHint) {
	s.hints = hints
}

// SetInfo sets optional right-aligned info text.
func (s *StatusBar) SetInfo(info string) {
	s.info = info
}

// SetWidth updates the bar width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// View renders the status bar.
func (s StatusBar) View() string {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Accent)

	helpStyle := lipgloss.NewStyle().
		Foreground(tui.DimColor)

	var parts []string
	for _, h := range s.hints {
		parts = append(parts, keyStyle.Render(h.Key)+" "+helpStyle.Render(h.Help))
	}

	left := strings.Join(parts, "  ")
	right := helpStyle.Render(s.info)

	gap := s.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return tui.StatusBarStyle.Width(s.width).Render(
		left + strings.Repeat(" ", gap) + right,
	)
}
