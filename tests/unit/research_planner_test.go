package unit

import (
	"context"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/agent"
	ragents "github.com/ilyaskhan/term-agent/internal/workflows/research/agents"
)

func TestResearchPlannerAgent_DecomposeAndValidateDAG(t *testing.T) {
	planner := ragents.NewResearchPlannerAgent("planner-test")

	ctx := context.Background()
	plan, err := planner.Decompose(ctx, "Transformers in Medical Imaging")
	if err != nil {
		t.Fatalf("unexpected decomposition error: %v", err)
	}

	if plan == nil {
		t.Fatalf("expected non-nil plan")
	}

	if len(plan.Tasks) < 4 {
		t.Errorf("expected at least 4 tasks in research plan, got %d", len(plan.Tasks))
	}

	if err := ragents.ValidateDAG(plan); err != nil {
		t.Errorf("expected valid DAG plan, got validation error: %v", err)
	}

	// Verify questions conversion helper
	questions, err := ragents.GenerateQuestions("proj-1", plan)
	if err != nil {
		t.Fatalf("failed to generate questions from plan: %v", err)
	}

	if len(questions) != len(plan.Tasks) {
		t.Errorf("expected %d questions matching tasks count, got %d", len(plan.Tasks), len(questions))
	}
}

func TestResearchPlannerAgent_InvalidPlanDAG(t *testing.T) {
	invalidPlan := &agent.Plan{
		ID:   "invalid-plan",
		Goal: "Test",
		Tasks: []*agent.TaskSpec{
			{
				ID:           "task-1",
				Description:  "Step 1",
				Dependencies: []string{"non-existent-task"},
			},
		},
	}

	if err := ragents.ValidateDAG(invalidPlan); err == nil {
		t.Errorf("expected error for missing dependency in DAG validation")
	}
}

func TestResearchPlannerAgent_Replan(t *testing.T) {
	planner := ragents.NewResearchPlannerAgent("planner-test")
	ctx := context.Background()

	plan, err := planner.Decompose(ctx, "Quantum Computing Error Correction")
	if err != nil {
		t.Fatalf("failed to decompose plan: %v", err)
	}

	replanned, err := planner.Replan(ctx, plan, "task-lit-search", "arXiv rate limit exceeded")
	if err != nil {
		t.Fatalf("failed to replan: %v", err)
	}

	if len(replanned.Tasks) <= len(plan.Tasks) {
		t.Errorf("expected fallback task injected during replanning")
	}
}
