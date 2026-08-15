package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ilyaskhan/term-agent/internal/tui/styles"
)

// ResearchFindingSummary represents a summarized research finding for display.
type ResearchFindingSummary struct {
	ID             string
	Topic          string
	ClaimsCount    int
	EvidenceCount  int
	VerifiedClaims int
}

// ResearchView displays dedicated Research Agent Mode project status, provenance, and paper preview.
type ResearchView struct {
	styles       *styles.Styles
	width        int
	height       int
	ProjectTitle string
	TemplateName string
	SourcesCount int
	Findings     []ResearchFindingSummary
	PaperPreview string
	// CommandLog holds lines of output from slash commands, shown in the view.
	CommandLog []string
}

// NewResearchView creates ResearchView instance with initial placeholder state.
func NewResearchView(s *styles.Styles) *ResearchView {
	return &ResearchView{
		styles:       s,
		ProjectTitle: "Autonomous Agent Orchestration and Scalable DAG Execution",
		TemplateName: "IEEE Standard Research Skeleton",
		SourcesCount: 8,
		Findings: []ResearchFindingSummary{
			{ID: "find-1", Topic: "DAG Scheduler Latency Benchmark", ClaimsCount: 4, EvidenceCount: 6, VerifiedClaims: 4},
			{ID: "find-2", Topic: "SQLite WAL Concurrency and Provenance Storage", ClaimsCount: 3, EvidenceCount: 5, VerifiedClaims: 3},
			{ID: "find-3", Topic: "Token Budget Management in Multi-Agent Loops", ClaimsCount: 5, EvidenceCount: 7, VerifiedClaims: 4},
		},
		PaperPreview: `# Abstract
This paper presents term-agent, an open-source terminal-centric agent runtime designed for autonomous DAG task decomposition, evidence provenance, and structured paper synthesis.

# 1. Introduction
Autonomous software and research agents require robust execution guarantees, topological scheduling, and verifiable provenance chains...

# 2. Related Work
Prior agentic systems suffer from context drift and unverified hallucination. We address this via Claim-Evidence-Source tracking...`,
	}
}

// SetSize updates dimensions.
func (v *ResearchView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// AddLog appends a line of command output to the view log.
func (v *ResearchView) AddLog(line string) {
	v.CommandLog = append(v.CommandLog, line)
}

// Clear removes all command log entries.
func (v *ResearchView) Clear() {
	v.CommandLog = nil
}

// UpdateProject updates research mode view data.
func (v *ResearchView) UpdateProject(title, template string, sources int, findings []ResearchFindingSummary, paper string) {
	v.ProjectTitle = title
	v.TemplateName = template
	v.SourcesCount = sources
	v.Findings = findings
	v.PaperPreview = paper
}

// Render draws the research project dashboard and paper preview.
// If CommandLog is non-empty, it renders the log output instead of the static dashboard.
func (v *ResearchView) Render() string {
	var b strings.Builder

	panelStyle := lipgloss.NewStyle().
		Width(v.width - 4).
		Height(v.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(v.styles.Tokens.BorderColor)).
		Padding(1)

	if len(v.CommandLog) > 0 {
		b.WriteString(v.styles.PanelTitle.Render("Research Mode — Command Output"))
		b.WriteString("\n\n")
		for _, line := range v.CommandLog {
			b.WriteString(line)
			b.WriteString("\n\n")
		}
		return panelStyle.Render(b.String())
	}

	b.WriteString(v.styles.PanelTitle.Render("Research Agent Mode — Workspace & Provenance Dashboard"))
	b.WriteString("\n\n")

	// Project Header
	header := fmt.Sprintf("Topic: %s  |  Template: %s  |  Sources: %d\n\n",
		v.styles.BadgeActive.Render(v.ProjectTitle),
		v.styles.Badge.Render(v.TemplateName),
		v.SourcesCount,
	)
	b.WriteString(header)

	// Findings & Provenance Table
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(v.styles.Tokens.PrimaryColor)).Render("Gathered Findings & Verified Claims"))
	b.WriteString("\n")

	tableHeader := fmt.Sprintf("%-10s %-45s %-12s %-12s %-15s\n", "ID", "Finding Topic", "Claims", "Evidence", "Provenance State")
	b.WriteString(v.styles.DiffContext.Render(tableHeader))
	b.WriteString(v.styles.DiffContext.Render(strings.Repeat("-", 95)) + "\n")

	for _, f := range v.Findings {
		provStatus := v.styles.TaskComplete.Render(fmt.Sprintf("Verified (%d/%d)", f.VerifiedClaims, f.ClaimsCount))
		if f.VerifiedClaims < f.ClaimsCount {
			provStatus = v.styles.TaskRunning.Render(fmt.Sprintf("Partial (%d/%d)", f.VerifiedClaims, f.ClaimsCount))
		}

		b.WriteString(fmt.Sprintf("%-10s %-45s %-12d %-12d %-15s\n",
			f.ID,
			f.Topic,
			f.ClaimsCount,
			f.EvidenceCount,
			provStatus,
		))
	}

	b.WriteString("\n\n")
	// Paper Synthesis Preview Box
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(v.styles.Tokens.SecondaryColor)).Render("Generated Research Paper Preview"))
	b.WriteString("\n")

	paperBox := lipgloss.NewStyle().
		Background(lipgloss.Color(v.styles.Tokens.SurfaceColor)).
		Foreground(lipgloss.Color(v.styles.Tokens.TextColor)).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(v.styles.Tokens.BorderColor)).
		Padding(1).
		Render(v.PaperPreview)

	b.WriteString(paperBox)

	return panelStyle.Render(b.String())
}
