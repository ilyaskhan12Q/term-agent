package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds non-secret and secret configuration for term-agent.
// Configuration precedence order: CLI Flags > Environment Variables > Config File (~/.config/termagent/config.toml) > Defaults.
type Config struct {
	// General Settings
	LogLevel     string `toml:"log_level"`
	DatabasePath string `toml:"database_path"`
	WorkspaceDir string `toml:"workspace_dir"`

	// Model Configuration
	DefaultProvider string `toml:"default_provider"`
	DefaultModel    string `toml:"default_model"`

	// API Keys (Loaded from environment only, never committed to config files)
	OpenAIAPIKey    string `toml:"-"`
	AnthropicAPIKey string `toml:"-"`
	GeminiAPIKey    string `toml:"-"`

	// Scheduler & Limits
	MaxParallelWorkers int `toml:"max_parallel_workers"`
	ContextWindowLimit int `toml:"context_window_limit"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	defaultDB := filepath.Join(home, ".config", "termagent", "termagent.db")

	return &Config{
		LogLevel:           "info",
		DatabasePath:       defaultDB,
		WorkspaceDir:       ".",
		DefaultProvider:    "openai",
		DefaultModel:       "gpt-4o",
		MaxParallelWorkers: 4,
		ContextWindowLimit: 128000,
	}
}

// Load loads configuration following the precedence hierarchy.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Environment variable overrides
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg.OpenAIAPIKey = key
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.AnthropicAPIKey = key
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		cfg.GeminiAPIKey = key
	}
	if dbPath := os.Getenv("TERMAGENT_DB_PATH"); dbPath != "" {
		cfg.DatabasePath = dbPath
	}

	return cfg, nil
}

// Validate checks for configuration sanity.
func (c *Config) Validate() error {
	if c.DatabasePath == "" {
		return fmt.Errorf("database_path cannot be empty")
	}
	return nil
}
