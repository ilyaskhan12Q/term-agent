package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/config"
	"github.com/ilyaskhan/term-agent/internal/context"
	"github.com/ilyaskhan/term-agent/internal/security"
	"github.com/ilyaskhan/term-agent/internal/workspace"
)

func TestConfigDefaults(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.DefaultProvider != "openai" {
		t.Errorf("expected default provider openai, got %s", cfg.DefaultProvider)
	}
	if cfg.MaxParallelWorkers != 4 {
		t.Errorf("expected 4 workers, got %d", cfg.MaxParallelWorkers)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("config validation failed: %v", err)
	}
}

func TestTokenEstimation(t *testing.T) {
	estimator := &context.SimpleEstimator{}
	text := "Hello term-agent world"
	tokens := estimator.CountTokens(text)
	if tokens <= 0 {
		t.Errorf("expected positive token count for text, got %d", tokens)
	}
}

func TestWorkspacePathValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	validFile := filepath.Join(tmpDir, "sub", "file.txt")
	validated, err := security.ValidateWorkspacePath(tmpDir, validFile)
	if err != nil {
		t.Errorf("expected valid path validation, got err: %v", err)
	}
	if validated == "" {
		t.Error("expected non-empty validated path")
	}

	// Test traversal attempt
	_, err = security.ValidateWorkspacePath(tmpDir, "../../../etc/passwd")
	if err == nil {
		t.Error("expected path traversal detection error, got nil")
	}
}

func TestStateHashing(t *testing.T) {
	hash := workspace.HashBytes([]byte("test content"))
	if len(hash) != 64 { // SHA-256 hex string length
		t.Errorf("expected 64 character hex hash, got length %d", len(hash))
	}
}
