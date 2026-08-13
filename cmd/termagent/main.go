package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ilyaskhan/term-agent/internal/app"
	"github.com/ilyaskhan/term-agent/internal/config"
)

const Version = "0.1.0-dev"

func parseFlags(args []string) (*config.CLIFlags, bool) {
	fs := flag.NewFlagSet("term-agent", flag.ContinueOnError)

	var flags config.CLIFlags

	fs.StringVar(&flags.WorkspaceDir, "workspace", "", "Path to target workspace directory")
	fs.StringVar(&flags.WorkspaceDir, "w", "", "Path to target workspace directory (shorthand)")

	fs.StringVar(&flags.Model, "model", "", "Default LLM model (e.g. gpt-4o, claude-3-5-sonnet-20241022, gemini-1.5-pro)")
	fs.StringVar(&flags.Model, "m", "", "Default LLM model (shorthand)")

	fs.StringVar(&flags.Provider, "provider", "", "Default LLM provider (openai, anthropic, gemini)")
	fs.StringVar(&flags.Provider, "p", "", "Default LLM provider (shorthand)")

	fs.StringVar(&flags.SessionID, "session", "", "Resume existing session by ID")
	fs.StringVar(&flags.SessionID, "s", "", "Resume existing session by ID (shorthand)")

	fs.StringVar(&flags.ConfigPath, "config", "", "Path to custom config file (default ~/.config/termagent/config.toml)")
	fs.StringVar(&flags.ConfigPath, "c", "", "Path to custom config file (shorthand)")

	fs.StringVar(&flags.LogLevel, "log-level", "", "Logging level (debug, info, warn, error)")

	fs.BoolVar(&flags.DryRun, "dry-run", false, "Run in dry-run mode without modifying disk or committing transactions")
	fs.BoolVar(&flags.Debug, "debug", false, "Enable verbose debug mode")

	fs.BoolVar(&flags.Version, "version", false, "Print version information and exit")
	fs.BoolVar(&flags.Version, "v", false, "Print version information and exit (shorthand)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: term-agent [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, false
	}

	if flags.Version {
		fmt.Printf("term-agent v%s\n", Version)
		return nil, false
	}

	return &flags, true
}

func main() {
	flags, shouldRun := parseFlags(os.Args[1:])
	if !shouldRun {
		return
	}

	application, err := app.Bootstrap(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error bootstrapping term-agent: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := application.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error running term-agent: %v\n", err)
		os.Exit(1)
	}
}
