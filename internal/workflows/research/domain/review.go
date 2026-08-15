package domain

import (
	"errors"
	"time"
)

// ReviewVerdict represents the decision outcome of the ReviewerAgent audit.
type ReviewVerdict string

const (
	ReviewVerdictApproved      ReviewVerdict = "APPROVED"
	ReviewVerdictNeedsRevision ReviewVerdict = "NEEDS_REVISION"
	ReviewVerdictRejected      ReviewVerdict = "REJECTED"
)

// PaperReview encapsulates the results of an adversarial claim & hallucination audit.
type PaperReview struct {
	ID                 string        `json:"id"`
	PaperID            string        `json:"paper_id"`
	ProjectID          string        `json:"project_id"`
	FidelityScore      float64       `json:"fidelity_score"`     // 0.0 - 1.0 (Higher is better)
	HallucinationRisk  float64       `json:"hallucination_risk"` // 0.0 - 1.0 (Lower is better)
	Verdict            ReviewVerdict `json:"verdict"`
	UncitedClaims      []string      `json:"uncited_claims,omitempty"`
	ContradictedClaims []string      `json:"contradicted_claims,omitempty"`
	ReviewComments     []string      `json:"review_comments,omitempty"`
	ReviewedAt         time.Time     `json:"reviewed_at"`
}

// NewPaperReview constructs a new PaperReview entity.
func NewPaperReview(id, paperID, projectID string, score float64, verdict ReviewVerdict) (*PaperReview, error) {
	if id == "" || paperID == "" || projectID == "" {
		return nil, errors.New("review ID, paper ID, and project ID are required")
	}
	if score < 0.0 || score > 1.0 {
		return nil, errors.New("fidelity score must be between 0.0 and 1.0")
	}

	hallucinationRisk := 1.0 - score
	if hallucinationRisk < 0.0 {
		hallucinationRisk = 0.0
	}

	return &PaperReview{
		ID:                 id,
		PaperID:            paperID,
		ProjectID:          projectID,
		FidelityScore:      score,
		HallucinationRisk:  hallucinationRisk,
		Verdict:            verdict,
		UncitedClaims:      []string{},
		ContradictedClaims: []string{},
		ReviewComments:     []string{},
		ReviewedAt:         time.Now(),
	}, nil
}
