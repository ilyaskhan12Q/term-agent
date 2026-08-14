package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateWorkspacePath verifies that targetPath resides within workspaceRoot using lexical and symlink resolution.
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

	// 1. Lexical check
	rel, err := filepath.Rel(absRoot, absFull)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("path traversal attack detected: %s escapes workspace root %s", targetPath, absRoot)
	}

	// 2. Symlink resolution check
	absRootReal, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		absRootReal = absRoot
	}

	absFullReal, err := resolveSymlinksOrExistingParent(absFull)
	if err != nil {
		return "", fmt.Errorf("symlink resolution failed: %w", err)
	}

	relReal, err := filepath.Rel(absRootReal, absFullReal)
	if err != nil || strings.HasPrefix(relReal, "..") || relReal == ".." {
		return "", fmt.Errorf("symlink escape attack detected: %s resolves to %s outside workspace root %s", targetPath, absFullReal, absRootReal)
	}

	return absFull, nil
}

// resolveSymlinksOrExistingParent evaluates symlinks for path or its deepest existing parent directory.
func resolveSymlinksOrExistingParent(path string) (string, error) {
	realPath, err := filepath.EvalSymlinks(path)
	if err == nil {
		return realPath, nil
	}

	// Path doesn't exist yet (e.g. creating a file). Find existing ancestor directory.
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	for dir != "." && dir != "/" {
		realDir, err := filepath.EvalSymlinks(dir)
		if err == nil {
			return filepath.Join(realDir, base), nil
		}
		base = filepath.Join(filepath.Base(dir), base)
		dir = filepath.Dir(dir)
	}

	return path, nil
}
