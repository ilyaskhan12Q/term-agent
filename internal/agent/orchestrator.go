package agent

import (
	"context"
)

// Orchestrator defines the contract for coordinating multi-agent task execution.
type Orchestrator interface {
	PlanTask(ctx context.Context, userRequest string) (*Plan, error)
	DispatchWorker(ctx context.Context, task *TaskSpec) (Agent, error)
	SynthesizeResults(ctx context.Context, taskResults []*TaskResult) (string, error)
}

// Plan represents an architectural breakdown of tasks.
type Plan struct {
	ID        string
	Goal      string
	Tasks     []*TaskSpec
	CreatedAt int64
}

// TaskSpec represents an individual sub-task in a plan.
type TaskSpec struct {
	ID           string
	Description  string
	Dependencies []string
	Role         AgentRole
	AssignedTo   string
}

// TaskResult contains the execution output of a sub-task.
type TaskResult struct {
	TaskID    string
	Success   bool
	Output    string
	Error     string
	Mutations []string
}
