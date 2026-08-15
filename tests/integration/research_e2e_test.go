package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ilyaskhan/term-agent/internal/events"
	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/persistence/repository"
	"github.com/ilyaskhan/term-agent/internal/workflows/research"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/templates"
)

func TestResearchMode_EndToEndFullPipeline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-e2e-research-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "e2e_research.db")
	db, err := persistence.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	bus := events.NewInMemoryEventBus()
	wf := research.NewResearchWorkflow()
	wf.SetDatabase(db)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	topic := "High-Throughput Parallel Processing in Distributed Ledger Systems"
	if err := wf.Initialize(ctx, topic); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	res, err := wf.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if res == nil || res.Output == "" {
		t.Fatal("expected non-empty output result from end-to-end workflow execution")
	}

	dataMap, ok := res.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{} in result data, got %T", res.Data)
	}

	projectID, _ := dataMap["project_id"].(string)
	paperID, _ := dataMap["paper_id"].(string)
	if projectID == "" || paperID == "" {
		t.Fatalf("expected valid project_id and paper_id in output data")
	}

	// 1. Verify Database State Persistence
	repo := repository.NewSQLiteResearchRepository(db)
	project, err := repo.GetProject(ctx, projectID)
	if err != nil || project == nil {
		t.Fatalf("failed to retrieve persisted project from DB: %v", err)
	}
	if project.Status != domain.ProjectStatusCompleted {
		t.Errorf("expected project status COMPLETED, got %s", project.Status)
	}

	paper, err := repo.GetPaperByProject(ctx, projectID)
	if err != nil || paper == nil {
		t.Fatalf("failed to retrieve persisted paper from DB: %v", err)
	}
	if len(paper.Sections) == 0 {
		t.Errorf("expected synthesized sections in persisted paper")
	}

	// 2. Verify Multi-Format Export End-to-End
	writer := templates.NewPaperWriter()
	formats := []struct {
		fmt        templates.ExportFormat
		ext        string
		containStr string
	}{
		{templates.ExportFormatMarkdown, ".md", "# "},
		{templates.ExportFormatLaTeX, ".tex", "\\documentclass"},
		{templates.ExportFormatHTML, ".html", "<!DOCTYPE html>"},
		{templates.ExportFormatJSON, ".json", "\"title\":"},
	}

	for _, item := range formats {
		outPath := filepath.Join(tmpDir, "paper_export"+item.ext)
		if err := writer.ExportToFile(paper, item.fmt, outPath); err != nil {
			t.Errorf("failed to export paper in format %s to %s: %v", item.fmt, outPath, err)
			continue
		}

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("failed to read exported file %s: %v", outPath, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("exported file %s is empty", outPath)
		}
	}
}

func BenchmarkResearchWorkflowExecution(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "termagent-bench-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	db, err := persistence.Open(dbPath)
	if err != nil {
		b.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	bus := events.NewInMemoryEventBus()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wf := research.NewResearchWorkflow()
		wf.SetDatabase(db)
		_ = wf.Initialize(ctx, "Benchmarking Autonomous Research Synthesis Execution")
		_, _ = wf.Execute(ctx, bus)
	}
}
