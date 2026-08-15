package domain

import (
	"errors"
	"time"
)

// SourceType categorizes origin of evidence.
type SourceType string

const (
	SourceTypeAcademicPaper SourceType = "ACADEMIC_PAPER"
	SourceTypeDocumentation SourceType = "DOCUMENTATION"
	SourceTypeWebPage       SourceType = "WEB_PAGE"
	SourceTypeDataset       SourceType = "DATASET"
	SourceTypeBook          SourceType = "BOOK"
)

// Source represents an external reference or published document.
type Source struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"project_id"`
	Title      string     `json:"title"`
	URI        string     `json:"uri"`
	Authors    []string   `json:"authors"`
	Year       int        `json:"year"`
	Publisher  string     `json:"publisher"`
	SourceType SourceType `json:"source_type"`
	TrustScore float64    `json:"trust_score"` // Range: 0.0 to 1.0
	FetchedAt  time.Time  `json:"fetched_at"`
}

// EvidenceVerification indicates the verification state of evidence.
type EvidenceVerification string

const (
	EvidenceStatusVerified     EvidenceVerification = "VERIFIED"
	EvidenceStatusUnverified   EvidenceVerification = "UNVERIFIED"
	EvidenceStatusMismatch     EvidenceVerification = "MISMATCH"
	EvidenceStatusContradicted EvidenceVerification = "CONTRADICTED"
)

// Evidence represents a verified excerpt, data point, or quote extracted from a Source.
type Evidence struct {
	ID                 string               `json:"id"`
	ProjectID          string               `json:"project_id"`
	SourceID           string               `json:"source_id"`
	Snippet            string               `json:"snippet"`
	Location           string               `json:"location"` // e.g. Page 4, Section 3.2, Line 120
	VerificationStatus EvidenceVerification `json:"verification_status"`
	ExtractorAgentID   string               `json:"extractor_agent_id"`
	CreatedAt          time.Time            `json:"created_at"`
}

// ClaimStrength represents the logical validity of a claim.
type ClaimStrength string

const (
	ClaimDirect      ClaimStrength = "DIRECT"
	ClaimInferential ClaimStrength = "INFERENTIAL"
	ClaimSpeculative ClaimStrength = "SPECULATIVE"
)

// Claim represents an assertion made by research agents, backed by one or more Evidence items.
type Claim struct {
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	Statement   string        `json:"statement"`
	EvidenceIDs []string      `json:"evidence_ids"`
	Strength    ClaimStrength `json:"strength"`
	CreatedAt   time.Time     `json:"created_at"`
}

// Citation links a statement in a paper section to a verified Source and Evidence ID.
type Citation struct {
	CitationID string `json:"citation_id"`
	SourceID   string `json:"source_id"`
	EvidenceID string `json:"evidence_id"`
	RefText    string `json:"ref_text"`
}

// NewSource constructs a Source entity with validation.
func NewSource(id, projectID, title, uri string, sourceType SourceType, trustScore float64) (*Source, error) {
	if id == "" || projectID == "" {
		return nil, errors.New("source ID and project ID are required")
	}
	if title == "" {
		return nil, errors.New("source title is required")
	}
	if trustScore < 0.0 || trustScore > 1.0 {
		trustScore = 1.0
	}
	return &Source{
		ID:         id,
		ProjectID:  projectID,
		Title:      title,
		URI:        uri,
		Authors:    []string{},
		SourceType: sourceType,
		TrustScore: trustScore,
		FetchedAt:  time.Now(),
	}, nil
}

// NewEvidence constructs an Evidence entity with validation.
func NewEvidence(id, projectID, sourceID, snippet, location, agentID string) (*Evidence, error) {
	if id == "" || projectID == "" || sourceID == "" {
		return nil, errors.New("evidence ID, project ID, and source ID are required")
	}
	if snippet == "" {
		return nil, errors.New("evidence snippet cannot be empty")
	}
	return &Evidence{
		ID:                 id,
		ProjectID:          projectID,
		SourceID:           sourceID,
		Snippet:            snippet,
		Location:           location,
		VerificationStatus: EvidenceStatusUnverified,
		ExtractorAgentID:   agentID,
		CreatedAt:          time.Now(),
	}, nil
}
