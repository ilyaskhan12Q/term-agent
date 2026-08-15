package unit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ragents "github.com/ilyaskhan/term-agent/internal/workflows/research/agents"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/provenance"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

func TestTemplateEngine_EmbeddedTemplates(t *testing.T) {
	eng, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}

	all := eng.ListTemplates()
	if len(all) < 4 {
		t.Errorf("expected at least 4 registered templates, got %d", len(all))
	}

	expectedIDs := []string{
		"academic_research",
		"technical_survey",
		"executive_briefing",
		"system_architecture",
	}

	for _, id := range expectedIDs {
		tmpl, err := eng.GetTemplate(id)
		if err != nil {
			t.Errorf("expected template %s to exist, error: %v", id, err)
		}
		if len(tmpl.Sections) == 0 {
			t.Errorf("template %s has no sections defined", id)
		}
	}
}

func TestPaperWriter_CompileFormats(t *testing.T) {
	eng, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}
	agent := ragents.NewSynthesisAgent("synth-writer-1", eng)

	project, err := domain.NewResearchProject("proj-writer-1", "sess-w", "High-Throughput Vector Search Architecture", "Evaluation of HNSW and DiskANN index strategies", "system_architecture")
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	finding := &domain.ResearchFinding{
		TaskID:  "task-w1",
		AgentID: "worker-w1",
		Findings: []string{
			"HNSW graph indices achieve sub-millisecond p99 recall @ 95%.",
			"DiskANN compresses vector embeddings with 4x memory reduction.",
		},
	}

	tracker := provenance.NewProvenanceTracker()
	src, _ := domain.NewSource("src-w1", "proj-writer-1", "DiskANN Benchmarks", "https://arxiv.org/abs/2401.99999", domain.SourceTypeAcademicPaper, 0.98)
	_, _ = tracker.RegisterSource(*src)

	paper, err := agent.SynthesizePaper(context.Background(), project, []*domain.ResearchFinding{finding}, tracker)
	if err != nil {
		t.Fatalf("SynthesizePaper failed: %v", err)
	}

	writer := templates.NewPaperWriter()

	// 1. Markdown
	mdOutput, err := writer.Compile(paper, templates.ExportFormatMarkdown)
	if err != nil {
		t.Fatalf("Markdown compilation failed: %v", err)
	}
	if !strings.Contains(mdOutput, "# High-Throughput Vector Search Architecture") {
		t.Errorf("expected Markdown output to contain paper title")
	}

	// 2. LaTeX
	texOutput, err := writer.Compile(paper, templates.ExportFormatLaTeX)
	if err != nil {
		t.Fatalf("LaTeX compilation failed: %v", err)
	}
	if !strings.Contains(texOutput, "\\documentclass[11pt, a4paper]{article}") {
		t.Errorf("expected LaTeX output to contain documentclass statement")
	}
	if !strings.Contains(texOutput, "\\title{\\textbf{High-Throughput Vector Search Architecture}}") {
		t.Errorf("expected LaTeX output to contain formatted title")
	}
	if !strings.Contains(texOutput, "\\end{document}") {
		t.Errorf("expected LaTeX output to end with \\end{document}")
	}

	// 3. HTML
	htmlOutput, err := writer.Compile(paper, templates.ExportFormatHTML)
	if err != nil {
		t.Fatalf("HTML compilation failed: %v", err)
	}
	if !strings.Contains(htmlOutput, "<!DOCTYPE html>") {
		t.Errorf("expected HTML output to contain DOCTYPE header")
	}
	if !strings.Contains(htmlOutput, "<h1>High-Throughput Vector Search Architecture</h1>") {
		t.Errorf("expected HTML output to contain h1 title tag")
	}

	// 4. JSON
	jsonOutput, err := writer.Compile(paper, templates.ExportFormatJSON)
	if err != nil {
		t.Fatalf("JSON compilation failed: %v", err)
	}
	var unmarshaled domain.ResearchPaper
	if err := json.Unmarshal([]byte(jsonOutput), &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal compiled JSON paper: %v", err)
	}
	if unmarshaled.Title != paper.Title {
		t.Errorf("expected JSON unmarshaled title %q, got %q", paper.Title, unmarshaled.Title)
	}
}

func TestPaperWriter_ExportToFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "paper_writer_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	paper := &domain.ResearchPaper{
		ID:             "paper-export-1",
		ProjectID:      "proj-export-1",
		TemplateID:     "executive_briefing",
		Title:          "Quantum Cryptography Roadmap",
		MarkdownOutput: "# Quantum Cryptography Roadmap\n\nExecutive briefing content.",
		Status:         domain.PaperStatusPassed,
	}

	writer := templates.NewPaperWriter()

	formats := []struct {
		format   templates.ExportFormat
		filename string
	}{
		{templates.ExportFormatMarkdown, "paper.md"},
		{templates.ExportFormatLaTeX, "paper.tex"},
		{templates.ExportFormatHTML, "paper.html"},
		{templates.ExportFormatJSON, "paper.json"},
	}

	for _, f := range formats {
		targetPath := filepath.Join(tempDir, f.filename)
		err := writer.ExportToFile(paper, f.format, targetPath)
		if err != nil {
			t.Errorf("ExportToFile failed for format %s: %v", f.format, err)
		}

		info, err := os.Stat(targetPath)
		if err != nil || info.Size() == 0 {
			t.Errorf("expected exported file %s to exist and be non-empty", targetPath)
		}
	}
}

func TestParseExportFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected templates.ExportFormat
		isErr    bool
	}{
		{"markdown", templates.ExportFormatMarkdown, false},
		{"MD", templates.ExportFormatMarkdown, false},
		{"latex", templates.ExportFormatLaTeX, false},
		{"TEX", templates.ExportFormatLaTeX, false},
		{"html", templates.ExportFormatHTML, false},
		{"json", templates.ExportFormatJSON, false},
		{"invalid", "", true},
	}

	for _, tc := range tests {
		fmt, err := templates.ParseExportFormat(tc.input)
		if tc.isErr && err == nil {
			t.Errorf("expected error for input %q, got nil", tc.input)
		}
		if !tc.isErr && err != nil {
			t.Errorf("unexpected error for input %q: %v", tc.input, err)
		}
		if fmt != tc.expected {
			t.Errorf("expected format %s for input %q, got %s", tc.expected, tc.input, fmt)
		}
	}
}
