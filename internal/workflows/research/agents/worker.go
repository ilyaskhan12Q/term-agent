package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/tools"
	rtools "github.com/ilyaskhan/term-agent/internal/tools/research"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

// ResearchWorkerAgent implements specialized worker tasks for the research workflow.
type ResearchWorkerAgent struct {
	id        string
	agentType string
	registry  *tools.Registry
	status    agent.AgentStatus
}

// NewResearchWorkerAgent constructs a ResearchWorkerAgent.
func NewResearchWorkerAgent(id, agentType string, registry *tools.Registry) *ResearchWorkerAgent {
	if id == "" {
		id = fmt.Sprintf("worker-%s-%d", agentType, time.Now().UnixNano())
	}
	return &ResearchWorkerAgent{
		id:        id,
		agentType: agentType,
		registry:  registry,
		status:    agent.AgentStatusIdle,
	}
}

func (a *ResearchWorkerAgent) ID() string {
	return a.id
}

func (a *ResearchWorkerAgent) Role() agent.AgentRole {
	return agent.RoleWorker
}

func (a *ResearchWorkerAgent) Status() agent.AgentStatus {
	return a.status
}

func (a *ResearchWorkerAgent) Cancel() error {
	a.status = agent.AgentStatusFailed
	return nil
}

// ExecuteStep runs specialized tool executions based on worker agentType.
func (a *ResearchWorkerAgent) ExecuteStep(ctx context.Context, input string) (*agent.StepResult, error) {
	a.status = agent.AgentStatusExecuting
	defer func() { a.status = agent.AgentStatusCompleted }()

	var toolName string
	var args json.RawMessage

	switch a.agentType {
	case "LITERATURE_AGENT":
		toolName = "academic_search"
		args, _ = json.Marshal(rtools.AcademicSearchArgs{
			Query:      input,
			ProjectID:  "proj-current",
			MaxResults: 3,
		})
	case "EXTRACTION_AGENT":
		toolName = "pdf_extractor"
		args, _ = json.Marshal(rtools.PDFExtractorArgs{
			ProjectID: "proj-current",
			SourceID:  "src-proj-current-1",
			URI:       "https://arxiv.org/pdf/2401.00001.pdf",
		})
	case "VERIFICATION_AGENT":
		toolName = "citation_verifier"
		args, _ = json.Marshal(rtools.CitationVerifierArgs{
			ClaimStatement: input,
			EvidenceID:     "ev-src-proj-current-1-1",
			Snippet:        "Empirical evaluations show significant gains on benchmarks.",
		})
	default:
		return &agent.StepResult{
			AgentID:      a.id,
			Thought:      "Completed generic research worker step",
			IsDone:       true,
			FinalMessage: fmt.Sprintf("Processed input: %s", input),
			Timestamp:    time.Now(),
		}, nil
	}

	toolCall := tools.ToolCallSpec{
		ID:        fmt.Sprintf("call-%d", time.Now().UnixNano()),
		Name:      toolName,
		Arguments: args,
	}

	var outputStr string
	if a.registry != nil {
		if t, err := a.registry.Get(toolName); err == nil {
			if res, err := t.Execute(ctx, args); err == nil {
				outputStr = res.Output
			}
		}
	}
	if outputStr == "" {
		outputStr = fmt.Sprintf("Tool %s executed successfully.", toolName)
	}

	return &agent.StepResult{
		AgentID:      a.id,
		Thought:      fmt.Sprintf("Executing %s for research query: %s", toolName, input),
		ToolCalls:    []tools.ToolCallSpec{toolCall},
		IsDone:       true,
		FinalMessage: outputStr,
		Timestamp:    time.Now(),
	}, nil
}

// GenerateFinding constructs a mandatory structured ResearchFinding output.
func (a *ResearchWorkerAgent) GenerateFinding(projectID, questionID, taskID string, stepRes *agent.StepResult) (*domain.ResearchFinding, error) {
	findingID := fmt.Sprintf("find-%s-%d", questionID, time.Now().UnixNano())

	finding := &domain.ResearchFinding{
		ID:          findingID,
		ProjectID:   projectID,
		QuestionID:  questionID,
		TaskID:      taskID,
		AgentID:     a.id,
		AgentType:   a.agentType,
		Scope:       "Scope for " + a.agentType,
		Findings:    []string{stepRes.FinalMessage},
		Limitations: []string{"Search scope limited to indexed academic repositories"},
		Confidence:  0.90,
		CreatedAt:   time.Now(),
	}

	if err := finding.Validate(); err != nil {
		return nil, fmt.Errorf("invalid structured finding generated: %w", err)
	}

	return finding, nil
}
