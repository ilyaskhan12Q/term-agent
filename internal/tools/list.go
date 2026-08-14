package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/workspace"
)

// ListWorkspaceArgs represents arguments for the list_workspace tool.
type ListWorkspaceArgs struct {
	Extension string `json:"extension,omitempty"` // e.g. ".go", ".json", ".md"
	MaxFiles  int    `json:"max_files,omitempty"` // Default 500
}

// ListWorkspaceTool implements Tool for discovering workspace files cleanly.
type ListWorkspaceTool struct {
	workspaceRoot string
}

// NewListWorkspaceTool constructs a tool instance for listing workspace files.
func NewListWorkspaceTool(workspaceRoot string) *ListWorkspaceTool {
	return &ListWorkspaceTool{workspaceRoot: workspaceRoot}
}

// Spec returns JSON schema specification for list_workspace.
func (t *ListWorkspaceTool) Spec() ToolSpec {
	return ToolSpec{
		Name:        "list_workspace",
		Description: "Lists non-ignored files inside the workspace with metadata (path, size, line count, binary flag).",
		RiskLevel:   RiskLevelRead,
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"extension": {
					"type": "string",
					"description": "Optional file extension filter (e.g. '.go', '.py')"
				},
				"max_files": {
					"type": "integer",
					"description": "Maximum number of files to return (default 500)"
				}
			}
		}`),
	}
}

// ValidateArgs validates list_workspace parameters.
func (t *ListWorkspaceTool) ValidateArgs(args json.RawMessage) error {
	var a ListWorkspaceArgs
	if len(args) > 0 {
		if err := json.Unmarshal(args, &a); err != nil {
			return fmt.Errorf("invalid arguments format: %w", err)
		}
	}
	return nil
}

// Execute scans workspace directory and returns filtered file metadata list.
func (t *ListWorkspaceTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	start := time.Now()
	var a ListWorkspaceArgs
	if len(args) > 0 {
		_ = json.Unmarshal(args, &a)
	}

	maxFiles := a.MaxFiles
	if maxFiles <= 0 {
		maxFiles = 500
	}

	scanner := workspace.NewDefaultScanner()
	files, err := scanner.DiscoverFiles(ctx, t.workspaceRoot)
	if err != nil {
		return &ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	ext := strings.ToLower(a.Extension)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	var filtered []workspace.FileInfo
	for _, f := range files {
		if ext != "" && strings.ToLower(filepath.Ext(f.RelPath)) != ext {
			continue
		}
		filtered = append(filtered, f)
		if len(filtered) >= maxFiles {
			break
		}
	}

	outputJSON, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("serialization failed: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	return &ToolResult{
		Output:   string(outputJSON),
		Duration: time.Since(start),
		IsError:  false,
	}, nil
}
