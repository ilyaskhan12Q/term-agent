package security

import (
	"context"
)

// SecurityPolicy defines the contract for runtime security validation.
type SecurityPolicy interface {
	ValidatePath(workspaceRoot string, targetPath string) error
	EvaluateCommand(ctx context.Context, command string) (*Classification, error)
	AuthorizeAction(action ActionType, resource string) (bool, error)
}

// ActionType represents types of actions subject to policy checks.
type ActionType string

const (
	ActionFileRead   ActionType = "FILE_READ"
	ActionFileWrite  ActionType = "FILE_WRITE"
	ActionFileDelete ActionType = "FILE_DELETE"
	ActionShellExec  ActionType = "SHELL_EXEC"
	ActionNetAccess  ActionType = "NET_ACCESS"
)
