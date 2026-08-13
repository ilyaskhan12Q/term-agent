package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// CLIFlags holds parsed command line arguments.
type CLIFlags struct {
	ConfigPath   string
	WorkspaceDir string
	Model        string
	Provider     string
	SessionID    string
	LogLevel     string
	DryRun       bool
	Debug        bool
	Version      bool
}

// Config holds non-secret and secret configuration for term-agent.
// Precedence order: CLI Flags > Environment Variables > Config File > Defaults.
type Config struct {
	// General Settings
	LogLevel     string `toml:"log_level"`
	DatabasePath string `toml:"database_path"`
	WorkspaceDir string `toml:"workspace_dir"`
	ConfigFile   string `toml:"-"`
	SessionID    string `toml:"-"`
	DryRun       bool   `toml:"dry_run"`
	Debug        bool   `toml:"debug"`

	// Model Configuration
	DefaultProvider string `toml:"default_provider"`
	DefaultModel    string `toml:"default_model"`

	// API Keys (Loaded from environment only, never written to config files)
	OpenAIAPIKey    string `toml:"-"`
	AnthropicAPIKey string `toml:"-"`
	GeminiAPIKey    string `toml:"-"`

	// Scheduler & Limits
	MaxParallelWorkers int `toml:"max_parallel_workers"`
	ContextWindowLimit int `toml:"context_window_limit"`
}

// DefaultConfig returns default configuration values.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	defaultDB := filepath.Join(home, ".config", "termagent", "termagent.db")
	defaultConfig := filepath.Join(home, ".config", "termagent", "config.toml")

	return &Config{
		LogLevel:           "info",
		DatabasePath:       defaultDB,
		WorkspaceDir:       ".",
		ConfigFile:         defaultConfig,
		DefaultProvider:    "openai",
		DefaultModel:       "gpt-4o",
		MaxParallelWorkers: 5,
		ContextWindowLimit: 128000,
		DryRun:             false,
		Debug:              false,
	}
}

// LoadWithOptions loads configuration following the strict precedence chain:
// CLI Flags > Environment Variables > Config File > Defaults.
func LoadWithOptions(flags *CLIFlags) (*Config, error) {
	cfg := DefaultConfig()

	// 1. Determine config file path
	configPath := cfg.ConfigFile
	if flags != nil && flags.ConfigPath != "" {
		configPath = flags.ConfigPath
	}

	// 2. Read Config File if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var fileCfg Config
			if err := toml.Unmarshal(data, &fileCfg); err == nil {
				if fileCfg.LogLevel != "" {
					cfg.LogLevel = fileCfg.LogLevel
				}
				if fileCfg.DatabasePath != "" {
					cfg.DatabasePath = fileCfg.DatabasePath
				}
				if fileCfg.WorkspaceDir != "" {
					cfg.WorkspaceDir = fileCfg.WorkspaceDir
				}
				if fileCfg.DefaultProvider != "" {
					cfg.DefaultProvider = fileCfg.DefaultProvider
				}
				if fileCfg.DefaultModel != "" {
					cfg.DefaultModel = fileCfg.DefaultModel
				}
				if fileCfg.MaxParallelWorkers > 0 {
					cfg.MaxParallelWorkers = fileCfg.MaxParallelWorkers
				}
				if fileCfg.ContextWindowLimit > 0 {
					cfg.ContextWindowLimit = fileCfg.ContextWindowLimit
				}
			}
		}
	}

	// 3. Environment Variable Overrides
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg.OpenAIAPIKey = key
		RegisterSecret(key)
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.AnthropicAPIKey = key
		RegisterSecret(key)
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		cfg.GeminiAPIKey = key
		RegisterSecret(key)
	}
	if dbPath := os.Getenv("TERMAGENT_DB_PATH"); dbPath != "" {
		cfg.DatabasePath = dbPath
	}
	if level := os.Getenv("TERMAGENT_LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}
	if provider := os.Getenv("TERMAGENT_PROVIDER"); provider != "" {
		cfg.DefaultProvider = provider
	}
	if model := os.Getenv("TERMAGENT_MODEL"); model != "" {
		cfg.DefaultModel = model
	}

	// 4. CLI Flags Overrides (Highest Precedence)
	if flags != nil {
		if flags.WorkspaceDir != "" {
			cfg.WorkspaceDir = flags.WorkspaceDir
		}
		if flags.Model != "" {
			cfg.DefaultModel = flags.Model
		}
		if flags.Provider != "" {
			cfg.DefaultProvider = flags.Provider
		}
		if flags.LogLevel != "" {
			cfg.LogLevel = flags.LogLevel
		}
		if flags.Debug {
			cfg.LogLevel = "debug"
			cfg.Debug = true
		}
		if flags.SessionID != "" {
			cfg.SessionID = flags.SessionID
		}
		if flags.DryRun {
			cfg.DryRun = true
		}
	}

	// Clean path formatting
	if absWork, err := filepath.Abs(cfg.WorkspaceDir); err == nil {
		cfg.WorkspaceDir = absWork
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Load loads default configuration with environment variables.
func Load() (*Config, error) {
	return LoadWithOptions(nil)
}

// Validate checks configuration parameters.
func (c *Config) Validate() error {
	if c.DatabasePath == "" {
		return errors.New("database_path cannot be empty")
	}
	if c.WorkspaceDir == "" {
		return errors.New("workspace_dir cannot be empty")
	}
	return nil
}
