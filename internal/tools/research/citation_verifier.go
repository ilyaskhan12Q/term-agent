package research

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/security"
	"github.com/ilyaskhan/term-agent/internal/tools"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

type CitationVerifierArgs struct {
	ClaimStatement string `json:"claim_statement"`
	EvidenceID     string `json:"evidence_id"`
	Snippet        string `json:"snippet"`
}

type CitationVerificationResult struct {
	ClaimStatement     string                      `json:"claim_statement"`
	EvidenceID         string                      `json:"evidence_id"`
	VerificationStatus domain.EvidenceVerification `json:"verification_status"`
	MatchConfidence    float64                     `json:"match_confidence"`
	Reasoning          string                      `json:"reasoning"`
}

type CitationVerifierTool struct{}

func NewCitationVerifierTool() *CitationVerifierTool {
	return &CitationVerifierTool{}
}

func (t *CitationVerifierTool) Spec() tools.ToolSpec {
	params, _ := json.Marshal(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"claim_statement": map[string]interface{}{
				"type":        "string",
				"description": "The research claim or assertion to verify",
			},
			"evidence_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the evidence being cited",
			},
			"snippet": map[string]interface{}{
				"type":        "string",
				"description": "Exact text snippet extracted from source document",
			},
		},
		"required": []string{"claim_statement", "snippet"},
	})

	return tools.ToolSpec{
		Name:        "citation_verifier",
		Description: "Verifies whether a research claim is directly supported, contradicted, or unverified by an evidence snippet using entailment and negation analysis.",
		Parameters:  params,
		RiskLevel:   tools.RiskLevelRead,
	}
}

func (t *CitationVerifierTool) ValidateArgs(args json.RawMessage) error {
	var a CitationVerifierArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return fmt.Errorf("invalid arguments for citation_verifier: %w", err)
	}
	if strings.TrimSpace(a.ClaimStatement) == "" {
		return errors.New("claim_statement is required")
	}
	if strings.TrimSpace(a.Snippet) == "" {
		return errors.New("snippet is required")
	}
	return nil
}

func (t *CitationVerifierTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	start := time.Now()
	var a CitationVerifierArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    err.Error(),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	claimClean := security.SanitizeUntrustedInput(a.ClaimStatement)
	snippetClean := security.SanitizeUntrustedInput(a.Snippet)

	status, confidence, reasoning := EvaluateClaimEvidenceEntailment(claimClean, snippetClean)

	res := CitationVerificationResult{
		ClaimStatement:     claimClean,
		EvidenceID:         a.EvidenceID,
		VerificationStatus: status,
		MatchConfidence:    confidence,
		Reasoning:          reasoning,
	}

	outJSON, err := json.Marshal(res)
	if err != nil {
		return &tools.ToolResult{
			Output:   "",
			Error:    fmt.Sprintf("failed to serialize verification result: %v", err),
			Duration: time.Since(start),
			IsError:  true,
		}, nil
	}

	return &tools.ToolResult{
		Output:   string(outJSON),
		Duration: time.Since(start),
		IsError:  false,
	}, nil
}

// EvaluateClaimEvidenceEntailment performs token overlap, negation check, and contradiction detection.
func EvaluateClaimEvidenceEntailment(claim, snippet string) (domain.EvidenceVerification, float64, string) {
	claimLower := strings.ToLower(claim)
	snippetLower := strings.ToLower(snippet)

	// Tokenize words
	reWord := regexp.MustCompile(`\b\w+\b`)
	claimWords := reWord.FindAllString(claimLower, -1)
	snippetWords := reWord.FindAllString(snippetLower, -1)

	if len(claimWords) == 0 || len(snippetWords) == 0 {
		return domain.EvidenceStatusUnverified, 0.0, "Empty or non-textual input provided."
	}

	snippetMap := make(map[string]bool)
	for _, w := range snippetWords {
		if len(w) > 2 { // filter tiny stop words
			snippetMap[w] = true
		}
	}

	matchCount := 0
	totalKeywords := 0
	for _, w := range claimWords {
		if len(w) > 2 {
			totalKeywords++
			if snippetMap[w] {
				matchCount++
			}
		}
	}

	overlapRatio := 0.0
	if totalKeywords > 0 {
		overlapRatio = float64(matchCount) / float64(totalKeywords)
	}

	// Detect negation mismatch
	claimHasNegation := containsNegation(claimLower)
	snippetHasNegation := containsNegation(snippetLower)

	if claimHasNegation != snippetHasNegation && overlapRatio > 0.4 {
		return domain.EvidenceStatusMismatch, 0.85,
			"Contradiction detected: Evidence snippet contains a negation mismatch relative to the claim."
	}

	// Status resolution based on overlap confidence
	if overlapRatio >= 0.50 {
		return domain.EvidenceStatusVerified, mathMin(0.98, 0.50+overlapRatio*0.50),
			fmt.Sprintf("Direct support confirmed: High semantic token overlap (%.1f%%).", overlapRatio*100)
	}

	if overlapRatio >= 0.25 {
		return domain.EvidenceStatusUnverified, 0.60,
			fmt.Sprintf("Partial support: Moderate keyword overlap (%.1f%%) requiring further verification.", overlapRatio*100)
	}

	return domain.EvidenceStatusMismatch, 0.30,
		fmt.Sprintf("Unverified/Mismatch: Insufficient keyword overlap (%.1f%%) between claim and evidence snippet.", overlapRatio*100)
}

func containsNegation(s string) bool {
	negations := []string{"not", "no", "never", "fail", "failed", "unlikely", "disprove", "reject", "absence"}
	for _, neg := range negations {
		if strings.Contains(s, neg) {
			return true
		}
	}
	return false
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
