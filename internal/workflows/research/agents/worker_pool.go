package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/persistence/repository"
	"github.com/ilyaskhan/term-agent/internal/tools"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
)

// WorkerPoolOptions configures parallel execution parameters for specialist researchers.
type WorkerPoolOptions struct {
	MaxConcurrency int
	ToolRegistry   *tools.Registry
	ResearchRepo   *repository.SQLiteResearchRepository
	Tracker        *provenance.ProvenanceTracker
}

// WorkerPool manages parallel execution of specialist research workers.
type WorkerPool struct {
	maxConcurrency int
	toolRegistry   *tools.Registry
	researchRepo   *repository.SQLiteResearchRepository
	tracker        *provenance.ProvenanceTracker
}

// NewWorkerPool creates a new WorkerPool instance.
func NewWorkerPool(opts WorkerPoolOptions) *WorkerPool {
	concurrency := opts.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 4
	}
	return &WorkerPool{
		maxConcurrency: concurrency,
		toolRegistry:   opts.ToolRegistry,
		researchRepo:   opts.ResearchRepo,
		tracker:        opts.Tracker,
	}
}

// ExecuteTasks runs worker tasks concurrently using a bounded worker pool.
func (p *WorkerPool) ExecuteTasks(ctx context.Context, projectID string, tasks []*agent.TaskSpec) ([]*domain.ResearchFinding, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	// Filter tasks assigned to worker agents
	var workerTasks []*agent.TaskSpec
	for _, t := range tasks {
		if t.AssignedTo != "SYNTHESIS_AGENT" && t.AssignedTo != "PLANNER_AGENT" {
			workerTasks = append(workerTasks, t)
		}
	}

	if len(workerTasks) == 0 {
		return nil, nil
	}

	sem := make(chan struct{}, p.maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var findings []*domain.ResearchFinding
	var execErrors []error

	for i, task := range workerTasks {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, t *agent.TaskSpec) {
			defer wg.Done()
			defer func() { <-sem }()

			workerID := fmt.Sprintf("worker-%s-%d-%d", t.AssignedTo, idx+1, time.Now().UnixNano())
			worker := NewResearchWorkerAgent(workerID, t.AssignedTo, p.toolRegistry)

			stepRes, err := worker.ExecuteStepWithProject(ctx, projectID, t.Description)
			if err != nil {
				mu.Lock()
				execErrors = append(execErrors, fmt.Errorf("task %s failed: %w", t.ID, err))
				mu.Unlock()
				return
			}

			questionID := fmt.Sprintf("q-%s-%d", projectID, idx+1)
			finding, err := worker.GenerateFinding(projectID, questionID, t.ID, stepRes)
			if err != nil {
				mu.Lock()
				execErrors = append(execErrors, fmt.Errorf("finding generation for task %s failed: %w", t.ID, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			findings = append(findings, finding)

			// Persist to repository if available
			if p.researchRepo != nil {
				_ = p.researchRepo.SaveFinding(ctx, finding)
			}

			// Register provenance if tracker available
			if p.tracker != nil {
				for _, src := range finding.Sources {
					_, _ = p.tracker.RegisterSource(src)
				}
				for _, ev := range finding.Evidence {
					_ = p.tracker.RegisterEvidence(ev)
				}
				for _, cl := range finding.Claims {
					_ = p.tracker.RegisterClaim(cl)
				}
			}
			mu.Unlock()
		}(i, task)
	}

	wg.Wait()

	if len(execErrors) > 0 && len(findings) == 0 {
		return nil, fmt.Errorf("all %d worker tasks failed: %v", len(execErrors), execErrors[0])
	}

	return findings, nil
}
