package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ilyaskhan/term-agent/internal/agent"
	"github.com/ilyaskhan/term-agent/internal/model"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
)

var (
	citationTagRegex  = regexp.MustCompile(`\[\d+\]`)
	empiricalNumRegex = regexp.MustCompile(`\b\d+(?:\.\d+)?%|\b\d+(?:\.\d+)?x|\b\d+ (?:ms|seconds|users|nodes|percent|GB|MB|FPS)\b`)
)

// ReviewerAgent performs adversarial audits on synthesized research papers to detect hallucinations and unverified claims.
type ReviewerAgent struct {
	id            string
	modelProvider model.ModelProvider
	status        agent.AgentStatus
}

// NewReviewerAgent constructs a ReviewerAgent instance.
func NewReviewerAgent(id string) *ReviewerAgent {
	if id == "" {
		id = "reviewer-agent-1"
	}
	return &ReviewerAgent{
		id:     id,
		status: agent.AgentStatusIdle,
	}
}

// WithModelProvider sets an optional AI model provider for LLM-driven adversarial review.
func (a *ReviewerAgent) WithModelProvider(provider model.ModelProvider) *ReviewerAgent {
	a.modelProvider = provider
	return a
}

// AuditPaper performs a comprehensive claim-verification and hallucination-risk evaluation on a ResearchPaper.
func (a *ReviewerAgent) AuditPaper(ctx context.Context, paper *domain.ResearchPaper, tracker *provenance.ProvenanceTracker, summary *provenance.VerificationSummary) (*domain.PaperReview, error) {
	if paper == nil {
		return nil, fmt.Errorf("cannot audit nil paper")
	}

	a.status = agent.AgentStatusExecuting
	defer func() { a.status = agent.AgentStatusCompleted }()

	reviewID := fmt.Sprintf("rev-%s", paper.ID)

	var uncitedClaims []string
	var contradictedClaims []string
	var reviewComments []string

	totalPenalties := 0.0

	// 1. Process Mismatches & Contradictions from Verification Summary Results
	if summary != nil {
		for _, res := range summary.Results {
			if res.VerificationStatus == domain.EvidenceStatusMismatch || res.VerificationStatus == domain.EvidenceStatusContradicted {
				msg := fmt.Sprintf("Claim '%s' contradicted or mismatched by evidence (Status: %s)", res.ClaimStatement, res.VerificationStatus)
				contradictedClaims = append(contradictedClaims, msg)
				reviewComments = append(reviewComments, fmt.Sprintf("[Contradiction Critical] %s", msg))
				totalPenalties += 0.25
			}
		}
	}

	// 2. Scan Sections for Empirical Statements without Inline Citations
	for _, sec := range paper.Sections {
		lines := strings.Split(sec.Content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "*(") {
				continue
			}

			// If line contains empirical metric/number or strong claim keywords but no [N] citation
			hasNumber := empiricalNumRegex.MatchString(trimmed)
			hasCitation := citationTagRegex.MatchString(trimmed)
			hasVerifiedBadge := strings.Contains(trimmed, "Verified: Direct Support")

			if hasNumber && !hasCitation && !hasVerifiedBadge {
				claimSnippet := trimmed
				if len(claimSnippet) > 90 {
					claimSnippet = claimSnippet[:90] + "..."
				}
				uncitedMsg := fmt.Sprintf("Section '%s': Uncited quantitative claim: %q", sec.Title, claimSnippet)
				uncitedClaims = append(uncitedClaims, uncitedMsg)
				reviewComments = append(reviewComments, fmt.Sprintf("[Uncited Claim] %s", uncitedMsg))
				totalPenalties += 0.08
			}
		}

		// Check for empty placeholder sections
		if strings.TrimSpace(sec.Content) == "" || strings.Contains(sec.Content, "pending synthesis") {
			emptyMsg := fmt.Sprintf("Section '%s' (%s) contains no synthesized content", sec.Title, sec.ID)
			reviewComments = append(reviewComments, fmt.Sprintf("[Incomplete Section] %s", emptyMsg))
			totalPenalties += 0.15
		}
	}

	// 3. Compute Fidelity Score and Hallucination Risk
	fidelityScore := 1.0 - totalPenalties
	if fidelityScore < 0.0 {
		fidelityScore = 0.0
	}
	if fidelityScore > 1.0 {
		fidelityScore = 1.0
	}

	// 4. Determine Verdict
	var verdict domain.ReviewVerdict
	if len(contradictedClaims) == 0 && fidelityScore >= 0.80 {
		verdict = domain.ReviewVerdictApproved
		reviewComments = append(reviewComments, "[Audit Outcome] Paper meets academic fidelity standards with verified provenance.")
	} else if fidelityScore >= 0.50 {
		verdict = domain.ReviewVerdictNeedsRevision
		reviewComments = append(reviewComments, "[Audit Outcome] Paper requires revisions to address uncited claims or minor contradictions.")
	} else {
		verdict = domain.ReviewVerdictRejected
		reviewComments = append(reviewComments, "[Audit Outcome] Paper rejected due to high hallucination risk or unverified contradictions.")
	}

	// 5. Optional LLM Adversarial Review Enhancement
	if a.modelProvider != nil {
		prompt := fmt.Sprintf("Perform adversarial peer review on this paper:\nTitle: %s\nContent:\n%s\nProvide 2 key review notes.",
			paper.Title, paper.MarkdownOutput)
		req := &model.CompletionRequest{
			Messages: []model.Message{
				{Role: model.RoleUser, Content: prompt},
			},
			Temperature: 0.1,
			MaxTokens:   256,
		}
		if resp, err := a.modelProvider.GenerateCompletion(ctx, req); err == nil && strings.TrimSpace(resp.Content) != "" {
			reviewComments = append(reviewComments, fmt.Sprintf("[LLM Peer Review Note] %s", strings.TrimSpace(resp.Content)))
		}
	}

	review, err := domain.NewPaperReview(reviewID, paper.ID, paper.ProjectID, fidelityScore, verdict)
	if err != nil {
		return nil, fmt.Errorf("failed to create paper review: %w", err)
	}

	review.UncitedClaims = uncitedClaims
	review.ContradictedClaims = contradictedClaims
	review.ReviewComments = reviewComments

	return review, nil
}
