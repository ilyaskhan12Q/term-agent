package workspace

import (
	"context"
)

// FileDiscovery defines the contract for scanning workspace files respecting gitignore rules.
type FileDiscovery interface {
	DiscoverFiles(ctx context.Context, root string) ([]string, error)
	IsIgnored(path string) bool
}
