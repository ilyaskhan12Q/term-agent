package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

// ExportFormat represents the target export format for paper compilation.
type ExportFormat string

const (
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatLaTeX    ExportFormat = "latex"
	ExportFormatHTML     ExportFormat = "html"
	ExportFormatJSON     ExportFormat = "json"
)

// ParseExportFormat normalizes raw format strings (e.g., "tex", "md") into ExportFormat.
func ParseExportFormat(input string) (ExportFormat, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "markdown", "md":
		return ExportFormatMarkdown, nil
	case "latex", "tex":
		return ExportFormatLaTeX, nil
	case "html":
		return ExportFormatHTML, nil
	case "json":
		return ExportFormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported export format: %s (supported: markdown, latex, html, json)", input)
	}
}

// PaperWriter handles multi-format research paper compilation and file export.
type PaperWriter struct{}

// NewPaperWriter constructs a new PaperWriter compiler instance.
func NewPaperWriter() *PaperWriter {
	return &PaperWriter{}
}

// Compile converts a ResearchPaper into the specified ExportFormat string.
func (w *PaperWriter) Compile(paper *domain.ResearchPaper, format ExportFormat) (string, error) {
	if paper == nil {
		return "", fmt.Errorf("cannot compile nil paper")
	}

	switch format {
	case ExportFormatMarkdown:
		return w.CompileMarkdown(paper), nil
	case ExportFormatLaTeX:
		return w.CompileLaTeX(paper), nil
	case ExportFormatHTML:
		return w.CompileHTML(paper), nil
	case ExportFormatJSON:
		return w.CompileJSON(paper)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// ExportToFile compiles paper into the requested format and writes it to filePath.
func (w *PaperWriter) ExportToFile(paper *domain.ResearchPaper, format ExportFormat, filePath string) error {
	compiled, err := w.Compile(paper, format)
	if err != nil {
		return fmt.Errorf("failed to compile paper: %w", err)
	}

	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if err := os.WriteFile(filePath, []byte(compiled), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// CompileMarkdown generates clean Markdown representation of the paper.
func (w *PaperWriter) CompileMarkdown(paper *domain.ResearchPaper) string {
	if paper.MarkdownOutput != "" {
		return paper.MarkdownOutput
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", paper.Title))
	sb.WriteString(fmt.Sprintf("**Project ID:** `%s` | **Template:** `%s` | **Status:** `%s` | **Generated:** `%s`\n\n",
		paper.ProjectID, paper.TemplateID, paper.Status, paper.CreatedAt.Format(time.RFC3339)))
	sb.WriteString("---\n\n")

	for _, sec := range paper.Sections {
		sb.WriteString(fmt.Sprintf("## %s\n\n", sec.Title))
		if strings.TrimSpace(sec.Content) != "" {
			sb.WriteString(sec.Content)
			sb.WriteString("\n\n")
		} else {
			sb.WriteString("*(Section content pending synthesis)*\n\n")
		}
	}

	return sb.String()
}

// CompileLaTeX generates a complete, valid LaTeX document.
func (w *PaperWriter) CompileLaTeX(paper *domain.ResearchPaper) string {
	var sb strings.Builder

	sb.WriteString("\\documentclass[11pt, a4paper]{article}\n")
	sb.WriteString("\\usepackage[utf8]{inputenc}\n")
	sb.WriteString("\\usepackage{hyperref}\n")
	sb.WriteString("\\usepackage{booktabs}\n")
	sb.WriteString("\\usepackage{geometry}\n")
	sb.WriteString("\\geometry{margin=1in}\n\n")

	titleEscaped := escapeLaTeX(paper.Title)
	sb.WriteString(fmt.Sprintf("\\title{\\textbf{%s}}\n", titleEscaped))
	sb.WriteString("\\author{term-agent Research Mode}\n")
	sb.WriteString(fmt.Sprintf("\\date{%s}\n\n", paper.CreatedAt.Format("January 2, 2006")))

	sb.WriteString("\\begin{document}\n")
	sb.WriteString("\\maketitle\n\n")

	for _, sec := range paper.Sections {
		secTitle := escapeLaTeX(sec.Title)
		if strings.ToLower(sec.ID) == "abstract" {
			sb.WriteString("\\begin{abstract}\n")
			sb.WriteString(escapeLaTeX(sec.Content))
			sb.WriteString("\n\\end{abstract}\n\n")
		} else {
			sb.WriteString(fmt.Sprintf("\\section{%s}\n", secTitle))
			lines := strings.Split(sec.Content, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					sb.WriteString("\n")
					continue
				}
				if strings.HasPrefix(trimmed, "#") {
					headerText := strings.TrimLeft(trimmed, "# ")
					sb.WriteString(fmt.Sprintf("\\subsection{%s}\n", escapeLaTeX(headerText)))
				} else {
					sb.WriteString(escapeLaTeX(trimmed))
					sb.WriteString("\n")
				}
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\\end{document}\n")
	return sb.String()
}

// CompileHTML generates a standalone, styled HTML5 research document.
func (w *PaperWriter) CompileHTML(paper *domain.ResearchPaper) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("<meta charset=\"UTF-8\">\n")
	sb.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString(fmt.Sprintf("<title>%s</title>\n", htmlEscape(paper.Title)))
	sb.WriteString("<style>\n")
	sb.WriteString("  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1a1a1a; max-width: 900px; margin: 0 auto; padding: 2rem; background-color: #fafafa; }\n")
	sb.WriteString("  .container { background: #ffffff; padding: 3rem; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); border: 1px solid #e5e7eb; }\n")
	sb.WriteString("  h1 { font-size: 2.25rem; color: #111827; margin-bottom: 0.5rem; line-height: 1.25; }\n")
	sb.WriteString("  .metadata { font-size: 0.875rem; color: #6b7280; border-bottom: 1px solid #e5e7eb; padding-bottom: 1rem; margin-bottom: 2rem; }\n")
	sb.WriteString("  h2 { font-size: 1.5rem; color: #1f2937; margin-top: 2rem; margin-bottom: 1rem; border-bottom: 2px solid #f3f4f6; padding-bottom: 0.5rem; }\n")
	sb.WriteString("  h3 { font-size: 1.2rem; color: #374151; margin-top: 1.5rem; }\n")
	sb.WriteString("  p { margin-bottom: 1rem; color: #374151; }\n")
	sb.WriteString("  ul, ol { margin-bottom: 1rem; padding-left: 1.5rem; }\n")
	sb.WriteString("  li { margin-bottom: 0.5rem; }\n")
	sb.WriteString("  code { background-color: #f3f4f6; padding: 0.2rem 0.4rem; border-radius: 4px; font-family: monospace; font-size: 0.9em; }\n")
	sb.WriteString("  blockquote { border-left: 4px solid #3b82f6; padding-left: 1rem; margin: 1rem 0; color: #4b5563; font-style: italic; background: #eff6ff; padding: 0.75rem 1rem; border-radius: 0 4px 4px 0; }\n")
	sb.WriteString("  .badge-verified { color: #059669; font-weight: 600; font-size: 0.85em; }\n")
	sb.WriteString("  .badge-contradiction { color: #dc2626; font-weight: 600; font-size: 0.85em; }\n")
	sb.WriteString("  @media print { body { background: white; padding: 0; } .container { box-shadow: none; border: none; padding: 0; } }\n")
	sb.WriteString("</style>\n</head>\n<body>\n")
	sb.WriteString("<div class=\"container\">\n")
	sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", htmlEscape(paper.Title)))
	sb.WriteString(fmt.Sprintf("<div class=\"metadata\"><strong>Project ID:</strong> %s &bull; <strong>Template:</strong> %s &bull; <strong>Status:</strong> %s &bull; <strong>Generated:</strong> %s</div>\n",
		htmlEscape(paper.ProjectID), htmlEscape(paper.TemplateID), htmlEscape(string(paper.Status)), paper.CreatedAt.Format(time.RFC822)))

	for _, sec := range paper.Sections {
		sb.WriteString(fmt.Sprintf("<h2>%s</h2>\n", htmlEscape(sec.Title)))
		if strings.TrimSpace(sec.Content) != "" {
			paragraphs := strings.Split(sec.Content, "\n\n")
			for _, p := range paragraphs {
				pTrimmed := strings.TrimSpace(p)
				if pTrimmed == "" {
					continue
				}
				if strings.HasPrefix(pTrimmed, "- ") || strings.HasPrefix(pTrimmed, "* ") {
					sb.WriteString("<ul>\n")
					lines := strings.Split(pTrimmed, "\n")
					for _, line := range lines {
						item := strings.TrimLeft(line, "-* ")
						sb.WriteString(fmt.Sprintf("  <li>%s</li>\n", formatHTMLInline(item)))
					}
					sb.WriteString("</ul>\n")
				} else {
					sb.WriteString(fmt.Sprintf("<p>%s</p>\n", formatHTMLInline(pTrimmed)))
				}
			}
		} else {
			sb.WriteString("<p><em>(Section content pending synthesis)</em></p>\n")
		}
	}

	sb.WriteString("</div>\n</body>\n</html>\n")
	return sb.String()
}

// CompileJSON marshals ResearchPaper to a JSON string.
func (w *PaperWriter) CompileJSON(paper *domain.ResearchPaper) (string, error) {
	bytes, err := json.MarshalIndent(paper, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal paper to JSON: %w", err)
	}
	return string(bytes), nil
}

// Helper functions for escaping special characters

func escapeLaTeX(text string) string {
	r := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"&", "\\&",
		"%", "\\%",
		"$", "\\$",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
	)
	return r.Replace(text)
}

func htmlEscape(text string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return r.Replace(text)
}

func formatHTMLInline(text string) string {
	escaped := htmlEscape(text)
	escaped = strings.ReplaceAll(escaped, "*(Verified: Direct Support", "<span class=\"badge-verified\">*(Verified: Direct Support")
	escaped = strings.ReplaceAll(escaped, "*(Caution: Contradicted", "<span class=\"badge-contradiction\">*(Caution: Contradicted")
	escaped = strings.ReplaceAll(escaped, ")*", ")*</span>")
	return escaped
}
