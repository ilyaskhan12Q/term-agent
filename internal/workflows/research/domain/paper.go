package domain

import (
	"errors"
	"time"
)

// PaperStatus represents the draft/review state of a generated paper.
type PaperStatus string

const (
	PaperStatusDraft    PaperStatus = "DRAFT"
	PaperStatusReviewing PaperStatus = "REVIEWING"
	PaperStatusPassed   PaperStatus = "PASSED"
	PaperStatusRejected PaperStatus = "REJECTED"
)

// PaperSection represents an individual section of a generated research document.
type PaperSection struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Order     int        `json:"order"`
	Citations []Citation `json:"citations,omitempty"`
}

// ResearchPaper represents the final synthesized research document structured according to a Global Research Template.
type ResearchPaper struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	TemplateID     string         `json:"template_id"`
	Title          string         `json:"title"`
	Sections       []PaperSection `json:"sections"`
	MarkdownOutput string         `json:"markdown_output"`
	Status         PaperStatus    `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// NewResearchPaper constructs a validated ResearchPaper entity.
func NewResearchPaper(id, projectID, templateID, title string) (*ResearchPaper, error) {
	if id == "" || projectID == "" {
		return nil, errors.New("paper ID and project ID are required")
	}
	if title == "" {
		return nil, errors.New("paper title cannot be empty")
	}
	now := time.Now()
	return &ResearchPaper{
		ID:         id,
		ProjectID:  projectID,
		TemplateID: templateID,
		Title:      title,
		Sections:   []PaperSection{},
		Status:     PaperStatusDraft,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
