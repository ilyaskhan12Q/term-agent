package scheduler

import (
	"github.com/ilyaskhan/term-agent/internal/agent"
)

// TaskQueue defines the contract for thread-safe task queuing.
type TaskQueue interface {
	Enqueue(task *agent.TaskSpec) error
	Dequeue() (*agent.TaskSpec, error)
	Len() int
}
