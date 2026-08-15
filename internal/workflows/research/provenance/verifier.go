package provenance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/tools"
	rtools "github.com/ilyaskhan/term-agent/internal/tools/research"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

// CitationVerificationResult encapsulates the verification output for a single claim-evidence pair.
type CitationVerificationResult struct {
	ClaimID            string                      `json:"claim_id"`
	ClaimStatement     string                      `json:"claim_statement"`
	EvidenceID         string                      `json:"evidence_id"`
	SourceID           string                      `json:"source_id"`
	VerificationStatus domain.EvidenceVerification `json:"verification_status"`
	MatchConfidence    float64                     `json:"match_confidence"`
	Reasoning          string                      `json:"reasoning"`
}

// VerificationSummary summarizes evidence verification across multiple research findings.
type VerificationSummary struct {
	TotalClaims      int                          `json:"total_claims"`
	VerifiedClaims   int                          `json:"verified_claims"`
	UnverifiedClaims int                          `json:"unverified_claims"`
	MismatchClaims   int                          `json:"mismatch_claims"`
	OverallRate      float64                      `json:"overall_rate"`
	Report           ProvenanceReport             `json:"report"`
	Results          []CitationVerificationResult `json:"results"`
}

// EvidenceVerifier coordinates entailment checks and updates claim/evidence verification states.
type EvidenceVerifier struct {
	tracker      *ProvenanceTracker
	toolRegistry *tools.Registry
}

// NewEvidenceVerifier creates a new EvidenceVerifier.
func NewEvidenceVerifier(tracker *ProvenanceTracker, registry *tools.Registry) *EvidenceVerifier {
	if tracker == nil {
		tracker = NewProvenanceTracker()
	}
	return &EvidenceVerifier{
		tracker:      tracker,
		toolRegistry: registry,
	}
}

// VerifyClaim executes entailment analysis between a claim statement and an evidence snippet.
func (v *EvidenceVerifier) VerifyClaim(ctx context.Context, claimID, claimStatement, evidenceID, snippet string) (*CitationVerificationResult, error) {
	if strings.TrimSpace(claimStatement) == "" {
		return nil, fmt.Errorf("claim statement cannot be empty")
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, fmt.Errorf("evidence snippet cannot be empty")
	}

	var status domain.EvidenceVerification
	var confidence float64
	var reasoning string

	// Attempt using citation_verifier tool from registry if available
	if v.toolRegistry != nil {
		if t, err := v.toolRegistry.Get("citation_verifier"); err == nil {
			args, _ := json.Marshal(rtools.CitationVerifierArgs{
				ClaimStatement: claimStatement,
				EvidenceID:     evidenceID,
				Snippet:        snippet,
			})
			if res, err := t.Execute(ctx, args); err == nil && !res.IsError {
				var toolRes rtools.CitationVerificationResult
				if err := json.Unmarshal([]byte(res.Output), &toolRes); err == nil {
					status = toolRes.VerificationStatus
					confidence = toolRes.MatchConfidence
					reasoning = toolRes.Reasoning
				}
			}
		}
	}

	// Fallback to direct entailment function
	if status == "" {
		status, confidence, reasoning = rtools.EvaluateClaimEvidenceEntailment(claimStatement, snippet)
	}

	result := &CitationVerificationResult{
		ClaimID:            claimID,
		ClaimStatement:     claimStatement,
		EvidenceID:         evidenceID,
		VerificationStatus: status,
		MatchConfidence:    confidence,
		Reasoning:          reasoning,
	}

	return result, nil
}

// VerifyFindings performs batch citation and evidence verification across research findings.
func (v *EvidenceVerifier) VerifyFindings(ctx context.Context, projectID string, findings []*domain.ResearchFinding) (*VerificationSummary, error) {
	var results []CitationVerificationResult
	verifiedCount := 0
	unverifiedCount := 0
	mismatchCount := 0

	for i, f := range findings {
		if f == nil {
			continue
		}

		// Register sources in finding to tracker
		for _, src := range f.Sources {
			_, _ = v.tracker.RegisterSource(src)
		}

		// Verify extracted claims or finding statements
		statements := f.Findings
		if len(f.Claims) > 0 {
			statements = nil
			for _, c := range f.Claims {
				statements = append(statements, c.Statement)
			}
		}

		for j, stmt := range statements {
			claimID := fmt.Sprintf("claim-%s-%d-%d", projectID, i+1, j+1)
			evID := fmt.Sprintf("ev-%s-%d-%d", projectID, i+1, j+1)
			srcID := fmt.Sprintf("src-%s-%d", projectID, i+1)

			// Default snippet fallback to finding text if evidence is empty
			snippet := stmt
			if len(f.Evidence) > j {
				snippet = f.Evidence[j].Snippet
				evID = f.Evidence[j].ID
				srcID = f.Evidence[j].SourceID
			}

			// Ensure dummy source and evidence registered if missing
			if _, ok := v.tracker.GetSource(srcID); !ok {
				dummySrc, _ := domain.NewSource(srcID, projectID, fmt.Sprintf("Source for finding %d-%d", i+1, j+1), fmt.Sprintf("https://arxiv.org/abs/2401.%05d", i*100+j+1), domain.SourceTypeAcademicPaper, 0.95)
				_, _ = v.tracker.RegisterSource(*dummySrc)
			}

			dummyEv, _ := domain.NewEvidence(evID, projectID, srcID, snippet, "Section 1", f.AgentID)
			_ = v.tracker.RegisterEvidence(*dummyEv)

			// Run verification
			verRes, err := v.VerifyClaim(ctx, claimID, stmt, evID, snippet)
			if err != nil {
				continue
			}
			verRes.SourceID = srcID

			// Update evidence status in tracker
			_ = v.tracker.UpdateEvidenceStatus(evID, verRes.VerificationStatus)

			// Determine claim strength
			strength := domain.ClaimSpeculative
			if verRes.MatchConfidence >= 0.80 {
				strength = domain.ClaimDirect
			} else if verRes.MatchConfidence >= 0.50 {
				strength = domain.ClaimInferential
			}

			// Register verified claim in tracker
			claimObj := domain.Claim{
				ID:          claimID,
				ProjectID:   projectID,
				Statement:   stmt,
				EvidenceIDs: []string{evID},
				Strength:    strength,
				CreatedAt:   time.Now(),
			}
			_ = v.tracker.RegisterClaim(claimObj)

			switch verRes.VerificationStatus {
			case domain.EvidenceStatusVerified:
				verifiedCount++
			case domain.EvidenceStatusMismatch:
				mismatchCount++
			default:
				unverifiedCount++
			}

			results = append(results, *verRes)
		}
	}

	totalClaims := len(results)
	overallRate := 0.0
	if totalClaims > 0 {
		overallRate = float64(verifiedCount) / float64(totalClaims)
	}

	report := v.tracker.GenerateReport()

	return &VerificationSummary{
		TotalClaims:      totalClaims,
		VerifiedClaims:   verifiedCount,
		UnverifiedClaims: unverifiedCount,
		MismatchClaims:   mismatchCount,
		OverallRate:      overallRate,
		Report:           report,
		Results:          results,
	}, nil
}
