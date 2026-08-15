package security

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// SecurityPolicy defines the contract for runtime security validation.
type SecurityPolicy interface {
	ValidatePath(workspaceRoot string, targetPath string) (string, error)
	EvaluateCommand(ctx context.Context, command string) (*Classification, error)
	AuthorizeAction(action ActionType, resource string) (bool, error)
	IsSensitivePath(targetPath string) bool
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

// DefaultSecurityPolicy implements runtime security policy enforcement.
type DefaultSecurityPolicy struct {
	classifier  Classifier
	permissions PermissionSet
}

// NewDefaultSecurityPolicy creates a new SecurityPolicy instance.
func NewDefaultSecurityPolicy(perms PermissionSet) *DefaultSecurityPolicy {
	return &DefaultSecurityPolicy{
		classifier:  NewPOSIXClassifier(),
		permissions: perms,
	}
}

// ValidatePath checks if targetPath stays strictly within workspaceRoot boundaries.
func (p *DefaultSecurityPolicy) ValidatePath(workspaceRoot, targetPath string) (string, error) {
	if p.IsSensitivePath(targetPath) {
		return "", fmt.Errorf("security policy violation: access to sensitive path '%s' is blocked", targetPath)
	}
	return ValidateWorkspacePath(workspaceRoot, targetPath)
}

// EvaluateCommand evaluates risk classification of a shell command using POSIX AST inspection.
func (p *DefaultSecurityPolicy) EvaluateCommand(ctx context.Context, command string) (*Classification, error) {
	if !p.permissions.AllowShellExec {
		return &Classification{
			Command:        command,
			RiskLevel:      CommandRiskBlocked,
			Category:       "POLICY_DISABLED",
			RequiresPrompt: true,
			Reason:         "Shell execution is disabled by active permission set",
		}, nil
	}

	return p.classifier.Classify(command)
}

// AuthorizeAction checks permission set for the specified action type and resource.
func (p *DefaultSecurityPolicy) AuthorizeAction(action ActionType, resource string) (bool, error) {
	switch action {
	case ActionFileRead:
		if p.IsSensitivePath(resource) {
			return false, fmt.Errorf("sensitive file read denied: %s", resource)
		}
		return true, nil
	case ActionFileWrite:
		if !p.permissions.AllowFileWrite {
			return false, fmt.Errorf("file write permission disabled")
		}
		if p.IsSensitivePath(resource) {
			return false, fmt.Errorf("sensitive file write denied: %s", resource)
		}
		return true, nil
	case ActionFileDelete:
		if !p.permissions.AllowFileDelete {
			return false, fmt.Errorf("file deletion permission disabled")
		}
		if p.IsSensitivePath(resource) {
			return false, fmt.Errorf("sensitive file deletion denied: %s", resource)
		}
		return true, nil
	case ActionShellExec:
		if !p.permissions.AllowShellExec {
			return false, fmt.Errorf("shell execution permission disabled")
		}
		return true, nil
	case ActionNetAccess:
		if !p.permissions.AllowNetwork {
			return false, fmt.Errorf("network access permission disabled")
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown action type: %s", action)
	}
}

// IsSensitivePath checks if path references known sensitive files (.env, ~/.ssh, credentials, etc.).
func (p *DefaultSecurityPolicy) IsSensitivePath(targetPath string) bool {
	clean := filepath.Clean(strings.ToLower(targetPath))
	base := filepath.Base(clean)

	if strings.HasPrefix(base, ".env") {
		return true
	}
	if strings.Contains(clean, ".ssh") || base == "id_rsa" || base == "id_ed25519" || base == "id_dsa" || base == "id_ecdsa" {
		return true
	}
	if strings.Contains(clean, ".aws") || strings.Contains(clean, ".gnupg") {
		return true
	}
	if clean == "/etc/passwd" || clean == "/etc/shadow" || clean == "/etc/sudoers" {
		return true
	}

	return false
}
