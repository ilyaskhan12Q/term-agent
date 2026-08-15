// Package commands provides the slash-command input parser.
package commands

import "strings"

// ParsedCommand holds the result of parsing a raw input string.
type ParsedCommand struct {
	// IsCommand is true if the input starts with '/'.
	IsCommand bool
	// Name is the command name without the slash (e.g. "research").
	Name string
	// Args contains space-separated arguments after the command name.
	Args []string
	// Raw is the original, unmodified input string.
	Raw string
}

// Parse parses a raw user input string into a ParsedCommand.
// Returns IsCommand=false if the string does not start with '/'.
func Parse(input string) ParsedCommand {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return ParsedCommand{IsCommand: false, Raw: input}
	}

	// Strip leading slash and split on whitespace.
	parts := strings.Fields(strings.TrimPrefix(input, "/"))
	if len(parts) == 0 {
		return ParsedCommand{IsCommand: false, Raw: input}
	}

	return ParsedCommand{
		IsCommand: true,
		Name:      strings.ToLower(parts[0]),
		Args:      parts[1:],
		Raw:       input,
	}
}
