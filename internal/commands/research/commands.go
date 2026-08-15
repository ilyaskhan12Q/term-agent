// Package research provides slash commands for Research Mode in term-agent.
// All commands in this package are research-domain specific.
// To add a coding command domain, create a separate package (e.g. internal/commands/coding).
package research

import (
	"fmt"
	"strings"

	"github.com/ilyaskhan/term-agent/internal/commands"
)

// ResearchState holds mutable research session state that commands read/write.
// It is passed by pointer so all commands share and mutate a single state.
type ResearchState struct {
	// Topic is the current research objective.
	Topic string
	// Provider is the selected LLM provider name.
	Provider string
	// Model is the selected model name.
	Model string
	// Status is the current research lifecycle phase.
	Status string
	// Paused indicates whether execution is currently paused.
	Paused bool
	// Cancelled indicates whether execution has been cancelled.
	Cancelled bool
	// ExportFormat is the requested export format ("markdown", "latex", "pdf").
	ExportFormat string
	// SourceCount is the count of sources collected so far.
	SourceCount int
	// ActivePlan is a human-readable summary of the current task plan.
	ActivePlan string
}

// NewResearchState returns a default ResearchState.
func NewResearchState() *ResearchState {
	return &ResearchState{
		Status:       "IDLE",
		Provider:     "openai",
		Model:        "gpt-4o",
		ExportFormat: "markdown",
	}
}

// RegisterAll registers all research slash commands into the provided registry.
func RegisterAll(r *commands.Registry, state *ResearchState) {
	r.Register(&researchCmd{state: state})
	r.Register(&topicCmd{state: state})
	r.Register(&planCmd{state: state})
	r.Register(&sourcesCmd{state: state})
	r.Register(&statusCmd{state: state})
	r.Register(&pauseCmd{state: state})
	r.Register(&resumeCmd{state: state})
	r.Register(&cancelCmd{state: state})
	r.Register(&exportCmd{state: state})
	r.Register(&modelCmd{state: state})
	r.Register(&helpCmd{})
	r.Register(&clearCmd{})
}

// ---------------------------------------------------------------------------
// /research
// ---------------------------------------------------------------------------

type researchCmd struct{ state *ResearchState }

func (c *researchCmd) Name() string        { return "research" }
func (c *researchCmd) Aliases() []string   { return []string{"r"} }
func (c *researchCmd) Description() string { return "Start a new research session on a topic." }
func (c *researchCmd) Usage() string       { return "/research <topic>" }

func (c *researchCmd) Execute(args []string) commands.CommandResult {
	if len(args) == 0 {
		return commands.CommandResult{
			Output: "Usage: " + c.Usage() + "\nExample: /research transformer architecture memory scaling",
		}
	}
	topic := strings.Join(args, " ")
	c.state.Topic = topic
	c.state.Status = "PLANNING"
	c.state.Paused = false
	c.state.Cancelled = false
	return commands.CommandResult{
		Output:       fmt.Sprintf("Research session started.\nTopic: %s\nStatus: PLANNING\nProvider: %s / %s", topic, c.state.Provider, c.state.Model),
		SwitchView:   "RESEARCH_VIEW",
		UpdateStatus: fmt.Sprintf("Research: PLANNING — %s", topic),
	}
}

// ---------------------------------------------------------------------------
// /topic
// ---------------------------------------------------------------------------

type topicCmd struct{ state *ResearchState }

func (c *topicCmd) Name() string        { return "topic" }
func (c *topicCmd) Aliases() []string   { return nil }
func (c *topicCmd) Description() string { return "Show or update the current research topic." }
func (c *topicCmd) Usage() string       { return "/topic [new topic]" }

func (c *topicCmd) Execute(args []string) commands.CommandResult {
	if len(args) == 0 {
		if c.state.Topic == "" {
			return commands.CommandResult{Output: "No research topic is currently set. Use /research <topic> to begin."}
		}
		return commands.CommandResult{Output: fmt.Sprintf("Current topic: %s", c.state.Topic)}
	}
	c.state.Topic = strings.Join(args, " ")
	return commands.CommandResult{
		Output:       fmt.Sprintf("Topic updated: %s", c.state.Topic),
		UpdateStatus: fmt.Sprintf("Topic: %s", c.state.Topic),
	}
}

// ---------------------------------------------------------------------------
// /plan
// ---------------------------------------------------------------------------

type planCmd struct{ state *ResearchState }

func (c *planCmd) Name() string        { return "plan" }
func (c *planCmd) Aliases() []string   { return nil }
func (c *planCmd) Description() string { return "Show the current research task DAG plan." }
func (c *planCmd) Usage() string       { return "/plan" }

