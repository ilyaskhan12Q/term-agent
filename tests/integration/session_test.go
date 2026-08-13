package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/app"
	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/persistence/repository"
)

func TestSessionPersistenceAndRecovery(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test_session.db")
	db, err := persistence.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	sessionRepo := repository.NewSQLiteSessionRepository(db)
	manager := app.NewSessionManager(sessionRepo, nil)

	ctx := context.Background()

	// 1. Prepare initial session
	s1, err := manager.PrepareSession(ctx, "", "/tmp/workspace-1")
	if err != nil {
		t.Fatalf("failed to prepare new session: %v", err)
	}

	if s1.ID == "" || s1.Status != "ACTIVE" {
		t.Errorf("unexpected session status or ID: %+v", s1)
	}

	// 2. Prepare second session without closing first (simulating crash recovery)
	manager2 := app.NewSessionManager(sessionRepo, nil)
	s2, err := manager2.PrepareSession(ctx, "", "/tmp/workspace-2")
	if err != nil {
		t.Fatalf("failed to prepare second session: %v", err)
	}

	// 3. Verify s1 was updated to INTERRUPTED during crash recovery
	oldS1, err := sessionRepo.GetSession(ctx, s1.ID)
	if err != nil {
		t.Fatalf("failed to get old session: %v", err)
	}
	if oldS1.Status != "INTERRUPTED" {
		t.Errorf("expected old session status to be INTERRUPTED, got %s", oldS1.Status)
	}

	// 4. Test resuming specific session ID
	manager3 := app.NewSessionManager(sessionRepo, nil)
	sResumed, err := manager3.PrepareSession(ctx, s2.ID, "/tmp/workspace-2")
	if err != nil {
		t.Fatalf("failed to resume session %s: %v", s2.ID, err)
	}
	if sResumed.ID != s2.ID {
		t.Errorf("expected resumed session ID %s, got %s", s2.ID, sResumed.ID)
	}

	// 5. Test clean session close
	if err := manager3.CloseCurrentSession(ctx); err != nil {
		t.Fatalf("failed to close session: %v", err)
	}

	closedS, err := sessionRepo.GetSession(ctx, s2.ID)
	if err != nil {
		t.Fatalf("failed to fetch closed session: %v", err)
	}
	if closedS.Status != "CLOSED" {
		t.Errorf("expected status CLOSED, got %s", closedS.Status)
	}
}
