package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ilyaskhan/term-agent/internal/config"
	"github.com/ilyaskhan/term-agent/internal/events"
	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/persistence/repository"
)

// App encapsulates top-level application runtime components and lifecycle management.
type App struct {
	cfg            *config.Config
	db             *persistence.DB
	asyncWriter    *persistence.ChannelAsyncWriter
	eventBus       events.EventBus
	sessionRepo    *repository.SQLiteSessionRepository
	sessionManager *SessionManager
	logger         *slog.Logger
	state          State
	mu             sync.RWMutex
	startTime      time.Time
}

// Bootstrap constructs and initializes App components, checking cold startup latency.
func Bootstrap(flags *config.CLIFlags) (*App, error) {
	start := time.Now()

	// 1. Load Configuration
	cfg, err := config.LoadWithOptions(flags)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	// 2. Setup Secret-Redacting Logger
	logger := config.SetupLogger(cfg.LogLevel, os.Stderr)

	// 3. Open SQLite Database & Apply Migrations
	db, err := persistence.Open(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}

	// 4. Initialize Async Writer Channel
	asyncWriter := persistence.NewChannelAsyncWriter(db, 500, logger)

	// 5. Initialize Memory Event Bus
	eventBus := events.NewInMemoryEventBus()

	// 6. Initialize Repositories and Session Manager
	sessionRepo := repository.NewSQLiteSessionRepository(db)
	sessionManager := NewSessionManager(sessionRepo, logger)

	appInstance := &App{
		cfg:            cfg,
		db:             db,
		asyncWriter:    asyncWriter,
		eventBus:       eventBus,
		sessionRepo:    sessionRepo,
		sessionManager: sessionManager,
		logger:         logger,
		state:          StateInitializing,
		startTime:      start,
	}

	bootDuration := time.Since(start)
	logger.Debug("App bootstrapped successfully", "duration_ms", bootDuration.Milliseconds())

	return appInstance, nil
}

// Config returns loaded configuration.
func (a *App) Config() *config.Config {
	return a.cfg
}

// DB returns database handle.
func (a *App) DB() *persistence.DB {
	return a.db
}

// EventBus returns event bus handle.
func (a *App) EventBus() events.EventBus {
	return a.eventBus
}

// SessionManager returns session manager handle.
func (a *App) SessionManager() *SessionManager {
	return a.sessionManager
}

// State returns current app lifecycle state.
func (a *App) State() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

func (a *App) setState(s State) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state = s
}

// Run executes main app execution loop, handling SIGINT/SIGTERM OS signals.
func (a *App) Run(ctx context.Context) error {
	a.setState(StateRunning)

	// 1. Prepare / Resume active session
	session, err := a.sessionManager.PrepareSession(ctx, a.cfg.SessionID, a.cfg.WorkspaceDir)
	if err != nil {
		a.setState(StateStopped)
		return fmt.Errorf("failed to prepare session: %w", err)
	}

	a.logger.Info("Term-agent started",
		"session_id", session.ID,
		"workspace", a.cfg.WorkspaceDir,
		"model", a.cfg.DefaultModel,
		"provider", a.cfg.DefaultProvider,
	)

	// 2. Register Signal Handler for SIGINT and SIGTERM
	sigCtx, stopSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	// Wait for termination signal or context cancellation
	select {
	case <-sigCtx.Done():
		a.logger.Info("Received shutdown signal, terminating gracefully...")
	}

	return a.Shutdown(context.Background())
}

// Shutdown executes graceful component shutdown.
func (a *App) Shutdown(ctx context.Context) error {
	a.setState(StateShuttingDown)
	a.logger.Info("Shutting down term-agent subsystems...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var shutdownErr error

	// 1. Close Active Session
	if err := a.sessionManager.CloseCurrentSession(shutdownCtx); err != nil {
		a.logger.Error("Failed to close session cleanly", "error", err)
		shutdownErr = err
	}

	// 2. Drain and Close Async Writer Channel
	if a.asyncWriter != nil {
		if err := a.asyncWriter.Close(shutdownCtx); err != nil {
			a.logger.Error("Failed to close async db writer", "error", err)
		}
	}

	// 3. Close Database Connection Pool
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Error("Failed to close database", "error", err)
		}
	}

	a.setState(StateStopped)
	a.logger.Info("App shutdown complete")
	return shutdownErr
}
