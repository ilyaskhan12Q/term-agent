package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateWorkspacePath verifies that targetPath resides within workspaceRoot using lexical path resolution.
// Note: Full symlink resolution (filepath.EvalSymlinks) for existing files is evaluated at runtime in Phase 3/6.
func ValidateWorkspacePath(workspaceRoot, targetPath string) (string, error) {
	if workspaceRoot == "" {
		return "", fmt.Errorf("workspace root cannot be empty")
	}
	if targetPath == "" {
		return "", fmt.Errorf("target path cannot be empty")
	}

	absRoot, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("invalid workspace root: %w", err)
	}

	cleanTarget := filepath.Clean(targetPath)
	var fullPath string
	if filepath.IsAbs(cleanTarget) {
		fullPath = cleanTarget
	} else {
		fullPath = filepath.Join(absRoot, cleanTarget)
	}

	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	rel, err := filepath.Rel(absRoot, absFull)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("path traversal attack detected: %s escapes workspace root %s", targetPath, absRoot)
	}

	return absFull, nil
}
