package context

import (
	"context"

	"github.com/ilyaskhan/term-agent/internal/model"
)

// Compactor defines the contract for summarizing conversation history when context exceeds budget limit.
type Compactor interface {
	Summarize(ctx context.Context, messages []model.Message) (string, []model.Message, error)
}
