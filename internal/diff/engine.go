package diff

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
