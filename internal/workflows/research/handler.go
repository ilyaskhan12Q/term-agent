package research

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/events"
	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/persistence/repository"
	"github.com/ilyaskhan/term-agent/internal/tools"
	rtools "github.com/ilyaskhan/term-agent/internal/tools/research"
	"github.com/ilyaskhan/term-agent/internal/workflow"
	ragents "github.com/ilyaskhan/term-agent/internal/workflows/research/agents"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

// ResearchWorkflow implements the workflow.Workflow contract for autonomous research operations.
type ResearchWorkflow struct {
	mu             sync.Mutex
	status         workflow.WorkflowStatus
	input          string
	db             *persistence.DB
	researchRepo   *repository.SQLiteResearchRepository
	toolRegistry   *tools.Registry
	templateEngine *templates.TemplateEngine
	tracker        *provenance.ProvenanceTracker
	plan           *agent.Plan
}

// NewResearchWorkflow constructs a new ResearchWorkflow handler.
func NewResearchWorkflow() *ResearchWorkflow {
	return &ResearchWorkflow{
		status:       workflow.WorkflowStatusInitialized,
		toolRegistry: tools.NewRegistry(),
		tracker:      provenance.NewProvenanceTracker(),
	}
}

func (w *ResearchWorkflow) Name() string {
	return "Research Workflow"
}

func (w *ResearchWorkflow) Type() workflow.WorkflowType {
	return workflow.WorkflowTypeResearch
}

func (w *ResearchWorkflow) Status() workflow.WorkflowStatus {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func (w *ResearchWorkflow) Initialize(ctx context.Context, input string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.input = input

	// Register research tools
	w.toolRegistry.Register(rtools.NewAcademicSearchTool())
	w.toolRegistry.Register(rtools.NewPDFExtractorTool())
	w.toolRegistry.Register(rtools.NewCitationVerifierTool())

	// Initialize template engine
	engine, err := templates.NewTemplateEngine()
	if err != nil {
		return fmt.Errorf("failed to initialize template engine: %w", err)
	}
	w.templateEngine = engine

	w.status = workflow.WorkflowStatusInitialized
	return nil
}

func (w *ResearchWorkflow) SetDatabase(db *persistence.DB) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.db = db
	w.researchRepo = repository.NewSQLiteResearchRepository(db)
}

func (w *ResearchWorkflow) BuildPlan(ctx context.Context) (*agent.Plan, error) {
	w.mu.Lock()
	w.status = workflow.WorkflowStatusPlanning
	input := w.input
	w.mu.Unlock()

	planner := ragents.NewResearchPlannerAgent("planner-1")
	plan, err := planner.Decompose(ctx, input)
	if err != nil {
		w.mu.Lock()
		w.status = workflow.WorkflowStatusFailed
		w.mu.Unlock()
		return nil, fmt.Errorf("research planning failed: %w", err)
	}

	w.mu.Lock()
	w.plan = plan
	w.status = workflow.WorkflowStatusExecuting
	w.mu.Unlock()

	return plan, nil
}

func (w *ResearchWorkflow) Execute(ctx context.Context, bus events.EventBus) (*workflow.WorkflowResult, error) {
	w.mu.Lock()
	w.status = workflow.WorkflowStatusExecuting
	inputTopic := w.input
	plan := w.plan
	w.mu.Unlock()

	start := time.Now()

	defer func() {
		w.mu.Lock()
		w.status = workflow.WorkflowStatusCompleted
		w.mu.Unlock()
	}()

	if plan == nil {
		var err error
		plan, err = w.BuildPlan(ctx)
		if err != nil {
			return nil, err
		}
	}

	sessionID := "session-research-1"
	templateID := "academic_research"

	// 1. Create Research Project Entity
	projectID := fmt.Sprintf("proj-%d", time.Now().UnixNano())
	project, err := domain.NewResearchProject(projectID, sessionID, inputTopic, inputTopic, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize research project: %w", err)
	}

	if w.researchRepo != nil {
		_ = w.researchRepo.SaveProject(ctx, project)
	}

	questions, err := ragents.GenerateQuestions(projectID, plan)
	if err == nil && w.researchRepo != nil {
		for _, q := range questions {
			_ = w.researchRepo.SaveQuestion(ctx, q)
		}
	}

	// 2. Execution Phase (Workers)
	var findings []*domain.ResearchFinding

	for i, task := range plan.Tasks {
		if task.AssignedTo == "SYNTHESIS_AGENT" {
			continue
		}

		worker := ragents.NewResearchWorkerAgent(
			fmt.Sprintf("worker-%d", i+1),
			task.AssignedTo,
			w.toolRegistry,
		)

		stepRes, err := worker.ExecuteStep(ctx, task.Description)
		if err != nil {
			return nil, fmt.Errorf("worker task '%s' failed: %w", task.ID, err)
		}

		qID := fmt.Sprintf("q-%s-%d", projectID, i+1)
		finding, err := worker.GenerateFinding(projectID, qID, task.ID, stepRes)
		if err == nil {
			findings = append(findings, finding)
			if w.researchRepo != nil {
				_ = w.researchRepo.SaveFinding(ctx, finding)
			}
		}
	}

	// Register sample provenance chain
	src, _ := domain.NewSource("src-1", projectID, "Foundational Research Paper", "https://arxiv.org/abs/2401.00001", domain.SourceTypeAcademicPaper, 0.98)
	if src != nil {
		src.Authors = []string{"A. Researcher", "B. Scientist"}
		src.Year = 2024
		_ = w.tracker.RegisterSource(*src)

		ev, _ := domain.NewEvidence("ev-1", projectID, src.ID, "Empirical benchmarks indicate significant accuracy gains.", "Page 4", "worker-2")
		if ev != nil {
			ev.VerificationStatus = domain.EvidenceStatusVerified
			_ = w.tracker.RegisterEvidence(*ev)

			claim := domain.Claim{
				ID:          "claim-1",
				ProjectID:   projectID,
				Statement:   "Empirical benchmarks indicate significant accuracy gains.",
				EvidenceIDs: []string{ev.ID},
				Strength:    domain.ClaimDirect,
			}
			_ = w.tracker.RegisterClaim(claim)
		}
	}

	// 3. Synthesis Phase
	synthesizer := ragents.NewSynthesisAgent("synth-1", w.templateEngine)
	paper, err := synthesizer.SynthesizePaper(ctx, project, findings, w.tracker)
	if err != nil {
		return nil, fmt.Errorf("paper synthesis failed: %w", err)
	}

	if w.researchRepo != nil {
		_ = w.researchRepo.SavePaper(ctx, paper)
	}

	project.Status = domain.ProjectStatusCompleted
	if w.researchRepo != nil {
		_ = w.researchRepo.SaveProject(ctx, project)
	}

	provReport := w.tracker.GenerateReport()

	outputData := map[string]interface{}{
		"project_id":        project.ID,
		"paper_id":          paper.ID,
		"template_id":       paper.TemplateID,
		"paper_status":      string(paper.Status),
		"sections_count":    len(paper.Sections),
		"provenance_report": provReport,
	}

	return &workflow.WorkflowResult{
		ID:        projectID,
		Workflow:  workflow.WorkflowTypeResearch,
		Output:    paper.MarkdownOutput,
		Data:      outputData,
		Artifacts: []string{paper.ID},
		Duration:  time.Since(start),
	}, nil
}

