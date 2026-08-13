package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/persistence"
)

func TestSQLiteInitializationAndMigrations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-db-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test_termagent.db")
	db, err := persistence.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Verify all 11 required tables exist
	tables := []string{
		"sessions",
		"messages",
		"tool_calls",
		"tool_results",
		"tasks",
		"task_dependencies",
		"mutation_transactions",
		"mutations",
		"context_summaries",
		"model_usage",
		"events",
	}

	for _, table := range tables {
		var name string
		err := db.QueryRowContext(context.Background(),
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s was not created during migration: %v", table, err)
		}
	}
}
