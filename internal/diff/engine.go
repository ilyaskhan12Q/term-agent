package diff

import (
	"bytes"
	"strings"
)

// FileDiff represents calculated differences for a single file.
type FileDiff struct {
	OldPath   string
	NewPath   string
	Hunks     []Hunk
	Additions int
	Deletions int
}

// Hunk represents a cohesive block of diff changes.
type Hunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []Line
}

// LineType specifies the type of line in a diff hunk.
type LineType string

const (
	LineContext  LineType = "CONTEXT"
	LineAddition LineType = "ADDITION"
	LineDeletion LineType = "DELETION"
)

// Line represents an individual line in a diff hunk.
type Line struct {
	Type    LineType
	Content string
}

// DiffEngine defines the contract for computing file differences.
type DiffEngine interface {
	ComputeDiff(oldContent, newContent []byte, oldPath, newPath string) (*FileDiff, error)
}

// DefaultDiffEngine implements DiffEngine using a clean line-based LCS (Longest Common Subsequence) diff algorithm.
type DefaultDiffEngine struct{}

// NewDefaultDiffEngine constructs a new DefaultDiffEngine instance.
func NewDefaultDiffEngine() *DefaultDiffEngine {
	return &DefaultDiffEngine{}
}

// ComputeDiff calculates line-based unified diffs between old and new byte content.
func (e *DefaultDiffEngine) ComputeDiff(oldContent, newContent []byte, oldPath, newPath string) (*FileDiff, error) {
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	lcs := computeLCS(oldLines, newLines)

	var allLines []Line
	additions := 0
	deletions := 0

	i, j, k := 0, 0, 0
	for i < len(oldLines) || j < len(newLines) {
		if k < len(lcs) && i == lcs[k].oldIdx && j == lcs[k].newIdx {
			allLines = append(allLines, Line{Type: LineContext, Content: oldLines[i]})
			i++
			j++
			k++
		} else if i < len(oldLines) && (k >= len(lcs) || i < lcs[k].oldIdx) {
			allLines = append(allLines, Line{Type: LineDeletion, Content: oldLines[i]})
			deletions++
			i++
		} else if j < len(newLines) && (k >= len(lcs) || j < lcs[k].newIdx) {
			allLines = append(allLines, Line{Type: LineAddition, Content: newLines[j]})
			additions++
			j++
		}
	}

	hunks := groupIntoHunks(allLines, 3)

	return &FileDiff{
		OldPath:   oldPath,
		NewPath:   newPath,
		Hunks:     hunks,
		Additions: additions,
		Deletions: deletions,
	}, nil
}

type lcsPair struct {
	oldIdx int
	newIdx int
}

func computeLCS(a, b []string) []lcsPair {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var pairs []lcsPair
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			pairs = append(pairs, lcsPair{oldIdx: i - 1, newIdx: j - 1})
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	// Reverse pairs to maintain ascending order
	for l := 0; l < len(pairs)/2; l++ {
		pairs[l], pairs[len(pairs)-1-l] = pairs[len(pairs)-1-l], pairs[l]
	}

	return pairs
}

func groupIntoHunks(lines []Line, contextSize int) []Hunk {
	if len(lines) == 0 {
		return nil
	}

	var hunks []Hunk

	oldLineNum := 1
	newLineNum := 1

	var currentLines []Line
	oldStart, newStart := 1, 1
	oldLinesCount, newLinesCount := 0, 0
	inHunk := false
	contextBuffer := 0

	for _, line := range lines {
		switch line.Type {
		case LineDeletion:
			if !inHunk {
				inHunk = true
				oldStart = oldLineNum
				newStart = newLineNum
			}
			currentLines = append(currentLines, line)
			oldLinesCount++
			oldLineNum++
			contextBuffer = 0

		case LineAddition:
			if !inHunk {
				inHunk = true
				oldStart = oldLineNum
				newStart = newLineNum
			}
			currentLines = append(currentLines, line)
			newLinesCount++
			newLineNum++
			contextBuffer = 0

		case LineContext:
			if inHunk {
				currentLines = append(currentLines, line)
				oldLinesCount++
				newLinesCount++
				contextBuffer++

				if contextBuffer >= contextSize {
					hunks = append(hunks, Hunk{
						OldStart: oldStart,
						OldLines: oldLinesCount,
						NewStart: newStart,
						NewLines: newLinesCount,
						Lines:    currentLines,
					})
					currentLines = nil
					inHunk = false
					oldLinesCount = 0
					newLinesCount = 0
					contextBuffer = 0
				}
			}
			oldLineNum++
			newLineNum++
		}
	}

	if inHunk && len(currentLines) > 0 {
		hunks = append(hunks, Hunk{
			OldStart: oldStart,
			OldLines: oldLinesCount,
			NewStart: newStart,
			NewLines: newLinesCount,
			Lines:    currentLines,
		})
	}

	return hunks
}

func splitLines(content []byte) []string {
	if len(content) == 0 {
		return nil
	}
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	raw := strings.Split(string(normalized), "\n")
	if len(raw) > 0 && raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}
	return raw
}
