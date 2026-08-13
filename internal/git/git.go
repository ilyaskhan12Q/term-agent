package git

import (
	"context"
)

// GitStatus represents workspace git repository status.
type GitStatus struct {
	Branch         string
	HeadCommit     string
	IsClean        bool
	ModifiedFiles  []string
	UntrackedFiles []string
}

// Client defines the contract for Git inspection.
type Client interface {
	GetStatus(ctx context.Context, repoPath string) (*GitStatus, error)
	GetDiff(ctx context.Context, repoPath string) (string, error)
	GetCurrentBranch(ctx context.Context, repoPath string) (string, error)
}
