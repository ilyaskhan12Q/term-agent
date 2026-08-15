package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/commands"
	research "github.com/ilyaskhan/term-agent/internal/commands/research"
	"github.com/ilyaskhan/term-agent/internal/config"
	"github.com/ilyaskhan/term-agent/internal/tui/components"
	"github.com/ilyaskhan/term-agent/internal/tui/keymap"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
	"github.com/ilyaskhan/term-agent/internal/tui/views"
)

// ViewMode represents the active primary tab / panel in the TUI.
type ViewMode string

const (
	ViewAgentView    ViewMode = "AGENT_VIEW"
	ViewPlanView     ViewMode = "PLAN_VIEW"
	ViewDiffView     ViewMode = "DIFF_VIEW"
	ViewLogView      ViewMode = "LOG_VIEW"
	ViewResearchView ViewMode = "RESEARCH_VIEW"
)

// Model represents the complete Bubble Tea application state model.
type Model struct {
	ActiveView   ViewMode
	StatusMsg    string
	IsBusy       bool
	Width        int
	Height       int
	Config       *config.Config
	Styles       *styles.Styles
	KeyMap       keymap.KeyMap
	Header       *components.Header
	StatusBar    *components.StatusBar
	Prompt       *components.Prompt
	Spinner      *components.Spinner
	AgentView        *views.AgentView
	ConversationView *views.ConversationView
	PlanView         *views.PlanView
	DiffView         *views.DiffView
	LogView          *views.LogView
	ResearchView     *views.ResearchView
	ContextUsed      int
	ContextMax       int
	// CmdRegistry dispatches slash commands entered in the prompt.
	CmdRegistry *commands.Registry
	// ResearchState holds mutable research session state shared across commands.
	ResearchState *research.ResearchState
}

// NewModel initializes the root TUI model with default components and sub-views.
func NewModel(cfg *config.Config) Model {
	s := styles.NewStyles()
	km := keymap.DefaultKeyMap()

	// Bootstrap command registry with all research commands.
	reg := commands.NewRegistry()
	rs := research.NewResearchState()
	research.RegisterAll(reg, rs)

	m := Model{
		ActiveView:       ViewAgentView,
		StatusMsg:        "Ready",
		IsBusy:           false,
		Width:            120,
		Height:           35,
		Config:           cfg,
		Styles:           s,
		KeyMap:           km,
		Header:           components.NewHeader(s),
		StatusBar:        components.NewStatusBar(s),
		Prompt:           components.NewPrompt(s),
		Spinner:          components.NewSpinner(s),
		AgentView:        views.NewAgentView(s),
		ConversationView: views.NewConversationView(s),
		PlanView:         views.NewPlanView(s),
		DiffView:         views.NewDiffView(s),
		LogView:          views.NewLogView(s),
		ResearchView:     views.NewResearchView(s),
		ContextUsed:      14200,
		ContextMax:       128000,
		CmdRegistry:      reg,
		ResearchState:    rs,
	}

	m.Header.SetWidth(m.Width)
	m.StatusBar.SetWidth(m.Width)
	m.Prompt.SetWidth(m.Width)

	m.AgentView.SetSize(m.Width, m.Height-8)
	m.ConversationView.SetSize(m.Width, m.Height-8)
	m.PlanView.SetSize(m.Width, m.Height-8)
	m.DiffView.SetSize(m.Width, m.Height-8)
	m.LogView.SetSize(m.Width, m.Height-8)
	m.ResearchView.SetSize(m.Width, m.Height-8)

	return m
}

// Init initializes background command generators (spinner tick, event subscriptions).
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Init(),
		m.Prompt.Focus(),
	)
}

// SetView switches active view tab.
func (m *Model) SetView(view ViewMode) {
	m.ActiveView = view
}

// View renders the application user interface.
func (m Model) View() string {
	sessID := ""
	workspace := ""
	provider := "openai"
	model := "gpt-4o"

	if m.Config != nil {
		sessID = m.Config.SessionID
		workspace = m.Config.WorkspaceDir
		provider = m.Config.DefaultProvider
		model = m.Config.DefaultModel
	}
	if m.ResearchState != nil {
		if m.ResearchState.Provider != "" {
			provider = m.ResearchState.Provider
		}
		if m.ResearchState.Model != "" {
			model = m.ResearchState.Model
		}
		if m.ResearchState.SessionID != "" {
			sessID = m.ResearchState.SessionID
		}
	}

	headerStr := m.Header.Render(string(m.ActiveView), sessID, workspace)

	var activeBody string
	switch m.ActiveView {
	case ViewAgentView:
		activeBody = m.ConversationView.Render()
	case ViewPlanView:
		activeBody = m.PlanView.Render()
	case ViewDiffView:
		activeBody = m.DiffView.Render()
	case ViewLogView:
		activeBody = m.LogView.Render()
	case ViewResearchView:
		activeBody = m.ResearchView.Render()
	default:
		activeBody = m.ConversationView.Render()
	}

	promptStr := m.Prompt.Render()
	statusStr := m.StatusBar.Render(provider, model, m.IsBusy, m.StatusMsg, m.ContextUsed, m.ContextMax)

	return lipgloss.JoinVertical(lipgloss.Left,
		headerStr,
		activeBody,
		promptStr,
		statusStr,
	)
}
