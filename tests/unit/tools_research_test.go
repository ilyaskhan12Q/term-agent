package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rtools "github.com/ilyaskhan/term-agent/internal/tools/research"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

// ---------------------------------------------------------------------------
// AcademicSearchTool Tests
// ---------------------------------------------------------------------------

func TestAcademicSearchTool_ArXivAPI(t *testing.T) {
	atomResponse := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <id>http://arxiv.org/abs/1706.03762v7</id>
    <title>Attention Is All You Need</title>
    <summary>The dominant sequence transduction models are based on complex recurrent or convolutional neural networks...</summary>
    <published>2017-06-12T00:00:00Z</published>
    <author><name>Ashish Vaswani</name></author>
    <author><name>Noam Shazeer</name></author>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/2005.14165v4</id>
    <title>Language Models are Few-Shot Learners</title>
    <summary>Recent work has demonstrated substantial gains on many NLP tasks and benchmarks...</summary>
    <published>2020-05-28T00:00:00Z</published>
    <author><name>Tom B. Brown</name></author>
  </entry>
</feed>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(atomResponse))
	}))
	defer server.Close()

	tool := rtools.NewAcademicSearchToolWithURL(server.URL)
	args, _ := json.Marshal(rtools.AcademicSearchArgs{
		Query:      "transformer models",
		ProjectID:  "proj-123",
		MaxResults: 2,
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error executing academic search: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error tool result, got error: %s", res.Error)
	}

	var sources []domain.Source
	if err := json.Unmarshal([]byte(res.Output), &sources); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if sources[0].Title != "Attention Is All You Need" {
		t.Errorf("unexpected title for first paper: %q", sources[0].Title)
	}
	if sources[0].Year != 2017 {
		t.Errorf("expected year 2017, got %d", sources[0].Year)
	}
	if len(sources[0].Authors) != 2 {
		t.Errorf("expected 2 authors, got %d", len(sources[0].Authors))
	}
}

func TestAcademicSearchTool_Fallback(t *testing.T) {
	// Point to non-existent server to trigger network error fallback
	tool := rtools.NewAcademicSearchToolWithURL("http://127.0.0.1:59999/invalid")
	args, _ := json.Marshal(rtools.AcademicSearchArgs{
		Query:      "quantum computing",
		ProjectID:  "proj-456",
		MaxResults: 3,
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error result: %s", res.Error)
	}

	var sources []domain.Source
	if err := json.Unmarshal([]byte(res.Output), &sources); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if len(sources) != 3 {
		t.Fatalf("expected 3 fallback sources, got %d", len(sources))
	}
	if !strings.Contains(sources[0].Title, "quantum computing") {
		t.Errorf("fallback source title missing query: %q", sources[0].Title)
	}
}

func TestAcademicSearchTool_Validation(t *testing.T) {
	tool := rtools.NewAcademicSearchTool()

	// Missing query
	args1, _ := json.Marshal(map[string]string{"project_id": "proj-1"})
	if err := tool.ValidateArgs(args1); err == nil {
		t.Error("expected error for missing query")
	}

	// Missing project_id
	args2, _ := json.Marshal(map[string]string{"query": "scaling"})
	if err := tool.ValidateArgs(args2); err == nil {
		t.Error("expected error for missing project_id")
	}
}

// ---------------------------------------------------------------------------
// WebFetchTool Tests
// ---------------------------------------------------------------------------

func TestWebFetchTool_Success(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head><title>Deep Learning Overview</title></head>
<body>
  <script>alert("malicious script");</script>
  <style>body { color: red; }</style>
  <h1>Deep Learning Overview</h1>
  <p>Neural networks are computing systems inspired by biological neural networks.</p>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	tool := rtools.NewWebFetchToolWithClient(server.Client())
	args, _ := json.Marshal(rtools.WebFetchArgs{
		URI:       server.URL,
		ProjectID: "proj-web",
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error executing web_fetch: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error tool result, got error: %s", res.Error)
	}

	var fetchRes rtools.WebFetchResponse
	if err := json.Unmarshal([]byte(res.Output), &fetchRes); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if fetchRes.Title != "Deep Learning Overview" {
		t.Errorf("unexpected extracted title: %q", fetchRes.Title)
	}
	if fetchRes.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", fetchRes.StatusCode)
	}
	if strings.Contains(fetchRes.WrappedText, "malicious script") {
		t.Error("expected script tag contents to be stripped from wrapped text")
	}
	if !strings.Contains(fetchRes.WrappedText, "Neural networks are computing systems") {
		t.Error("wrapped text missing main body text")
	}
}

func TestWebFetchTool_ContentTypeGuard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))
	}))
	defer server.Close()

	tool := rtools.NewWebFetchToolWithClient(server.Client())
	args, _ := json.Marshal(rtools.WebFetchArgs{
		URI:       server.URL,
		ProjectID: "proj-png",
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error result for image/png content type")
	}
	if !strings.Contains(res.Error, "unsupported content-type") {
		t.Errorf("unexpected error message: %q", res.Error)
	}
}

