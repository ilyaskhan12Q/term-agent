package agent

import (
	"context"
	"time"

	"github.com/ilyaskhan/term-agent/internal/tools"
)

// AgentRole defines the responsibility scope of an agent.
type AgentRole string

const (
	RoleOrchestrator AgentRole = "ORCHESTRATOR"
	RolePlanner      AgentRole = "PLANNER"
	RoleWorker       AgentRole = "WORKER"
	RoleEvaluator    AgentRole = "EVALUATOR"
)

// AgentStatus represents the current state of an agent.
type AgentStatus string

const (
	AgentStatusIdle      AgentStatus = "IDLE"
	AgentStatusPlanning  AgentStatus = "PLANNING"
	AgentStatusExecuting AgentStatus = "EXECUTING"
	AgentStatusWaiting   AgentStatus = "WAITING_APPROVAL"
	AgentStatusCompleted AgentStatus = "COMPLETED"
	AgentStatusFailed    AgentStatus = "FAILED"
)

// Agent defines the fundamental contract for AI agents in term-agent.
type Agent interface {
	ID() string
	Role() AgentRole
	Status() AgentStatus
	ExecuteStep(ctx context.Context, input string) (*StepResult, error)
	Cancel() error
}

// StepResult represents the output of a single reasoning/execution step.
type StepResult struct {
	AgentID      string
	Thought      string
	ToolCalls    []tools.ToolCallSpec
	IsDone       bool
	FinalMessage string
	Timestamp    time.Time
}
