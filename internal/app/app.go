package app

import (
	"context"

	"github.com/ilyaskhan/term-agent/internal/config"
)

// App encapsulates the top-level application runtime state.
type App struct {
	cfg *config.Config
}

// New constructs a new App instance.
func New() *App {
	return &App{
		cfg: config.DefaultConfig(),
	}
}

// Run starts the application lifecycle.
func (a *App) Run(ctx context.Context, args []string) error {
	// Phase 0: Foundation skeleton
	return nil
}
