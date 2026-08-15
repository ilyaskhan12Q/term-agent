package unit

import (
	"context"
	"strings"
	"testing"

	ragents "github.com/ilyaskhan/term-agent/internal/workflows/research/agents"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

func TestSynthesisAgent_SynthesizePaper_DefaultTemplate(t *testing.T) {
	engine, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}
	agent := ragents.NewSynthesisAgent("synth-1", engine)

	project, err := domain.NewResearchProject("proj-1", "sess-1", "Quantum Computing Benchmarks", "Benchmarking quantum supremacy on noisy devices", "academic_research")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	finding := &domain.ResearchFinding{
		TaskID:  "task-1",
		AgentID: "worker-1",
		Findings: []string{
			"Quantum error mitigation reduces gate fidelity degradation.",
		},
	}

	tracker := provenance.NewProvenanceTracker()
	src, _ := domain.NewSource("src-1", "proj-1", "Quantum Supremacy", "https://arxiv.org/abs/2401.00001", domain.SourceTypeAcademicPaper, 0.95)
	_, _ = tracker.RegisterSource(*src)

	paper, err := agent.SynthesizePaper(context.Background(), project, []*domain.ResearchFinding{finding}, tracker)
	if err != nil {
		t.Fatalf("SynthesizePaper failed: %v", err)
	}

	if paper.Status != domain.PaperStatusPassed {
		t.Errorf("expected paper status PASSED, got %s", paper.Status)
	}

	if len(paper.Sections) == 0 {
		t.Errorf("expected non-empty paper sections")
	}

	if !strings.Contains(paper.MarkdownOutput, "# Quantum Computing Benchmarks") {
		t.Errorf("expected paper markdown to contain title header")
	}

	if !strings.Contains(paper.MarkdownOutput, "Quantum error mitigation reduces gate fidelity degradation.") {
		t.Errorf("expected paper markdown to contain finding statement")
	}
}

func TestSynthesisAgent_SynthesizePaperWithVerification_Annotations(t *testing.T) {
	engine, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}
	agent := ragents.NewSynthesisAgent("synth-1", engine)

	project, _ := domain.NewResearchProject("proj-2", "sess-2", "AI Alignment Survey", "Survey of scalable oversight techniques", "academic_research")
	tracker := provenance.NewProvenanceTracker()

	finding := &domain.ResearchFinding{
		TaskID:  "task-2",
		AgentID: "worker-2",
		Findings: []string{
			"Scalable oversight improves RLHF safety margins.",
		},
	}

	verifier := provenance.NewEvidenceVerifier(tracker, nil)
	verSummary, err := verifier.VerifyFindings(context.Background(), "proj-2", []*domain.ResearchFinding{finding})
	if err != nil {
		t.Fatalf("VerifyFindings failed: %v", err)
	}

	paper, err := agent.SynthesizePaperWithVerification(context.Background(), project, []*domain.ResearchFinding{finding}, tracker, verSummary)
	if err != nil {
		t.Fatalf("SynthesizePaperWithVerification failed: %v", err)
	}

	if !strings.Contains(paper.MarkdownOutput, "Verified: Direct Support") {
		t.Errorf("expected paper markdown to contain verification badge 'Verified: Direct Support'")
	}

	if !strings.Contains(paper.MarkdownOutput, "Provenance & Entailment Report") {
		t.Errorf("expected paper references to contain Provenance & Entailment Report")
	}
}

func TestSynthesisAgent_BuildSourceCitationMap(t *testing.T) {
	sources := []domain.Source{
		{ID: "src-2"},
		{ID: "src-1"},
		{ID: "src-3"},
	}

	m := ragents.BuildSourceCitationMap(sources)
	if m["src-1"] != 1 || m["src-2"] != 2 || m["src-3"] != 3 {
		t.Errorf("unexpected citation index mapping: %v", m)
	}
}
