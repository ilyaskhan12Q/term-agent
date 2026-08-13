package agent

import (
	"context"
)

// Worker defines the contract for sub-agent worker execution.
type Worker interface {
	Agent
	ExecuteTask(ctx context.Context, task *TaskSpec) (*TaskResult, error)
}
