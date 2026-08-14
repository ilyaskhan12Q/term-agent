package workspace

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ilyaskhan/term-agent/internal/security"
)

// DefaultMaxReadFileBytes defines the upper memory bound for file reading (5MB).
const DefaultMaxReadFileBytes int64 = 5 * 1024 * 1024

// ReadOptions defines parameters for reading file contents inside workspace bounds.
type ReadOptions struct {
	StartLine int   // 1-indexed (0 defaults to line 1)
	EndLine   int   // 1-indexed (0 reads to EOF or max limit)
	MaxBytes  int64 // 0 defaults to DefaultMaxReadFileBytes
}

// ReadResult represents the bounded file content payload returned to tools or agents.
type ReadResult struct {
	RelPath     string `json:"rel_path"`
	AbsPath     string `json:"abs_path"`
	TotalLines  int    `json:"total_lines"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	Content     string `json:"content"`
	IsTruncated bool   `json:"is_truncated"`
	SizeBytes   int64  `json:"size_bytes"`
}

// ReadWorkspaceFile safely validates workspace boundaries, checks binary type, and reads line-bounded text content.
func ReadWorkspaceFile(ctx context.Context, workspaceRoot, relPath string, opts ReadOptions) (*ReadResult, error) {
	absPath, err := security.ValidateWorkspacePath(workspaceRoot, relPath)
	if err != nil {
		return nil, fmt.Errorf("security error: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("target path %s is a directory, not a file", relPath)
	}

	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxReadFileBytes
	}

	if info.Size() > maxBytes && opts.StartLine == 0 && opts.EndLine == 0 {
		return nil, fmt.Errorf("file %s size (%d bytes) exceeds maximum limit (%d bytes); specify line range or increase limit", relPath, info.Size(), maxBytes)
	}

	if isBinaryFile(absPath) {
		return nil, fmt.Errorf("cannot read binary file %s as text", relPath)
	}

	f, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	startLine := opts.StartLine
	if startLine <= 0 {
		startLine = 1
	}

	endLine := opts.EndLine

	scanner := bufio.NewScanner(f)
	// Support long lines up to 1MB per line
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var sb strings.Builder
	currentLine := 0
	readLines := 0
	totalReadBytes := int64(0)
	isTruncated := false

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		currentLine++

		if currentLine < startLine {
			continue
		}

		if endLine > 0 && currentLine > endLine {
			break
		}

		lineText := scanner.Text()
		lineBytes := int64(len(lineText) + 1)

		if totalReadBytes+lineBytes > maxBytes {
			isTruncated = true
			break
		}

		if readLines > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(lineText)

		totalReadBytes += lineBytes
		readLines++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	actualEndLine := currentLine
	if endLine > 0 && currentLine > endLine {
		actualEndLine = endLine
	}

	return &ReadResult{
		RelPath:     relPath,
		AbsPath:     absPath,
		TotalLines:  currentLine,
		StartLine:   startLine,
		EndLine:     actualEndLine,
		Content:     sb.String(),
		IsTruncated: isTruncated,
		SizeBytes:   info.Size(),
	}, nil
}
