package workspace

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ilyaskhan/term-agent/internal/security"
)

// DefaultWorkspace implements the Workspace interface with path safety and state hashing.
type DefaultWorkspace struct {
	rootPath string
	scanner  FileDiscovery
}

// NewDefaultWorkspace initializes a workspace instance for a root path.
func NewDefaultWorkspace(rootPath string) (*DefaultWorkspace, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace root path: %w", err)
	}

	return &DefaultWorkspace{
		rootPath: absRoot,
		scanner:  NewDefaultScanner(),
	}, nil
}

// RootPath returns the absolute workspace root path.
func (w *DefaultWorkspace) RootPath() string {
	return w.rootPath
}

// IsWithinBoundary returns true if the target path is legally inside the workspace root.
func (w *DefaultWorkspace) IsWithinBoundary(path string) bool {
	_, err := security.ValidateWorkspacePath(w.rootPath, path)
	return err == nil
}

// ReadFile reads the full contents of a file inside workspace bounds.
func (w *DefaultWorkspace) ReadFile(ctx context.Context, relativePath string) ([]byte, error) {
	res, err := ReadWorkspaceFile(ctx, w.rootPath, relativePath, ReadOptions{})
	if err != nil {
		return nil, err
	}
	return []byte(res.Content), nil
}

// ComputeStateHash calculates a SHA-256 hash representing the current workspace state.
func (w *DefaultWorkspace) ComputeStateHash(ctx context.Context) (string, error) {
	files, err := w.scanner.DiscoverFiles(ctx, w.rootPath)
	if err != nil {
		return "", err
	}
	return ComputeWorkspaceHash(files)
}

// DiscoverFiles returns all indexed non-ignored files inside the workspace.
func (w *DefaultWorkspace) DiscoverFiles(ctx context.Context) ([]FileInfo, error) {
	return w.scanner.DiscoverFiles(ctx, w.rootPath)
}
