package diff

// DiffRenderer defines the contract for formatting diffs for TUI rendering.
type DiffRenderer interface {
	RenderTerminal(diff *FileDiff) (string, error)
}
