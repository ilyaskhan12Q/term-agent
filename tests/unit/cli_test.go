package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/config"
)

func TestCLIConfigPrecedence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create custom config.toml
	configFile := filepath.Join(tmpDir, "config.toml")
	tomlData := []byte(`
default_provider = "anthropic"
default_model = "claude-3-5-sonnet-20241022"
log_level = "warn"
`)
	if err := os.WriteFile(configFile, tomlData, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set Environment Variable
	t.Setenv("TERMAGENT_MODEL", "gemini-1.5-pro")

	// CLI Flags (Highest Precedence)
	flags := &config.CLIFlags{
		ConfigPath:   configFile,
		Model:        "gpt-4o",
		WorkspaceDir: tmpDir,
		DryRun:       true,
	}

	cfg, err := config.LoadWithOptions(flags)
	if err != nil {
		t.Fatalf("failed to load config with options: %v", err)
	}

	// 1. Model overridden by CLI Flag (gpt-4o over env gemini-1.5-pro and file claude-3-5-sonnet)
	if cfg.DefaultModel != "gpt-4o" {
		t.Errorf("expected model gpt-4o from CLI flag, got %s", cfg.DefaultModel)
	}

	// 2. Provider loaded from Config File (anthropic)
	if cfg.DefaultProvider != "anthropic" {
		t.Errorf("expected provider anthropic from config file, got %s", cfg.DefaultProvider)
	}

	// 3. DryRun set by CLI Flag
	if !cfg.DryRun {
		t.Errorf("expected dry_run to be true from CLI flag")
	}

	// 4. Absolute Workspace directory
	absTmp, _ := filepath.Abs(tmpDir)
	if cfg.WorkspaceDir != absTmp {
		t.Errorf("expected workspace %s, got %s", absTmp, cfg.WorkspaceDir)
	}
}
