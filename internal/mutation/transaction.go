package mutation

import (
	"context"
	"time"
)

// TransactionStatus represents the state of a mutation transaction.
type TransactionStatus string

const (
	TxStatusProposed        TransactionStatus = "PROPOSED"
	TxStatusValidated       TransactionStatus = "VALIDATED"
	TxStatusWaitingApproval TransactionStatus = "WAITING_APPROVAL"
	TxStatusApproved        TransactionStatus = "APPROVED"
	TxStatusCommitting      TransactionStatus = "COMMITTING"
	TxStatusCommitted       TransactionStatus = "COMMITTED"
	TxStatusRejected        TransactionStatus = "REJECTED"
	TxStatusRollingBack     TransactionStatus = "ROLLING_BACK"
	TxStatusRolledBack      TransactionStatus = "ROLLED_BACK"
	TxStatusFailed          TransactionStatus = "FAILED"
	TxStatusConflict        TransactionStatus = "CONFLICT"
)

// Transaction represents a atomic batch of file mutations.
type Transaction struct {
	ID        string
	SessionID string
	Status    TransactionStatus
	Mutations []*FileMutation
	Snapshots []*FileSnapshot
	CreatedAt time.Time
}

// MutationEngine defines the contract for safe, transactional filesystem mutations.
type MutationEngine interface {
	BeginTransaction(ctx context.Context, sessionID string) (*Transaction, error)
	StageMutation(tx *Transaction, mut *FileMutation) error
	CommitTransaction(ctx context.Context, tx *Transaction) error
	RollbackTransaction(ctx context.Context, tx *Transaction) error
}
