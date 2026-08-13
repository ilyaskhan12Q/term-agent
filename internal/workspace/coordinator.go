package workspace

import (
	"context"
)

// Coordinator coordinates concurrent access and operations on a workspace.
type Coordinator interface {
	AcquireLock(ctx context.Context, path string) (UnlockFunc, error)
	GetWorkspace(path string) (Workspace, error)
}

// UnlockFunc releases an acquired file lock.
type UnlockFunc func()
