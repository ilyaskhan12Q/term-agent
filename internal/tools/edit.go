package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilyaskhan/term-agent/internal/diff"
	"github.com/ilyaskhan/term-agent/internal/mutation"
	"github.com/ilyaskhan/term-agent/internal/security"
	"github.com/ilyaskhan/term-agent/internal/workspace"
)

// EditFileArgs represents arguments for the edit_file tool.
type EditFileArgs struct {
	Path       string `json:"path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	BeforeHash string `json:"before_hash,omitempty"`
}

// EditFileTool implements Tool for targeted string block replacement in existing workspace text files.
type EditFileTool struct {
	engine        mutation.MutationEngine
	diffEngine    diff.DiffEngine
	workspaceRoot string
}

// NewEditFileTool constructs a new EditFileTool instance.
func NewEditFileTool(workspaceRoot string, engine mutation.MutationEngine) *EditFileTool {
	return &EditFileTool{
		engine:        engine,
		diffEngine:    diff.NewDefaultDiffEngine(),
		workspaceRoot: workspaceRoot,
	}
}

// Spec returns JSON schema specification for edit_file.
func (t *EditFileTool) Spec() ToolSpec {
	return ToolSpec{
		Name:        "edit_file",
		Description: "Edits an existing workspace file by replacing an exact block of text (old_string) with new text (new_string).",
		RiskLevel:   RiskLevelMutation,
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "Relative path to file inside workspace"
				},
				"old_string": {
					"type": "string",
					"description": "Exact text block to replace"
				},
				"new_string": {
					"type": "string",
					"description": "New text block to insert in place of old_string"
				},
				"before_hash": {
					"type": "string",
					"description": "Optional original SHA-256 hash of file for optimistic concurrency verification"
				}
			},
			"required": ["path", "old_string", "new_string"]
		}`),
	}
}

// ValidateArgs checks if mandatory parameters are present.
func (t *EditFileTool) ValidateArgs(args json.RawMessage) error {
	var a EditFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid arguments format: %w", err)
	}
	if a.Path == "" {
		return fmt.Errorf("path parameter is required")
	}
	if a.OldString == "" {
		return fmt.Errorf("old_string parameter is required")
	}
	return nil
}

// Execute performs target string block replacement within a transaction.
func (t *EditFileTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	start := time.Now()
	var a EditFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	// Read existing file content using ReadWorkspaceFile
	readRes, err := workspace.ReadWorkspaceFile(ctx, t.workspaceRoot, a.Path, workspace.ReadOptions{})
	if err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("failed to read target file for editing: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}
	currentContent := readRes.Content
	if !strings.Contains(currentContent, a.OldString) {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("target old_string was not found inside file: %s", a.Path),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	newContent := strings.Replace(currentContent, a.OldString, a.NewString, 1)

	sessionID := uuid.New().String()
	tx, err := t.engine.BeginTransaction(ctx, sessionID)
	if err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("failed to begin transaction: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	origHash := a.BeforeHash
	if origHash == "" {
		absPath, err := security.ValidateWorkspacePath(t.workspaceRoot, a.Path)
		if err == nil {
			if diskBytes, err := os.ReadFile(absPath); err == nil {
				origHash = workspace.HashBytes(diskBytes)
			}
		}
		if origHash == "" {
			origHash = workspace.HashBytes([]byte(currentContent))
		}
	}

	mut := &mutation.FileMutation{
		Path:         a.Path,
		Type:         mutation.MutationModify,
		OriginalHash: origHash,
		NewContent:   []byte(newContent),
	}

	if err := t.engine.StageMutation(tx, mut); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("mutation staging failed: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	fileDiff, _ := t.diffEngine.ComputeDiff([]byte(currentContent), []byte(newContent), a.Path, a.Path)
	diffText := diff.RenderUnifiedDiff(fileDiff)

	if err := t.engine.CommitTransaction(ctx, tx); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("transaction commit failed: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	output := fmt.Sprintf("Successfully committed edit to %s.\n\nDiff:\n%s", a.Path, diffText)
	return &ToolResult{
		Output:   output,
		Duration: time.Since(start),
		IsError:  false,
	}, nil
}
