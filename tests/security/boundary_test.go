package security_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/security"
)

func TestPathTraversalPrevention(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-sec-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	traversalPaths := []string{
		"../etc/passwd",
		"../../sensitive.key",
		"/etc/shadow",
		"foo/bar/../../../../../../tmp/escape",
	}

	for _, p := range traversalPaths {
		_, err := security.ValidateWorkspacePath(tmpDir, p)
		if err == nil {
			t.Errorf("SECURITY FAILURE: path traversal %s was allowed outside workspace root %s", p, tmpDir)
		}
	}
}

func TestSymlinkEscape(t *testing.T) {
	workspaceDir, err := os.MkdirTemp("", "termagent-sec-ws-*")
	if err != nil {
		t.Fatalf("failed to create temp ws dir: %v", err)
	}
	defer os.RemoveAll(workspaceDir)

	outsideDir, err := os.MkdirTemp("", "termagent-sec-outside-*")
	if err != nil {
		t.Fatalf("failed to create temp outside dir: %v", err)
	}
	defer os.RemoveAll(outsideDir)

	// Create secret file outside workspace
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("super-secret-key"), 0600); err != nil {
		t.Fatalf("failed to write outside file: %v", err)
	}

	// Create symlink inside workspace pointing outside
	symlinkPath := filepath.Join(workspaceDir, "symlink_secret.txt")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Validate path must fail
	_, err = security.ValidateWorkspacePath(workspaceDir, "symlink_secret.txt")
	if err == nil {
		t.Fatalf("SECURITY FAILURE: symlink pointing outside workspace (%s -> %s) was allowed!", symlinkPath, outsideFile)
	}
}

func TestDangerousCommandClassificationPlaceholder(t *testing.T) {
	// Security boundary placeholder for non-naive command classification (Phase 6 implementation)
	t.Log("Placeholder for non-naive command classification: commands like rm -rf, sudo, network pipes must be classified as BLOCKED or REQUIRES_USER")
}
