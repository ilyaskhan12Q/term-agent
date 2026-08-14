package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ilyaskhan/term-agent/internal/workspace"
)

// ReadFileArgs represents arguments for the read_file tool.
type ReadFileArgs struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

// ReadFileTool implements Tool for reading workspace text files cleanly within security boundaries.
type ReadFileTool struct {
	workspaceRoot string
}

// NewReadFileTool creates a new ReadFileTool instance for a workspace root.
func NewReadFileTool(workspaceRoot string) *ReadFileTool {
	return &ReadFileTool{workspaceRoot: workspaceRoot}
}

// Spec returns JSON schema specification for read_file.
func (t *ReadFileTool) Spec() ToolSpec {
	return ToolSpec{
		Name:        "read_file",
		Description: "Reads text contents of a file inside the workspace. Supports optional line range pagination.",
		RiskLevel:   RiskLevelRead,
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "Relative path to the file inside workspace root"
				},
				"start_line": {
					"type": "integer",
					"description": "Optional 1-indexed starting line number"
				},
				"end_line": {
					"type": "integer",
					"description": "Optional 1-indexed ending line number"
				}
			},
			"required": ["path"]
		}`),
	}
}

// ValidateArgs checks if mandatory path argument is provided.
func (t *ReadFileTool) ValidateArgs(args json.RawMessage) error {
	var a ReadFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid arguments format: %w", err)
	}
	if a.Path == "" {
		return fmt.Errorf("path parameter is required")
	}
	return nil
}

// Execute performs the bounded file read inside workspace root.
func (t *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	start := time.Now()
	var a ReadFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	opts := workspace.ReadOptions{
		StartLine: a.StartLine,
		EndLine:   a.EndLine,
	}

	res, err := workspace.ReadWorkspaceFile(ctx, t.workspaceRoot, a.Path, opts)
	if err != nil {
		return &ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	outputJSON, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return &ToolResult{
			Output:   res.Content,
			Duration: time.Since(start),
			IsError:  false,
		}, nil
	}

	return &ToolResult{
		Output:   string(outputJSON),
		Duration: time.Since(start),
		IsError:  false,
	}, nil
}
