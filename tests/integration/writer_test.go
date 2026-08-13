package integration

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ilyaskhan/term-agent/internal/persistence"
)

func TestChannelAsyncWriter_EnqueueAndDrain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-writer-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := persistence.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	writer := persistence.NewChannelAsyncWriter(db, 100, nil)

	count := 50
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		sessionID := "session-" + string(rune('A'+i%26)) + string(rune('0'+i%10))
		err := writer.Enqueue(func(ctx context.Context, db *persistence.DB) error {
			defer wg.Done()
			_, err := db.ExecContext(ctx, "INSERT OR IGNORE INTO sessions (id, title, workspace_path, status) VALUES (?, ?, ?, ?)",
				sessionID, "Test Session", "/tmp/workspace", "ACTIVE")
			return err
		})
		if err != nil {
			t.Fatalf("enqueue failed: %v", err)
		}
	}

	wg.Wait()

	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := writer.Close(closeCtx); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	var rowCount int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM sessions").Scan(&rowCount)
	if err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}

	if rowCount == 0 {
		t.Errorf("expected inserted sessions, got 0")
	}
}
