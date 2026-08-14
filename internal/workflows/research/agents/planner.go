package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

// ResearchPlannerAgent handles research task decomposition into DAG plans.
type ResearchPlannerAgent struct {
	id     string
	status agent.AgentStatus
}

// NewResearchPlannerAgent constructs a ResearchPlannerAgent instance.
func NewResearchPlannerAgent(id string) *ResearchPlannerAgent {
	if id == "" {
		id = "research-planner-1"
	}
	return &ResearchPlannerAgent{
		id:     id,
		status: agent.AgentStatusIdle,
	}
}

func (a *ResearchPlannerAgent) ID() string {
	return a.id
}

func (a *ResearchPlannerAgent) Role() agent.AgentRole {
	return agent.RolePlanner
}

func (a *ResearchPlannerAgent) Status() agent.AgentStatus {
	return a.status
}

func (a *ResearchPlannerAgent) Cancel() error {
	a.status = agent.AgentStatusFailed
	return nil
}

func (a *ResearchPlannerAgent) ExecuteStep(ctx context.Context, input string) (*agent.StepResult, error) {
	a.status = agent.AgentStatusExecuting
	defer func() { a.status = agent.AgentStatusCompleted }()

	return &agent.StepResult{
		AgentID:      a.id,
		Thought:      "Decomposing research objective into literature, extraction, verification, and synthesis task DAG.",
		IsDone:       true,
		FinalMessage: fmt.Sprintf("Successfully generated research plan for objective: %s", input),
		Timestamp:    time.Now(),
	}, nil
}

// Decompose creates a structured research plan DAG with research sub-questions.
func (a *ResearchPlannerAgent) Decompose(ctx context.Context, goal string) (*agent.Plan, error) {
	planID := fmt.Sprintf("plan-research-%d", time.Now().UnixNano())

	// Questions breakdown
	questions := []string{
		fmt.Sprintf("What is the state of academic literature regarding %s?", goal),
		fmt.Sprintf("What are the empirical evidence and performance metrics for %s?", goal),
		fmt.Sprintf("What are the primary contradictions or limitations in current %s approaches?", goal),
	}

	tasks := []*agent.TaskSpec{
		{
			ID:           "task-lit-search",
			Description:  questions[0],
			Dependencies: []string{},
			Role:         agent.RoleWorker,
			AssignedTo:   "LITERATURE_AGENT",
		},
		{
			ID:           "task-evidence-extract",
			Description:  questions[1],
			Dependencies: []string{"task-lit-search"},
			Role:         agent.RoleWorker,
			AssignedTo:   "EXTRACTION_AGENT",
		},
		{
			ID:           "task-citation-verify",
			Description:  questions[2],
			Dependencies: []string{"task-evidence-extract"},
			Role:         agent.RoleWorker,
			AssignedTo:   "VERIFICATION_AGENT",
		},
		{
			ID:           "task-paper-synthesis",
			Description:  fmt.Sprintf("Synthesize findings into paper draft for: %s", goal),
			Dependencies: []string{"task-citation-verify"},
			Role:         agent.RoleWorker,
			AssignedTo:   "SYNTHESIS_AGENT",
		},
	}

	return &agent.Plan{
		ID:        planID,
		Goal:      goal,
		Tasks:     tasks,
		CreatedAt: time.Now().Unix(),
	}, nil
}

func (a *ResearchPlannerAgent) Replan(ctx context.Context, plan *agent.Plan, failedTaskID string, reason string) (*agent.Plan, error) {
	if plan == nil {
		return nil, fmt.Errorf("invalid plan")
	}
	// Simple replan retry strategy
	return plan, nil
}

// GenerateQuestions helper converts plan tasks to domain.ResearchQuestion entities.
func GenerateQuestions(projectID string, plan *agent.Plan) ([]*domain.ResearchQuestion, error) {
	var questions []*domain.ResearchQuestion
	for i, task := range plan.Tasks {
		qID := fmt.Sprintf("q-%s-%d", projectID, i+1)
		q, err := domain.NewResearchQuestion(qID, projectID, task.Description, task.AssignedTo, i+1)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}
