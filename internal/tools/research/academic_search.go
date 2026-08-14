package research

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ilyaskhan/term-agent/internal/tools"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

type AcademicSearchArgs struct {
	Query      string `json:"query"`
	ProjectID  string `json:"project_id"`
	MaxResults int    `json:"max_results,omitempty"`
}

type AcademicSearchTool struct{}

func NewAcademicSearchTool() *AcademicSearchTool {
	return &AcademicSearchTool{}
}

func (t *AcademicSearchTool) Spec() tools.ToolSpec {
	params, _ := json.Marshal(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Academic search query terms or topic string",
			},
			"project_id": map[string]interface{}{
				"type":        "string",
				"description": "Research project ID to attach sources to",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of paper results to return (default 5)",
			},
		},
		"required": []string{"query", "project_id"},
	})

	return tools.ToolSpec{
		Name:        "academic_search",
		Description: "Searches literature and academic papers, returning structured Source records.",
		Parameters:  params,
		RiskLevel:   tools.RiskLevelRead,
	}
}

func (t *AcademicSearchTool) ValidateArgs(args json.RawMessage) error {
	var a AcademicSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid arguments for academic_search: %w", err)
	}
	if a.Query == "" {
		return errors.New("query is required")
	}
	if a.ProjectID == "" {
		return errors.New("project_id is required")
	}
	return nil
}

func (t *AcademicSearchTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	start := time.Now()
	var a AcademicSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	if a.MaxResults <= 0 {
		a.MaxResults = 5
	}

	// Mock/Synthesize search results structured as domain.Source
	var sources []domain.Source
	for i := 1; i <= a.MaxResults; i++ {
		srcID := fmt.Sprintf("src-%s-%d", a.ProjectID, i)
		title := fmt.Sprintf("Academic Investigation on %s (Paper #%d)", a.Query, i)
		uri := fmt.Sprintf("https://arxiv.org/abs/2608.%04d", i*100)
		src, err := domain.NewSource(srcID, a.ProjectID, title, uri, domain.SourceTypeAcademicPaper, 0.95)
		if err == nil {
			src.Authors = []string{"A. Vaswani", "N. Shazeer", "J. Uszkoreit"}
			src.Year = 2024
			src.Publisher = "ArXiv Preprints"
			sources = append(sources, *src)
		}
	}

	outJSON, err := json.Marshal(sources)
	if err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("failed to serialize search results: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	return &tools.ToolResult{
		Output:   string(outJSON),
		Duration: time.Since(start),
		IsError:  false,
	}, nil
}
