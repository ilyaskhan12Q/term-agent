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
	w.toolRegistry.Register(rtools.NewWebFetchTool())
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

	// Ensure parent session exists in sessions table for DB FK constraint
	if w.db != nil {
		_, _ = w.db.ExecContext(ctx, "INSERT OR IGNORE INTO sessions (id, workspace_path, status, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", sessionID, "/tmp", "ACTIVE")
	}

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

	// 2. Execution Phase (Parallel Workers)
	pool := ragents.NewWorkerPool(ragents.WorkerPoolOptions{
		MaxConcurrency: 4,
		ToolRegistry:   w.toolRegistry,
		ResearchRepo:   w.researchRepo,
		Tracker:        w.tracker,
	})

	findings, err := pool.ExecuteTasks(ctx, projectID, plan.Tasks)
	if err != nil {
		return nil, fmt.Errorf("worker execution failed: %w", err)
	}

	// 2b. Evidence & Citation Verification Phase
	verifier := provenance.NewEvidenceVerifier(w.tracker, w.toolRegistry)
	verSummary, _ := verifier.VerifyFindings(ctx, projectID, findings)

	// 3. Synthesis Phase
	synthesizer := ragents.NewSynthesisAgent("synth-1", w.templateEngine)
	paper, err := synthesizer.SynthesizePaperWithVerification(ctx, project, findings, w.tracker, verSummary)
	if err != nil {
		return nil, fmt.Errorf("paper synthesis failed: %w", err)
	}

	// 4. Reviewer & Hallucination Audit Phase
	reviewer := ragents.NewReviewerAgent("reviewer-1")
	review, err := reviewer.AuditPaper(ctx, paper, w.tracker, verSummary)
	if err == nil && review != nil {
		if review.Verdict == domain.ReviewVerdictApproved || review.FidelityScore >= 0.60 {
			paper.Status = domain.PaperStatusPassed
		} else if review.Verdict == domain.ReviewVerdictNeedsRevision {
			paper.Status = domain.PaperStatusReviewing
		} else {
			paper.Status = domain.PaperStatusRejected
		}
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
		"project_id":           project.ID,
		"paper_id":             paper.ID,
		"template_id":          paper.TemplateID,
		"paper_status":         string(paper.Status),
		"sections_count":       len(paper.Sections),
		"provenance_report":    provReport,
		"verification_summary": verSummary,
		"paper_review":         review,
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
