package unit

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/events"
	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/workflow"
	"github.com/ilyaskhan/term-agent/internal/workflows/research"
	ragents "github.com/ilyaskhan/term-agent/internal/workflows/research/agents"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

func TestResearchWorkflow_FullPipeline(t *testing.T) {
	ctx := context.Background()
	bus := events.NewInMemoryEventBus()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_workflow.db")

	db, err := persistence.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	wf := research.NewResearchWorkflow()
	wf.SetDatabase(db)

	if wf.Name() != "Research Workflow" {
		t.Errorf("Expected workflow name 'Research Workflow', got %s", wf.Name())
	}
	if wf.Type() != workflow.WorkflowTypeResearch {
		t.Errorf("Expected workflow type RESEARCH, got %s", wf.Type())
	}

	topic := "Quantum Machine Learning Algorithms for High-Dimensional Optimization"
	if err := wf.Initialize(ctx, topic); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	res, err := wf.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if res == nil {
		t.Fatal("Expected non-nil WorkflowResult")
	}
	if res.Workflow != workflow.WorkflowTypeResearch {
		t.Errorf("Expected result workflow type RESEARCH, got %s", res.Workflow)
	}
	if res.Output == "" {
		t.Error("Expected non-empty markdown output")
	}

	dataMap, ok := res.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{} in res.Data, got %T", res.Data)
	}
	if dataMap["paper_status"] != "PASSED" {
		t.Errorf("Expected paper_status PASSED, got %v", dataMap["paper_status"])
	}
}

func TestResearchPlannerAgent(t *testing.T) {
	ctx := context.Background()
	planner := ragents.NewResearchPlannerAgent("planner-test")

	plan, err := planner.Decompose(ctx, "Investigate Transformers for Time Series")
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}
	if plan == nil || len(plan.Tasks) == 0 {
		t.Fatal("Expected plan with decomposed tasks")
	}
}

func TestSynthesisAgent(t *testing.T) {
	ctx := context.Background()
	engine, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	synthesizer := ragents.NewSynthesisAgent("synth-test", engine)
	if synthesizer.ID() != "synth-test" {
		t.Errorf("Expected agent ID 'synth-test', got %s", synthesizer.ID())
	}

	res, err := synthesizer.ExecuteStep(ctx, "Synthesize test paper")
	if err != nil || !res.IsDone {
		t.Errorf("ExecuteStep failed or not done: %v", err)
	}
}
