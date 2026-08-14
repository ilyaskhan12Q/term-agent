package domain

import (
	"errors"
	"time"
)

// ProjectStatus represents the overall status of a research project.
type ProjectStatus string

const (
	ProjectStatusCreated      ProjectStatus = "CREATED"
	ProjectStatusPlanning     ProjectStatus = "PLANNING"
	ProjectStatusExecuting    ProjectStatus = "EXECUTING"
	ProjectStatusSynthesizing ProjectStatus = "SYNTHESIZING"
	ProjectStatusDrafting     ProjectStatus = "DRAFTING"
	ProjectStatusReviewing    ProjectStatus = "REVIEWING"
	ProjectStatusCompleted    ProjectStatus = "COMPLETED"
	ProjectStatusFailed       ProjectStatus = "FAILED"
)

// ResearchProject represents the top-level research objective and execution lifecycle.
type ResearchProject struct {
	ID         string        `json:"id"`
	SessionID  string        `json:"session_id"`
	Title      string        `json:"title"`
	Objective  string        `json:"objective"`
	TemplateID string        `json:"template_id"`
	Status     ProjectStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// NewResearchProject constructs a new ResearchProject with default template and status.
func NewResearchProject(id, sessionID, title, objective, templateID string) (*ResearchProject, error) {
	if id == "" {
		return nil, errors.New("research project ID cannot be empty")
	}
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}
	if objective == "" {
		return nil, errors.New("research objective cannot be empty")
	}
	if templateID == "" {
		templateID = "academic_research"
	}
	now := time.Now()
	return &ResearchProject{
		ID:         id,
		SessionID:  sessionID,
		Title:      title,
		Objective:  objective,
		TemplateID: templateID,
		Status:     ProjectStatusCreated,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
