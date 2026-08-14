package workspace

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ilyaskhan/term-agent/internal/security"
)

// FileInfo contains comprehensive metadata for an indexed file inside the workspace.
type FileInfo struct {
	RelPath   string      `json:"rel_path"`
	AbsPath   string      `json:"abs_path"`
	SizeBytes int64       `json:"size_bytes"`
	Mode      os.FileMode `json:"mode"`
	IsBinary  bool        `json:"is_binary"`
	ModTime   time.Time   `json:"mod_time"`
	LineCount int         `json:"line_count"`
}

// FileDiscovery defines the contract for scanning workspace files respecting gitignore rules.
type FileDiscovery interface {
	DiscoverFiles(ctx context.Context, root string) ([]FileInfo, error)
	IsIgnored(relPath string) bool
}

// DefaultIgnoredDirs returns common directories ignored during workspace discovery.
var DefaultIgnoredDirs = map[string]bool{
	".git":          true,
	".svn":          true,
	".hg":           true,
	"node_modules":  true,
	"vendor":        true,
	"bin":           true,
	"obj":           true,
	"dist":          true,
	"build":         true,
	".termagent":    true,
	"__pycache__":   true,
	".pytest_cache": true,
	".idea":         true,
	".vscode":       true,
}

// DefaultIgnoredFiles returns common filenames ignored during workspace discovery.
var DefaultIgnoredFiles = map[string]bool{
	".DS_Store":         true,
	"thumbs.db":         true,
	"package-lock.json": false, // explicitly indexed
}

// DefaultScanner implements FileDiscovery with ignore rules, binary detection, and path validation.
type DefaultScanner struct {
	customIgnores map[string]bool
	maxFileSize   int64 // Maximum single file size to index line count for (default: 5MB)
}

// NewDefaultScanner initializes a new workspace file discovery scanner.
func NewDefaultScanner() *DefaultScanner {
	return &DefaultScanner{
		customIgnores: make(map[string]bool),
		maxFileSize:   5 * 1024 * 1024, // 5MB
	}
}

// DiscoverFiles walks the workspace root, validates path boundaries, and indexes all valid non-ignored files.
func (s *DefaultScanner) DiscoverFiles(ctx context.Context, root string) ([]FileInfo, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace root: %w", err)
	}

	// Load custom ignore patterns from .gitignore / .termagentignore if present
	ignorePatterns := s.loadIgnorePatterns(absRoot)

	var files []FileInfo

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // Skip unreadable paths gracefully
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}

		if rel == "." {
			return nil
		}

		base := d.Name()

		if d.IsDir() {
			if DefaultIgnoredDirs[base] || s.isPatternIgnored(rel, ignorePatterns) {
				return filepath.SkipDir
			}
			return nil
		}

		if DefaultIgnoredFiles[base] || s.isPatternIgnored(rel, ignorePatterns) {
			return nil
		}

		// Security check: path boundary validation
		validatedAbs, err := security.ValidateWorkspacePath(absRoot, rel)
		if err != nil {
			return nil // Skip paths escaping workspace root (e.g. symlink traps)
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		fileInfo := FileInfo{
			RelPath:   filepath.ToSlash(rel),
			AbsPath:   validatedAbs,
			SizeBytes: info.Size(),
			Mode:      info.Mode(),
			ModTime:   info.ModTime(),
		}

		// Detect binary file type
		fileInfo.IsBinary = isBinaryFile(validatedAbs)

		// Calculate line count for text files under size threshold
		if !fileInfo.IsBinary && info.Size() <= s.maxFileSize {
			fileInfo.LineCount = countLines(validatedAbs)
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil && err != context.Canceled {
		return nil, fmt.Errorf("workspace scan failed: %w", err)
	}

	return files, nil
}

// IsIgnored checks whether a given relative path matches default or loaded ignore patterns.
func (s *DefaultScanner) IsIgnored(relPath string) bool {
	cleanRel := filepath.Clean(relPath)
	parts := strings.Split(cleanRel, string(filepath.Separator))
	for _, part := range parts {
		if DefaultIgnoredDirs[part] || DefaultIgnoredFiles[part] {
			return true
		}
	}
	return false
}

// loadIgnorePatterns reads .gitignore and .termagentignore files inside the workspace root.
func (s *DefaultScanner) loadIgnorePatterns(root string) []string {
	var patterns []string
	ignoreFiles := []string{".gitignore", ".termagentignore"}

	for _, filename := range ignoreFiles {
		filePath := filepath.Join(root, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			patterns = append(patterns, line)
		}
	}

	return patterns
}

// isPatternIgnored checks if a path matches simple wildcard ignore patterns.
func (s *DefaultScanner) isPatternIgnored(relPath string, patterns []string) bool {
	normalized := filepath.ToSlash(relPath)
	for _, pattern := range patterns {
		pattern = strings.TrimPrefix(pattern, "/")
		if matched, err := filepath.Match(pattern, normalized); err == nil && matched {
			return true
		}
		if matched, err := filepath.Match(pattern, filepath.Base(normalized)); err == nil && matched {
			return true
		}
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(normalized, strings.TrimSuffix(pattern, "/")) {
			return true
		}
	}
	return false
}

// isBinaryFile checks the first 512 bytes of a file for null bytes (\x00).
func isBinaryFile(absPath string) bool {
	f, err := os.Open(absPath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}

	return bytes.IndexByte(buf[:n], 0) != -1
}

// countLines counts newlines in a text file.
func countLines(absPath string) int {
	f, err := os.Open(absPath)
	if err != nil {
		return 0
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)
		if err == io.EOF {
			break
		}
		if err != nil {
			return count
		}
	}
	return count
}
