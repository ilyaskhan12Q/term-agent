package provenance

import (
	"errors"
	"fmt"
	"strings"
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
	ConflictCount    int     `json:"conflict_count"`
}

// ConflictPair records conflicting evidence statements found during provenance checks.
type ConflictPair struct {
	EvidenceIDA string `json:"evidence_id_a"`
	EvidenceIDB string `json:"evidence_id_b"`
	Reason      string `json:"reason"`
}

// ProvenanceTracker maintains in-memory provenance tracking and validation graphs.
type ProvenanceTracker struct {
	mu       sync.RWMutex
	sources  map[string]domain.Source
	evidence map[string]domain.Evidence
	claims   map[string]domain.Claim
	uriMap   map[string]string // URI -> Source ID mapping for deduplication
}

// NewProvenanceTracker initializes a new ProvenanceTracker.
func NewProvenanceTracker() *ProvenanceTracker {
	return &ProvenanceTracker{
		sources:  make(map[string]domain.Source),
		evidence: make(map[string]domain.Evidence),
		claims:   make(map[string]domain.Claim),
		uriMap:   make(map[string]string),
	}
}

// RegisterSource records a research Source entity with URI deduplication.
func (t *ProvenanceTracker) RegisterSource(s domain.Source) (string, error) {
	if s.ID == "" {
		return "", errors.New("source ID cannot be empty")
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	// URI deduplication check
	if s.URI != "" {
		normURI := strings.ToLower(strings.TrimSpace(s.URI))
		if existingID, ok := t.uriMap[normURI]; ok {
			return existingID, nil // Return existing source ID
		}
		t.uriMap[normURI] = s.ID
	}

	t.sources[s.ID] = s
	return s.ID, nil
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

// UpdateEvidenceStatus updates the verification status of a registered evidence entity.
func (t *ProvenanceTracker) UpdateEvidenceStatus(id string, status domain.EvidenceVerification) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	ev, exists := t.evidence[id]
	if !exists {
		return fmt.Errorf("evidence ID '%s' not registered", id)
	}
	ev.VerificationStatus = status
	t.evidence[id] = ev
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

// DetectEvidenceConflicts identifies evidence pairs with contradictory claims or status mismatches.
func (t *ProvenanceTracker) DetectEvidenceConflicts() []ConflictPair {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var conflicts []ConflictPair
	evList := make([]domain.Evidence, 0, len(t.evidence))
	for _, ev := range t.evidence {
		evList = append(evList, ev)
	}

	for i := 0; i < len(evList); i++ {
		for j := i + 1; j < len(evList); j++ {
			evA := evList[i]
			evB := evList[j]

			if evA.VerificationStatus == domain.EvidenceStatusMismatch || evB.VerificationStatus == domain.EvidenceStatusMismatch {
				conflicts = append(conflicts, ConflictPair{
					EvidenceIDA: evA.ID,
					EvidenceIDB: evB.ID,
					Reason:      "Verification status mismatch flagged on evidence record.",
				})
			}
		}
	}

	return conflicts
}

// GenerateReport computes total coverage, verification statistics, and conflict counts.
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

	conflicts := t.DetectEvidenceConflicts()

	return ProvenanceReport{
		TotalSources:     totalSrc,
		TotalEvidence:    totalEv,
		TotalClaims:      totalClaims,
		SupportedClaims:  supportedClaims,
		CoverageScore:    coverageScore,
		VerificationRate: verificationRate,
		ConflictCount:    len(conflicts),
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
			authors = strings.Join(src.Authors, ", ")
		}
		res += fmt.Sprintf("[%d] %s. *%s*. %s (%d). [URI: %s]\n", idx, authors, src.Title, src.Publisher, src.Year, src.URI)
		idx++
	}
	return res
}
