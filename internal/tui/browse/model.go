package browse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/xoai/sage-wiki/internal/tui"
	"github.com/xoai/sage-wiki/internal/tui/components"
)

type fileEntry struct {
	name    string
	path    string
	section string
}

type viewMode int

const (
	modeTree viewMode = iota
	modeArticle
)

// articleLoadedMsg carries the rendered article content.
type articleLoadedMsg struct{ content string }

// Model is the Browse tab.
type Model struct {
	files     []fileEntry
	cursor    int
	mode      viewMode
	viewport  viewport.Model
	statusBar components.StatusBar
	width     int
	height    int

	projectDir string
	outputDir  string
}

// New creates a browse model.
func New(projectDir, outputDir string) Model {
	vp := viewport.New(80, 20)
	sb := components.NewStatusBar(80)

	m := Model{
		viewport:   vp,
		statusBar:  sb,
		projectDir: projectDir,
		outputDir:  outputDir,
	}
	m.scanFiles()
	m.updateHints()
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.width - 2
		m.viewport.Height = m.height - 4
		m.statusBar.SetWidth(m.width)

	case tea.KeyMsg:
		if m.mode == modeTree {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.files)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.files) > 0 {
					m.mode = modeArticle
					m.updateHints()
					return m, m.loadArticle()
				}
			}
		} else {
			switch msg.String() {
			case "backspace", "esc":
				m.mode = modeTree
				m.updateHints()
				return m, nil
			default:
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}

	case articleLoadedMsg:
		m.viewport.SetContent(msg.content)
		m.viewport.GotoTop()
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var content string
	if m.mode == modeTree {
		content = m.renderTree()
	} else {
		content = m.viewport.View()
	}

	border := tui.BorderStyle.
		Width(m.width - 2).
		Height(m.height - 4)

	return lipgloss.JoinVertical(lipgloss.Left,
		border.Render(content),
		m.statusBar.View(),
	)
}

// --- Helpers ---

func (m *Model) scanFiles() {
	absOutput := filepath.Join(m.projectDir, m.outputDir)
	m.files = nil

	for _, section := range []string{"concepts", "summaries", "outputs"} {
		dir := filepath.Join(absOutput, section)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			m.files = append(m.files, fileEntry{
				name:    strings.TrimSuffix(e.Name(), ".md"),
				path:    filepath.Join(section, e.Name()),
				section: section,
			})
		}
	}
}

// Refresh rescans files (called on tab switch).
func (m *Model) Refresh() {
	m.scanFiles()
}

func (m Model) renderTree() string {
	var b strings.Builder
	currentSection := ""
	visibleHeight := m.height - 6
	start := 0
	if m.cursor >= visibleHeight {
		start = m.cursor - visibleHeight + 1
	}

	for i := start; i < len(m.files) && i < start+visibleHeight; i++ {
		f := m.files[i]
		if f.section != currentSection {
			currentSection = f.section
			header := tui.DimStyle.Render(strings.ToUpper(currentSection))
			b.WriteString("\n" + header + "\n")
		}

		if i == m.cursor {
			b.WriteString(tui.SelectedStyle.Render("▸ "+f.name) + "\n")
		} else {
			b.WriteString("  " + f.name + "\n")
		}
	}

	if len(m.files) == 0 {
		b.WriteString(tui.DimStyle.Render("\n  No articles yet. Run: sage-wiki compile"))
	}

	return b.String()
}

func (m *Model) updateHints() {
	if m.mode == modeTree {
		m.statusBar.SetHints([]components.KeyHint{
			{Key: "↑↓", Help: "navigate"},
			{Key: "enter", Help: "open"},
		})
		m.statusBar.SetInfo(fmt.Sprintf("%d articles", len(m.files)))
	} else {
		m.statusBar.SetHints([]components.KeyHint{
			{Key: "↑↓", Help: "scroll"},
			{Key: "esc", Help: "back to list"},
		})
		m.statusBar.SetInfo("")
	}
}

func (m Model) loadArticle() tea.Cmd {
	if m.cursor >= len(m.files) {
		return nil
	}
	f := m.files[m.cursor]
	return func() tea.Msg {
		absPath := filepath.Join(m.projectDir, m.outputDir, f.path)
		data, err := os.ReadFile(absPath)
		if err != nil {
			return articleLoadedMsg{content: "Could not load: " + err.Error()}
		}

		content := string(data)
		if strings.HasPrefix(content, "---\n") {
			if end := strings.Index(content[4:], "\n---"); end >= 0 {
				content = strings.TrimSpace(content[4+end+4:])
			}
		}

		rendered, err := glamour.Render(content, "dark")
		if err != nil {
			rendered = content
		}
		return articleLoadedMsg{content: rendered}
	}
}
