package agent

import (
	"context"
)

// Planner defines the contract for decomposing complex requests into task DAGs.
type Planner interface {
	Decompose(ctx context.Context, goal string) (*Plan, error)
	Replan(ctx context.Context, plan *Plan, failedTaskID string, reason string) (*Plan, error)
}
