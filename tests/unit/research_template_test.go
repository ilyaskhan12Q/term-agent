package unit

import (
	"testing"

	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

func TestTemplateEngineLoading(t *testing.T) {
	eng, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to initialize template engine: %v", err)
	}

	acad, err := eng.GetTemplate("academic_research")
	if err != nil {
		t.Fatalf("failed to get academic_research template: %v", err)
	}
	if len(acad.Sections) < 5 {
		t.Errorf("expected at least 5 sections in academic_research template, got %d", len(acad.Sections))
	}

	survey, err := eng.GetTemplate("technical_survey")
	if err != nil {
		t.Fatalf("failed to get technical_survey template: %v", err)
	}
	if survey.Name == "" {
		t.Errorf("expected non-empty template name for technical_survey")
	}
}

func TestTemplateSkeletonAndValidation(t *testing.T) {
	eng, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to initialize template engine: %v", err)
	}

	skel, err := eng.CreatePaperSkeleton("academic_research")
	if err != nil {
		t.Fatalf("failed to create paper skeleton: %v", err)
	}
	if len(skel) == 0 {
		t.Errorf("expected non-empty section list in skeleton")
	}

	paper, err := domain.NewResearchPaper("paper-1", "proj-1", "academic_research", "Test Paper")
	if err != nil {
		t.Fatalf("failed to create paper: %v", err)
	}
	paper.Sections = skel

	// Validation should fail because content is empty
	if err := eng.ValidatePaperCompleteness(paper); err == nil {
		t.Errorf("expected validation failure for empty paper sections")
	}

	// Populate sections
	for i := range paper.Sections {
		paper.Sections[i].Content = "Populated content for section " + paper.Sections[i].Title
	}

	// Validation should now pass
	if err := eng.ValidatePaperCompleteness(paper); err != nil {
		t.Errorf("unexpected validation failure for populated paper: %v", err)
	}
}
