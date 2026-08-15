package security_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/security"
)

func TestDefaultSecurityPolicySensitivePaths(t *testing.T) {
	policy := security.NewDefaultSecurityPolicy(security.DefaultPermissions())

	sensitive := []string{
		".env",
		".env.production",
		"config/.env.local",
		"~/.ssh/id_rsa",
		"id_ed25519",
		"/etc/shadow",
		"/etc/passwd",
		"~/.aws/credentials",
	}

	for _, p := range sensitive {
		if !policy.IsSensitivePath(p) {
			t.Errorf("SECURITY FAILURE: path '%s' was not flagged as sensitive", p)
		}

		_, err := policy.ValidatePath("/workspace", p)
		if err == nil {
			t.Errorf("SECURITY FAILURE: ValidatePath allowed sensitive path '%s'", p)
		}
	}
}

func TestDefaultSecurityPolicyActionAuthorization(t *testing.T) {
	perms := security.PermissionSet{
		AllowFileWrite:  true,
		AllowFileDelete: false,
		AllowShellExec:  true,
		AllowNetwork:    false,
	}

	policy := security.NewDefaultSecurityPolicy(perms)

	// Normal write should be authorized
	ok, err := policy.AuthorizeAction(security.ActionFileWrite, "main.go")
	if !ok || err != nil {
		t.Fatalf("expected file write to be authorized, got ok=%v, err=%v", ok, err)
	}

	// Sensitive write should be denied
	ok, err = policy.AuthorizeAction(security.ActionFileWrite, ".env")
	if ok || err == nil {
		t.Fatalf("expected sensitive file write to be denied")
	}

	// File delete should be denied by permissions
	ok, err = policy.AuthorizeAction(security.ActionFileDelete, "file.txt")
	if ok || err == nil {
		t.Fatalf("expected file delete to be denied by permissions")
	}

	// Network access should be denied
	ok, err = policy.AuthorizeAction(security.ActionNetAccess, "https://api.github.com")
	if ok || err == nil {
		t.Fatalf("expected network access to be denied by permissions")
	}
}

func TestDefaultSecurityPolicyCommandEvaluation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-pol-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	policy := security.NewDefaultSecurityPolicy(security.DefaultPermissions())
	ctx := context.Background()

	// Safe command
	res, err := policy.EvaluateCommand(ctx, "ls -la")
	if err != nil || res.RiskLevel != security.CommandRiskAllowed {
		t.Errorf("expected ls -la to be ALLOWED, got %s (err=%v)", res.RiskLevel, err)
	}

	// Blocked command
	res, err = policy.EvaluateCommand(ctx, "sudo rm -rf /")
	if err != nil || res.RiskLevel != security.CommandRiskBlocked {
		t.Errorf("expected sudo rm -rf / to be BLOCKED, got %s", res.RiskLevel)
	}

	// Path escape check in policy ValidatePath
	_, err = policy.ValidatePath(tmpDir, "../outside.txt")
	if err == nil {
		t.Errorf("expected path escape to return error")
	}

	// Normal inside path check
	validPath, err := policy.ValidatePath(tmpDir, "sub/file.txt")
	if err != nil {
		t.Errorf("unexpected error validating inside path: %v", err)
	}
	expectedPath := filepath.Join(tmpDir, "sub/file.txt")
	if validPath != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, validPath)
	}
}
