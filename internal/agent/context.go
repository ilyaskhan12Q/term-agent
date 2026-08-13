package agent

// AgentContext holds execution context and parameters for an agent instance.
type AgentContext struct {
	SessionID     string
	WorkspaceRoot string
	MaxSteps      int
	CurrentStep   int
	Metadata      map[string]string
}
