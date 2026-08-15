package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/mutation"
	"github.com/ilyaskhan/term-agent/internal/tools"
	"github.com/ilyaskhan/term-agent/internal/workspace"
)

func TestIntegration_MultiFileAtomicTransaction(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-integ-mut-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup pre-existing file
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Old Readme\n"), 0644); err != nil {
		t.Fatalf("failed to prepare README.md: %v", err)
	}

	engine := mutation.NewDefaultMutationEngine(tmpDir)
	tx, err := engine.BeginTransaction(context.Background(), "multi-file-session")
	if err != nil {
		t.Fatalf("BeginTransaction failed: %v", err)
	}

	// 1. Mutation 1: Modify existing README.md
	mut1 := &mutation.FileMutation{
		Path:         "README.md",
		Type:         mutation.MutationModify,
		OriginalHash: workspace.HashBytes([]byte("# Old Readme\n")),
		NewContent:   []byte("# New Project Readme\n"),
	}
	if err := engine.StageMutation(tx, mut1); err != nil {
		t.Fatalf("StageMutation 1 failed: %v", err)
	}

	// 2. Mutation 2: Create main.go
	mut2 := &mutation.FileMutation{
		Path:       "main.go",
		Type:       mutation.MutationCreate,
		NewContent: []byte("package main\n\nfunc main() {}\n"),
	}
	if err := engine.StageMutation(tx, mut2); err != nil {
		t.Fatalf("StageMutation 2 failed: %v", err)
	}

	if len(tx.Mutations) != 2 {
		t.Fatalf("expected 2 staged mutations, got %d", len(tx.Mutations))
	}

	// 3. Approve and commit
	if err := engine.ApproveTransaction(tx); err != nil {
		t.Fatalf("ApproveTransaction failed: %v", err)
	}

	if err := engine.CommitTransaction(context.Background(), tx); err != nil {
		t.Fatalf("CommitTransaction failed: %v", err)
	}

	// Verify atomic changes
	readmeContent, _ := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	if string(readmeContent) != "# New Project Readme\n" {
		t.Errorf("README.md mutation failed: got %q", string(readmeContent))
	}

	mainContent, _ := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	if string(mainContent) != "package main\n\nfunc main() {}\n" {
		t.Errorf("main.go creation failed: got %q", string(mainContent))
	}
}

func TestIntegration_ConcurrentToolEditsAndOCC(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-occ-concurrent-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "shared.txt")
	if err := os.WriteFile(filePath, []byte("counter: 0\n"), 0644); err != nil {
		t.Fatalf("failed to write shared.txt: %v", err)
	}

	engine := mutation.NewDefaultMutationEngine(tmpDir)
	editTool := tools.NewEditFileTool(tmpDir, engine)

	var wg sync.WaitGroup
	successCount := 0
	conflictCount := 0
	var mu sync.Mutex

	// Concurrent edits attempting to replace counter string
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			args, _ := json.Marshal(map[string]string{
				"path":       "shared.txt",
				"old_string": "counter: 0",
				"new_string": "counter: 1",
			})
			res, err := editTool.Execute(context.Background(), args)
			mu.Lock()
			defer mu.Unlock()
			if err == nil && !res.IsError {
				successCount++
			} else {
				conflictCount++
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("expected exactly 1 edit to succeed under OCC, got %d", successCount)
	}
	if conflictCount != 4 {
		t.Errorf("expected 4 edits to conflict under OCC, got %d", conflictCount)
	}
}
