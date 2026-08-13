package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ilyaskhan/term-agent/internal/persistence"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

// SQLiteSessionRepository implements SessionRepository backed by SQLite.
type SQLiteSessionRepository struct {
	db *persistence.DB
}

// NewSQLiteSessionRepository initializes a new SQLiteSessionRepository.
func NewSQLiteSessionRepository(db *persistence.DB) *SQLiteSessionRepository {
	return &SQLiteSessionRepository{db: db}
}

// CreateSession persists a new session record.
func (r *SQLiteSessionRepository) CreateSession(ctx context.Context, s *SessionRecord) error {
	if s == nil || s.ID == "" {
		return fmt.Errorf("invalid session record")
	}

	if s.Status == "" {
		s.Status = "ACTIVE"
	}
	now := time.Now().Format(time.RFC3339)
	if s.CreatedAt == "" {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	query := `
		INSERT INTO sessions (id, title, workspace_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.Title, s.WorkspacePath, s.Status, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSession fetches a session by its unique ID.
func (r *SQLiteSessionRepository) GetSession(ctx context.Context, id string) (*SessionRecord, error) {
	query := `
		SELECT id, title, workspace_path, status, created_at, updated_at
		FROM sessions
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var s SessionRecord
	err := row.Scan(&s.ID, &s.Title, &s.WorkspacePath, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil
}

// ListSessions retrieves all sessions ordered by updated_at descending.
func (r *SQLiteSessionRepository) ListSessions(ctx context.Context) ([]*SessionRecord, error) {
	query := `
		SELECT id, title, workspace_path, status, created_at, updated_at
		FROM sessions
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*SessionRecord
	for rows.Next() {
		var s SessionRecord
		if err := rows.Scan(&s.ID, &s.Title, &s.WorkspacePath, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		sessions = append(sessions, &s)
	}

	return sessions, rows.Err()
}

// UpdateSessionStatus updates the status of an existing session.
func (r *SQLiteSessionRepository) UpdateSessionStatus(ctx context.Context, id string, status string) error {
	now := time.Now().Format(time.RFC3339)
	query := `
		UPDATE sessions
		SET status = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// GetActiveSession returns the latest session with status 'ACTIVE'.
func (r *SQLiteSessionRepository) GetActiveSession(ctx context.Context) (*SessionRecord, error) {
	query := `
		SELECT id, title, workspace_path, status, created_at, updated_at
		FROM sessions
		WHERE status = 'ACTIVE'
		ORDER BY updated_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query)

	var s SessionRecord
	err := row.Scan(&s.ID, &s.Title, &s.WorkspacePath, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}
	return &s, nil
}

// RecoverOrphanedSessions updates any 'ACTIVE' session to 'INTERRUPTED' upon crash recovery.
func (r *SQLiteSessionRepository) RecoverOrphanedSessions(ctx context.Context) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	query := `
		UPDATE sessions
		SET status = 'INTERRUPTED', updated_at = ?
		WHERE status = 'ACTIVE'
	`
	res, err := r.db.ExecContext(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("failed to recover orphaned sessions: %w", err)
	}
	return res.RowsAffected()
}
