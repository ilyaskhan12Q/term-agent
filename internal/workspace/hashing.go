package workspace

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashBytes computes the SHA-256 hash of a file or content snippet.
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
