package persistence

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations/000001_initial_schema.sql
var initialSchemaSQL string

//go:embed migrations/000002_research_workflow.sql
var researchWorkflowSQL string

// DB encapsulates the SQLite database connection pool and schema migrations.
type DB struct {
	*sql.DB
	dbPath string
}

// Open initializes SQLite database at dbPath, enabling WAL mode and foreign key constraints.
func Open(dbPath string) (*DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	database := &DB{
		DB:     db,
		dbPath: dbPath,
	}

	if err := database.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return database, nil
}

// Migrate executes schema migrations.
func (d *DB) Migrate(ctx context.Context) error {
	if _, err := d.ExecContext(ctx, initialSchemaSQL); err != nil {
		return fmt.Errorf("failed to execute initial schema migration: %w", err)
	}
	if _, err := d.ExecContext(ctx, researchWorkflowSQL); err != nil {
		return fmt.Errorf("failed to execute research workflow migration: %w", err)
	}
	return nil
}
