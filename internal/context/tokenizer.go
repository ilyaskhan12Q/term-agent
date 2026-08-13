package context

// Tokenizer defines the contract for estimating token counts.
type Tokenizer interface {
	CountTokens(text string) int
}

// SimpleEstimator provides baseline token estimation (approx 4 chars per token).
type SimpleEstimator struct{}

func (e *SimpleEstimator) CountTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4
}