func (c *planCmd) Execute(args []string) commands.CommandResult {
	if c.state.Topic == "" {
		return commands.CommandResult{Output: "No active research session. Use /research <topic> to begin."}
	}
	plan := c.state.ActivePlan
	if plan == "" {
		plan = "Plan not yet generated. Research is in status: " + c.state.Status
	}
	return commands.CommandResult{
		Output:     fmt.Sprintf("Research Plan\nTopic: %s\n\n%s", c.state.Topic, plan),
		SwitchView: "RESEARCH_VIEW",
	}
}

// ---------------------------------------------------------------------------
// /sources
// ---------------------------------------------------------------------------

type sourcesCmd struct{ state *ResearchState }

func (c *sourcesCmd) Name() string        { return "sources" }
func (c *sourcesCmd) Aliases() []string   { return nil }
func (c *sourcesCmd) Description() string { return "Show sources collected so far in the research." }
func (c *sourcesCmd) Usage() string       { return "/sources" }

func (c *sourcesCmd) Execute(args []string) commands.CommandResult {
	if c.state.Topic == "" {
		return commands.CommandResult{Output: "No active research session."}
	}
	return commands.CommandResult{
		Output:     fmt.Sprintf("Sources collected: %d\nStatus: %s", c.state.SourceCount, c.state.Status),
		SwitchView: "RESEARCH_VIEW",
	}
}

// ---------------------------------------------------------------------------
// /status
// ---------------------------------------------------------------------------

type statusCmd struct{ state *ResearchState }

func (c *statusCmd) Name() string        { return "status" }
func (c *statusCmd) Aliases() []string   { return []string{"s"} }
func (c *statusCmd) Description() string { return "Show the current research session status." }
func (c *statusCmd) Usage() string       { return "/status" }

func (c *statusCmd) Execute(args []string) commands.CommandResult {
	if c.state.Topic == "" {
		return commands.CommandResult{Output: "No active research session. Status: IDLE"}
	}
	paused := ""
	if c.state.Paused {
		paused = " (PAUSED)"
	}
	out := fmt.Sprintf(
		"Research Status\n  Topic:    %s\n  Phase:    %s%s\n  Provider: %s / %s\n  Sources:  %d\n  Export:   %s",
		c.state.Topic,
		c.state.Status,
		paused,
		c.state.Provider,
		c.state.Model,
		c.state.SourceCount,
		c.state.ExportFormat,
	)
	return commands.CommandResult{Output: out}
}

// ---------------------------------------------------------------------------
// /pause
// ---------------------------------------------------------------------------

type pauseCmd struct{ state *ResearchState }

func (c *pauseCmd) Name() string        { return "pause" }
func (c *pauseCmd) Aliases() []string   { return nil }
func (c *pauseCmd) Description() string { return "Pause the current research session." }
func (c *pauseCmd) Usage() string       { return "/pause" }

func (c *pauseCmd) Execute(args []string) commands.CommandResult {
	if c.state.Topic == "" {
		return commands.CommandResult{Output: "No active research session to pause."}
	}
	if c.state.Paused {
		return commands.CommandResult{Output: "Research is already paused. Use /resume to continue."}
	}
	c.state.Paused = true
	return commands.CommandResult{
		Output:       "Research paused. Use /resume to continue.",
		UpdateStatus: fmt.Sprintf("Research: PAUSED — %s", c.state.Topic),
	}
}

// ---------------------------------------------------------------------------
// /resume
// ---------------------------------------------------------------------------

type resumeCmd struct{ state *ResearchState }

func (c *resumeCmd) Name() string        { return "resume" }
func (c *resumeCmd) Aliases() []string   { return nil }
func (c *resumeCmd) Description() string { return "Resume a paused research session." }
func (c *resumeCmd) Usage() string       { return "/resume" }

func (c *resumeCmd) Execute(args []string) commands.CommandResult {
	if c.state.Topic == "" {
		return commands.CommandResult{Output: "No active research session to resume."}
	}
	if !c.state.Paused {
		return commands.CommandResult{Output: fmt.Sprintf("Research is not paused. Current status: %s", c.state.Status)}
	}
	c.state.Paused = false
	return commands.CommandResult{
		Output:       fmt.Sprintf("Research resumed. Continuing from: %s", c.state.Status),
		UpdateStatus: fmt.Sprintf("Research: %s — %s", c.state.Status, c.state.Topic),
	}
}

// ---------------------------------------------------------------------------
// /cancel
// ---------------------------------------------------------------------------

type cancelCmd struct{ state *ResearchState }

func (c *cancelCmd) Name() string        { return "cancel" }
func (c *cancelCmd) Aliases() []string   { return nil }
func (c *cancelCmd) Description() string { return "Cancel and discard the current research session." }
func (c *cancelCmd) Usage() string       { return "/cancel" }

