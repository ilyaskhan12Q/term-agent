package unit

import (
	"testing"

	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
)

func TestProvenanceTrackerFlow(t *testing.T) {
	tracker := provenance.NewProvenanceTracker()

	// 1. Register Source
	src, err := domain.NewSource("src-1", "proj-1", "Attention is All You Need", "https://arxiv.org/abs/1706.03762", domain.SourceTypeAcademicPaper, 0.99)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}
	src.Authors = []string{"Vaswani et al."}
	src.Year = 2017

	if err := tracker.RegisterSource(*src); err != nil {
		t.Fatalf("failed to register source: %v", err)
	}

	// 2. Register Evidence linked to invalid source (should error)
	invalidEv, _ := domain.NewEvidence("ev-invalid", "proj-1", "non-existent-src", "Snippet text", "Page 1", "agent-1")
	if err := tracker.RegisterEvidence(*invalidEv); err == nil {
		t.Errorf("expected error when registering evidence for non-existent source")
	}

	// 3. Register valid Evidence
	ev, err := domain.NewEvidence("ev-1", "proj-1", src.ID, "Self-attention reduces sequential computation", "Page 2", "agent-1")
	if err != nil {
		t.Fatalf("failed to create evidence: %v", err)
	}
	ev.VerificationStatus = domain.EvidenceStatusVerified

	if err := tracker.RegisterEvidence(*ev); err != nil {
		t.Fatalf("failed to register evidence: %v", err)
	}

	// 4. Register Claim linked to Evidence
	claim := domain.Claim{
		ID:          "claim-1",
		ProjectID:   "proj-1",
		Statement:   "Transformers perform parallel attention operations.",
		EvidenceIDs: []string{ev.ID},
		Strength:    domain.ClaimDirect,
	}
	if err := tracker.RegisterClaim(claim); err != nil {
		t.Fatalf("failed to register claim: %v", err)
	}

	// 5. Trace Claim Provenance
	c, evs, srcs, err := tracker.TraceClaim("claim-1")
	if err != nil {
		t.Fatalf("failed to trace claim: %v", err)
	}
	if c.Statement != claim.Statement {
		t.Errorf("claim statement mismatch")
	}
	if len(evs) != 1 || evs[0].ID != ev.ID {
		t.Errorf("expected 1 evidence item matching ev-1")
	}
	if len(srcs) != 1 || srcs[0].ID != src.ID {
		t.Errorf("expected 1 source matching src-1")
	}

	// 6. Generate Report
	report := tracker.GenerateReport()
	if report.TotalSources != 1 || report.TotalEvidence != 1 || report.TotalClaims != 1 {
		t.Errorf("unexpected count in report: %+v", report)
	}
	if report.CoverageScore != 1.0 || report.VerificationRate != 1.0 {
		t.Errorf("expected 1.0 coverage and verification rates, got %+v", report)
	}

	// 7. Verify Bibliography Output
	bib := tracker.BuildBibliographyFormats()
	if bib == "" {
		t.Errorf("expected non-empty bibliography string")
	}
}
