package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// MessageKind defines the type of message rendered in the conversation stream.
type MessageKind string

const (
	KindUser         MessageKind = "USER"
	KindAssistant    MessageKind = "ASSISTANT"
	KindSystem       MessageKind = "SYSTEM"
	KindTool         MessageKind = "TOOL"
	KindResearchTree MessageKind = "RESEARCH_TREE"
)

// ConversationMessage represents a single item in the OpenCode linear chat viewport.
type ConversationMessage struct {
	Kind      string
	Author    string
	Timestamp string
	Content   string
	// Tool execution specifics
	ToolName   string
	ToolArgs   string
	ToolStatus string
	ToolOutput string
	// Research activity tree specifics
	ResearchTopic string
	SubAgentTasks map[string]string // taskName -> status ("complete", "running", "waiting")
}

// ConversationView implements OpenCode's primary linear conversation viewport.
type ConversationView struct {
	styles   *styles.Styles
	viewport viewport.Model
	messages []ConversationMessage
	width    int
	height   int
	ready    bool
}

// NewConversationView constructs the unified OpenCode conversation view.
func NewConversationView(s *styles.Styles) *ConversationView {
	vp := viewport.New(120, 25)
	vp.SetContent("Welcome to term-agent (OpenCode UX Mode)\nType a prompt or run a slash command like /research, /plan, /model, /help")

	cv := &ConversationView{
		styles:   s,
		viewport: vp,
		messages: make([]ConversationMessage, 0),
		width:    120,
		height:   25,
		ready:    true,
	}

	// Add initial welcome messages
	cv.AddSystemMessage("⚡ TERM-AGENT Research & Agent Terminal initialized.")
	cv.AddSystemMessage("Type /research <topic> to start autonomous research or /help for command registry.")

	return cv
}

// SetSize updates the viewport dimensions on terminal resize.
func (cv *ConversationView) SetSize(w, h int) {
	cv.width = w
	cv.height = h
	cv.viewport.Width = w - 2
	cv.viewport.Height = h - 2
	cv.rebuildViewport()
}

// AddUserMessage appends a user prompt message to the stream.
func (cv *ConversationView) AddUserMessage(content string) {
	cv.messages = append(cv.messages, ConversationMessage{
		Kind:    string(KindUser),
		Author:  "User",
		Content: content,
	})
	cv.rebuildViewport()
	cv.viewport.GotoBottom()
}

// AddAssistantMessage appends an assistant text response to the stream.
func (cv *ConversationView) AddAssistantMessage(content string) {
	cv.messages = append(cv.messages, ConversationMessage{
		Kind:    string(KindAssistant),
		Author:  "Assistant",
		Content: content,
	})
	cv.rebuildViewport()
	cv.viewport.GotoBottom()
}

// AddSystemMessage appends system alerts or command outputs to the stream.
func (cv *ConversationView) AddSystemMessage(content string) {
	cv.messages = append(cv.messages, ConversationMessage{
		Kind:    string(KindSystem),
		Author:  "System",
		Content: content,
	})
	cv.rebuildViewport()
	cv.viewport.GotoBottom()
}

// AddToolMessage appends inline tool execution blocks to the stream.
func (cv *ConversationView) AddToolMessage(name, args, status, output string) {
	cv.messages = append(cv.messages, ConversationMessage{
		Kind:       string(KindTool),
		Author:     "Tool",
		ToolName:   name,
		ToolArgs:   args,
		ToolStatus: status,
		ToolOutput: output,
	})
	cv.rebuildViewport()
	cv.viewport.GotoBottom()
}

// AddResearchTreeMessage appends an inline research sub-agent activity tree to the stream.
func (cv *ConversationView) AddResearchTreeMessage(topic string, tasks map[string]string) {
	cv.messages = append(cv.messages, ConversationMessage{
		Kind:          string(KindResearchTree),
		Author:        "ResearchAgent",
		ResearchTopic: topic,
		SubAgentTasks: tasks,
	})
	cv.rebuildViewport()
	cv.viewport.GotoBottom()
}

// Clear resets all messages in the conversation view.
func (cv *ConversationView) Clear() {
	cv.messages = nil
	cv.AddSystemMessage("Conversation cleared.")
}

// Update handles scroll events in the viewport.
func (cv *ConversationView) Update(msg tea.Msg) (*ConversationView, tea.Cmd) {
	var cmd tea.Cmd
	cv.viewport, cmd = cv.viewport.Update(msg)
	return cv, cmd
}

// rebuildViewport formats all messages into a unified linear text buffer for rendering.
func (cv *ConversationView) rebuildViewport() {
	var b strings.Builder

	for i, msg := range cv.messages {
		if i > 0 {
			b.WriteString("\n\n")
		}

		switch MessageKind(msg.Kind) {
		case KindUser:
			promptLabel := cv.styles.PromptPrefix.Render("❯ ")
			userText := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(msg.Content)
			b.WriteString(promptLabel + userText)

		case KindAssistant:
			header := cv.styles.BadgeActive.Render("Assistant")
			b.WriteString(header + "\n" + msg.Content)

		case KindSystem:
			sysStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8A8A8A"))
			b.WriteString(sysStyle.Render("── " + msg.Content))

		case KindTool:
			boxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4A4A4A")).
				Padding(0, 1)

			toolTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cv.styles.Tokens.SecondaryColor)).Render("🔧 Tool: " + msg.ToolName)
			statusBadge := cv.styles.Badge.Render(msg.ToolStatus)
			argsLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render("Args: " + msg.ToolArgs)

			toolBody := toolTitle + " " + statusBadge + "\n" + argsLine
			if msg.ToolOutput != "" {
				toolBody += "\n" + cv.styles.DiffContext.Render("Output: "+msg.ToolOutput)
			}
			b.WriteString(boxStyle.Render(toolBody))

		case KindResearchTree:
			b.WriteString(cv.renderResearchTree(msg.ResearchTopic, msg.SubAgentTasks))

		default:
			b.WriteString(msg.Content)
		}
	}

	cv.viewport.SetContent(b.String())
}

// renderResearchTree formats a clean OpenCode inline research activity tree.
func (cv *ConversationView) renderResearchTree(topic string, tasks map[string]string) string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cv.styles.Tokens.PrimaryColor)).Render(fmt.Sprintf("Researching: %q", topic))
	b.WriteString(header + "\n")

	if len(tasks) == 0 {
		tasks = map[string]string{
			"Literature Review":      "complete",
			"Methodology Analysis":   "running",
			"Evidence Verification":  "waiting",
			"Document Synthesis":     "waiting",
		}
	}

	for task, status := range tasks {
		var icon string
		var statusStyle lipgloss.Style

		switch strings.ToLower(status) {
		case "complete", "completed", "done", "ok":
			icon = "✓"
			statusStyle = cv.styles.TaskComplete
		case "running", "executing", "in_progress":
			icon = "●"
			statusStyle = cv.styles.TaskRunning
		default:
			icon = "⟳"
			statusStyle = cv.styles.TaskPending
		}

		taskLabel := fmt.Sprintf("  - %-25s", task)
		badge := statusStyle.Render(fmt.Sprintf("%s %s", icon, status))
		b.WriteString(taskLabel + badge + "\n")
	}

	return b.String()
}

// Render draws the conversation view inside a bordered viewport container.
func (cv *ConversationView) Render() string {
	boxStyle := lipgloss.NewStyle().
		Width(cv.width).
		Height(cv.height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(cv.styles.Tokens.BorderColor)).
		Padding(0, 1)

	return boxStyle.Render(cv.viewport.View())
}
