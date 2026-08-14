package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/workspace"
)

// SearchWorkspaceArgs represents arguments for the search_workspace tool.
type SearchWorkspaceArgs struct {
	Query           string `json:"query"`
	IsRegex         bool   `json:"is_regex,omitempty"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty"`
	Extension       string `json:"extension,omitempty"`
	MaxMatches      int    `json:"max_matches,omitempty"` // Default 100
}

// SearchMatch represents a single matching line in a workspace file.
type SearchMatch struct {
	RelPath     string `json:"rel_path"`
	LineNumber  int    `json:"line_number"`
	LineContent string `json:"line_content"`
}

// SearchWorkspaceTool implements Tool for searching text files inside the workspace.
type SearchWorkspaceTool struct {
	workspaceRoot string
}

// NewSearchWorkspaceTool constructs a tool instance for workspace search.
func NewSearchWorkspaceTool(workspaceRoot string) *SearchWorkspaceTool {
	return &SearchWorkspaceTool{workspaceRoot: workspaceRoot}
}

// Spec returns JSON schema specification for search_workspace.
func (t *SearchWorkspaceTool) Spec() ToolSpec {
	return ToolSpec{
		Name:        "search_workspace",
		Description: "Searches text files in the workspace for literal text or regex patterns. Returns matching lines and line numbers.",
		RiskLevel:   RiskLevelRead,
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Search term or regular expression pattern"
				},
				"is_regex": {
					"type": "boolean",
					"description": "Treat query as a regular expression pattern"
				},
				"case_insensitive": {
					"type": "boolean",
					"description": "Perform case-insensitive matching"
				},
				"extension": {
					"type": "string",
					"description": "Optional file extension filter (e.g. '.go', '.md')"
				},
				"max_matches": {
					"type": "integer",
					"description": "Maximum number of total matches to return (default 100)"
				}
			},
			"required": ["query"]
		}`),
	}
}

// ValidateArgs checks if mandatory query argument is provided.
func (t *SearchWorkspaceTool) ValidateArgs(args json.RawMessage) error {
	var a SearchWorkspaceArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid arguments format: %w", err)
	}
	if a.Query == "" {
		return fmt.Errorf("query parameter is required")
	}
	return nil
}

// Execute performs workspace search across indexed text files.
func (t *SearchWorkspaceTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	start := time.Now()
	var a SearchWorkspaceArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	maxMatches := a.MaxMatches
	if maxMatches <= 0 {
		maxMatches = 100
	}

	ext := strings.ToLower(a.Extension)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	var re *regexp.Regexp
	var err error

	pattern := a.Query
	if a.CaseInsensitive {
		pattern = "(?i)" + pattern
	}

	if a.IsRegex {
		re, err = regexp.Compile(pattern)
		if err != nil {
			return &ToolResult{
				Output:   "",
				Error:    fmt.Sprintf("invalid regex pattern: %v", err),
				Duration: time.Since(start),
				IsError:  true,
			}, nil
		}
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

	var matches []SearchMatch
	totalMatches := 0

fileLoop:
	for _, f := range files {
		if f.IsBinary {
			continue
		}

		if ext != "" && strings.ToLower(filepath.Ext(f.RelPath)) != ext {
			continue
		}

		fileObj, err := os.Open(f.AbsPath)
		if err != nil {
			continue
		}

		fileScanner := bufio.NewScanner(fileObj)
		buf := make([]byte, 64*1024)
		fileScanner.Buffer(buf, 1024*1024)

		lineNum := 0
		for fileScanner.Scan() {
			select {
			case <-ctx.Done():
				fileObj.Close()
				break fileLoop
			default:
			}

			lineNum++
			lineText := fileScanner.Text()

			matched := false
			if a.IsRegex && re != nil {
				matched = re.MatchString(lineText)
			} else {
				if a.CaseInsensitive {
					matched = strings.Contains(strings.ToLower(lineText), strings.ToLower(a.Query))
				} else {
					matched = strings.Contains(lineText, a.Query)
				}
			}

			if matched {
				matches = append(matches, SearchMatch{
					RelPath:     f.RelPath,
					LineNumber:  lineNum,
					LineContent: strings.TrimSpace(lineText),
				})
				totalMatches++

				if totalMatches >= maxMatches {
					fileObj.Close()
					break fileLoop
				}
			}
		}
		fileObj.Close()
	}

	outputJSON, err := json.MarshalIndent(matches, "", "  ")
	if err != nil {
		return &ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("serialization error: %v", err),
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
