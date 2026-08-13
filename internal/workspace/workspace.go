package workspace

import (
	"context"
)

// Workspace defines the contract for workspace state inspection and safety boundaries.
type Workspace interface {
	RootPath() string
	IsWithinBoundary(path string) bool
	ReadFile(ctx context.Context, relativePath string) ([]byte, error)
	ComputeStateHash(ctx context.Context) (string, error)
}
