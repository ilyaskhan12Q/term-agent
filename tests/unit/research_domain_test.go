package unit

import (
	"testing"

	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

func TestResearchProjectValidation(t *testing.T) {
	_, err := domain.NewResearchProject("", "sess-1", "Title", "Objective", "academic_research")
	if err == nil {
		t.Errorf("expected error for empty project ID")
	}

	p, err := domain.NewResearchProject("proj-1", "sess-1", "Quantum Computing", "Investigate QPU error rates", "")
	if err != nil {
		t.Fatalf("unexpected error creating project: %v", err)
	}
	if p.TemplateID != "academic_research" {
		t.Errorf("expected default templateID 'academic_research', got '%s'", p.TemplateID)
	}
	if p.Status != domain.ProjectStatusCreated {
		t.Errorf("expected initial status 'CREATED', got '%s'", p.Status)
	}
}

func TestResearchFindingValidation(t *testing.T) {
	finding := &domain.ResearchFinding{
		ID:         "find-1",
		ProjectID:  "proj-1",
		QuestionID: "q-1",
		AgentType:  "LITERATURE",
		Confidence: 1.5, // Invalid > 1.0
	}

	if err := finding.Validate(); err == nil {
		t.Errorf("expected validation error for confidence > 1.0")
	}

	finding.Confidence = 0.95
	if err := finding.Validate(); err != nil {
		t.Errorf("unexpected error for valid finding: %v", err)
	}
}

func TestProvenanceEntities(t *testing.T) {
	src, err := domain.NewSource("src-1", "proj-1", "Attention is All You Need", "https://arxiv.org/abs/1706.03762", domain.SourceTypeAcademicPaper, 0.98)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}
	if src.TrustScore != 0.98 {
		t.Errorf("expected trust score 0.98, got %f", src.TrustScore)
	}

	ev, err := domain.NewEvidence("ev-1", "proj-1", src.ID, "Self-attention mechanism allows parallel processing", "Page 3", "agent-lit")
	if err != nil {
		t.Fatalf("failed to create evidence: %v", err)
	}
	if ev.VerificationStatus != domain.EvidenceStatusUnverified {
		t.Errorf("expected initial evidence status UNVERIFIED, got %s", ev.VerificationStatus)
	}
}
