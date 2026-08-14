package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// TaskNode represents a task item in the scheduler DAG.
type TaskNode struct {
	ID           string
	Title        string
	Agent        string
	Status       string // "PENDING", "RUNNING", "COMPLETE", "FAILED"
	Dependencies []string
}

// PlanView displays interactive task dependency DAG tree.
type PlanView struct {
	styles *styles.Styles
	width  int
	height int
	tasks  []TaskNode
}

// NewPlanView creates PlanView instance.
func NewPlanView(s *styles.Styles) *PlanView {
	return &PlanView{
		styles: s,
		tasks: []TaskNode{
			{ID: "task-1", Title: "Decompose Research Topic into Sub-Goals", Agent: "planner-agent", Status: "COMPLETE", Dependencies: nil},
			{ID: "task-2", Title: "Search Academic Literature & Papers", Agent: "literature-agent", Status: "RUNNING", Dependencies: []string{"task-1"}},
			{ID: "task-3", Title: "Extract Claims & PDF Evidence Snippets", Agent: "extraction-agent", Status: "PENDING", Dependencies: []string{"task-2"}},
			{ID: "task-4", Title: "Verify Citations & Evidence Consistency", Agent: "verification-agent", Status: "PENDING", Dependencies: []string{"task-3"}},
			{ID: "task-5", Title: "Synthesize Paper & Apply Skeleton Template", Agent: "synthesis-agent", Status: "PENDING", Dependencies: []string{"task-4"}},
		},
	}
}

// SetSize updates dimensions.
func (v *PlanView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// SetTasks replaces current DAG task tree.
func (v *PlanView) SetTasks(tasks []TaskNode) {
	v.tasks = tasks
}

// Render draws the task DAG graph view.
func (v *PlanView) Render() string {
	var b strings.Builder
	b.WriteString(v.styles.PanelTitle.Render("🌿 Task Execution DAG Plan"))
	b.WriteString("\n\n")

	for i, task := range v.tasks {
		statusStr := ""
		switch task.Status {
		case "COMPLETE":
			statusStr = v.styles.TaskComplete.Render("[✔ COMPLETE]")
		case "RUNNING":
			statusStr = v.styles.TaskRunning.Render("[◐ RUNNING ]")
		case "FAILED":
			statusStr = v.styles.TaskFailed.Render("[❌ FAILED  ]")
		default:
			statusStr = v.styles.TaskPending.Render("[⏳ PENDING ]")
		}

		branch := "├─"
		if i == len(v.tasks)-1 {
			branch = "└─"
		}

		agentBadge := v.styles.Badge.Render(task.Agent)
		depsStr := ""
		if len(task.Dependencies) > 0 {
			depsStr = v.styles.DiffContext.Render(" (depends on: " + strings.Join(task.Dependencies, ", ") + ")")
		}

		b.WriteString(fmt.Sprintf("%s %s  %-40s  %s%s\n", branch, statusStr, task.Title, agentBadge, depsStr))
		if i < len(v.tasks)-1 {
			b.WriteString("│\n")
		}
	}

	content := b.String()

	panelStyle := lipgloss.NewStyle().
		Width(v.width - 4).
		Height(v.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(v.styles.Tokens.BorderColor)).
		Padding(1)

	return panelStyle.Render(content)
}
