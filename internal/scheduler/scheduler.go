package scheduler

import (
	"context"

	"github.com/ilyaskhan/term-agent/internal/agent"
)

// Scheduler defines the contract for orchestrating concurrent agent task execution.
type Scheduler interface {
	Schedule(ctx context.Context, plan *agent.Plan) (<-chan *agent.TaskResult, error)
	Cancel(ctx context.Context) error
	Status() SchedulerStatus
}

// SchedulerStatus represents worker pool metrics and queue state.
type SchedulerStatus struct {
	TotalWorkers   int
	ActiveWorkers  int
	QueuedTasks    int
	CompletedTasks int
	FailedTasks    int
}
