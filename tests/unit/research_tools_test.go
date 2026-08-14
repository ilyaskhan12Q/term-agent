package unit

import (
	"context"
	"encoding/json"
	"testing"

	rtools "github.com/ilyaskhan/term-agent/internal/tools/research"
)

func TestAcademicSearchTool(t *testing.T) {
	tool := rtools.NewAcademicSearchTool()
	if tool.Spec().Name != "academic_search" {
		t.Errorf("expected tool name 'academic_search', got '%s'", tool.Spec().Name)
	}

	invalidArgs := json.RawMessage(`{"query": ""}`)
	if err := tool.ValidateArgs(invalidArgs); err == nil {
		t.Errorf("expected validation error for empty query and missing project_id")
	}

	validArgs := json.RawMessage(`{"query": "Quantum Machine Learning", "project_id": "proj-1", "max_results": 2}`)
	if err := tool.ValidateArgs(validArgs); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	res, err := tool.Execute(context.Background(), validArgs)
	if err != nil {
		t.Fatalf("failed to execute academic_search tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %s", res.Error)
	}
	if res.Output == "" {
		t.Errorf("expected non-empty output")
	}
}

func TestPDFExtractorTool(t *testing.T) {
	tool := rtools.NewPDFExtractorTool()

	validArgs := json.RawMessage(`{"project_id": "proj-1", "source_id": "src-1", "file_path": "/tmp/paper.pdf"}`)
	if err := tool.ValidateArgs(validArgs); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	res, err := tool.Execute(context.Background(), validArgs)
	if err != nil {
		t.Fatalf("failed to execute pdf_extractor tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool execution error: %s", res.Error)
	}
}

func TestCitationVerifierTool(t *testing.T) {
	tool := rtools.NewCitationVerifierTool()

	validArgs := json.RawMessage(`{"claim_statement": "Transformers improve parallelization", "snippet": "Transformers allow significant parallelization during training", "evidence_id": "ev-1"}`)
	if err := tool.ValidateArgs(validArgs); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	res, err := tool.Execute(context.Background(), validArgs)
	if err != nil {
		t.Fatalf("failed to execute citation_verifier tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %s", res.Error)
	}
}