func TestWebFetchTool_Paywalled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	tool := rtools.NewWebFetchToolWithClient(server.Client())
	args, _ := json.Marshal(rtools.WebFetchArgs{
		URI:       server.URL,
		ProjectID: "proj-paywall",
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error wrapped paywall response, got error: %s", res.Error)
	}

	var fetchRes rtools.WebFetchResponse
	if err := json.Unmarshal([]byte(res.Output), &fetchRes); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if !fetchRes.IsPaywalled {
		t.Error("expected IsPaywalled=true for 403 Forbidden")
	}
}

// ---------------------------------------------------------------------------
// PDFExtractorTool Tests
// ---------------------------------------------------------------------------

func TestPDFExtractorTool_LocalFile(t *testing.T) {
	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")

	// Write mock PDF content with page marker and text instruction
	mockPDF := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n/Page 1\n(The scaling behavior of attention mechanisms shows sub-quadratic bounds.)\n%%EOF"
	if err := os.WriteFile(pdfPath, []byte(mockPDF), 0644); err != nil {
		t.Fatalf("failed to write mock PDF: %v", err)
	}

	tool := rtools.NewPDFExtractorTool()
	args, _ := json.Marshal(rtools.PDFExtractorArgs{
		ProjectID: "proj-pdf",
		SourceID:  "src-pdf-1",
		FilePath:  pdfPath,
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error tool result, got: %s", res.Error)
	}

	var pdfRes rtools.PDFExtractorResult
	if err := json.Unmarshal([]byte(res.Output), &pdfRes); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if pdfRes.TotalItems == 0 {
		t.Fatal("expected at least 1 evidence item extracted")
	}
	if !strings.Contains(pdfRes.WrappedText, "scaling behavior") {
		t.Errorf("extracted text missing expected snippet: %q", pdfRes.WrappedText)
	}
}

func TestPDFExtractorTool_RemoteContentTypeGuard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html>Not a PDF</html>"))
	}))
	defer server.Close()

	tool := rtools.NewPDFExtractorToolWithClient(server.Client())
	args, _ := json.Marshal(rtools.PDFExtractorArgs{
		ProjectID: "proj-pdf",
		SourceID:  "src-pdf-2",
		URI:       server.URL,
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected tool error when remote content is not PDF")
	}
	if !strings.Contains(res.Error, "unexpected content-type") {
		t.Errorf("unexpected error message: %q", res.Error)
	}
}

// ---------------------------------------------------------------------------
// CitationVerifierTool Tests
// ---------------------------------------------------------------------------

func TestCitationVerifierTool_Entailment(t *testing.T) {
	// Case 1: Direct Support (high token overlap, no negation mismatch)
	claim1 := "Transformer models use self-attention to process sequences in parallel."
	snippet1 := "Self-attention mechanisms allow transformer architectures to compute sequence representations in parallel without recurrence."

	status1, conf1, reasoning1 := rtools.EvaluateClaimEvidenceEntailment(claim1, snippet1)
	if status1 != domain.EvidenceStatusVerified {
		t.Errorf("expected status VERIFIED, got %s (reasoning: %s)", status1, reasoning1)
	}
	if conf1 < 0.5 {
		t.Errorf("expected confidence >= 0.5, got %f", conf1)
	}

	// Case 2: Contradiction / Negation Mismatch
	claim2 := "The algorithm does not scale to large datasets."
	snippet2 := "The algorithm scales efficiently to large datasets across distributed nodes."

	status2, _, reasoning2 := rtools.EvaluateClaimEvidenceEntailment(claim2, snippet2)
	if status2 != domain.EvidenceStatusMismatch {
		t.Errorf("expected status MISMATCH for negation contradiction, got %s (reasoning: %s)", status2, reasoning2)
	}

	// Case 3: Unverified / Low Overlap
	claim3 := "Quantum annealing provides exponential speedup for linear algebra."
	snippet3 := "The database index uses B-trees for efficient range queries."

	status3, conf3, _ := rtools.EvaluateClaimEvidenceEntailment(claim3, snippet3)
	if status3 != domain.EvidenceStatusMismatch && status3 != domain.EvidenceStatusUnverified {
		t.Errorf("expected status MISMATCH or UNVERIFIED for low overlap, got %s", status3)
	}
	if conf3 >= 0.7 {
		t.Errorf("expected low confidence for unrelated snippet, got %f", conf3)
	}
}

func TestCitationVerifierTool_Execute(t *testing.T) {
	tool := rtools.NewCitationVerifierTool()
	args, _ := json.Marshal(rtools.CitationVerifierArgs{
		ClaimStatement: "Sparse attention reduces memory overhead from O(N^2) to O(N sqrt(N)).",
		EvidenceID:     "ev-999",
		Snippet:        "By applying sparse attention patterns, memory overhead is bounded by O(N sqrt(N)) instead of O(N^2).",
	})

	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %s", res.Error)
	}

	var verRes rtools.CitationVerificationResult
	if err := json.Unmarshal([]byte(res.Output), &verRes); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if verRes.VerificationStatus != domain.EvidenceStatusVerified {
		t.Errorf("expected status VERIFIED, got %s", verRes.VerificationStatus)
	}
	if verRes.EvidenceID != "ev-999" {
		t.Errorf("expected evidence ID ev-999, got %q", verRes.EvidenceID)
	}
}
