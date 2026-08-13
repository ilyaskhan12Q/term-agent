package repository

import (
	"context"

	"github.com/ilyaskhan/term-agent/internal/model"
)

// SessionRecord represents database record for a session.
type SessionRecord struct {
	ID            string
	Title         string
	WorkspacePath string
	Status        string
	CreatedAt     string
	UpdatedAt     string
}

// SessionRepository defines the contract for session persistence.
type SessionRepository interface {
	CreateSession(ctx context.Context, s *SessionRecord) error
	GetSession(ctx context.Context, id string) (*SessionRecord, error)
	ListSessions(ctx context.Context) ([]*SessionRecord, error)
}

// MessageRepository defines the contract for message persistence.
type MessageRepository interface {
	SaveMessage(ctx context.Context, sessionID string, msg *model.Message) error
	GetMessages(ctx context.Context, sessionID string) ([]*model.Message, error)
}
