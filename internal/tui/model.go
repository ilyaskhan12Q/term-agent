package tui

// ViewMode represents the active primary tab / panel in the TUI.
type ViewMode string

const (
	ViewAgentView ViewMode = "AGENT_VIEW"
	ViewDiffView  ViewMode = "DIFF_VIEW"
	ViewPlanView  ViewMode = "PLAN_VIEW"
	ViewLogView   ViewMode = "LOG_VIEW"
)

// UIState represents the baseline state structure for Bubble Tea model.
type UIState struct {
	ActiveView ViewMode
	StatusMsg  string
	IsBusy     bool
}
