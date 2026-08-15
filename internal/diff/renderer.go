package diff

import (
	"fmt"
	"strings"
)

// RenderUnifiedDiff renders a FileDiff into a standard Git-compatible unified diff patch format.
func RenderUnifiedDiff(diff *FileDiff) string {
	if diff == nil || (len(diff.Hunks) == 0 && diff.Additions == 0 && diff.Deletions == 0) {
		return ""
	}

	var sb strings.Builder

	oldPath := diff.OldPath
	if oldPath == "" {
		oldPath = "/dev/null"
	} else {
		oldPath = "a/" + oldPath
	}

	newPath := diff.NewPath
	if newPath == "" {
		newPath = "/dev/null"
	} else {
		newPath = "b/" + newPath
	}

	sb.WriteString(fmt.Sprintf("--- %s\n", oldPath))
	sb.WriteString(fmt.Sprintf("+++ %s\n", newPath))

	for _, hunk := range diff.Hunks {
		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunk.OldStart, hunk.OldLines, hunk.NewStart, hunk.NewLines))
		for _, line := range hunk.Lines {
			switch line.Type {
			case LineAddition:
				sb.WriteString("+" + line.Content + "\n")
			case LineDeletion:
				sb.WriteString("-" + line.Content + "\n")
			case LineContext:
				sb.WriteString(" " + line.Content + "\n")
			}
		}
	}

	return sb.String()
}
