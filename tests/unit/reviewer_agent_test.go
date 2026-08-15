package unit

import (
	"context"
	"testing"

	ragents "github.com/ilyaskhan/term-agent/internal/workflows/research/agents"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
)

func TestReviewerAgent_ApprovedPaper(t *testing.T) {
	reviewer := ragents.NewReviewerAgent("rev-test-1")
	tracker := provenance.NewProvenanceTracker()

	paper, err := domain.NewResearchPaper("paper-1", "proj-1", "academic_research", "Verified Distributed Consensus")
	if err != nil {
		t.Fatalf("unexpected error creating paper: %v", err)
	}

	paper.Sections = []domain.PaperSection{
		{
			ID:      "abstract",
			Title:   "Abstract",
			Content: "This study evaluates distributed consensus algorithm performance.",
			Order:   1,
		},
		{
			ID:      "methodology",
			Title:   "Methodology",
			Content: "We analyze Raft and Paxos protocol variants under network partitioning [1]. Verified: Direct Support",
			Order:   2,
		},
	}
	paper.MarkdownOutput = "# Verified Distributed Consensus\n\n## Abstract\nThis study evaluates..."

	summary := &provenance.VerificationSummary{
		TotalClaims:    2,
		VerifiedClaims: 2,
		OverallRate:    1.0,
		Results:        nil,
	}

	review, err := reviewer.AuditPaper(context.Background(), paper, tracker, summary)
	if err != nil {
		t.Fatalf("AuditPaper returned error: %v", err)
	}

	if review.Verdict != domain.ReviewVerdictApproved {
		t.Errorf("expected verdict APPROVED, got %s", review.Verdict)
	}

	if review.FidelityScore < 0.80 {
		t.Errorf("expected fidelity score >= 0.80, got %f", review.FidelityScore)
	}

	if review.HallucinationRisk > 0.20 {
		t.Errorf("expected hallucination risk <= 0.20, got %f", review.HallucinationRisk)
	}
}

func TestReviewerAgent_UncitedClaimsAndContradictions(t *testing.T) {
	reviewer := ragents.NewReviewerAgent("rev-test-2")
	tracker := provenance.NewProvenanceTracker()

	paper, err := domain.NewResearchPaper("paper-2", "proj-1", "technical_survey", "Unverified Benchmark Paper")
	if err != nil {
		t.Fatalf("unexpected error creating paper: %v", err)
	}

	paper.Sections = []domain.PaperSection{
		{
			ID:      "benchmarks",
			Title:   "Benchmarks",
			Content: "System A achieves 50000000000000 ms throughput under heavy load.", // Uncited metric
			Order:   1,
		},
		{
			ID:      "conclusion",
			Title:   "Conclusion",
			Content: "(Section content pending synthesis)", // Incomplete section
			Order:   2,
		},
	}

	summary := &provenance.VerificationSummary{
		TotalClaims: 1,
		Results: []provenance.CitationVerificationResult{
			{
				ClaimID:            "claim-1",
				ClaimStatement:     "Latency under partition",
				VerificationStatus: domain.EvidenceStatusMismatch,
				MatchConfidence:    0.20,
				Reasoning:          "Network delays exceeded SLA limits",
			},
		},
	}

	review, err := reviewer.AuditPaper(context.Background(), paper, tracker, summary)
	if err != nil {
		t.Fatalf("AuditPaper returned error: %v", err)
	}

	if len(review.ContradictedClaims) == 0 {
		t.Errorf("expected contradicted claims to be flagged")
	}

	if len(review.UncitedClaims) == 0 {
		t.Errorf("expected uncited claims to be flagged")
	}

	if review.Verdict == domain.ReviewVerdictApproved {
		t.Errorf("expected paper to NOT be APPROVED due to errors, got %s", review.Verdict)
	}
}

func TestReviewerAgent_NilPaper(t *testing.T) {
	reviewer := ragents.NewReviewerAgent("rev-test-3")
	_, err := reviewer.AuditPaper(context.Background(), nil, nil, nil)
	if err == nil {
		t.Errorf("expected error for nil paper, got nil")
	}
}
