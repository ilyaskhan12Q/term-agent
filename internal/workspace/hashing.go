package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// HashBytes computes the SHA-256 hash of a file or content snippet.
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ComputeWorkspaceHash produces a deterministic SHA-256 hash representing the state of indexed workspace files.
func ComputeWorkspaceHash(files []FileInfo) (string, error) {
	// Sort files by RelPath for deterministic hashing
	sorted := make([]FileInfo, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RelPath < sorted[j].RelPath
	})

	h := sha256.New()
	for _, f := range sorted {
		entry := fmt.Sprintf("%s:%d:%d\n", f.RelPath, f.SizeBytes, f.ModTime.UnixNano())
		h.Write([]byte(entry))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
