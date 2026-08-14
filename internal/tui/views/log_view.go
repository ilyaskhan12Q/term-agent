package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// LogEventItem represents an event bus log item.
type LogEventItem struct {
	Timestamp string
	Level     string // "INFO", "WARN", "DEBUG", "ERROR"
	Component string
	Message   string
}

// LogView displays system event logs.
type LogView struct {
	styles *styles.Styles
	width  int
	height int
	logs   []LogEventItem
}

// NewLogView creates LogView instance.
func NewLogView(s *styles.Styles) *LogView {
	return &LogView{
		styles: s,
		logs: []LogEventItem{
			{Timestamp: "00:00:00", Level: "INFO", Component: "app", Message: "Term-agent bootstrap completed in 12ms"},
			{Timestamp: "00:00:01", Level: "INFO", Component: "persistence", Message: "SQLite database initialized with migrations applied"},
			{Timestamp: "00:00:01", Level: "INFO", Component: "eventbus", Message: "InMemoryEventBus subscription handlers active"},
			{Timestamp: "00:00:02", Level: "DEBUG", Component: "workflow", Message: "ResearchWorkflow registered in WorkflowRegistry"},
		},
	}
}

// SetSize updates dimensions.
func (v *LogView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// AddLog appends a new event log item.
func (v *LogView) AddLog(item LogEventItem) {
	v.logs = append(v.logs, item)
}

// Render draws the event logs view.
func (v *LogView) Render() string {
	var b strings.Builder
	b.WriteString(v.styles.PanelTitle.Render("📋 System Event & Diagnostic Logs"))
	b.WriteString("\n\n")

	maxItems := v.height - 6
	if maxItems <= 0 {
		maxItems = 10
	}

	startIdx := 0
	if len(v.logs) > maxItems {
		startIdx = len(v.logs) - maxItems
	}

	for _, item := range v.logs[startIdx:] {
		lvlStyle := v.styles.Badge
		switch item.Level {
		case "ERROR":
			lvlStyle = v.styles.TaskFailed
		case "WARN":
			lvlStyle = v.styles.TaskRunning
		case "INFO":
			lvlStyle = v.styles.TaskComplete
		case "DEBUG":
			lvlStyle = v.styles.DiffContext
		}

		timeTag := v.styles.DiffContext.Render("[" + item.Timestamp + "]")
		compTag := v.styles.Badge.Render(item.Component)
		lvlTag := lvlStyle.Render(fmt.Sprintf("[%s]", item.Level))

		b.WriteString(fmt.Sprintf("%s %s %s %s\n", timeTag, lvlTag, compTag, item.Message))
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
