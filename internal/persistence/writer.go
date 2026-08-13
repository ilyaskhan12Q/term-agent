package persistence

import (
	"context"
)

// WriteTask represents an asynchronous database write operation.
type WriteTask func(ctx context.Context, db *DB) error

// AsyncWriter defines the contract for controlled async database writes to avoid blocking TUI rendering.
type AsyncWriter interface {
	Enqueue(task WriteTask) error
	Close(ctx context.Context) error
}
