package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/xoai/sage-wiki/internal/tui"
)

// StreamView is a viewport that appends content incrementally during streaming
// and renders the final content via glamour when done.
type StreamView struct {
	viewport  viewport.Model
	raw       string // accumulated raw text
	streaming bool
	width     int
	height    int
	focused   bool
	title     string
}

// NewStreamView creates a new streaming viewport.
func NewStreamView(width, height int) StreamView {
	vp := viewport.New(width-2, height-2)
	return StreamView{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// AppendToken adds a token during streaming.
func (s *StreamView) AppendToken(token string) {
	s.raw += token
	s.streaming = true
	s.viewport.SetContent(s.raw)
	s.viewport.GotoBottom()
}

// Finish renders the full content via glamour.
func (s *StreamView) Finish() {
	s.streaming = false
	rendered, err := glamour.Render(s.raw, "dark")
	if err != nil {
		rendered = s.raw
	}
	s.viewport.SetContent(rendered)
	s.viewport.GotoTop()
}

// Clear resets the content.
func (s *StreamView) Clear() {
	s.raw = ""
	s.streaming = false
	s.viewport.SetContent("")
}

// Raw returns the accumulated raw text.
func (s *StreamView) Raw() string {
	return s.raw
}

// SetSize updates dimensions.
func (s *StreamView) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.Width = width - 2
	s.viewport.Height = height - 2
}

// SetFocused sets focus state.
func (s *StreamView) SetFocused(focused bool) {
	s.focused = focused
}

// SetTitle sets the component title.
func (s *StreamView) SetTitle(title string) {
	s.title = title
}

// Update handles messages.
func (s *StreamView) Update(msg tea.Msg) (*StreamView, tea.Cmd) {
	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

// View renders the stream view.
func (s StreamView) View() string {
	border := tui.BorderStyle
	if s.focused {
		border = tui.ActiveBorderStyle
	}

	title := s.title
	if title == "" {
		title = "Output"
	}
	if s.streaming {
		title += " (streaming...)"
	}

	header := tui.TitleStyle.Render(title)

	content := border.
		Width(s.width - 2).
		Height(s.height - 2).
		Render(s.viewport.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}
