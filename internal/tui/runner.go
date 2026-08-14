package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilyaskhan/term-agent/internal/config"
)

// Run launches the interactive Bubble Tea terminal user interface.
func Run(cfg *config.Config) error {
	m := NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
