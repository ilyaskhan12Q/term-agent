package domain

import (
	"errors"
	"time"
)

// QuestionStatus represents the status of an individual research question.
type QuestionStatus string

const (
	QuestionStatusPending    QuestionStatus = "PENDING"
	QuestionStatusInProgress QuestionStatus = "IN_PROGRESS"
	QuestionStatusCompleted  QuestionStatus = "COMPLETED"
	QuestionStatusFailed     QuestionStatus = "FAILED"
)

// ResearchQuestion represents a decomposed research sub-question within a project.
type ResearchQuestion struct {
	ID        string         `json:"id"`
	ProjectID string         `json:"project_id"`
	Question  string         `json:"question"`
	Scope     string         `json:"scope"`
	Priority  int            `json:"priority"`
	Status    QuestionStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

// NewResearchQuestion constructs a validated ResearchQuestion instance.
func NewResearchQuestion(id, projectID, question, scope string, priority int) (*ResearchQuestion, error) {
	if id == "" {
		return nil, errors.New("research question ID cannot be empty")
	}
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}
	if question == "" {
		return nil, errors.New("question text cannot be empty")
	}
	if priority <= 0 {
		priority = 1
	}
	return &ResearchQuestion{
		ID:        id,
		ProjectID: projectID,
		Question:  question,
		Scope:     scope,
		Priority:  priority,
		Status:    QuestionStatusPending,
		CreatedAt: time.Now(),
	}, nil
}
