package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// DiffView renders file diffs for user approval.
type DiffView struct {
	styles        *styles.Styles
	width         int
	height        int
	TransactionID string
	FilePath      string
	DiffText      string
	HasPending    bool
}

// NewDiffView creates DiffView instance.
func NewDiffView(s *styles.Styles) *DiffView {
	sampleDiff := `--- a/internal/config/config.go
+++ b/internal/config/config.go
@@ -10,6 +10,8 @@ type Config struct {
 	DefaultModel string
 	LogLevel     string
+	// Research Agent Mode Configuration
+	EnableResearchMode bool
+	MaxLiteratureDepth int
 }`
	return &DiffView{
		styles:        s,
		TransactionID: "tx-8f92a1",
		FilePath:      "internal/config/config.go",
		DiffText:      sampleDiff,
		HasPending:    true,
	}
}

// SetSize updates dimensions.
func (v *DiffView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// SetDiff updates pending diff data.
func (v *DiffView) SetDiff(txID, path, diff string) {
	v.TransactionID = txID
	v.FilePath = path
	v.DiffText = diff
	v.HasPending = true
}

// Render draws the diff view.
func (v *DiffView) Render() string {
	var b strings.Builder
	b.WriteString(v.styles.PanelTitle.Render("🔍 Diff Review & Transaction Authorization"))
	b.WriteString("\n\n")

	if !v.HasPending {
		b.WriteString(v.styles.DiffContext.Render("No pending transaction diffs waiting for approval."))
	} else {
		headerInfo := fmt.Sprintf("Transaction: %s  |  Target File: %s\n\n",
			v.styles.BadgeActive.Render(v.TransactionID),
			v.styles.Badge.Render(v.FilePath),
		)
		b.WriteString(headerInfo)

		lines := strings.Split(v.DiffText, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "+") {
				b.WriteString(v.styles.DiffAdd.Render(line) + "\n")
			} else if strings.HasPrefix(line, "-") {
				b.WriteString(v.styles.DiffRemove.Render(line) + "\n")
			} else if strings.HasPrefix(line, "@") {
				b.WriteString(v.styles.DiffHeader.Render(line) + "\n")
			} else {
				b.WriteString(v.styles.DiffContext.Render(line) + "\n")
			}
		}

		b.WriteString("\n\n")
		actionBox := lipgloss.NewStyle().
			Background(lipgloss.Color(v.styles.Tokens.SurfaceColor)).
			Padding(0, 1).
			Render("Press [ y / Enter ] to APPROVE & APPLY  |  Press [ n / Esc ] to REJECT")
		b.WriteString(actionBox)
	}

	content := b.String()

	panelStyle := lipgloss.NewStyle().
		Width(v.width - 4).
		Height(v.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(v.styles.Tokens.BorderColor)).
		Padding(1)

	return panelStyle.Render(content)
}
