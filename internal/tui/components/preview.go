package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/xoai/sage-wiki/internal/tui"
)

// Preview renders markdown content in a scrollable viewport.
type Preview struct {
	viewport viewport.Model
	title    string
	width    int
	height   int
	focused  bool
}

// NewPreview creates a new preview component.
func NewPreview(width, height int) Preview {
	vp := viewport.New(width-2, height-2) // account for border
	return Preview{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// SetContent renders markdown and displays it.
func (p *Preview) SetContent(title string, markdown string) {
	p.title = title
	rendered, err := glamour.Render(markdown, "dark")
	if err != nil {
		rendered = markdown
	}
	p.viewport.SetContent(rendered)
	p.viewport.GotoTop()
}

// SetSize updates the viewport dimensions.
func (p *Preview) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 2
	p.viewport.Height = height - 2
}

// SetFocused sets focus state for border styling.
func (p *Preview) SetFocused(focused bool) {
	p.focused = focused
}

// Update handles messages.
func (p *Preview) Update(msg tea.Msg) (*Preview, tea.Cmd) {
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return p, cmd
}

// View renders the preview.
func (p Preview) View() string {
	border := tui.BorderStyle
	if p.focused {
		border = tui.ActiveBorderStyle
	}

	title := tui.TitleStyle.Render(p.title)
	if p.title == "" {
		title = tui.DimStyle.Render("Preview")
	}

	content := border.
		Width(p.width - 2).
		Height(p.height - 2).
		Render(p.viewport.View())

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}
