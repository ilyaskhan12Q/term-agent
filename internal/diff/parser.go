package diff

// UnifiedDiffParser defines the contract for parsing raw unified diff strings.
type UnifiedDiffParser interface {
	Parse(unifiedDiff string) ([]*FileDiff, error)
}
