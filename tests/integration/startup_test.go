package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ilyaskhan/term-agent/internal/app"
	"github.com/ilyaskhan/term-agent/internal/config"
)

func TestColdStartupLatency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-startup-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "startup.db")

	flags := &config.CLIFlags{
		WorkspaceDir: tmpDir,
		LogLevel:     "error",
	}
	t.Setenv("TERMAGENT_DB_PATH", dbPath)

	start := time.Now()
	application, err := app.Bootstrap(flags)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	duration := time.Since(start)

	maxDuration := 150 * time.Millisecond
	if raceEnabled {
		maxDuration = 500 * time.Millisecond
	}

	if duration > maxDuration {
		t.Errorf("cold startup latency exceeded %v budget: got %v", maxDuration, duration)
	}

	if application.State() != app.StateInitializing {
		t.Errorf("expected state INITIALIZING, got %s", application.State())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	if application.State() != app.StateStopped {
		t.Errorf("expected state STOPPED after shutdown, got %s", application.State())
	}
}
