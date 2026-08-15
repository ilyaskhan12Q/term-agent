package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilyaskhan/term-agent/internal/commands"
	"github.com/ilyaskhan/term-agent/internal/tui/views"
)

// Msg represents custom events sent to the TUI update loop.
type Msg interface{}

// AgentStateMsg represents a status update event from an agent.
type AgentStateMsg struct {
	AgentID string
	Status  string
	Thought string
}

// DiffUpdateMsg represents a diff update ready for review.
type DiffUpdateMsg struct {
	TransactionID string
	FilePath      string
	UnifiedDiff   string
}

// EventLogMsg represents system event log event.
type EventLogMsg struct {
	Timestamp string
	Level     string
	Component string
	Message   string
}

// PromptSubmitMsg represents user submitting input from prompt box.
type PromptSubmitMsg struct {
	Text string
}

// CommandResultMsg delivers a slash-command execution result to the update loop.
type CommandResultMsg struct {
	Result commands.CommandResult
}

// Update handles incoming events, keyboard input, and state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var isSubmitKey bool

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		m.Header.SetWidth(m.Width)
		m.StatusBar.SetWidth(m.Width)
		m.Prompt.SetWidth(m.Width)

		bodyH := m.Height - 8
		if bodyH < 5 {
			bodyH = 5
		}

		m.AgentView.SetSize(m.Width, bodyH)
		m.ConversationView.SetSize(m.Width, bodyH)
		m.PlanView.SetSize(m.Width, bodyH)
		m.DiffView.SetSize(m.Width, bodyH)
		m.LogView.SetSize(m.Width, bodyH)
		m.ResearchView.SetSize(m.Width, bodyH)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.KeyMap.TabNext):
			m.rotateTab(1)
		case key.Matches(msg, m.KeyMap.TabPrev):
			m.rotateTab(-1)

		case key.Matches(msg, m.KeyMap.SelectTab1):
			m.ActiveView = ViewAgentView
		case key.Matches(msg, m.KeyMap.SelectTab2):
			m.ActiveView = ViewPlanView
		case key.Matches(msg, m.KeyMap.SelectTab3):
			m.ActiveView = ViewDiffView
		case key.Matches(msg, m.KeyMap.SelectTab4):
			m.ActiveView = ViewLogView
		case key.Matches(msg, m.KeyMap.SelectTab5):
			m.ActiveView = ViewResearchView

		case key.Matches(msg, m.KeyMap.Submit):
			isSubmitKey = true
			val := m.Prompt.Value()
			m.Prompt.Reset()
			if val != "" {
				// Add user message to conversation view and legacy agent log
				m.ConversationView.AddUserMessage(val)
				m.AgentView.AddLog(views.AgentLogItem{
					Timestamp: time.Now().Format("15:04:05"),
					AgentID:   "user",
					Type:      "answer",
					Content:   val,
				})

				parsed := commands.Parse(val)
				if parsed.IsCommand {
					// /help is intercepted here so the registry can provide its own output.
					var result commands.CommandResult
					if parsed.Name == "help" || parsed.Name == "h" || parsed.Name == "?" {
						result = commands.CommandResult{Output: m.CmdRegistry.HelpText()}
					} else {
						result = m.CmdRegistry.Dispatch(parsed.Name, parsed.Args)
					}

					if parsed.Name == "research" && m.ResearchState != nil && m.ResearchState.Topic != "" {
						m.ConversationView.AddResearchTreeMessage(m.ResearchState.Topic, m.ResearchState.SubAgents)
					}

					cmds = append(cmds, func() tea.Msg { return CommandResultMsg{Result: result} })
				} else {
					// Plain text input: route to orchestrator.
					m.StatusMsg = fmt.Sprintf("Processing: %s...", truncate(val, 40))
					m.IsBusy = true
					m.ConversationView.AddAssistantMessage("Analyzing request and building task execution plan...")
					m.ConversationView.AddToolMessage("orchestrator", fmt.Sprintf("query=%q", val), "COMPLETED", "Task DAG generated with sub-agent steps.")
					if len(val) > 5 {
						m.AgentView.AddLog(views.AgentLogItem{
							Timestamp: time.Now().Format("15:04:05"),
							AgentID:   "orchestrator",
							Type:      "thought",
							Content:   "Analyzing request and building task DAG plan...",
						})
					}
				}
			}

		case key.Matches(msg, m.KeyMap.Clear):
			m.ConversationView.Clear()
			m.ResearchView.Clear()

		case key.Matches(msg, m.KeyMap.ApproveDiff):
			if m.ActiveView == ViewDiffView && m.DiffView.HasPending {
				m.DiffView.HasPending = false
				m.StatusMsg = fmt.Sprintf("Transaction %s APPROVED and applied cleanly", m.DiffView.TransactionID)
				m.LogView.AddLog(views.LogEventItem{
					Timestamp: time.Now().Format("15:04:05"),
					Level:     "INFO",
					Component: "diff_engine",
					Message:   fmt.Sprintf("Approved transaction %s for %s", m.DiffView.TransactionID, m.DiffView.FilePath),
				})
			}

		case key.Matches(msg, m.KeyMap.RejectDiff):
			if m.ActiveView == ViewDiffView && m.DiffView.HasPending {
				m.DiffView.HasPending = false
				m.StatusMsg = fmt.Sprintf("Transaction %s REJECTED and discarded", m.DiffView.TransactionID)
				m.LogView.AddLog(views.LogEventItem{
					Timestamp: time.Now().Format("15:04:05"),
					Level:     "WARN",
					Component: "diff_engine",
					Message:   fmt.Sprintf("Rejected transaction %s for %s", m.DiffView.TransactionID, m.DiffView.FilePath),
				})
			}
		}

	case AgentStateMsg:
		m.AgentView.AddLog(views.AgentLogItem{
			Timestamp: time.Now().Format("15:04:05"),
			AgentID:   msg.AgentID,
			Type:      "thought",
			Content:   msg.Thought,
		})
		m.ConversationView.AddAssistantMessage(fmt.Sprintf("[%s] %s", msg.AgentID, msg.Thought))

	case DiffUpdateMsg:
		m.DiffView.SetDiff(msg.TransactionID, msg.FilePath, msg.UnifiedDiff)
		m.ActiveView = ViewDiffView
		m.StatusMsg = "Pending transaction diff requires approval"

	case EventLogMsg:
		m.LogView.AddLog(views.LogEventItem{
			Timestamp: msg.Timestamp,
			Level:     msg.Level,
			Component: msg.Component,
			Message:   msg.Message,
		})

	case CommandResultMsg:
		result := msg.Result
		if result.Output == "__CLEAR__" {
			m.ConversationView.Clear()
			m.ResearchView.Clear()
		} else if result.Output != "" {
			m.ConversationView.AddSystemMessage(result.Output)
			m.ResearchView.AddLog(result.Output)
		}
		if result.SwitchView != "" {
			switch result.SwitchView {
			case "RESEARCH_VIEW":
				m.ActiveView = ViewResearchView
			case "PLAN_VIEW":
				m.ActiveView = ViewPlanView
			case "AGENT_VIEW", "CONVERSATION_VIEW":
				m.ActiveView = ViewAgentView
			}
		}
		if result.UpdateStatus != "" {
			m.StatusMsg = result.UpdateStatus
		}
		if result.Quit {
			return m, tea.Quit
		}
	}

	// Update active prompt component (skip if processing Submit keypress to avoid newline injection)
	if !isSubmitKey {
		newPrompt, promptCmd := m.Prompt.Update(msg)
		m.Prompt = newPrompt
		cmds = append(cmds, promptCmd)
	}

	// Update conversation view viewport (scroll navigation)
	newConvView, convCmd := m.ConversationView.Update(msg)
	m.ConversationView = newConvView
	cmds = append(cmds, convCmd)

	// Update spinner
	newSpinner, spinnerCmd := m.Spinner.Update(msg)
	m.Spinner = newSpinner
	cmds = append(cmds, spinnerCmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) rotateTab(delta int) {
	tabList := []ViewMode{ViewAgentView, ViewPlanView, ViewDiffView, ViewLogView, ViewResearchView}
	currIdx := 0
	for i, t := range tabList {
		if t == m.ActiveView {
			currIdx = i
			break
		}
	}
	newIdx := (currIdx + delta + len(tabList)) % len(tabList)
	m.ActiveView = tabList[newIdx]
}

// truncate shortens s to at most maxLen runes, appending "..." if trimmed.
func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}
