package tui

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
	UnifiedDiff   string
}
