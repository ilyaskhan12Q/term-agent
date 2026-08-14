package provenance

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

// ProvenanceReport provides a summary of evidence integrity and claim coverage across a project.
type ProvenanceReport struct {
	TotalSources     int     `json:"total_sources"`
	TotalEvidence    int     `json:"total_evidence"`
	TotalClaims      int     `json:"total_claims"`
	SupportedClaims  int     `json:"supported_claims"`
	CoverageScore    float64 `json:"coverage_score"`    // SupportedClaims / TotalClaims
	VerificationRate float64 `json:"verification_rate"` // Verified Evidence / Total Evidence
}

// ProvenanceTracker maintains in-memory provenance tracking and validation graphs.
type ProvenanceTracker struct {
	mu       sync.RWMutex
	sources  map[string]domain.Source
	evidence map[string]domain.Evidence
	claims   map[string]domain.Claim
}

// NewProvenanceTracker initializes a new ProvenanceTracker.
func NewProvenanceTracker() *ProvenanceTracker {
	return &ProvenanceTracker{
		sources:  make(map[string]domain.Source),
		evidence: make(map[string]domain.Evidence),
		claims:   make(map[string]domain.Claim),
	}
}

// RegisterSource records a research Source entity.
func (t *ProvenanceTracker) RegisterSource(s domain.Source) error {
	if s.ID == "" {
		return errors.New("source ID cannot be empty")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sources[s.ID] = s
	return nil
}

// RegisterEvidence records an Evidence entity linked to a registered Source.
func (t *ProvenanceTracker) RegisterEvidence(e domain.Evidence) error {
	if e.ID == "" {
		return errors.New("evidence ID cannot be empty")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.sources[e.SourceID]; !exists {
		return fmt.Errorf("referenced source ID '%s' is not registered", e.SourceID)
	}
	t.evidence[e.ID] = e
	return nil
}

// RegisterClaim records a Claim entity linked to registered Evidence IDs.
func (t *ProvenanceTracker) RegisterClaim(c domain.Claim) error {
	if c.ID == "" {
		return errors.New("claim ID cannot be empty")
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, evID := range c.EvidenceIDs {
		if _, exists := t.evidence[evID]; !exists {
			return fmt.Errorf("referenced evidence ID '%s' is not registered", evID)
		}
	}
	t.claims[c.ID] = c
	return nil
}

// GetSource retrieves a source by ID.
func (t *ProvenanceTracker) GetSource(id string) (domain.Source, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.sources[id]
	return s, ok
}

// TraceClaim returns the full provenance chain (Claim -> Evidence -> Source) for a given claim ID.
func (t *ProvenanceTracker) TraceClaim(claimID string) (*domain.Claim, []domain.Evidence, []domain.Source, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	claim, ok := t.claims[claimID]
	if !ok {
		return nil, nil, nil, fmt.Errorf("claim '%s' not found", claimID)
	}

	var evList []domain.Evidence
	srcMap := make(map[string]domain.Source)

	for _, evID := range claim.EvidenceIDs {
		ev, ok := t.evidence[evID]
		if ok {
			evList = append(evList, ev)
			if src, ok := t.sources[ev.SourceID]; ok {
				srcMap[src.ID] = src
			}
		}
	}

	var srcList []domain.Source
	for _, src := range srcMap {
		srcList = append(srcList, src)
	}

	return &claim, evList, srcList, nil
}

// GenerateReport computes total coverage and verification statistics.
func (t *ProvenanceTracker) GenerateReport() ProvenanceReport {
	t.mu.RLock()
	defer t.mu.RUnlock()

	totalSrc := len(t.sources)
	totalEv := len(t.evidence)
	totalClaims := len(t.claims)

	verifiedEv := 0
	for _, ev := range t.evidence {
		if ev.VerificationStatus == domain.EvidenceStatusVerified {
			verifiedEv++
		}
	}

	supportedClaims := 0
	for _, c := range t.claims {
		if len(c.EvidenceIDs) > 0 {
			supportedClaims++
		}
	}

	coverageScore := 0.0
	if totalClaims > 0 {
		coverageScore = float64(supportedClaims) / float64(totalClaims)
	}

	verificationRate := 0.0
	if totalEv > 0 {
		verificationRate = float64(verifiedEv) / float64(totalEv)
	}

	return ProvenanceReport{
		TotalSources:     totalSrc,
		TotalEvidence:    totalEv,
		TotalClaims:      totalClaims,
		SupportedClaims:  supportedClaims,
		CoverageScore:    coverageScore,
		VerificationRate: verificationRate,
	}
}

// BuildBibliographyFormats returns a formatted Markdown bibliography block of all registered sources.
func (t *ProvenanceTracker) BuildBibliographyFormats() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.sources) == 0 {
		return "## References\n\n*No sources recorded.*"
	}

	res := "## References\n\n"
	idx := 1
	for _, src := range t.sources {
		authors := "Unknown Authors"
		if len(src.Authors) > 0 {
			authors = fmt.Sprintf("%v", src.Authors)
		}
		res += fmt.Sprintf("[%d] %s. *%s*. %s (%d). [URI: %s]\n", idx, authors, src.Title, src.Publisher, src.Year, src.URI)
		idx++
	}
	return res
}
