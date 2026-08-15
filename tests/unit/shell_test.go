package unit_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ilyaskhan/term-agent/internal/security"
	"github.com/ilyaskhan/term-agent/internal/tools"
)

func TestExecuteShellToolSuccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-shell-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	policy := security.NewDefaultSecurityPolicy(security.FullPermissions())
	tool := tools.NewExecuteShellTool(tmpDir, policy)

	args := tools.ExecuteShellArgs{
		Command: "echo 'hello term-agent'",
	}
	argsBytes, _ := json.Marshal(args)

	res, err := tool.Execute(context.Background(), argsBytes)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error tool result, got error: %s", res.Error)
	}

	var resp tools.ShellExecutionResponse
	if err := json.Unmarshal([]byte(res.Output), &resp); err != nil {
		t.Fatalf("failed to parse shell tool response JSON: %v", err)
	}

	if resp.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", resp.ExitCode)
	}
	if !strings.Contains(resp.Stdout, "hello term-agent") {
		t.Errorf("expected stdout to contain 'hello term-agent', got '%s'", resp.Stdout)
	}
	if resp.Risk.RiskLevel != security.CommandRiskAllowed {
		t.Errorf("expected ALLOWED risk level, got %s", resp.Risk.RiskLevel)
	}
}

func TestExecuteShellToolBlockedBySecurityPolicy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-shell-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	policy := security.NewDefaultSecurityPolicy(security.DefaultPermissions())
	tool := tools.NewExecuteShellTool(tmpDir, policy)

	args := tools.ExecuteShellArgs{
		Command: "sudo rm -rf /",
	}
	argsBytes, _ := json.Marshal(args)

	res, err := tool.Execute(context.Background(), argsBytes)
	if err != nil {
		t.Fatalf("unexpected system execution error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected tool result to be flagged as error when blocked by policy")
	}
	if !strings.Contains(res.Error, "command blocked by security policy") {
		t.Errorf("expected security policy error message, got: %s", res.Error)
	}
}

func TestExecuteShellToolTimeoutHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-shell-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	policy := security.NewDefaultSecurityPolicy(security.FullPermissions())
	tool := tools.NewExecuteShellTool(tmpDir, policy)

	// Command sleeps for 5 seconds, timeout set to 1 second
	args := tools.ExecuteShellArgs{
		Command:        "sleep 5",
		TimeoutSeconds: 1,
	}
	argsBytes, _ := json.Marshal(args)

	start := time.Now()
	res, err := tool.Execute(context.Background(), argsBytes)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected system error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected tool result to indicate error on timeout")
	}
	if !strings.Contains(res.Error, "timed out") {
		t.Errorf("expected timeout message in error, got: %s", res.Error)
	}
	if duration > 3*time.Second {
		t.Errorf("execution took too long (%v), timeout should have terminated process earlier", duration)
	}
}

func TestExecuteShellToolWorkingDirectoryConstraint(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-shell-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subDir: %v", err)
	}

	policy := security.NewDefaultSecurityPolicy(security.FullPermissions())
	tool := tools.NewExecuteShellTool(tmpDir, policy)

	// Attempting working_dir outside workspace
	argsOutside := tools.ExecuteShellArgs{
		Command:    "pwd",
		WorkingDir: "../outside",
	}
	argsOutsideBytes, _ := json.Marshal(argsOutside)

	resOutside, _ := tool.Execute(context.Background(), argsOutsideBytes)
	if !resOutside.IsError {
		t.Fatalf("expected working_dir outside workspace to fail")
	}
	if !strings.Contains(resOutside.Error, "path traversal attack detected") && !strings.Contains(resOutside.Error, "invalid working directory") {
		t.Errorf("expected traversal error, got: %s", resOutside.Error)
	}

	// Valid working_dir inside workspace
	argsInside := tools.ExecuteShellArgs{
		Command:    "pwd",
		WorkingDir: "subdir",
	}
	argsInsideBytes, _ := json.Marshal(argsInside)

	resInside, err := tool.Execute(context.Background(), argsInsideBytes)
	if err != nil || resInside.IsError {
		t.Fatalf("unexpected error running inside working_dir: %v / error: %s", err, resInside.Error)
	}

	var resp tools.ShellExecutionResponse
	_ = json.Unmarshal([]byte(resInside.Output), &resp)
	evalDir, _ := filepath.EvalSymlinks(subDir)
	evalStdout, _ := filepath.EvalSymlinks(strings.TrimSpace(resp.Stdout))
	if evalStdout != evalDir && strings.TrimSpace(resp.Stdout) != subDir {
		t.Errorf("expected working dir %s or %s, got %s", subDir, evalDir, resp.Stdout)
	}
}

func TestExecuteShellToolOutputTruncation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-shell-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	policy := security.NewDefaultSecurityPolicy(security.FullPermissions())
	tool := tools.NewExecuteShellTool(tmpDir, policy)

	// Generate large output (> 1MB)
	args := tools.ExecuteShellArgs{
		Command: "yes 'This is a long repetitive output line for memory truncation testing' | head -n 50000",
	}
	argsBytes, _ := json.Marshal(args)

	res, err := tool.Execute(context.Background(), argsBytes)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	var resp tools.ShellExecutionResponse
	if err := json.Unmarshal([]byte(res.Output), &resp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if !resp.Truncated {
		t.Errorf("expected response output to be flagged as Truncated")
	}
	if !strings.Contains(resp.Stdout, "OUTPUT TRUNCATED") {
		t.Errorf("expected stdout to contain truncation notice")
	}
}
