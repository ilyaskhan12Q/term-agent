package mutation

// FileSnapshot captures file state before mutation for atomicity and rollback.
type FileSnapshot struct {
	Path        string
	Exists      bool
	Content     []byte
	ContentHash string
}
