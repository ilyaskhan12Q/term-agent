package unit_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyaskhan/term-agent/internal/tools"
	"github.com/ilyaskhan/term-agent/internal/workspace"
)

func TestWorkspaceScannerAndIgnoreRules(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-ws-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	// tmpDir/
	//   main.go
	//   README.md
	//   .gitignore (ignores *.tmp)
	//   build.tmp (should be ignored)
	//   node_modules/ (should be ignored by default)
	//     dep.js
	//   binary.bin (contains null byte)

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Project\nWelcome\n"), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.tmp\n"), 0644); err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "build.tmp"), []byte("temp content"), 0644); err != nil {
		t.Fatalf("failed to write build.tmp: %v", err)
	}

	nodeDir := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatalf("failed to mkdir node_modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeDir, "dep.js"), []byte("console.log('hi')"), 0644); err != nil {
		t.Fatalf("failed to write dep.js: %v", err)
	}

	// Create binary file with null byte
	binContent := []byte{0x7f, 'E', 'L', 'F', 0x00, 0x01, 0x02}
	if err := os.WriteFile(filepath.Join(tmpDir, "binary.bin"), binContent, 0644); err != nil {
		t.Fatalf("failed to write binary.bin: %v", err)
	}

	scanner := workspace.NewDefaultScanner()
	files, err := scanner.DiscoverFiles(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("DiscoverFiles failed: %v", err)
	}

	foundMap := make(map[string]workspace.FileInfo)
	for _, f := range files {
		foundMap[f.RelPath] = f
	}

	// Verify main.go is discovered and non-binary
	if mainInfo, ok := foundMap["main.go"]; !ok {
		t.Errorf("expected main.go to be discovered")
	} else if mainInfo.IsBinary {
		t.Errorf("expected main.go to be classified as text")
	}

	// Verify binary.bin is discovered and binary
	if binInfo, ok := foundMap["binary.bin"]; !ok {
		t.Errorf("expected binary.bin to be discovered")
	} else if !binInfo.IsBinary {
		t.Errorf("expected binary.bin to be classified as binary")
	}

	// Verify build.tmp is ignored by .gitignore
	if _, ok := foundMap["build.tmp"]; ok {
		t.Errorf("build.tmp should have been ignored by .gitignore")
	}

	// Verify node_modules/dep.js is ignored by DefaultIgnoredDirs
	if _, ok := foundMap["node_modules/dep.js"]; ok {
		t.Errorf("node_modules/dep.js should have been ignored by default directory rules")
	}
}

func TestReadWorkspaceFilePagination(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-read-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "lines.txt"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write lines.txt: %v", err)
	}

	// Read lines 2 to 4
	opts := workspace.ReadOptions{
		StartLine: 2,
		EndLine:   4,
	}

	res, err := workspace.ReadWorkspaceFile(context.Background(), tmpDir, "lines.txt", opts)
	if err != nil {
		t.Fatalf("ReadWorkspaceFile failed: %v", err)
	}

	expected := "line 2\nline 3\nline 4"
	if res.Content != expected {
		t.Errorf("expected content %q, got %q", expected, res.Content)
	}

	if res.StartLine != 2 || res.EndLine != 4 {
		t.Errorf("expected range 2-4, got %d-%d", res.StartLine, res.EndLine)
	}
}

func TestWorkspaceStateHash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-hash-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ws, err := workspace.NewDefaultWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("failed to construct DefaultWorkspace: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write a.txt: %v", err)
	}

	hash1, err := ws.ComputeStateHash(context.Background())
	if err != nil {
		t.Fatalf("ComputeStateHash failed: %v", err)
	}

	if hash1 == "" {
		t.Errorf("expected non-empty state hash")
	}

	// Mutate workspace
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("world"), 0644); err != nil {
		t.Fatalf("failed to write b.txt: %v", err)
	}

	hash2, err := ws.ComputeStateHash(context.Background())
	if err != nil {
		t.Fatalf("ComputeStateHash 2 failed: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("state hash should change after workspace file addition")
	}
}

func TestWorkspaceToolsExecution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "termagent-tools-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("{\"port\": 8080}\n"), 0644); err != nil {
		t.Fatalf("failed to write config.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "server.go"), []byte("package main\n\nimport \"fmt\"\n\nfunc startServer() {\n\tfmt.Println(\"port: 8080\")\n}\n"), 0644); err != nil {
		t.Fatalf("failed to write server.go: %v", err)
	}

	// 1. Test ListWorkspaceTool
	listTool := tools.NewListWorkspaceTool(tmpDir)
	listRes, err := listTool.Execute(context.Background(), json.RawMessage(`{"extension": ".go"}`))
	if err != nil || listRes.IsError {
		t.Fatalf("ListWorkspaceTool execution failed: %v, errStr: %s", err, listRes.Error)
	}

	var files []workspace.FileInfo
	if err := json.Unmarshal([]byte(listRes.Output), &files); err != nil {
		t.Fatalf("failed to parse ListWorkspaceTool JSON: %v", err)
	}

	if len(files) != 1 || files[0].RelPath != "server.go" {
		t.Errorf("expected 1 file (server.go), got %v", files)
	}

	// 2. Test ReadFileTool
	readTool := tools.NewReadFileTool(tmpDir)
	readRes, err := readTool.Execute(context.Background(), json.RawMessage(`{"path": "config.json"}`))
	if err != nil || readRes.IsError {
		t.Fatalf("ReadFileTool execution failed: %v, errStr: %s", err, readRes.Error)
	}

	var readPayload workspace.ReadResult
	if err := json.Unmarshal([]byte(readRes.Output), &readPayload); err != nil {
		t.Fatalf("failed to parse ReadFileTool JSON: %v", err)
	}

	if readPayload.Content != "{\"port\": 8080}\n" {
		t.Errorf("unexpected content: %q", readPayload.Content)
	}

	// 3. Test SearchWorkspaceTool
	searchTool := tools.NewSearchWorkspaceTool(tmpDir)
	searchRes, err := searchTool.Execute(context.Background(), json.RawMessage(`{"query": "8080"}`))
	if err != nil || searchRes.IsError {
		t.Fatalf("SearchWorkspaceTool execution failed: %v, errStr: %s", err, searchRes.Error)
	}

	var matches []tools.SearchMatch
	if err := json.Unmarshal([]byte(searchRes.Output), &matches); err != nil {
		t.Fatalf("failed to parse SearchWorkspaceTool JSON: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("expected 2 matches for '8080', got %d", len(matches))
	}
}
