package domain

import (
	"errors"
	"time"
)

// ResearchFinding represents the mandatory structured result produced by any research agent execution step.
// Uncontrolled raw prose outputs are explicitly disallowed.
type ResearchFinding struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	QuestionID  string     `json:"question_id"`
	TaskID      string     `json:"task_id"`
	AgentID     string     `json:"agent_id"`
	AgentType   string     `json:"agent_type"`
	Scope       string     `json:"scope"`
	Findings    []string   `json:"findings"`
	Evidence    []Evidence `json:"evidence"`
	Sources     []Source   `json:"sources"`
	Claims      []Claim    `json:"claims"`
	Limitations []string   `json:"limitations"`
	Confidence  float64    `json:"confidence"` // Range: 0.0 to 1.0
	CreatedAt   time.Time  `json:"created_at"`
}

// Validate checks whether the ResearchFinding fulfills structured output integrity requirements.
func (rf *ResearchFinding) Validate() error {
	if rf.ID == "" {
		return errors.New("finding ID is required")
	}
	if rf.ProjectID == "" {
		return errors.New("finding project ID is required")
	}
	if rf.QuestionID == "" {
		return errors.New("finding question ID is required")
	}
	if rf.AgentType == "" {
		return errors.New("finding agent type is required")
	}
	if rf.Confidence < 0.0 || rf.Confidence > 1.0 {
		return errors.New("finding confidence score must be between 0.0 and 1.0")
	}
	return nil
}
