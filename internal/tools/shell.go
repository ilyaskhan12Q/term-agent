package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/ilyaskhan/term-agent/internal/security"
)

const (
	DefaultShellTimeout = 30 * time.Second
	MaxShellTimeout     = 300 * time.Second
	MaxOutputBytes      = 1024 * 1024 // 1 MB buffer limit
)

// ShellToolSpec defines the schema for the execute_shell tool.
var ShellToolSpec = ToolSpec{
	Name:        "execute_shell",
	Description: "Executes a shell command inside the workspace after POSIX security policy evaluation.",
	RiskLevel:   RiskLevelShell,
	Parameters: json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "Shell command line to execute"
			},
			"timeout_seconds": {
				"type": "integer",
				"description": "Optional execution timeout limit in seconds (default 30s, max 300s)"
			},
			"working_dir": {
				"type": "string",
				"description": "Optional working directory relative to workspace root"
			}
		},
		"required": ["command"]
	}`),
}

// ExecuteShellArgs represents input parameters for execute_shell.
type ExecuteShellArgs struct {
	Command        string `json:"command"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	WorkingDir     string `json:"working_dir,omitempty"`
}

// ShellExecutionResponse represents structured output from shell execution.
type ShellExecutionResponse struct {
	Command   string                   `json:"command"`
	ExitCode  int                      `json:"exit_code"`
	Stdout    string                   `json:"stdout"`
	Stderr    string                   `json:"stderr"`
	Risk      *security.Classification `json:"security_classification"`
	Duration  string                   `json:"duration"`
	Truncated bool                     `json:"truncated"`
}

// ExecuteShellTool implements Tool for secure shell execution inside the workspace boundary.
type ExecuteShellTool struct {
	workspaceRoot string
	policy        security.SecurityPolicy
}

// NewExecuteShellTool constructs a new ExecuteShellTool.
func NewExecuteShellTool(workspaceRoot string, policy security.SecurityPolicy) *ExecuteShellTool {
	if policy == nil {
		policy = security.NewDefaultSecurityPolicy(security.DefaultPermissions())
	}
	return &ExecuteShellTool{
		workspaceRoot: workspaceRoot,
		policy:        policy,
	}
}

// Spec returns the JSON tool specification.
func (t *ExecuteShellTool) Spec() ToolSpec {
	return ShellToolSpec
}

// ValidateArgs checks if mandatory parameters are present.
func (t *ExecuteShellTool) ValidateArgs(args json.RawMessage) error {
	var a ExecuteShellArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid shell execution arguments: %w", err)
	}
	if a.Command == "" {
		return fmt.Errorf("command parameter is required")
	}
	return nil
}

// Execute evaluates security policy and runs the command within bounded limits.
func (t *ExecuteShellTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	start := time.Now()
	var a ExecuteShellArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	// 1. Resolve & Validate Working Directory
	targetDir := t.workspaceRoot
	if a.WorkingDir != "" {
		validatedDir, err := t.policy.ValidatePath(t.workspaceRoot, a.WorkingDir)
		if err != nil {
			return &ToolResult{
				Output:   "",
				Error:    fmt.Sprintf("invalid working directory: %v", err),
				Duration: time.Since(start),
				IsError:  true,
			}, nil
		}
		targetDir = validatedDir
	}

	// 2. Evaluate Command Security Classification
	classification, err := t.policy.EvaluateCommand(ctx, a.Command)
	if err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("security policy evaluation error: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	if classification.RiskLevel == security.CommandRiskBlocked {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("command blocked by security policy: %s", classification.Reason),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	// 3. Compute Bounded Execution Timeout
	timeout := DefaultShellTimeout
	if a.TimeoutSeconds > 0 {
		requestedTimeout := time.Duration(a.TimeoutSeconds) * time.Second
		if requestedTimeout > MaxShellTimeout {
			timeout = MaxShellTimeout
		} else {
			timeout = requestedTimeout
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 4. Set Up Command Process
	cmd := exec.CommandContext(execCtx, "sh", "-c", a.Command)
	cmd.Dir = targetDir
	cmd.WaitDelay = 200 * time.Millisecond

	stdoutBuf := &boundedBuffer{maxSize: MaxOutputBytes}
	stderrBuf := &boundedBuffer{maxSize: MaxOutputBytes}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	runErr := cmd.Run()
	duration := time.Since(start)

	// Check timeout
	if execCtx.Err() == context.DeadlineExceeded {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("command execution timed out after %v (limit: %v)", duration.Round(time.Millisecond), timeout),
			Duration: duration,
			IsError:  true,
		}, nil
	}

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	resp := ShellExecutionResponse{
		Command:   a.Command,
		ExitCode:  exitCode,
		Stdout:    stdoutBuf.String(),
		Stderr:    stderrBuf.String(),
		Risk:      classification,
		Duration:  duration.Round(time.Millisecond).String(),
		Truncated: stdoutBuf.truncated || stderrBuf.truncated,
	}

	respBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return &ToolResult{
			Output:   stdoutBuf.String(),
			Error:    stderrBuf.String(),
			Duration: duration,
			IsError:  exitCode != 0,
		}, nil
	}

	return &ToolResult{
		Output:   string(respBytes),
		Duration: duration,
		IsError:  exitCode != 0,
	}, nil
}

// boundedBuffer caps captured buffer size to prevent memory explosion.
type boundedBuffer struct {
	buf       bytes.Buffer
	maxSize   int
	truncated bool
}

func (b *boundedBuffer) Write(p []byte) (n int, err error) {
	if b.buf.Len()+len(p) > b.maxSize {
		allowed := b.maxSize - b.buf.Len()
		if allowed > 0 {
			b.buf.Write(p[:allowed])
		}
		b.truncated = true
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *boundedBuffer) String() string {
	s := b.buf.String()
	if b.truncated {
		s += "\n[OUTPUT TRUNCATED: Exceeded 1MB limit]"
	}
	return s
}

var _ io.Writer = (*boundedBuffer)(nil)
