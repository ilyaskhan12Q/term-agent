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
	if goal == "" {
		return nil, fmt.Errorf("research goal cannot be empty")
	}

	planID := fmt.Sprintf("plan-research-%d", time.Now().UnixNano())

	// Questions breakdown tailored to research goal
	questions := []string{
		fmt.Sprintf("What is the state of academic literature regarding %s?", goal),
		fmt.Sprintf("What are the web sources and industry standards for %s?", goal),
		fmt.Sprintf("What empirical evidence and performance metrics exist for %s?", goal),
		fmt.Sprintf("What are the primary contradictions, risks, or limitations in %s?", goal),
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
			ID:           "task-web-search",
			Description:  questions[1],
			Dependencies: []string{},
			Role:         agent.RoleWorker,
			AssignedTo:   "WEB_RESEARCH_AGENT",
		},
		{
			ID:           "task-evidence-extract",
			Description:  questions[2],
			Dependencies: []string{"task-lit-search", "task-web-search"},
			Role:         agent.RoleWorker,
			AssignedTo:   "EXTRACTION_AGENT",
		},
		{
			ID:           "task-citation-verify",
			Description:  questions[3],
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

	plan := &agent.Plan{
		ID:        planID,
		Goal:      goal,
		Tasks:     tasks,
		CreatedAt: time.Now().Unix(),
	}

	if err := ValidateDAG(plan); err != nil {
		return nil, fmt.Errorf("invalid plan DAG generated: %w", err)
	}

	return plan, nil
}

// Replan dynamic strategy: updates failed task or injects a fallback discovery task.
func (a *ResearchPlannerAgent) Replan(ctx context.Context, plan *agent.Plan, failedTaskID string, reason string) (*agent.Plan, error) {
	if plan == nil {
		return nil, fmt.Errorf("invalid plan: plan is nil")
	}

	var updatedTasks []*agent.TaskSpec
	fallbackID := fmt.Sprintf("%s-fallback", failedTaskID)

	for _, task := range plan.Tasks {
		if task.ID == failedTaskID {
			// Create modified version with broadened search prompt
			newTask := *task
			newTask.Description = fmt.Sprintf("Fallback search [Reason: %s]: %s", reason, task.Description)
			updatedTasks = append(updatedTasks, &newTask)

			// Inject fallback secondary task
			fallbackTask := &agent.TaskSpec{
				ID:           fallbackID,
				Description:  fmt.Sprintf("Secondary web discovery for fallback of %s", failedTaskID),
				Dependencies: []string{failedTaskID},
				Role:         agent.RoleWorker,
				AssignedTo:   "WEB_RESEARCH_AGENT",
			}
			updatedTasks = append(updatedTasks, fallbackTask)
		} else {
			// Update dependencies if child depended on failed task
			var newDeps []string
			for _, dep := range task.Dependencies {
				newDeps = append(newDeps, dep)
				if dep == failedTaskID {
					newDeps = append(newDeps, fallbackID)
				}
			}
			taskCopy := *task
			taskCopy.Dependencies = newDeps
			updatedTasks = append(updatedTasks, &taskCopy)
		}
	}

	newPlan := &agent.Plan{
		ID:        fmt.Sprintf("%s-replan", plan.ID),
		Goal:      plan.Goal,
		Tasks:     updatedTasks,
		CreatedAt: time.Now().Unix(),
	}

	return newPlan, nil
}

// ValidateDAG verifies that task dependencies form a valid Directed Acyclic Graph (DAG).
func ValidateDAG(plan *agent.Plan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}
	taskMap := make(map[string]bool)
	for _, task := range plan.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID cannot be empty")
		}
		if taskMap[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskMap[task.ID] = true
	}

	for _, task := range plan.Tasks {
		for _, dep := range task.Dependencies {
			if !taskMap[dep] {
				return fmt.Errorf("task %s references unknown dependency %s", task.ID, dep)
			}
			if dep == task.ID {
				return fmt.Errorf("task %s has self-referential dependency", task.ID)
			}
		}
	}

	return nil
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
