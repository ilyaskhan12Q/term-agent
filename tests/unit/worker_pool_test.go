package unit

import (
	"context"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/tools"
	ragents "github.com/ilyaskhan/term-agent/internal/workflows/research/agents"
)

func TestWorkerPool_ParallelExecution(t *testing.T) {
	ctx := context.Background()
	registry := tools.NewRegistry()

	pool := ragents.NewWorkerPool(ragents.WorkerPoolOptions{
		MaxConcurrency: 3,
		ToolRegistry:   registry,
	})

	tasks := []*agent.TaskSpec{
		{
			ID:          "task-1",
			Description: "Search quantum computing literature",
			AssignedTo:  "LITERATURE_AGENT",
		},
		{
			ID:          "task-2",
			Description: "Fetch web results on quantum error correction",
			AssignedTo:  "WEB_RESEARCH_AGENT",
		},
		{
			ID:          "task-3",
			Description: "Extract PDF evidence from arXiv source",
			AssignedTo:  "EXTRACTION_AGENT",
		},
		{
			ID:          "task-4",
			Description: "Verify claim against evidence",
			AssignedTo:  "VERIFICATION_AGENT",
		},
		{
			ID:          "task-5",
			Description: "Synthesize final draft",
			AssignedTo:  "SYNTHESIS_AGENT", // Should be ignored by WorkerPool
		},
	}

	findings, err := pool.ExecuteTasks(ctx, "proj-pool-test", tasks)
	if err != nil {
		t.Fatalf("ExecuteTasks failed: %v", err)
	}

	// SYNTHESIS_AGENT is filtered out, so 4 findings expected
	if len(findings) != 4 {
		t.Errorf("Expected 4 findings from worker agents, got %d", len(findings))
	}

	for _, f := range findings {
		if f.ProjectID != "proj-pool-test" {
			t.Errorf("Expected project ID 'proj-pool-test', got %s", f.ProjectID)
		}
		if f.Confidence <= 0 {
			t.Errorf("Expected positive confidence score, got %f", f.Confidence)
		}
		if len(f.Findings) == 0 {
			t.Error("Expected non-empty findings content")
		}
	}
}
