package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/ilyaskhan/term-agent/internal/persistence/repository"
)

// State represents the current lifecycle state of the application.
type State string

const (
	StateInitializing State = "INITIALIZING"
	StateRunning      State = "RUNNING"
	StateShuttingDown State = "SHUTTING_DOWN"
	StateStopped      State = "STOPPED"
)

// Lifecycle defines the application startup and shutdown contracts.
type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	State() State
}

// SessionManager manages active and historical session state across app lifecycle execution.
type SessionManager struct {
	repo          *repository.SQLiteSessionRepository
	activeSession *repository.SessionRecord
	mu            sync.RWMutex
	logger        *slog.Logger
}

// NewSessionManager initializes SessionManager.
func NewSessionManager(repo *repository.SQLiteSessionRepository, logger *slog.Logger) *SessionManager {
	if logger == nil {
		logger = slog.Default()
	}
	return &SessionManager{
		repo:   repo,
		logger: logger,
	}
}

// PrepareSession resumes an existing session if sessionID is supplied, or creates a new active session.
func (m *SessionManager) PrepareSession(ctx context.Context, requestedSessionID, workspacePath string) (*repository.SessionRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. If explicit session ID requested, retrieve it
	if requestedSessionID != "" {
		s, err := m.repo.GetSession(ctx, requestedSessionID)
		if err != nil {
			return nil, fmt.Errorf("requested session %s not found: %w", requestedSessionID, err)
		}

		if s.Status != "ACTIVE" {
			if err := m.repo.UpdateSessionStatus(ctx, s.ID, "ACTIVE"); err != nil {
				return nil, fmt.Errorf("failed to reactivate session: %w", err)
			}
			s.Status = "ACTIVE"
		}
		m.activeSession = s
		m.logger.Info("Resumed existing session", "session_id", s.ID, "workspace", s.WorkspacePath)
		return s, nil
	}

	// 2. Mark any previously orphaned active sessions as INTERRUPTED
	recovered, err := m.repo.RecoverOrphanedSessions(ctx)
	if err != nil {
		m.logger.Warn("Failed to clean up orphaned sessions", "error", err)
	} else if recovered > 0 {
		m.logger.Info("Recovered orphaned sessions", "count", recovered)
	}

	// 3. Create a brand new active session
	newID := uuid.New().String()
	s := &repository.SessionRecord{
		ID:            newID,
		Title:         fmt.Sprintf("Session %s", newID[:8]),
		WorkspacePath: workspacePath,
		Status:        "ACTIVE",
	}

	if err := m.repo.CreateSession(ctx, s); err != nil {
		return nil, fmt.Errorf("failed to create initial session: %w", err)
	}

	m.activeSession = s
	m.logger.Info("Initialized new session", "session_id", s.ID, "workspace", workspacePath)
	return s, nil
}

// ActiveSession returns current loaded session record.
func (m *SessionManager) ActiveSession() *repository.SessionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeSession
}

// CloseCurrentSession marks active session as CLOSED upon clean shutdown.
func (m *SessionManager) CloseCurrentSession(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeSession == nil {
		return nil
	}

	if err := m.repo.UpdateSessionStatus(ctx, m.activeSession.ID, "CLOSED"); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	m.logger.Info("Closed session successfully", "session_id", m.activeSession.ID)
	m.activeSession.Status = "CLOSED"
	return nil
}
