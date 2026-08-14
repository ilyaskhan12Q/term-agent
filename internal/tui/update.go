package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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

// Update handles incoming events, keyboard input, and state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
			val := m.Prompt.Value()
			if val != "" {
				m.Prompt.Reset()
				m.StatusMsg = fmt.Sprintf("Processing prompt: %s...", val)
				m.IsBusy = true

				m.AgentView.AddLog(views.AgentLogItem{
					Timestamp: time.Now().Format("15:04:05"),
					AgentID:   "user",
					Type:      "answer",
					Content:   val,
				})

				// If input contains research keywords, automatically jump to Research View
				if len(val) > 5 {
					m.AgentView.AddLog(views.AgentLogItem{
						Timestamp: time.Now().Format("15:04:05"),
						AgentID:   "orchestrator",
						Type:      "thought",
						Content:   "Analyzing request and building task DAG plan...",
					})
				}
			}

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
	}

	// Update active prompt component
	newPrompt, promptCmd := m.Prompt.Update(msg)
	m.Prompt = newPrompt
	cmds = append(cmds, promptCmd)

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
