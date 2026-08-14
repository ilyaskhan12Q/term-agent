package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// StatusBar Component renders bottom telemetry & shortcut bar.
type StatusBar struct {
	styles *styles.Styles
	width  int
}

// NewStatusBar initializes StatusBar.
func NewStatusBar(s *styles.Styles) *StatusBar {
	return &StatusBar{styles: s}
}

// SetWidth updates layout width.
func (sb *StatusBar) SetWidth(w int) {
	sb.width = w
}

// Render draws the bottom status bar with provider, model, context usage, state, and quick help.
func (sb *StatusBar) Render(provider, model string, isBusy bool, statusMsg string, contextUsed, contextMax int) string {
	stateStr := sb.styles.StatusSuccess.Render("● READY")
	if isBusy {
		stateStr = sb.styles.StatusBusy.Render("◐ RUNNING...")
	}

	provStr := sb.styles.StatusItem.Render("Provider:") + sb.styles.StatusVal.Render(provider)
	modelStr := sb.styles.StatusItem.Render("Model:") + sb.styles.StatusVal.Render(model)

	ctxPct := 0
	if contextMax > 0 {
		ctxPct = (contextUsed * 100) / contextMax
	}
	ctxStr := sb.styles.StatusItem.Render("Context:") + sb.styles.StatusVal.Render(fmt.Sprintf("%d/%d tokens (%d%%)", contextUsed, contextMax, ctxPct))

	left := lipgloss.JoinHorizontal(lipgloss.Center, stateStr, " | ", provStr, " | ", modelStr, " | ", ctxStr)

	help := sb.styles.StatusItem.Render("Tab: Switch | Enter: Send | q: Quit")

	if statusMsg != "" {
		msgStr := sb.styles.StatusItem.Render("[" + statusMsg + "]")
		left = lipgloss.JoinHorizontal(lipgloss.Center, left, " ", msgStr)
	}

	gapWidth := sb.width - lipgloss.Width(left) - lipgloss.Width(help)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := strings.Repeat(" ", gapWidth)

	return sb.styles.StatusBar.Render(lipgloss.JoinHorizontal(lipgloss.Center, left, gap, help))
}
