package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	rtools "github.com/ilyaskhan/term-agent/internal/tools/research"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
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

func TestWebFetchTool(t *testing.T) {
	// Setup test server simulating web page response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><head><title>Test Research Page</title></head><body><h1>Main Section</h1><p>Transformers demonstrate state-of-the-art results on NLP tasks.</p></body></html>`))
	}))
	defer ts.Close()

	tool := rtools.NewWebFetchTool()
	tool.HTTPClient = ts.Client()

	if tool.Spec().Name != "web_fetch" {
		t.Errorf("expected tool name 'web_fetch', got '%s'", tool.Spec().Name)
	}

	validArgs := json.RawMessage(`{"uri": "` + ts.URL + `", "project_id": "proj-1"}`)
	if err := tool.ValidateArgs(validArgs); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	res, err := tool.Execute(context.Background(), validArgs)
	if err != nil {
		t.Fatalf("failed to execute web_fetch tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %s", res.Error)
	}

	var fetchResp rtools.WebFetchResponse
	if err := json.Unmarshal([]byte(res.Output), &fetchResp); err != nil {
		t.Fatalf("failed to parse web_fetch output: %v", err)
	}

	if fetchResp.Title != "Test Research Page" {
		t.Errorf("expected page title 'Test Research Page', got '%s'", fetchResp.Title)
	}

	if !strings.Contains(fetchResp.WrappedText, "<untrusted_content") {
		t.Errorf("expected security wrapping envelope around web page text")
	}
}

func TestPDFExtractorTool(t *testing.T) {
	// Create temporary PDF file for testing
	tmpFile, err := os.CreateTemp("", "test_paper_*.pdf")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("%PDF-1.4\n1 0 obj\n(Page 1 text (Transformers demonstrate state of the art accuracy.))\nendobj\n%%EOF\n")
	_ = tmpFile.Close()

	tool := rtools.NewPDFExtractorTool()

	validArgs := json.RawMessage(`{"project_id": "proj-1", "source_id": "src-1", "file_path": "` + tmpFile.Name() + `"}`)
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

	var pdfRes rtools.PDFExtractorResult
	if err := json.Unmarshal([]byte(res.Output), &pdfRes); err != nil {
		t.Fatalf("failed to parse pdf_extractor output: %v", err)
	}

	if pdfRes.TotalItems == 0 {
		t.Errorf("expected extracted evidence items")
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

	var verRes rtools.CitationVerificationResult
	if err := json.Unmarshal([]byte(res.Output), &verRes); err != nil {
		t.Fatalf("failed to parse citation_verifier output: %v", err)
	}

	if verRes.VerificationStatus != domain.EvidenceStatusVerified {
		t.Errorf("expected status VERIFIED, got '%s'", verRes.VerificationStatus)
	}
}

func TestProvenanceTrackerDeduplication(t *testing.T) {
	tracker := provenance.NewProvenanceTracker()

	src1, _ := domain.NewSource("src-1", "proj-1", "Attention is All You Need", "https://arxiv.org/abs/1706.03762", domain.SourceTypeAcademicPaper, 0.95)
	src2, _ := domain.NewSource("src-2", "proj-1", "Attention is All You Need Duplicate", "https://arxiv.org/abs/1706.03762", domain.SourceTypeAcademicPaper, 0.95)

	id1, err := tracker.RegisterSource(*src1)
	if err != nil {
		t.Fatalf("failed to register source 1: %v", err)
	}

	id2, err := tracker.RegisterSource(*src2)
	if err != nil {
		t.Fatalf("failed to register source 2: %v", err)
	}

	if id1 != id2 {
		t.Errorf("expected URI deduplication to return existing source ID '%s', got '%s'", id1, id2)
	}
}
