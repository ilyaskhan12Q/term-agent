package unit

import (
	"context"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/events"
	"github.com/ilyaskhan/term-agent/internal/workflow"
)

type mockWorkflow struct {
	name   string
	wType  workflow.WorkflowType
	status workflow.WorkflowStatus
}

func (m *mockWorkflow) Name() string                    { return m.name }
func (m *mockWorkflow) Type() workflow.WorkflowType     { return m.wType }
func (m *mockWorkflow) Status() workflow.WorkflowStatus { return m.status }
func (m *mockWorkflow) Initialize(ctx context.Context, input string) error {
	m.status = workflow.WorkflowStatusPlanning
	return nil
}
func (m *mockWorkflow) BuildPlan(ctx context.Context) (*agent.Plan, error) {
	return &agent.Plan{ID: "mock-plan", Goal: "test"}, nil
}
func (m *mockWorkflow) Execute(ctx context.Context, bus events.EventBus) (*workflow.WorkflowResult, error) {
	m.status = workflow.WorkflowStatusCompleted
	return &workflow.WorkflowResult{ID: "res-1", Workflow: m.wType, Output: "done"}, nil
}

func TestWorkflowRegistry(t *testing.T) {
	reg := workflow.NewRegistry()

	mwCoding := &mockWorkflow{name: "Coding Workflow", wType: workflow.WorkflowTypeCoding, status: workflow.WorkflowStatusInitialized}
	mwResearch := &mockWorkflow{name: "Research Workflow", wType: workflow.WorkflowTypeResearch, status: workflow.WorkflowStatusInitialized}

	reg.Register(mwCoding)
	reg.Register(mwResearch)

	list := reg.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 registered workflows, got %d", len(list))
	}

	w, err := reg.Get(workflow.WorkflowTypeResearch)
	if err != nil {
		t.Fatalf("failed to get research workflow: %v", err)
	}
	if w.Name() != "Research Workflow" {
		t.Errorf("expected workflow name 'Research Workflow', got '%s'", w.Name())
	}

	_, err = reg.Get(workflow.WorkflowType("NON_EXISTENT"))
	if err == nil {
		t.Errorf("expected error for non-existent workflow type, got nil")
	}
}