func (c *cancelCmd) Execute(args []string) commands.CommandResult {
	if c.state.Topic == "" {
		return commands.CommandResult{Output: "No active research session to cancel."}
	}
	topic := c.state.Topic
	c.state.Topic = ""
	c.state.Status = "IDLE"
	c.state.Paused = false
	c.state.Cancelled = true
	c.state.SourceCount = 0
	c.state.ActivePlan = ""
	return commands.CommandResult{
		Output:       fmt.Sprintf("Research session cancelled.\nTopic was: %s\nAll progress discarded.", topic),
		UpdateStatus: "Research: IDLE",
	}
}

// ---------------------------------------------------------------------------
// /export
// ---------------------------------------------------------------------------

type exportCmd struct{ state *ResearchState }

func (c *exportCmd) Name() string      { return "export" }
func (c *exportCmd) Aliases() []string { return nil }
func (c *exportCmd) Description() string {
	return "Export research output. Format: markdown | latex | pdf."
}
func (c *exportCmd) Usage() string { return "/export [format]" }

func (c *exportCmd) Execute(args []string) commands.CommandResult {
	if c.state.Topic == "" {
		return commands.CommandResult{Output: "No active research session to export."}
	}
	format := c.state.ExportFormat
	if len(args) > 0 {
		f := strings.ToLower(args[0])
		switch f {
		case "markdown", "md":
			format = "markdown"
		case "latex", "tex":
			format = "latex"
		case "pdf":
			format = "pdf"
		default:
			return commands.CommandResult{
				Output: fmt.Sprintf("Unknown export format %q. Supported: markdown, latex, pdf", args[0]),
			}
		}
		c.state.ExportFormat = format
	}
	return commands.CommandResult{
		Output: fmt.Sprintf("Export format set to: %s\nExport will be triggered when research reaches COMPLETED status.", format),
	}
}

// ---------------------------------------------------------------------------
// /model
// ---------------------------------------------------------------------------

type modelCmd struct{ state *ResearchState }

func (c *modelCmd) Name() string        { return "model" }
func (c *modelCmd) Aliases() []string   { return []string{"m"} }
func (c *modelCmd) Description() string { return "Show or change the active LLM provider and model." }
func (c *modelCmd) Usage() string       { return "/model [provider model]" }

func (c *modelCmd) Execute(args []string) commands.CommandResult {
	if len(args) == 0 {
		return commands.CommandResult{
			Output: fmt.Sprintf("Active provider: %s\nActive model:    %s\n\nTo change: /model <provider> <model>\nExample:   /model anthropic claude-3-5-sonnet-20241022\nProviders: openai, anthropic, gemini", c.state.Provider, c.state.Model),
		}
	}
	if len(args) < 2 {
		return commands.CommandResult{
			Output: "Usage: " + c.Usage() + "\nExample: /model openai gpt-4o",
		}
	}
	provider := strings.ToLower(args[0])
	switch provider {
	case "openai", "anthropic", "gemini":
	default:
		return commands.CommandResult{
			Output: fmt.Sprintf("Unknown provider %q. Supported: openai, anthropic, gemini", provider),
		}
	}
	c.state.Provider = provider
	c.state.Model = args[1]
	return commands.CommandResult{
		Output:       fmt.Sprintf("Provider set to: %s\nModel set to:    %s", c.state.Provider, c.state.Model),
		UpdateStatus: fmt.Sprintf("%s / %s", c.state.Provider, c.state.Model),
	}
}

// ---------------------------------------------------------------------------
// /help
// ---------------------------------------------------------------------------

type helpCmd struct{}

func (c *helpCmd) Name() string        { return "help" }
func (c *helpCmd) Aliases() []string   { return []string{"h", "?"} }
func (c *helpCmd) Description() string { return "Show available commands." }
func (c *helpCmd) Usage() string       { return "/help" }

// Execute is overridden by the TUI dispatcher which has access to Registry.HelpText().
// This stub returns a placeholder; the TUI dispatcher intercepts /help before Dispatch.
func (c *helpCmd) Execute(args []string) commands.CommandResult {
	return commands.CommandResult{Output: "Help output is generated by the command registry."}
}

// ---------------------------------------------------------------------------
// /clear
// ---------------------------------------------------------------------------

type clearCmd struct{}

func (c *clearCmd) Name() string        { return "clear" }
func (c *clearCmd) Aliases() []string   { return []string{"cls"} }
func (c *clearCmd) Description() string { return "Clear the current view output." }
func (c *clearCmd) Usage() string       { return "/clear" }

func (c *clearCmd) Execute(args []string) commands.CommandResult {
	return commands.CommandResult{
		Output:     "__CLEAR__",
		SwitchView: "",
	}
}
