// Package commands provides an extensible slash-command system for term-agent.
// Commands are registered by domain (research, coding, etc.) and dispatched from
// the TUI prompt by the parser. The registry is designed so new command domains
// can be added without restructuring existing commands.
package commands

import (
	"fmt"
	"sort"
	"strings"
)

// CommandResult is the outcome of executing a slash command.
type CommandResult struct {
	// Output is the text to display to the user.
	Output string
	// SwitchView instructs the TUI to switch to this view. Empty means no switch.
	SwitchView string
	// UpdateStatus replaces the TUI status bar message. Empty means no change.
	UpdateStatus string
	// Quit signals the application should exit.
	Quit bool
}

// Command is the interface all slash commands must implement.
type Command interface {
	// Name returns the canonical command name without the slash prefix (e.g. "research").
	Name() string
	// Aliases returns additional trigger names (may be empty).
	Aliases() []string
	// Description returns a one-line description for /help output.
	Description() string
	// Usage returns a usage string (e.g. "/research <topic>").
	Usage() string
	// Execute runs the command with the parsed arguments and returns a result.
	Execute(args []string) CommandResult
}

// Registry stores all registered commands and dispatches parsed input.
type Registry struct {
	commands map[string]Command // keyed by canonical name
	aliases  map[string]string  // alias -> canonical name
}

// NewRegistry constructs an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds a command. Panics if a name or alias conflicts with an existing registration.
func (r *Registry) Register(cmd Command) {
	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		panic(fmt.Sprintf("commands: duplicate command name %q", name))
	}
	r.commands[name] = cmd
	for _, alias := range cmd.Aliases() {
		if _, exists := r.aliases[alias]; exists {
			panic(fmt.Sprintf("commands: duplicate alias %q", alias))
		}
		r.aliases[alias] = name
	}
}

// Dispatch looks up and executes a command by name or alias with the given args.
// Returns an error result if the command is not found.
func (r *Registry) Dispatch(name string, args []string) CommandResult {
	canonical := name
	if mapped, ok := r.aliases[name]; ok {
		canonical = mapped
	}
	cmd, ok := r.commands[canonical]
	if !ok {
		return CommandResult{
			Output: fmt.Sprintf("Unknown command /%s. Type /help to see available commands.", name),
		}
	}
	return cmd.Execute(args)
}

// HelpText returns formatted help output listing all registered commands.
func (r *Registry) HelpText() string {
	var sb strings.Builder
	sb.WriteString("Available commands:\n\n")

	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		cmd := r.commands[name]
		aliasStr := ""
		if len(cmd.Aliases()) > 0 {
			aliasStr = fmt.Sprintf(" (aliases: /%s)", strings.Join(cmd.Aliases(), ", /"))
		}
		sb.WriteString(fmt.Sprintf("  %-12s %s%s\n", "/"+name, cmd.Description(), aliasStr))
	}
	return sb.String()
}

// Names returns a sorted list of all canonical command names.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Lookup returns a command by canonical name or alias, or nil if not found.
func (r *Registry) Lookup(name string) Command {
	canonical := name
	if mapped, ok := r.aliases[name]; ok {
		canonical = mapped
	}
	return r.commands[canonical]
}
