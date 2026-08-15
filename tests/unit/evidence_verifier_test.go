package unit

import (
	"context"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
)

func TestEvidenceVerifier_VerifyClaim_DirectSupport(t *testing.T) {
	tracker := provenance.NewProvenanceTracker()
	verifier := provenance.NewEvidenceVerifier(tracker, nil)

	claim := "Empirical benchmarks indicate significant accuracy gains in transformer models."
	snippet := "Empirical benchmarks indicate significant accuracy gains when tested across large benchmark datasets."

	res, err := verifier.VerifyClaim(context.Background(), "claim-1", claim, "ev-1", snippet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.VerificationStatus != domain.EvidenceStatusVerified {
		t.Errorf("expected status VERIFIED, got %s", res.VerificationStatus)
	}

	if res.MatchConfidence < 0.70 {
		t.Errorf("expected confidence >= 0.70, got %.2f", res.MatchConfidence)
	}
}

func TestEvidenceVerifier_VerifyClaim_Contradiction(t *testing.T) {
	tracker := provenance.NewProvenanceTracker()
	verifier := provenance.NewEvidenceVerifier(tracker, nil)

	claim := "Empirical benchmarks indicate significant accuracy gains in transformer models."
	snippet := "Empirical benchmarks fail to show any significant accuracy gains in transformer models."

	res, err := verifier.VerifyClaim(context.Background(), "claim-2", claim, "ev-2", snippet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.VerificationStatus != domain.EvidenceStatusMismatch {
		t.Errorf("expected status MISMATCH due to negation, got %s", res.VerificationStatus)
	}
}

func TestEvidenceVerifier_VerifyFindings(t *testing.T) {
	tracker := provenance.NewProvenanceTracker()
	verifier := provenance.NewEvidenceVerifier(tracker, nil)

	finding1 := &domain.ResearchFinding{
		TaskID:  "task-1",
		AgentID: "worker-1",
		Findings: []string{
			"Quantum error correction improves qubit coherence times.",
		},
		Evidence: []domain.Evidence{
			{
				ID:        "ev-101",
				ProjectID: "proj-1",
				SourceID:  "src-101",
				Snippet:   "Quantum error correction techniques significantly improve qubit coherence times.",
			},
		},
	}

	finding2 := &domain.ResearchFinding{
		TaskID:  "task-2",
		AgentID: "worker-2",
		Findings: []string{
			"Superconducting circuits fail to achieve quantum supremacy.",
		},
		Evidence: []domain.Evidence{
			{
				ID:        "ev-102",
				ProjectID: "proj-1",
				SourceID:  "src-102",
				Snippet:   "Superconducting circuits demonstrate clear quantum supremacy on benchmark tasks.",
			},
		},
	}

	summary, err := verifier.VerifyFindings(context.Background(), "proj-1", []*domain.ResearchFinding{finding1, finding2})
	if err != nil {
		t.Fatalf("unexpected error running VerifyFindings: %v", err)
	}

	if summary.TotalClaims != 2 {
		t.Fatalf("expected 2 total claims, got %d", summary.TotalClaims)
	}

	if summary.VerifiedClaims != 1 {
		t.Errorf("expected 1 verified claim, got %d", summary.VerifiedClaims)
	}

	if summary.MismatchClaims != 1 {
		t.Errorf("expected 1 mismatch claim, got %d", summary.MismatchClaims)
	}

	report := tracker.GenerateReport()
	if report.TotalClaims != 2 {
		t.Errorf("expected report to register 2 claims, got %d", report.TotalClaims)
	}
}
