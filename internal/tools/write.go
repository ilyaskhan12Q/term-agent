package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ilyaskhan/term-agent/internal/diff"
	"github.com/ilyaskhan/term-agent/internal/mutation"
)

// WriteFileArgs represents arguments for the write_file tool.
type WriteFileArgs struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	BeforeHash string `json:"before_hash,omitempty"`
}

// WriteFileTool implements Tool for transactional file writes and creations inside workspace.
type WriteFileTool struct {
	engine        mutation.MutationEngine
	diffEngine    diff.DiffEngine
	workspaceRoot string
}

// NewWriteFileTool constructs a new WriteFileTool instance.
func NewWriteFileTool(workspaceRoot string, engine mutation.MutationEngine) *WriteFileTool {
	return &WriteFileTool{
		engine:        engine,
		diffEngine:    diff.NewDefaultDiffEngine(),
		workspaceRoot: workspaceRoot,
	}
}

// Spec returns JSON schema specification for write_file.
func (t *WriteFileTool) Spec() ToolSpec {
	return ToolSpec{
		Name:        "write_file",
		Description: "Creates a new file or overwrites an existing file inside the workspace safely using transactional execution.",
		RiskLevel:   RiskLevelMutation,
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "Relative path to file inside workspace"
				},
				"content": {
					"type": "string",
					"description": "Full text content to write to file"
				},
				"before_hash": {
					"type": "string",
					"description": "Optional original SHA-256 hash of file for optimistic concurrency verification"
				}
			},
			"required": ["path", "content"]
		}`),
	}
}

// ValidateArgs checks if path and content arguments are valid.
func (t *WriteFileTool) ValidateArgs(args json.RawMessage) error {
	var a WriteFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid arguments format: %w", err)
	}
	if a.Path == "" {
		return fmt.Errorf("path parameter is required")
	}
	return nil
}

// Execute runs transactional file creation / write operation.
func (t *WriteFileTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	start := time.Now()
	var a WriteFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

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

	mutType := mutation.MutationModify
	if a.BeforeHash == "" {
		mutType = mutation.MutationCreate
	}

	mut := &mutation.FileMutation{
		Path:         a.Path,
		Type:         mutType,
		OriginalHash: a.BeforeHash,
		NewContent:   []byte(a.Content),
	}

	if err := t.engine.StageMutation(tx, mut); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("mutation staging failed: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	// Calculate unified diff for tool output transparency
	var oldContent []byte
	if len(tx.Snapshots) > 0 && tx.Snapshots[0].Exists {
		oldContent = tx.Snapshots[0].Content
	}
	fileDiff, _ := t.diffEngine.ComputeDiff(oldContent, []byte(a.Content), a.Path, a.Path)
	diffText := diff.RenderUnifiedDiff(fileDiff)

	// Commit transaction
	if err := t.engine.CommitTransaction(ctx, tx); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("transaction commit failed: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	output := fmt.Sprintf("Successfully committed write to %s.\n\nDiff:\n%s", a.Path, diffText)
	return &ToolResult{
		Output:   output,
		Duration: time.Since(start),
		IsError:  false,
	}, nil
}
