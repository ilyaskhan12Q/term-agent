package context

// BudgetCalculator calculates context window allocation and limits.
type BudgetCalculator interface {
	RemainingTokens(usedTokens, maxTokens int) int
	IsNearLimit(usedTokens, maxTokens int, thresholdRatio float64) bool
}
