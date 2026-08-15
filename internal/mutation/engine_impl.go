package mutation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ilyaskhan/term-agent/internal/security"
	"github.com/ilyaskhan/term-agent/internal/workspace"
)

var (
	ErrConcurrencyConflict     = errors.New("optimistic concurrency conflict: file on disk was modified since hash was captured")
	ErrInvalidStateTransition  = errors.New("invalid transaction state transition")
	ErrWorkspaceBoundaryEscape = errors.New("file path escapes workspace boundary")
)

// DefaultMutationEngine implements MutationEngine and RollbackManager contracts.
type DefaultMutationEngine struct {
	mu            sync.RWMutex
	workspaceRoot string
}

// NewDefaultMutationEngine constructs a new DefaultMutationEngine instance.
func NewDefaultMutationEngine(workspaceRoot string) *DefaultMutationEngine {
	return &DefaultMutationEngine{
		workspaceRoot: workspaceRoot,
	}
}

// BeginTransaction initializes a new empty transaction.
func (e *DefaultMutationEngine) BeginTransaction(ctx context.Context, sessionID string) (*Transaction, error) {
	tx := &Transaction{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Status:    TxStatusProposed,
		Mutations: make([]*FileMutation, 0),
		Snapshots: make([]*FileSnapshot, 0),
		CreatedAt: time.Now(),
	}
	return tx, nil
}

// StageMutation validates security boundaries, verifies OCC hash, captures snapshot, and stages mutation.
func (e *DefaultMutationEngine) StageMutation(tx *Transaction, mut *FileMutation) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if tx.Status != TxStatusProposed && tx.Status != TxStatusValidated {
		return fmt.Errorf("%w: cannot stage mutation in state %s", ErrInvalidStateTransition, tx.Status)
	}

	absPath, err := security.ValidateWorkspacePath(e.workspaceRoot, mut.Path)
	if err != nil {
		tx.Status = TxStatusFailed
		return fmt.Errorf("%w: %s (%v)", ErrWorkspaceBoundaryEscape, mut.Path, err)
	}

	// Capture snapshot and check OCC (Optimistic Concurrency Control)
	snapshot := &FileSnapshot{
		Path:   mut.Path,
		Exists: false,
	}

	if content, err := os.ReadFile(absPath); err == nil {
		snapshot.Exists = true
		snapshot.Content = content
		snapshot.ContentHash = workspace.HashBytes(content)

		// OCC Check: Verify file hasn't changed since agent read it
		if mut.Type != MutationCreate && mut.OriginalHash != "" && mut.OriginalHash != snapshot.ContentHash {
			tx.Status = TxStatusConflict
			return fmt.Errorf("%w for path %s (expected %s, got %s)", ErrConcurrencyConflict, mut.Path, mut.OriginalHash, snapshot.ContentHash)
		}
	} else {
		// File does not exist
		if mut.Type == MutationModify {
			tx.Status = TxStatusFailed
			return fmt.Errorf("cannot modify non-existent file: %s", mut.Path)
		}
	}

	if mut.ID == "" {
		mut.ID = uuid.New().String()
	}
	mut.TransactionID = tx.ID

	tx.Mutations = append(tx.Mutations, mut)
	tx.Snapshots = append(tx.Snapshots, snapshot)
	tx.Status = TxStatusValidated

	return nil
}

// RequestApproval transitions valid transaction to WAITING_APPROVAL state.
func (e *DefaultMutationEngine) RequestApproval(tx *Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if tx.Status != TxStatusValidated {
		return fmt.Errorf("%w: cannot request approval from state %s", ErrInvalidStateTransition, tx.Status)
	}
	tx.Status = TxStatusWaitingApproval
	return nil
}

// ApproveTransaction transitions WAITING_APPROVAL to APPROVED state.
func (e *DefaultMutationEngine) ApproveTransaction(tx *Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if tx.Status != TxStatusWaitingApproval && tx.Status != TxStatusValidated {
		return fmt.Errorf("%w: cannot approve transaction from state %s", ErrInvalidStateTransition, tx.Status)
	}
	tx.Status = TxStatusApproved
	return nil
}

// RejectTransaction transitions transaction to REJECTED state.
func (e *DefaultMutationEngine) RejectTransaction(tx *Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tx.Status = TxStatusRejected
	return nil
}

// CommitTransaction applies all staged mutations atomically.
func (e *DefaultMutationEngine) CommitTransaction(ctx context.Context, tx *Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if tx.Status != TxStatusApproved && tx.Status != TxStatusValidated {
		return fmt.Errorf("%w: cannot commit transaction in state %s", ErrInvalidStateTransition, tx.Status)
	}

	tx.Status = TxStatusCommitting

	for i, mut := range tx.Mutations {
		absPath, err := security.ValidateWorkspacePath(e.workspaceRoot, mut.Path)
		if err != nil {
			e.rollbackInternal(tx)
			tx.Status = TxStatusFailed
			return fmt.Errorf("security validation failed during commit: %w", err)
		}

		// Re-verify OCC check right before applying changes
		if mut.Type != MutationCreate && mut.OriginalHash != "" {
			if diskBytes, err := os.ReadFile(absPath); err == nil {
				currentHash := workspace.HashBytes(diskBytes)
				if currentHash != mut.OriginalHash {
					tx.Status = TxStatusConflict
					return fmt.Errorf("%w for path %s (expected %s, got %s)", ErrConcurrencyConflict, mut.Path, mut.OriginalHash, currentHash)
				}
			} else if mut.Type == MutationModify {
				tx.Status = TxStatusFailed
				return fmt.Errorf("cannot modify non-existent file during commit: %s", mut.Path)
			}
		}

		_ = i // placeholder if index unused
		switch mut.Type {
		case MutationCreate, MutationModify:
			dir := filepath.Dir(absPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				e.rollbackInternal(tx)
				tx.Status = TxStatusFailed
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			if err := os.WriteFile(absPath, mut.NewContent, 0644); err != nil {
				e.rollbackInternal(tx)
				tx.Status = TxStatusFailed
				return fmt.Errorf("failed to write file %s: %w", absPath, err)
			}

		case MutationDelete:
			if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
				e.rollbackInternal(tx)
				tx.Status = TxStatusFailed
				return fmt.Errorf("failed to delete file %s: %w", absPath, err)
			}
		}
	}

	tx.Status = TxStatusCommitted
	return nil
}

// RollbackTransaction restores workspace files from snapshots.
func (e *DefaultMutationEngine) RollbackTransaction(ctx context.Context, tx *Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.rollbackInternal(tx)
}

func (e *DefaultMutationEngine) rollbackInternal(tx *Transaction) error {
	tx.Status = TxStatusRollingBack

	var lastErr error
	for i := len(tx.Snapshots) - 1; i >= 0; i-- {
		snap := tx.Snapshots[i]
		absPath, err := security.ValidateWorkspacePath(e.workspaceRoot, snap.Path)
		if err != nil {
			lastErr = err
			continue
		}

		if snap.Exists {
			if err := os.WriteFile(absPath, snap.Content, 0644); err != nil {
				lastErr = err
			}
		} else {
			if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
				lastErr = err
			}
		}
	}

	tx.Status = TxStatusRolledBack
	return lastErr
}
