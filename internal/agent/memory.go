package agent

import (
	"context"
)

// Memory defines the contract for agent persistent and short-term memory management.
type Memory interface {
	Remember(ctx context.Context, key string, value string) error
	Recall(ctx context.Context, key string) (string, error)
	Clear(ctx context.Context) error
}
