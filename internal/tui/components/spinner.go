package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// Spinner Component wraps bubbles spinner animation.
type Spinner struct {
	model  spinner.Model
	styles *styles.Styles
}

// NewSpinner constructs a styled Spinner.
func NewSpinner(s *styles.Styles) *Spinner {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(s.Tokens.PrimaryColor))
	return &Spinner{
		model:  sp,
		styles: s,
	}
}

// Init starts spinner tick.
func (sp *Spinner) Init() tea.Cmd {
	return sp.model.Tick
}

// Update handles tick messages.
func (sp *Spinner) Update(msg tea.Msg) (*Spinner, tea.Cmd) {
	var cmd tea.Cmd
	sp.model, cmd = sp.model.Update(msg)
	return sp, cmd
}

// View renders spinner frame string.
func (sp *Spinner) View() string {
	return sp.model.View()
}
