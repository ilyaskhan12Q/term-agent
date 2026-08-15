package unit_test

import (
	"strings"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/diff"
)

func TestDefaultDiffEngine_ComputeAndRender(t *testing.T) {
	oldContent := []byte("line 1\nline 2\nline 3\n")
	newContent := []byte("line 1\nline 2 modified\nline 3\nline 4 added\n")

	engine := diff.NewDefaultDiffEngine()
	fileDiff, err := engine.ComputeDiff(oldContent, newContent, "test.txt", "test.txt")
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	if fileDiff.Additions != 2 {
		t.Errorf("expected 2 additions, got %d", fileDiff.Additions)
	}
	if fileDiff.Deletions != 1 {
		t.Errorf("expected 1 deletion, got %d", fileDiff.Deletions)
	}

	rendered := diff.RenderUnifiedDiff(fileDiff)
	if !strings.Contains(rendered, "--- a/test.txt") || !strings.Contains(rendered, "+++ b/test.txt") {
		t.Errorf("rendered diff missing standard file header: %s", rendered)
	}
	if !strings.Contains(rendered, "-line 2") || !strings.Contains(rendered, "+line 2 modified") {
		t.Errorf("rendered diff missing expected line changes: %s", rendered)
	}
	if !strings.Contains(rendered, "+line 4 added") {
		t.Errorf("rendered diff missing addition line: %s", rendered)
	}
}

func TestDiffEngine_EmptyAndNewFile(t *testing.T) {
	engine := diff.NewDefaultDiffEngine()

	// New file creation diff
	fileDiff, err := engine.ComputeDiff(nil, []byte("hello world\n"), "", "new.txt")
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	if fileDiff.Additions != 1 || fileDiff.Deletions != 0 {
		t.Errorf("expected 1 addition, 0 deletions for new file, got +%d -%d", fileDiff.Additions, fileDiff.Deletions)
	}

	rendered := diff.RenderUnifiedDiff(fileDiff)
	if !strings.Contains(rendered, "--- /dev/null") || !strings.Contains(rendered, "+++ b/new.txt") {
		t.Errorf("expected /dev/null header for file creation, got: %s", rendered)
	}
}
