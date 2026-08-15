package unit_test

import (
	"strings"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/commands"
	rcmds "github.com/ilyaskhan/term-agent/internal/commands/research"
)

// ---------------------------------------------------------------------------
// Parser tests
// ---------------------------------------------------------------------------

func TestParser_PlainText(t *testing.T) {
	p := commands.Parse("hello world")
	if p.IsCommand {
		t.Fatal("expected IsCommand=false for plain text")
	}
	if p.Raw != "hello world" {
		t.Errorf("unexpected Raw: %q", p.Raw)
	}
}

func TestParser_SlashCommand_NoArgs(t *testing.T) {
	p := commands.Parse("/status")
	if !p.IsCommand {
		t.Fatal("expected IsCommand=true")
	}
	if p.Name != "status" {
		t.Errorf("expected name 'status', got %q", p.Name)
	}
	if len(p.Args) != 0 {
		t.Errorf("expected 0 args, got %v", p.Args)
	}
}

func TestParser_SlashCommand_WithArgs(t *testing.T) {
	p := commands.Parse("/research transformer architecture memory scaling")
	if !p.IsCommand {
		t.Fatal("expected IsCommand=true")
	}
	if p.Name != "research" {
		t.Errorf("expected name 'research', got %q", p.Name)
	}
	if len(p.Args) != 4 {
		t.Errorf("expected 4 args, got %d: %v", len(p.Args), p.Args)
	}
}

func TestParser_CaseInsensitive(t *testing.T) {
	p := commands.Parse("/STATUS")
	if p.Name != "status" {
		t.Errorf("expected lowercase name, got %q", p.Name)
	}
}

func TestParser_EmptySlash(t *testing.T) {
	p := commands.Parse("/")
	if p.IsCommand {
		t.Fatal("bare slash should not be treated as a command")
	}
}

func TestParser_Whitespace(t *testing.T) {
	p := commands.Parse("   /plan   ")
	if !p.IsCommand {
		t.Fatal("expected IsCommand=true after trimming")
	}
	if p.Name != "plan" {
		t.Errorf("expected name 'plan', got %q", p.Name)
	}
}

// ---------------------------------------------------------------------------
// Registry tests
// ---------------------------------------------------------------------------

func TestRegistry_RegisterAndDispatch(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	result := reg.Dispatch("status", nil)
	if result.Output == "" {
		t.Fatal("expected non-empty output from /status")
	}
}

func TestRegistry_UnknownCommand(t *testing.T) {
	reg := commands.NewRegistry()
	result := reg.Dispatch("nonexistent", nil)
	if !strings.Contains(result.Output, "Unknown command") {
		t.Errorf("expected 'Unknown command' in output, got: %q", result.Output)
	}
}

func TestRegistry_HelpText_ContainsAllCommands(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	help := reg.HelpText()
	for _, name := range reg.Names() {
		if !strings.Contains(help, "/"+name) {
			t.Errorf("help text missing command /%s", name)
		}
	}
}

func TestRegistry_DuplicateRegistrationPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration, got none")
		}
	}()
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)
	rcmds.RegisterAll(reg, state) // second call must panic
}

// ---------------------------------------------------------------------------
// Individual command tests
// ---------------------------------------------------------------------------

func TestResearchCmd_NoArgs(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	result := reg.Dispatch("research", nil)
	if !strings.Contains(result.Output, "Usage:") {
		t.Errorf("expected usage hint, got: %q", result.Output)
	}
}

func TestResearchCmd_WithTopic(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	result := reg.Dispatch("research", []string{"transformer", "scaling"})
	if state.Topic != "transformer scaling" {
		t.Errorf("expected topic 'transformer scaling', got %q", state.Topic)
	}
	if result.SwitchView != "RESEARCH_VIEW" {
		t.Errorf("expected SwitchView RESEARCH_VIEW, got %q", result.SwitchView)
	}
	if state.Status != "PLANNING" {
		t.Errorf("expected status PLANNING, got %q", state.Status)
	}
}

func TestStatusCmd_NoActiveTopic(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	result := reg.Dispatch("status", nil)
	if !strings.Contains(result.Output, "IDLE") {
		t.Errorf("expected IDLE in output, got: %q", result.Output)
	}
}

