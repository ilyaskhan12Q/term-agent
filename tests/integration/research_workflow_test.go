package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ilyaskhan/term-agent/internal/events"
	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/workflow"
	"github.com/ilyaskhan/term-agent/internal/workflows/research"
)

func TestResearchWorkflowExecution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-research-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "research.db")
	db, err := persistence.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	bus := events.NewInMemoryEventBus()

	// Register workflow in generic registry
	registry := workflow.NewRegistry()
	researchWf := research.NewResearchWorkflow()
	researchWf.SetDatabase(db)
	registry.Register(researchWf)

	// Fetch from generic registry
	wf, err := registry.Get(workflow.WorkflowTypeResearch)
	if err != nil {
		t.Fatalf("failed to retrieve research workflow from registry: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	topic := "Evaluation of Multi-Agent Systems in Autonomous Code Generation"
	if err := wf.Initialize(ctx, topic); err != nil {
		t.Fatalf("workflow initialization failed: %v", err)
	}

	plan, err := wf.BuildPlan(ctx)
	if err != nil {
		t.Fatalf("workflow planning failed: %v", err)
	}
	if len(plan.Tasks) == 0 {
		t.Fatalf("expected plan tasks to be populated")
	}

	res, err := wf.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("workflow execution failed: %v", err)
	}

	if res.Workflow != workflow.WorkflowTypeResearch {
		t.Errorf("expected workflow type %s, got %s", workflow.WorkflowTypeResearch, res.Workflow)
	}
	if res.Output == "" {
		t.Errorf("expected non-empty paper output from research workflow")
	}
	if wf.Status() != workflow.WorkflowStatusCompleted {
		t.Errorf("expected workflow status COMPLETED, got %s", wf.Status())
	}
}
