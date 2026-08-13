package unit

import (
	"testing"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/scheduler"
)

func TestDependencyGraphValidDAG(t *testing.T) {
	tasks := []*agent.TaskSpec{
		{ID: "task-1", Dependencies: nil},
		{ID: "task-2", Dependencies: []string{"task-1"}},
		{ID: "task-3", Dependencies: []string{"task-2"}},
	}

	graph := scheduler.NewDependencyGraph(tasks)
	if err := graph.Validate(); err != nil {
		t.Errorf("expected valid DAG, got error: %v", err)
	}

	ready := graph.ReadyTasks(map[string]bool{})
	if len(ready) != 1 || ready[0].ID != "task-1" {
		t.Errorf("expected initial ready task 'task-1', got %v", ready)
	}
}

func TestDependencyGraphCycleDetection(t *testing.T) {
	tasks := []*agent.TaskSpec{
		{ID: "task-a", Dependencies: []string{"task-b"}},
		{ID: "task-b", Dependencies: []string{"task-a"}},
	}

	graph := scheduler.NewDependencyGraph(tasks)
	if err := graph.Validate(); err == nil {
		t.Error("expected error for cyclic dependency, got nil")
	}
}