func TestStatusCmd_WithTopic(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	state.Topic = "quantum error correction"
	state.Status = "EXECUTING"

	result := reg.Dispatch("status", nil)
	if !strings.Contains(result.Output, "quantum error correction") {
		t.Errorf("topic not in output: %q", result.Output)
	}
	if !strings.Contains(result.Output, "EXECUTING") {
		t.Errorf("status not in output: %q", result.Output)
	}
}

func TestPauseResume_Lifecycle(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	// Cannot pause without a topic.
	result := reg.Dispatch("pause", nil)
	if !strings.Contains(result.Output, "No active") {
		t.Errorf("expected 'No active' for pause without topic: %q", result.Output)
	}

	state.Topic = "LLM scaling laws"
	state.Status = "EXECUTING"

	// Pause.
	result = reg.Dispatch("pause", nil)
	if !state.Paused {
		t.Fatal("expected state.Paused=true after /pause")
	}

	// Double-pause guard.
	result = reg.Dispatch("pause", nil)
	if !strings.Contains(result.Output, "already paused") {
		t.Errorf("expected 'already paused' message, got: %q", result.Output)
	}

	// Resume.
	result = reg.Dispatch("resume", nil)
	if state.Paused {
		t.Fatal("expected state.Paused=false after /resume")
	}
}

func TestCancelCmd_ClearsState(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	state.Topic = "vector databases"
	state.Status = "EXECUTING"
	state.SourceCount = 5

	reg.Dispatch("cancel", nil)

	if state.Topic != "" {
		t.Errorf("expected empty topic after cancel, got %q", state.Topic)
	}
	if state.SourceCount != 0 {
		t.Errorf("expected 0 sources after cancel, got %d", state.SourceCount)
	}
	if !state.Cancelled {
		t.Fatal("expected state.Cancelled=true")
	}
}

func TestModelCmd_Show(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	result := reg.Dispatch("model", nil)
	if !strings.Contains(result.Output, "openai") {
		t.Errorf("expected default provider 'openai' in output: %q", result.Output)
	}
}

func TestModelCmd_SetProviderAndModel(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	result := reg.Dispatch("model", []string{"anthropic", "claude-3-5-sonnet-20241022"})
	if result.Output == "" {
		t.Fatal("expected non-empty output")
	}
	if state.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got %q", state.Provider)
	}
	if state.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("expected model 'claude-3-5-sonnet-20241022', got %q", state.Model)
	}
}

func TestModelCmd_UnknownProvider(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	result := reg.Dispatch("model", []string{"fakeai", "model-x"})
	if !strings.Contains(result.Output, "Unknown provider") {
		t.Errorf("expected 'Unknown provider', got: %q", result.Output)
	}
}

func TestExportCmd_SetFormat(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)
	state.Topic = "neural scaling" // need topic for export

	reg.Dispatch("export", []string{"latex"})
	if state.ExportFormat != "latex" {
		t.Errorf("expected ExportFormat 'latex', got %q", state.ExportFormat)
	}

	reg.Dispatch("export", []string{"md"})
	if state.ExportFormat != "markdown" {
		t.Errorf("expected ExportFormat 'markdown', got %q", state.ExportFormat)
	}
}

func TestExportCmd_InvalidFormat(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)
	state.Topic = "test"

	result := reg.Dispatch("export", []string{"docx"})
	if !strings.Contains(result.Output, "Unknown export format") {
		t.Errorf("expected 'Unknown export format', got: %q", result.Output)
	}
}

func TestAliases_Work(t *testing.T) {
	reg := commands.NewRegistry()
	state := rcmds.NewResearchState()
	rcmds.RegisterAll(reg, state)

	// /r is alias for /research
	result := reg.Dispatch("r", []string{"LLM"})
	if state.Topic != "LLM" {
		t.Errorf("alias /r did not trigger research command: %q", result.Output)
	}

	// /s is alias for /status
	result = reg.Dispatch("s", nil)
	if !strings.Contains(result.Output, "LLM") {
		t.Errorf("alias /s did not trigger status command: %q", result.Output)
	}

	// /m is alias for /model
	reg.Dispatch("m", []string{"gemini", "gemini-1.5-pro"})
	if state.Provider != "gemini" {
		t.Errorf("alias /m did not trigger model command")
	}

	// /cls is alias for /clear
	result = reg.Dispatch("cls", nil)
	if result.Output != "__CLEAR__" {
		t.Errorf("alias /cls should return __CLEAR__, got %q", result.Output)
	}
}
