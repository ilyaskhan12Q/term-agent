package mutation

// MutationType represents the operation performed on a file.
type MutationType string

const (
	MutationCreate MutationType = "CREATE"
	MutationModify MutationType = "MODIFY"
	MutationDelete MutationType = "DELETE"
)

// FileMutation describes an individual file change inside a transaction.
type FileMutation struct {
	ID            string
	TransactionID string
	Path          string
	Type          MutationType
	OriginalHash  string
	NewHash       string
	NewContent    []byte
}
