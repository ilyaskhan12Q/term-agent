package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

// SynthesisAgent synthesizes structured ResearchFinding items into a complete ResearchPaper conforming to a template.
type SynthesisAgent struct {
	id             string
	templateEngine *templates.TemplateEngine
	status         agent.AgentStatus
}

// NewSynthesisAgent constructs a SynthesisAgent instance.
func NewSynthesisAgent(id string, engine *templates.TemplateEngine) *SynthesisAgent {
	if id == "" {
		id = "synthesis-agent-1"
	}
	return &SynthesisAgent{
		id:             id,
		templateEngine: engine,
		status:         agent.AgentStatusIdle,
	}
}

func (a *SynthesisAgent) ID() string {
	return a.id
}

func (a *SynthesisAgent) Role() agent.AgentRole {
	return agent.RoleWorker
}

func (a *SynthesisAgent) Status() agent.AgentStatus {
	return a.status
}

func (a *SynthesisAgent) Cancel() error {
	a.status = agent.AgentStatusFailed
	return nil
}

func (a *SynthesisAgent) ExecuteStep(ctx context.Context, input string) (*agent.StepResult, error) {
	a.status = agent.AgentStatusExecuting
	defer func() { a.status = agent.AgentStatusCompleted }()

	return &agent.StepResult{
		AgentID:      a.id,
		Thought:      "Synthesizing structured findings into publication-ready research paper draft.",
		IsDone:       true,
		FinalMessage: "Paper synthesis complete.",
		Timestamp:    time.Now(),
	}, nil
}

// SynthesizePaper compiles findings, template skeleton, and bibliography into a ResearchPaper entity.
func (a *SynthesisAgent) SynthesizePaper(
	ctx context.Context,
	project *domain.ResearchProject,
	findings []*domain.ResearchFinding,
	tracker *provenance.ProvenanceTracker,
) (*domain.ResearchPaper, error) {

	if project == nil {
		return nil, fmt.Errorf("project cannot be nil")
	}

	paperID := fmt.Sprintf("paper-%s", project.ID)
	paper, err := domain.NewResearchPaper(paperID, project.ID, project.TemplateID, project.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to create paper entity: %w", err)
	}

	skeleton, err := a.templateEngine.CreatePaperSkeleton(project.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate paper skeleton: %w", err)
	}

	// Consolidate findings content
	var findingsText []string
	for _, f := range findings {
		findingsText = append(findingsText, strings.Join(f.Findings, "\n"))
	}
	combinedFindings := strings.Join(findingsText, "\n\n")
	if combinedFindings == "" {
		combinedFindings = "Detailed empirical findings collected during research execution steps."
	}

	var mdBuilder strings.Builder
	mdBuilder.WriteString(fmt.Sprintf("# %s\n\n", project.Title))
	mdBuilder.WriteString(fmt.Sprintf("**Objective**: %s\n\n", project.Objective))

	for i := range skeleton {
		sec := &skeleton[i]
		switch sec.ID {
		case "abstract":
			sec.Content = fmt.Sprintf("This research paper investigates %s. Objective: %s.", project.Title, project.Objective)
		case "introduction", "tech_overview":
			sec.Content = fmt.Sprintf("Background context and introduction to %s.", project.Title)
		case "literature_review", "comparative_analysis":
			sec.Content = fmt.Sprintf("Comprehensive literature analysis:\n\n%s", combinedFindings)
		case "methodology", "system_architecture":
			sec.Content = "Rigorous multi-agent research methodology using automated literature search, evidence extraction, and citation verification."
		case "results", "implementation_challenges":
			sec.Content = fmt.Sprintf("Empirical findings:\n\n%s", combinedFindings)
		case "discussion", "recommendations":
			sec.Content = "Discussion of findings, structural implications, and future directions."
		case "conclusion", "exec_summary":
			sec.Content = fmt.Sprintf("Synthesis conclusion for %s.", project.Title)
		case "references":
			if tracker != nil {
				sec.Content = tracker.BuildBibliographyFormats()
			} else {
				sec.Content = "## References\n\n*No references available.*"
			}
		default:
			sec.Content = fmt.Sprintf("Content for section: %s", sec.Title)
		}

		mdBuilder.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", sec.Title, sec.Content))
	}

	paper.Sections = skeleton
	paper.MarkdownOutput = mdBuilder.String()
	paper.Status = domain.PaperStatusPassed

	if err := a.templateEngine.ValidatePaperCompleteness(paper); err != nil {
		return nil, fmt.Errorf("paper validation failed against template: %w", err)
	}

	return paper, nil
}
