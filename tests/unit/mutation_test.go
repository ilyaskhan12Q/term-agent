package unit_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/mutation"
	"github.com/ilyaskhan/term-agent/internal/tools"
	"github.com/ilyaskhan/term-agent/internal/workspace"
)

func TestMutationEngine_LifecycleAndCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-mut-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	engine := mutation.NewDefaultMutationEngine(tmpDir)

	// 1. Begin Transaction
	tx, err := engine.BeginTransaction(context.Background(), "session-123")
	if err != nil {
		t.Fatalf("BeginTransaction failed: %v", err)
	}

	if tx.Status != mutation.TxStatusProposed {
		t.Errorf("expected initial status PROPOSED, got %s", tx.Status)
	}

	// 2. Stage Mutation (Create new file)
	mut := &mutation.FileMutation{
		Path:       "config.yaml",
		Type:       mutation.MutationCreate,
		NewContent: []byte("server:\n  port: 8080\n"),
	}

	if err := engine.StageMutation(tx, mut); err != nil {
		t.Fatalf("StageMutation failed: %v", err)
	}

	if tx.Status != mutation.TxStatusValidated {
		t.Errorf("expected status VALIDATED, got %s", tx.Status)
	}

	// 3. Request & Approve Transaction
	if err := engine.RequestApproval(tx); err != nil {
		t.Fatalf("RequestApproval failed: %v", err)
	}
	if tx.Status != mutation.TxStatusWaitingApproval {
		t.Errorf("expected status WAITING_APPROVAL, got %s", tx.Status)
	}

	if err := engine.ApproveTransaction(tx); err != nil {
		t.Fatalf("ApproveTransaction failed: %v", err)
	}
	if tx.Status != mutation.TxStatusApproved {
		t.Errorf("expected status APPROVED, got %s", tx.Status)
	}

	// 4. Commit Transaction
	if err := engine.CommitTransaction(context.Background(), tx); err != nil {
		t.Fatalf("CommitTransaction failed: %v", err)
	}

	if tx.Status != mutation.TxStatusCommitted {
		t.Errorf("expected status COMMITTED, got %s", tx.Status)
	}

	// Verify file created on disk
	content, err := os.ReadFile(filepath.Join(tmpDir, "config.yaml"))
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(content) != "server:\n  port: 8080\n" {
		t.Errorf("unexpected file content: %q", string(content))
	}
}

func TestMutationEngine_OptimisticConcurrencyConflict(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-occ-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "version.txt")
	if err := os.WriteFile(filePath, []byte("v1.0.0"), 0644); err != nil {
		t.Fatalf("failed to write version.txt: %v", err)
	}

	origHash := workspace.HashBytes([]byte("v1.0.0"))

	// Simulate external modification before transaction staging
	if err := os.WriteFile(filePath, []byte("v1.0.1-external"), 0644); err != nil {
		t.Fatalf("failed to overwrite version.txt: %v", err)
	}

	engine := mutation.NewDefaultMutationEngine(tmpDir)
	tx, _ := engine.BeginTransaction(context.Background(), "session-occ")

	mut := &mutation.FileMutation{
		Path:         "version.txt",
		Type:         mutation.MutationModify,
		OriginalHash: origHash,
		NewContent:   []byte("v2.0.0"),
	}

	err = engine.StageMutation(tx, mut)
	if !errors.Is(err, mutation.ErrConcurrencyConflict) {
		t.Fatalf("expected ErrConcurrencyConflict, got: %v", err)
	}

	if tx.Status != mutation.TxStatusConflict {
		t.Errorf("expected transaction status CONFLICT, got %s", tx.Status)
	}
}

func TestMutationEngine_Rollback(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-rollback-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "original.txt")
	if err := os.WriteFile(filePath, []byte("original content"), 0644); err != nil {
		t.Fatalf("failed to write original.txt: %v", err)
	}

	engine := mutation.NewDefaultMutationEngine(tmpDir)
	tx, _ := engine.BeginTransaction(context.Background(), "session-rb")

	mut := &mutation.FileMutation{
		Path:         "original.txt",
		Type:         mutation.MutationModify,
		OriginalHash: workspace.HashBytes([]byte("original content")),
		NewContent:   []byte("corrupted content"),
	}

	_ = engine.StageMutation(tx, mut)
	_ = engine.ApproveTransaction(tx)

	// Simulate write & rollback
	if err := engine.RollbackTransaction(context.Background(), tx); err != nil {
		t.Fatalf("RollbackTransaction failed: %v", err)
	}

	if tx.Status != mutation.TxStatusRolledBack {
		t.Errorf("expected status ROLLED_BACK, got %s", tx.Status)
	}

	restored, _ := os.ReadFile(filePath)
	if string(restored) != "original content" {
		t.Errorf("rollback failed to restore original content, got %q", string(restored))
	}
}

func TestWriteAndEditFileTools(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-tools-mut-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	engine := mutation.NewDefaultMutationEngine(tmpDir)

	// 1. Test WriteFileTool
	writeTool := tools.NewWriteFileTool(tmpDir, engine)
	writeRes, err := writeTool.Execute(context.Background(), json.RawMessage(`{
		"path": "app.go",
		"content": "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	}`))

	if err != nil || writeRes.IsError {
		t.Fatalf("WriteFileTool failed: %v, errStr: %s", err, writeRes.Error)
	}

	// 2. Test EditFileTool
	editTool := tools.NewEditFileTool(tmpDir, engine)
	editRes, err := editTool.Execute(context.Background(), json.RawMessage(`{
		"path": "app.go",
		"old_string": "println(\"hello\")",
		"new_string": "println(\"world\")"
	}`))

	if err != nil || editRes.IsError {
		t.Fatalf("EditFileTool failed: %v, errStr: %s", err, editRes.Error)
	}

	finalContent, err := os.ReadFile(filepath.Join(tmpDir, "app.go"))
	if err != nil {
		t.Fatalf("failed to read app.go: %v", err)
	}

	expected := "package main\n\nfunc main() {\n\tprintln(\"world\")\n}\n"
	if string(finalContent) != expected {
		t.Errorf("expected content %q, got %q", expected, string(finalContent))
	}
}
