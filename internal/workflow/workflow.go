package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/events"
)

// WorkflowType identifies the operational mode of a workflow.
type WorkflowType string

const (
	WorkflowTypeCoding   WorkflowType = "CODING"
	WorkflowTypeResearch WorkflowType = "RESEARCH"
)

// WorkflowStatus represents the current state of a workflow execution.
type WorkflowStatus string

const (
	WorkflowStatusInitialized WorkflowStatus = "INITIALIZED"
	WorkflowStatusPlanning    WorkflowStatus = "PLANNING"
	WorkflowStatusExecuting   WorkflowStatus = "EXECUTING"
	WorkflowStatusReviewing   WorkflowStatus = "REVIEWING"
	WorkflowStatusCompleted   WorkflowStatus = "COMPLETED"
	WorkflowStatusFailed      WorkflowStatus = "FAILED"
)

// WorkflowResult contains the output artifacts and payload of a workflow execution.
type WorkflowResult struct {
	ID        string        `json:"id"`
	Workflow  WorkflowType  `json:"workflow"`
	Output    string        `json:"output"`
	Data      interface{}   `json:"data,omitempty"`
	Artifacts []string      `json:"artifacts,omitempty"`
	Duration  time.Duration `json:"duration"`
	Error     error         `json:"-"`
}

// Workflow defines the execution contract for specialized agent workflows.
type Workflow interface {
	Name() string
	Type() WorkflowType
	Initialize(ctx context.Context, input string) error
	BuildPlan(ctx context.Context) (*agent.Plan, error)
	Execute(ctx context.Context, bus events.EventBus) (*WorkflowResult, error)
	Status() WorkflowStatus
}

// Registry manages registered workflows.
type Registry struct {
	mu        sync.RWMutex
	workflows map[WorkflowType]Workflow
}

// NewRegistry constructs a new workflow registry.
func NewRegistry() *Registry {
	return &Registry{
		workflows: make(map[WorkflowType]Workflow),
	}
}

// Register registers a workflow handler.
func (r *Registry) Register(w Workflow) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workflows[w.Type()] = w
}

// Get retrieves a workflow handler by type.
func (r *Registry) Get(wType WorkflowType) (Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.workflows[wType]
	if !ok {
		return nil, fmt.Errorf("workflow type not found: %s", wType)
	}
	return w, nil
}

// List returns all registered workflow types.
func (r *Registry) List() []WorkflowType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]WorkflowType, 0, len(r.workflows))
	for t := range r.workflows {
		types = append(types, t)
	}
	return types
}
