package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// AgentLogItem represents a single step / thought / tool action in agent timeline.
type AgentLogItem struct {
	Timestamp string
	AgentID   string
	Type      string // "thought", "tool", "answer", "error"
	Content   string
}

// AgentView displays real-time agent output feed.
type AgentView struct {
	styles *styles.Styles
	width  int
	height int
	logs   []AgentLogItem
}

// NewAgentView creates AgentView instance.
func NewAgentView(s *styles.Styles) *AgentView {
	return &AgentView{
		styles: s,
		logs: []AgentLogItem{
			{
				Timestamp: "00:00:01",
				AgentID:   "system",
				Type:      "answer",
				Content:   "term-agent engine initialized and ready. Type a research topic or task below to begin.",
			},
		},
	}
}

// SetSize updates dimensions.
func (v *AgentView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// AddLog appends a new message item.
func (v *AgentView) AddLog(item AgentLogItem) {
	v.logs = append(v.logs, item)
}

// Clear resets the log feed.
func (v *AgentView) Clear() {
	v.logs = []AgentLogItem{}
}

// Render renders the agent activity view.
func (v *AgentView) Render() string {
	var b strings.Builder
	b.WriteString(v.styles.PanelTitle.Render("🤖 Agent Activity Stream"))
	b.WriteString("\n\n")

	// Render log history
	maxItems := v.height - 6
	if maxItems <= 0 {
		maxItems = 10
	}

	startIdx := 0
	if len(v.logs) > maxItems {
		startIdx = len(v.logs) - maxItems
	}

	for _, item := range v.logs[startIdx:] {
		timeTag := v.styles.DiffContext.Render("[" + item.Timestamp + "]")
		agentTag := v.styles.Badge.Render(item.AgentID)

		switch item.Type {
		case "thought":
			b.WriteString(fmt.Sprintf("%s %s %s\n", timeTag, agentTag, v.styles.ThoughtBox.Render(item.Content)))
		case "tool":
			b.WriteString(fmt.Sprintf("%s %s %s\n", timeTag, agentTag, v.styles.ToolBox.Render("🔧 "+item.Content)))
		case "answer":
			b.WriteString(fmt.Sprintf("%s %s %s\n", timeTag, agentTag, lipgloss.NewStyle().Foreground(lipgloss.Color(v.styles.Tokens.TextColor)).Render("💬 "+item.Content)))
		case "error":
			b.WriteString(fmt.Sprintf("%s %s %s\n", timeTag, agentTag, v.styles.TaskFailed.Render("❌ "+item.Content)))
		default:
			b.WriteString(fmt.Sprintf("%s %s %s\n", timeTag, agentTag, item.Content))
		}
		b.WriteString("\n")
	}

	content := b.String()

	panelStyle := lipgloss.NewStyle().
		Width(v.width - 4).
		Height(v.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(v.styles.Tokens.BorderColor)).
		Padding(1)

	return panelStyle.Render(content)
}
