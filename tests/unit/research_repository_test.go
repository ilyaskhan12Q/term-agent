package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/persistence/repository"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

func setupTestDB(t *testing.T) (*persistence.DB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "term-agent-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := persistence.Open(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open database: %v", err)
	}

	// Apply migration 000001 & 000002
	schema01, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_initial_schema.sql"))
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to read migration 000001: %v", err)
	}
	schema02, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000002_research_workflow.sql"))
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to read migration 000002: %v", err)
	}

	if _, err := db.Exec(string(schema01)); err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to apply migration 000001: %v", err)
	}
	if _, err := db.Exec(string(schema02)); err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to apply migration 000002: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
	return db, cleanup
}

func TestResearchRepositoryOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create prerequisite session
	sessRepo := repository.NewSQLiteSessionRepository(db)
	sessRecord := &repository.SessionRecord{
		ID:            "sess-100",
		Title:         "Research Session",
		WorkspacePath: "/tmp/workspace",
		Status:        "ACTIVE",
	}
	if err := sessRepo.CreateSession(ctx, sessRecord); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	repo := repository.NewSQLiteResearchRepository(db)

	// 2. Test Project CRUD
	proj, err := domain.NewResearchProject("proj-100", "sess-100", "LLM Hallucinations", "Analyze mitigation strategies", "academic_research")
	if err != nil {
		t.Fatalf("failed to construct project: %v", err)
	}

	if err := repo.SaveProject(ctx, proj); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	fetchedProj, err := repo.GetProject(ctx, "proj-100")
	if err != nil {
		t.Fatalf("failed to fetch project: %v", err)
	}
	if fetchedProj.Title != proj.Title {
		t.Errorf("expected title '%s', got '%s'", proj.Title, fetchedProj.Title)
	}

	// 3. Test Question Persistence & Query
	q1, err := domain.NewResearchQuestion("q-1", "proj-100", "What causes hallucinations in transformers?", "Transformer architecture", 1)
	if err != nil {
		t.Fatalf("failed to construct question: %v", err)
	}
	if err := repo.SaveQuestion(ctx, q1); err != nil {
		t.Fatalf("failed to save question: %v", err)
	}

	questions, err := repo.ListQuestions(ctx, "proj-100")
	if err != nil {
		t.Fatalf("failed to list questions: %v", err)
	}
	if len(questions) != 1 || questions[0].Question != q1.Question {
		t.Errorf("expected 1 question matching q1, got %d", len(questions))
	}

	// 4. Test Structured Finding Persistence
	finding := &domain.ResearchFinding{
		ID:         "find-100",
		ProjectID:  "proj-100",
		QuestionID: "q-1",
		TaskID:     "task-100",
		AgentID:    "agent-lit",
		AgentType:  "LITERATURE",
		Findings:   []string{"Attention mechanisms suffer from positional degradation"},
		Confidence: 0.92,
		CreatedAt:  time.Now(),
	}
	if err := repo.SaveFinding(ctx, finding); err != nil {
		t.Fatalf("failed to save finding: %v", err)
	}

	findings, err := repo.ListFindings(ctx, "proj-100")
	if err != nil {
		t.Fatalf("failed to list findings: %v", err)
	}
	if len(findings) != 1 || findings[0].Confidence != 0.92 {
		t.Errorf("expected 1 finding with confidence 0.92, got %d", len(findings))
	}

	// 5. Test Paper Persistence
	paper, err := domain.NewResearchPaper("paper-100", "proj-100", "academic_research", "LLM Hallucinations: A Survey")
	if err != nil {
		t.Fatalf("failed to create paper: %v", err)
	}
	paper.MarkdownOutput = "# Abstract\nThis paper examines..."
	if err := repo.SavePaper(ctx, paper); err != nil {
		t.Fatalf("failed to save paper: %v", err)
	}

	fetchedPaper, err := repo.GetPaperByProject(ctx, "proj-100")
	if err != nil {
		t.Fatalf("failed to fetch paper: %v", err)
	}
	if fetchedPaper.Title != paper.Title || fetchedPaper.MarkdownOutput != paper.MarkdownOutput {
		t.Errorf("paper title or markdown output mismatch")
	}
}
