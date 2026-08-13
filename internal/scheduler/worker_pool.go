package scheduler

import (
	"context"

	"github.com/ilyaskhan/term-agent/internal/agent"
)

// WorkerPool defines the contract for bounded parallel execution workers.
type WorkerPool interface {
	Submit(ctx context.Context, task *agent.TaskSpec, worker agent.Worker) error
	Shutdown(ctx context.Context) error
	Capacity() int
}
