package context

import (
	"context"

	"github.com/ilyaskhan/term-agent/internal/model"
)

// ContextManager manages message history, token budgets, and context window compaction.
type ContextManager interface {
	AddMessage(msg model.Message) error
	GetMessages() []model.Message
	TokenCount() int
	Compact(ctx context.Context) error
}
