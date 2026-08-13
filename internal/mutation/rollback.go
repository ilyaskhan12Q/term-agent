package mutation

import (
	"context"
)

// RollbackManager defines the contract for undoing uncommitted or rejected transactions.
type RollbackManager interface {
	Rollback(ctx context.Context, tx *Transaction) error
}
