package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// Header Component renders the top status/tab bar.
type Header struct {
	styles *styles.Styles
	width  int
}

// NewHeader initializes Header component.
func NewHeader(s *styles.Styles) *Header {
	return &Header{styles: s}
}

// SetWidth updates layout width.
func (h *Header) SetWidth(w int) {
	h.width = w
}

// Render draws the top header with active tab highlighting.
func (h *Header) Render(activeView string, sessionID string, workspace string) string {
	title := h.styles.HeaderTitle.Render("⚡ TERM-AGENT")

	tabs := []struct {
		ID   string
		Name string
	}{
		{"AGENT_VIEW", "1:Agent"},
		{"PLAN_VIEW", "2:Plan DAG"},
		{"DIFF_VIEW", "3:Diff Review"},
		{"LOG_VIEW", "4:Event Logs"},
		{"RESEARCH_VIEW", "5:Research Mode"},
	}

	var renderedTabs []string
	for _, t := range tabs {
		if t.ID == activeView {
			renderedTabs = append(renderedTabs, h.styles.ActiveTab.Render(t.Name))
		} else {
			renderedTabs = append(renderedTabs, h.styles.InactiveTab.Render(t.Name))
		}
	}

	tabRow := strings.Join(renderedTabs, h.styles.TabSeparator.Render(" | "))

	wsInfo := ""
	if workspace != "" {
		wsInfo = h.styles.Badge.Render(fmt.Sprintf("dir: %s", workspace))
	}
	if sessionID != "" {
		shortID := sessionID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		wsInfo = lipgloss.JoinHorizontal(lipgloss.Center, wsInfo, " ", h.styles.BadgeActive.Render(fmt.Sprintf("session: %s", shortID)))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", tabRow)

	gapWidth := h.width - lipgloss.Width(left) - lipgloss.Width(wsInfo)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := strings.Repeat(" ", gapWidth)

	return h.styles.HeaderBar.Render(lipgloss.JoinHorizontal(lipgloss.Center, left, gap, wsInfo))
}
