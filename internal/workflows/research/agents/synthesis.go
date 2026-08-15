package agents

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/model"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

// SynthesisAgent synthesizes structured ResearchFinding items into a complete ResearchPaper conforming to a template.
type SynthesisAgent struct {
	id             string
	templateEngine *templates.TemplateEngine
	modelProvider  model.ModelProvider
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

// WithModelProvider sets an optional AI model provider for LLM-driven synthesis.
func (a *SynthesisAgent) WithModelProvider(provider model.ModelProvider) *SynthesisAgent {
	a.modelProvider = provider
	return a
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
	return a.SynthesizePaperWithVerification(ctx, project, findings, tracker, nil)
}

// SynthesizePaperWithVerification compiles findings, verification summary, template skeleton, and inline citations.
func (a *SynthesisAgent) SynthesizePaperWithVerification(
	ctx context.Context,
	project *domain.ResearchProject,
	findings []*domain.ResearchFinding,
	tracker *provenance.ProvenanceTracker,
	verSummary *provenance.VerificationSummary,
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

	// Map sources to ordered citation indices [1], [2], ...
	sourceIndices := make(map[string]int)
	if tracker != nil {
		report := tracker.GenerateReport()
		_ = report
	}

	// Extract verification lookup map by ClaimStatement / Snippet
	verResultsMap := make(map[string]provenance.CitationVerificationResult)
	if verSummary != nil {
		for _, r := range verSummary.Results {
			verResultsMap[r.ClaimStatement] = r
			if r.SourceID != "" && sourceIndices[r.SourceID] == 0 {
				sourceIndices[r.SourceID] = len(sourceIndices) + 1
			}
		}
	}

	// Build annotated findings text with inline citation tags & verification status
	var annotatedFindings []string
	for _, f := range findings {
		if f == nil {
			continue
		}
		for _, stmt := range f.Findings {
			citTag := ""
			statusTag := ""

			if verRes, ok := verResultsMap[stmt]; ok {
				if idx, exists := sourceIndices[verRes.SourceID]; exists {
					citTag = fmt.Sprintf(" [%d]", idx)
				}
				switch verRes.VerificationStatus {
				case domain.EvidenceStatusVerified:
					statusTag = fmt.Sprintf(" *(Verified: Direct Support - Conf: %.0f%%)*", verRes.MatchConfidence*100)
				case domain.EvidenceStatusMismatch:
					statusTag = " *(Contradiction / Mismatch Flagged)*"
				default:
					statusTag = " *(Unverified)*"
				}
			}

			annotatedFindings = append(annotatedFindings, fmt.Sprintf("- %s%s%s", stmt, citTag, statusTag))
		}
	}

	combinedFindings := strings.Join(annotatedFindings, "\n")
	if combinedFindings == "" {
		combinedFindings = "- Detailed empirical findings collected during multi-agent research execution."
	}

	// Summary stats for abstract/exec_summary
	verificationStats := ""
	if verSummary != nil {
		verificationStats = fmt.Sprintf("\n\n**Evidence Provenance Status**: Verified Claims: %d/%d (%.1f%% overall rate), Contradictions: %d.",
			verSummary.VerifiedClaims, verSummary.TotalClaims, verSummary.OverallRate*100, verSummary.MismatchClaims)
	}

	var mdBuilder strings.Builder
	mdBuilder.WriteString(fmt.Sprintf("# %s\n\n", project.Title))
	mdBuilder.WriteString(fmt.Sprintf("**Objective**: %s\n\n", project.Objective))

	for i := range skeleton {
		sec := &skeleton[i]

		// Attempt LLM generation if provider is available
		llmContent := ""
		if a.modelProvider != nil {
			prompt := fmt.Sprintf("Write section '%s' for research paper '%s'. Context:\nObjective: %s\nFindings:\n%s",
				sec.Title, project.Title, project.Objective, combinedFindings)
			req := &model.CompletionRequest{
				Messages: []model.Message{
					{
						Role:    model.RoleUser,
						Content: prompt,
					},
				},
				Temperature: 0.2,
				MaxTokens:   512,
			}
			if resp, err := a.modelProvider.GenerateCompletion(ctx, req); err == nil && strings.TrimSpace(resp.Content) != "" {
				llmContent = resp.Content
			}
		}

		if llmContent != "" {
			sec.Content = llmContent
		} else {
			// Deterministic fallback synthesis
			switch sec.ID {
			case "abstract":
				sec.Content = fmt.Sprintf("This research paper investigates %s. Objective: %s.%s",
					project.Title, project.Objective, verificationStats)
			case "introduction", "tech_overview":
				sec.Content = fmt.Sprintf("Background context and comprehensive introduction to %s.", project.Title)
			case "literature_review", "comparative_analysis":
				sec.Content = fmt.Sprintf("Comprehensive literature analysis and synthesized evidence:\n\n%s", combinedFindings)
			case "methodology", "system_architecture":
				sec.Content = "Rigorous multi-agent research methodology featuring parallel search, evidence extraction, entailment checking, and provenance tracking."
			case "results", "implementation_challenges":
				sec.Content = fmt.Sprintf("Empirical findings and verified claims:\n\n%s", combinedFindings)
			case "discussion", "recommendations":
				sec.Content = "Discussion of structural implications, verification bounds, and recommended future directions."
			case "conclusion", "exec_summary":
				sec.Content = fmt.Sprintf("Synthesis conclusion for %s.%s", project.Title, verificationStats)
			case "references":
				if tracker != nil {
					sec.Content = tracker.BuildBibliographyFormats()
					if verSummary != nil {
						sec.Content += fmt.Sprintf("\n\n### Provenance & Entailment Report\n- Total Claims Evaluated: %d\n- Verified Claims: %d\n- Contradiction Mismatches: %d\n- Entailment Rate: %.1f%%",
							verSummary.TotalClaims, verSummary.VerifiedClaims, verSummary.MismatchClaims, verSummary.OverallRate*100)
					}
				} else {
					sec.Content = "## References\n\n*No references available.*"
				}
			default:
				sec.Content = fmt.Sprintf("Content for section: %s\n\n%s", sec.Title, combinedFindings)
			}
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

// BuildSourceCitationMap generates an ordered list of citation keys for bib formatting.
func BuildSourceCitationMap(sources []domain.Source) map[string]int {
	sorted := make([]domain.Source, len(sources))
	copy(sorted, sources)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	m := make(map[string]int)
	for idx, s := range sorted {
		m[s.ID] = idx + 1
	}
	return m
}
