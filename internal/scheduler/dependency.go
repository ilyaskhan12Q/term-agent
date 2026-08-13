package scheduler

import (
	"fmt"

	"github.com/ilyaskhan/term-agent/internal/agent"
)

// DependencyGraph manages task dependencies and topological ordering.
type DependencyGraph struct {
	tasks map[string]*agent.TaskSpec
}

// NewDependencyGraph constructs a dependency graph.
func NewDependencyGraph(tasks []*agent.TaskSpec) *DependencyGraph {
	m := make(map[string]*agent.TaskSpec)
	for _, t := range tasks {
		m[t.ID] = t
	}
	return &DependencyGraph{tasks: m}
}

// Validate checks for cycles in the dependency graph using topological sorting.
func (g *DependencyGraph) Validate() error {
	if len(g.tasks) == 0 {
		return nil
	}

	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for id := range g.tasks {
		inDegree[id] = 0
	}

	for id, t := range g.tasks {
		for _, dep := range t.Dependencies {
			// If dependency exists in the task map
			if _, exists := g.tasks[dep]; exists {
				adj[dep] = append(adj[dep], id)
				inDegree[id]++
			}
		}
	}

	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visitedCount := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visitedCount++

		for _, neighbor := range adj[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if visitedCount != len(g.tasks) {
		return fmt.Errorf("cyclic dependency detected in task graph")
	}

	return nil
}

// ReadyTasks returns tasks whose dependencies are satisfied.
func (g *DependencyGraph) ReadyTasks(completedIDs map[string]bool) []*agent.TaskSpec {
	var ready []*agent.TaskSpec
	for _, t := range g.tasks {
		if completedIDs[t.ID] {
			continue
		}
		depsMet := true
		for _, dep := range t.Dependencies {
			if !completedIDs[dep] {
				depsMet = false
				break
			}
		}
		if depsMet {
			ready = append(ready, t)
		}
	}
	return ready
}
